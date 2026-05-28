# 并发安全设计与验证

> 性能优化引入了 `sync.Pool` 复用 `ASTEvaluator`、以及在多个 `Evaluate` 调用之间共享只读 `PreCompiled` 缓存。本文档说明并发模型，并以 Go 1.25 转正的 `testing/synctest` 给出确定性验证。
>
> 本项目 `go.mod` 声明 `go 1.24`，以保证与下游消费者（如 PAI-Rec）的 toolchain 对齐；并发测试文件头部加了 `//go:build go1.25` ，仅在 Go 1.25+ 编译运行，Go 1.24 用户不会被该依赖阻塞。

对应代码：

- [evaluable_expression.go](file:///Users/dingjing/work/go-antlr-valuate/evaluable_expression.go)
- [parser/precompile.go](file:///Users/dingjing/work/go-antlr-valuate/parser/precompile.go)
- [parser/ast_evaluator.go](file:///Users/dingjing/work/go-antlr-valuate/parser/ast_evaluator.go)
- [evaluable_expression_concurrent_test.go](file:///Users/dingjing/work/go-antlr-valuate/evaluable_expression_concurrent_test.go)

---

## 1. 并发使用契约

`EvaluableExpression` 的对外契约：

> 解析得到一个 `EvaluableExpression` 后，**可以从多个 goroutine 并发调用 `Evaluate`**。每次调用传入的 `params` map 必须在该次调用期间不被外部 mutate。

允许的用法：

```go
expr, _ := valuate.NewEvaluableExpressionWithFunctions(text, funcs)

for i := 0; i < N; i++ {
    go func(p map[string]interface{}) {
        v, _ := expr.Evaluate(p)
        ...
    }(myParams[i])
}
```

不允许的用法：

- 在 `Evaluate` 进行的同时，从另一个 goroutine 写入同一个 `params` map。
- 在 `NewEvaluableExpressionWithFunctions` 返回之后，再去修改作为 `functions` 入参传进来的 map（会和 `pre.Functions` 共享底层）。

---

## 2. 并发模型

### 2.1 各组件的所有权

| 组件 | 写入时机 | 读取时机 | 跨 goroutine 共享？ |
| --- | --- | --- | --- |
| `PreCompiled.Literals` 等所有 map | `PreCompile` 期间 | `Evaluate` 期间 | ✅ 只读共享 |
| `EvaluableExpression.pool` (`sync.Pool`) | 全程 | 全程 | ✅ pool 自身并发安全 |
| `ASTEvaluator` 实例 | 单次 `Evaluate` 内 | 单次 `Evaluate` 内 | ❌ 通过 pool Get/Put 串行使用 |
| `paramsMap` | 调用方持有 | `Evaluate` 期间 | ❌ caller-owned |

### 2.2 关键不变量

1. **`PreCompiled` 一旦构造完成就不再修改**。`PreCompile` 在 `NewEvaluableExpressionWithFunctions` 里同步完成，之后所有 `Evaluate` 只读这些 map。Go 的 map 在"无并发写入"前提下并发读是安全的。
2. **`ASTEvaluator` 永不在两个 goroutine 间同时使用**。`sync.Pool.Get` 把对象在多 goroutine 间分发，但任何一个 evaluator 实例的生命周期完全包含在一次 `Evaluate` 调用内：

   ```go
   v := e.pool.Get()           // 取出
   ast.Reset(params, e.pre)    // 仅本 goroutine 写
   ast.Visit(...)              // 仅本 goroutine 读写
   ast.Reset(nil, nil)         // 清引用
   e.pool.Put(ast)             // 放回
   ```

3. **归还前显式清空引用**：`Reset(nil, nil)` 防止 evaluator 在池中长期持有 caller 的 `paramsMap`，避免：
   - GC 不可达对象被错误 pin 住；
   - 下一次 `Get` 拿到对象时，`paramsMap` 还指向上一次调用的 map，造成读到陈旧数据。

### 2.3 `BuildInFunctionsMap` 的处理

`parser.BuildInFunctionsMap` 是包级变量。`NewEvaluableExpressionWithFunctions` 不直接共享它，而是 **新 make 一个 `merged` map**，逐个 copy 进去：

```go
merged := make(map[string]parser.ExpressionFunction, len(functions)+len(parser.BuildInFunctionsMap))
for k, v := range parser.BuildInFunctionsMap { merged[k] = v }
for k, v := range functions { merged[k] = v }
pre := parser.PreCompile(expressionContext, merged)
```

因此即便用户在多个 goroutine 里并发 `New...`，也只会并发读 `BuildInFunctionsMap`，不会写。

---

## 3. 验证：`testing/synctest`

Go 1.25 把 `testing/synctest` 转正（在 1.24 仍为 `GOEXPERIMENT=synctest` 实验特性，且 API 为 `synctest.Run`）。它的关键特性：

- `synctest.Test(t, fn)` 把 `fn` 放进一个 **bubble**，bubble 内启动的 goroutine 由 synctest 调度器驱动。
- 调度是 **确定性** 的：测试每次运行的 goroutine 交错顺序是受控的，更容易暴露隐藏的竞态。
- 配合 `go test -race` 能同时拿到 race detector 的覆盖。

[evaluable_expression_concurrent_test.go](file:///Users/dingjing/work/go-antlr-valuate/evaluable_expression_concurrent_test.go) 给出两组验证：

### 3.1 单实例 × 多 goroutine（`TestConcurrentEvaluate_Synctest`）

`goroutines = 64`、`iterations = 200`，6 个子用例覆盖：

| 子用例 | 触发的代码路径 |
| --- | --- |
| `constant_folded_bool` | `pre.Literals[ctx]` fast-path（折叠后的 bool） |
| `constant_folded_arith` | 折叠后的 float64 |
| `single_param` | `pre.VarNames` + `paramsMap` 共享读 |
| `ternary_short_circuit` | 三目短路 + 字面量 fast-path |
| `rank_score_business` | 复杂业务表达式 + 共享 `paramsMap` |
| `max_index_business` | `pre.ResolvedFuncs` + 用户函数 + 数组遍历 |

```go
synctest.Test(t, func(t *testing.T) {
    expr, _ := NewEvaluableExpressionWithFunctions(tc.expr, tc.funcs)
    var wg sync.WaitGroup
    wg.Add(goroutines)
    for g := 0; g < goroutines; g++ {
        go func() {
            defer wg.Done()
            for i := 0; i < iterations; i++ {
                got, _ := expr.Evaluate(tc.params)
                // 校验结果
            }
        }()
    }
    wg.Wait()
})
```

浮点用例用 `1e-9` epsilon 比较，避免 IEEE-754 求和顺序差异（`rankScoreWant = 0.8*1.0 + 0.3*0.0 + 12.0/35.0*0.15`）。

### 3.2 并发 Parse + Evaluate（`TestConcurrentParseAndEvaluate_Synctest`）

验证 **并发地** 调 `NewEvaluableExpressionWithFunctions` + `Evaluate` 不会破坏 `BuildInFunctionsMap` 等包级状态。`goroutines = 32`，每个 goroutine 独立创建自己的 `expr` 并求值业务表达式。

---

## 4. 跑测方法

```bash
# 常规正确性
go test ./...

# 数据竞争检测（synctest 子用例同时受益）
go test -race ./...

# 仅跑并发用例
go test -race -run TestConcurrent ./...
```

预期：全绿，无 race report。

---

## 5. 已知限制 / 注意事项

1. **caller 的 `params` map 必须在单次 `Evaluate` 内不可变**。如果业务侧需要并发更新参数，应在调用前做一次浅拷贝，或自行加锁。
2. **`functions` 入参在 `New...` 返回后不应被 mutate**。`merged` map 内部的函数指针来自传入 map，新增/删除条目无法被 `pre.Functions` 感知。
3. **`PreCompiled` 的 map 不应被外部修改**。`PreCompile` 不导出修改入口，但通过反射等方式仍可触达，调用方不应这么做。
4. **synctest 的 bubble 内不能调用真实网络/磁盘 IO 的阻塞操作**。本测试只跑 CPU + map 访问，未触碰这一限制。

---

## 6. 总结

- 并发安全的支柱：**只读共享 `PreCompiled` + 单次 `Evaluate` 独占的 evaluator + `sync.Pool` 中转**。
- 验证手段：`testing/synctest` 提供确定性 goroutine 调度，配合 `-race` 给出强保证。
- 在 64 goroutine × 200 iteration × 6 场景 + 32 goroutine × 50 iteration 的并发 Parse 场景下，全部用例稳定通过、race detector 无告警。
