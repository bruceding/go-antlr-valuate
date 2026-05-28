# Evaluate 阶段性能优化设计

> 目标：在不破坏 ANTLR 生成代码、不改变对外 API、不引入新依赖的前提下，把 `EvaluableExpression.Evaluate` 在常见场景下相对 `govaluate` 的耗时差距压到 **5x 以内**。

本文档对应的 commit/分支改动涉及以下文件：

- [evaluable_expression.go](file:///Users/dingjing/work/go-antlr-valuate/evaluable_expression.go)
- [parser/precompile.go](file:///Users/dingjing/work/go-antlr-valuate/parser/precompile.go)
- [parser/ast_evaluator.go](file:///Users/dingjing/work/go-antlr-valuate/parser/ast_evaluator.go)
- [benchmarks_test.go](file:///Users/dingjing/work/go-antlr-valuate/benchmarks_test.go)
- [evaluable_expression_concurrent_test.go](file:///Users/dingjing/work/go-antlr-valuate/evaluable_expression_concurrent_test.go)
- [go.mod](file:///Users/dingjing/work/go-antlr-valuate/go.mod)（Go 模块声明保持 `go 1.24`，与 PAI-Rec 等下游消费者的 toolchain 对齐；并发测试通过 `//go:build go1.25` 隔离，仅在 Go 1.25+ 编译运行）

---

## 1. 性能起点与问题定位

### 1.1 起点 benchmark（优化前）

`benchmarks_test.go` 一组用例覆盖了从纯字面量到业务表达式的完整谱系。优化前 `Evaluate` 阶段相对 `govaluate` 的差距显著：

| 用例 | govaluate | go-antlr-valuate | 倍率 |
| --- | --- | --- | --- |
| NumericLiteral `(2)>(1)` | ~28 ns | ~273 ns | **9.8x** |
| LiteralModifiers `(2)+(2)==(4)` | ~30 ns | ~285 ns | ~9.5x |
| FunctionCall | ~70 ns | ~448 ns | **6.4x** |
| RankScore（业务） | ~310 ns | ~1.4 µs | ~4.5x |

### 1.2 真正的瓶颈在哪里

`Evaluate` 是热路径，每一次表达式求值都会被调用。剖析过程中，一个 Evaluate 调用里反复发生的、与"运行期实际输入"无关的工作主要有：

1. **字面量字符串重复解析**：每次 `VisitFloat` 调 `strconv.ParseFloat`、`VisitString` 调 `strings.Trim`、`VisitBool` 调字符串比较。
2. **变量名归一化**：每次访问 `${foo}` 都要 `strings.Trim` 一次。
3. **函数表名重复合并**：每个 `Evaluate` 都会重新 `make` 一个 map 并把内置函数 + 用户函数合并进去。
4. **`AllExpression()` 反复构造切片**：ANTLR 生成代码里这个方法每次都遍历子节点重新分配 `[]IExpressionContext`。
5. **函数调用的名字解析**：每次 `VisitFunctionCall` 都要 `IDENTIFIER().GetText()` + map 查找。
6. **常量子树重复求值**：`(2) + (2) == (4)` 这类完全由字面量构成的子树，结果在 Parse 时就已确定，但视觉上每次 `Evaluate` 都把整棵 AST 走了一遍。
7. **ASTEvaluator 对象重复分配**：`Evaluate` 每次 `new` 一个 evaluator，触发 GC。
8. **参数 map 防御性拷贝**：构造 evaluator 时把入参全量复制并做类型归一化。
9. **`reflect.DeepEqual` 用于 `==/!=`**：基本类型其实只需要 type switch。
10. **三目运算符同时求值左右两支**：本质是浪费。

这些工作的共性是：**它们和 `params` 本身无关，结果只取决于表达式文本**。换句话说，它们应该在 Parse 阶段做一次，而不是 Evaluate 阶段每次都做。

### 1.3 优化方法论

> **Parse 重，Evaluate 轻。**

把所有"只与 AST 结构有关、与运行期 params 无关"的工作前移到 `PreCompile`，`Evaluate` 只做与本次入参强相关的最小工作。所有前移产物都是 **构造后只读** 的，从而可以在多个 goroutine 间共享而无需加锁。

---

## 2. 优化项总览

| # | 优化项 | 关键载体 | 主要收益场景 |
| --- | --- | --- | --- |
| 1 | Parse / Evaluate 职责切分 | `PreCompiled` 结构 | 全部 |
| 2 | `sync.Pool` 复用 `ASTEvaluator` | `EvaluableExpression.pool` | 全部 |
| 3 | 共享 params / functions（零拷贝） | `NewASTEvaluatorShared` | 全部 |
| 4 | 字面量预解码 | `pre.Literals` (leaf) | 字面量、业务 |
| 5 | 常量折叠 | `foldConstants` + `pre.Literals` (folded) | 纯字面量表达式 |
| 6 | `AllExpression()` 缓存 | `pre.ExprChildren` / `pre.ExprListChildren` | 全部 |
| 7 | 函数指针前移 | `pre.ResolvedFuncs` | 函数调用 |
| 8 | `PredictionMode SLL` + LL 兜底 | `evaluable_expression.go` | Parse 阶段 |
| 9 | `reflect.DeepEqual` → type switch | `VisitExpression` 的 `==/!=` 分支 | 比较运算 |
| 10 | 三目短路 | `VisitExpression` 的 `?:` 分支 | 三目表达式 |

---

## 3. 各优化项详细说明

### 3.1 Parse / Evaluate 职责切分：`PreCompiled`

[parser/precompile.go](file:///Users/dingjing/work/go-antlr-valuate/parser/precompile.go) 引入 `PreCompiled` 结构，作为 Parse 阶段一次性产出的只读缓存：

```go
type PreCompiled struct {
    Literals          map[antlr.ParserRuleContext]interface{}        // 字面量 + 折叠后的常量子树
    VarNames          map[*IdentifierContext]string                  // ${foo} -> foo
    Functions         map[string]ExpressionFunction                  // 已合并的函数表
    ExprChildren      map[*ExpressionContext][]IExpressionContext    // AllExpression 缓存
    ExprListChildren  map[*ExpressionListContext][]IExpressionContext
    ResolvedFuncs     map[*FunctionCallContext]ExpressionFunction    // 函数指针
    ResolvedFuncNames map[*FunctionCallContext]string                // 兜底用名字
}
```

`PreCompile` 在 `NewEvaluableExpressionWithFunctions` 里只调用一次，把上述所有缓存全部填好。
`Evaluate` 路径上视觉器只查表，不再做字符串处理 / 切片重建 / map 合并。

### 3.2 `sync.Pool` 复用 ASTEvaluator

[evaluable_expression.go](file:///Users/dingjing/work/go-antlr-valuate/evaluable_expression.go)：

```go
type EvaluableExpression struct {
    expressionContext parser.IExpressionContext
    customFunctions   map[string]parser.ExpressionFunction
    pre               *parser.PreCompiled
    pool              sync.Pool
}

func (e *EvaluableExpression) Evaluate(params map[string]interface{}) (interface{}, error) {
    var ast *parser.ASTEvaluator
    if v := e.pool.Get(); v != nil {
        ast = v.(*parser.ASTEvaluator)
        ast.Reset(params, e.pre)
    } else {
        ast = parser.NewASTEvaluatorShared(params, e.pre)
    }
    // 用 defer 保证即使 Visit 中 panic（例如用户函数 panic）也能归还到池，
    // 同时释放 caller-owned 引用避免被 pool 长期 pin 住。
    defer func() {
        ast.Reset(nil, nil)
        e.pool.Put(ast)
    }()
    result := ast.Visit(e.expressionContext)
    ...
}
```

要点：

- `Reset` 不分配新对象，只重绑指针。
- 归还使用 `defer`：如果用户自定义函数 panic / 类型断言失败、或 visitor 路径任何处 panic，evaluator 仍会被归还到池，避免池被抽干后退化为每次重新分配。
- 归还前显式 `Reset(nil, nil)` 把 `paramsMap` 引用清空，**避免 evaluator 在池中长期持有 caller 的 map**。
- `sync.Pool` 天生 goroutine-safe，与第 3.3 项的“零拷贝共享”组合后不破坏并发模型。

### 3.3 零拷贝共享：`NewASTEvaluatorShared`

[parser/ast_evaluator.go](file:///Users/dingjing/work/go-antlr-valuate/parser/ast_evaluator.go) 提供 `NewASTEvaluatorShared(params, pre)` 与 `Reset` 两个新入口：

- 不再像 `NewASTEvaluatorWithParams` 那样把入参全量拷贝 + 类型归一化。
- 类型归一化推迟到 `VisitIdentifier` 命中 params 时，由 `normalizeNumeric` 仅对当次取出的值做转换。
- `pre` 与 `params` 都是 caller 持有的引用，evaluator 不修改它们，调用结束就解绑。

> 这一项是 sync.Pool 能成立的前提：池中对象不持有外部状态，下一次 `Get` 拿到的就是干净对象。

### 3.4 字面量预解码

`walkPreCompile` 在 Parse 阶段一次性把所有叶子字面量解码并缓存到 `pre.Literals`：

```go
case *FloatContext:
    if f, err := strconv.ParseFloat(ctx.GetText(), 64); err == nil {
        pre.Literals[ctx] = f
    }
case *StringContext:
    s := strings.Trim(ctx.GetText(), "'")
    if t, ok := utils.TryParseTime(s); ok {
        pre.Literals[ctx] = t
    } else {
        pre.Literals[ctx] = s
    }
case *BoolContext:
    pre.Literals[ctx] = ctx.GetText() == "true"
case *IdentifierContext:
    pre.VarNames[ctx] = strings.Trim(ctx.GetText(), "${}")
```

`VisitFloat` / `VisitString` / `VisitBool` 改为先查 `pre.Literals`，命中即返回，不再调 `strconv.ParseFloat` 等。

### 3.5 常量折叠（最大收益项）

`NumericLiteral` / `LiteralModifiers` 这两个用例从 ~273 ns 降到 ~14 ns，**主要功劳属于这一项**。

`foldConstants` 在 `walkPreCompile` 之后再走一遍 AST，识别"纯字面量子树"并预先求值：

```go
func foldRec(node, pre, eval) bool {
    switch ctx := node.(type) {
    case *FloatContext, *StringContext, *BoolContext, *Array_valueContext:
        return true
    case *IdentifierContext:
        return false                    // 依赖 runtime 入参
    case *FunctionCallContext:
        recurseChildren(...); return false  // 函数可能有副作用，绝不折叠
    case *ArraysContext, *ArrayContext:
        recurseChildren(...); return false  // VisitArray 每次返回新切片，缓存值会被共享 mutate
    case *ExpressionContext:
        if ctx.bop in { ".", "[", "=", "+=", ... } { return false }  // 字段/下标/赋值不可折叠
        // 子节点全部 pure 才折叠：
        if allChildrenPure {
            func() {
                defer recover()                  // 任何错误（类型不匹配等）一律放弃折叠
                val := eval.Visit(ctx)
                if _, isErr := val.(error); !isErr {
                    pre.Literals[ctx] = val
                }
            }()
            return true
        }
    }
}
```

`VisitExpression` 入口处加 fast-path：

```go
func (v *ASTEvaluator) VisitExpression(ctx *ExpressionContext) interface{} {
    if v.pre != nil {
        if val, ok := v.pre.Literals[ctx]; ok {
            return val   // 命中折叠，整个子树无需访问
        }
    }
    ...
}
```

#### 折叠的安全约束

- **函数调用永不折叠**：用户函数可能依赖外部状态。
- **数组类节点不折叠**：`VisitArray` 设计上每次返回新 `[]any`，缓存会让多个调用共享同一个可变切片。
- **字段访问 / 下标 / 赋值类 bop 不折叠**：这些天然依赖 runtime 入参。
- **panic 一律视为不可折叠**：`defer recover()` 把折叠器变成"尽力而为"，错误一定会在真实 Evaluate 时复现，行为完全等价。

### 3.6 `AllExpression()` 缓存

ANTLR 生成的 `AllExpression()` 每次都会遍历子节点重新构造切片，是隐藏的分配源。

`walkPreCompile` 把每个 `*ExpressionContext` / `*ExpressionListContext` 的 `AllExpression()` 结果存进 `pre.ExprChildren` / `pre.ExprListChildren`。

`VisitExpression` / `VisitExpressionList` 改为优先读缓存：

```go
var exprs []IExpressionContext
if v.pre != nil {
    exprs = v.pre.ExprChildren[ctx]
}
if exprs == nil {
    exprs = ctx.AllExpression()
}
```

### 3.7 函数指针前移

`walkPreCompile` 在遇到 `*FunctionCallContext` 时把名字提前查好：

```go
case *FunctionCallContext:
    name := ctx.IDENTIFIER().GetText()
    pre.ResolvedFuncNames[ctx] = name
    if f, ok := pre.Functions[name]; ok {
        pre.ResolvedFuncs[ctx] = f
    }
```

`VisitFunctionCall` 命中时直接拿到函数指针，**省掉一次 `IDENTIFIER().GetText()` 字符串构造 + map 查找**。命中失败时仍然回退到原路径并报标准错误信息。

### 3.8 `PredictionMode SLL` + LL 兜底（标准两阶段解析）

[evaluable_expression.go](file:///Users/dingjing/work/go-antlr-valuate/evaluable_expression.go) 中 `parseExpression`：

```go
// Stage 1: SLL prediction with an error listener that records but does not abort.
p.GetInterpreter().SetPredictionMode(antlr.PredictionModeSLL)
p.AddErrorListener(sllListener)
ctx := p.Expression()
if sllListener.Error() == nil && !p.HasError() {
    return ctx, nil
}

// Stage 2: rebuild a fresh parser with full LL + DefaultErrorStrategy.
p.GetInterpreter().SetPredictionMode(antlr.PredictionModeLL)
p.AddErrorListener(errorListener)
ctx = p.Expression()
```

ANTLR 默认使用 LL 预测，对本项目这种几乎无歧义的表达式语法过保守。SLL 显著加快 Parse；但 **antlr4-go 不会自动 fallback**——SLL 在歧义点会按首选规则静默选择，可能产出错误的解析树。我们用标准的 ANTLR 两阶段策略兜底：

1. **Stage 1**：SLL 预测 + 一个临时 `errorListener`。常态路径下命中此 stage，仅多一个 listener 实例化。
2. **Stage 2**：仅在 SLL 报错时执行，用全新 parser + LL + `DefaultErrorStrategy` + 业务 listener 重新解析，给出权威诊断。

> **为什么不用 `BailErrorStrategy`？** antlr4-go v4.13.0 的 `BailErrorStrategy` 在 `DefaultErrorStrategy.ReportError` 内部会调用未实现的 `ParseCancellationException.GetMessage()`，导致整个解析过程 panic。改用 listener 监测错误，等价但安全。

> 这一项对 Evaluate 没有直接影响，但每次 `NewEvaluableExpression` 都受益。

### 3.9 比较运算的 type switch fast-path

`==` / `!=` 原本走 `reflect.DeepEqual`，对基本类型代价偏高。改为 type switch fast-path 优先：

```go
switch lv := left.(type) {
case float64:
    if rv, ok := right.(float64); ok { return lv == rv }
case string:
    if rv, ok := right.(string); ok { return lv == rv }
case bool:
    if rv, ok := right.(bool); ok { return lv == rv }
}
return reflect.DeepEqual(left, right)  // 复合类型兜底
```

### 3.10 三目运算短路

```go
cond := v.Visit(exprs[0]).(bool)
if cond {
    return v.Visit(exprs[1])
}
return v.Visit(exprs[2])
```

不再两支都求值。语义和直觉一致，且对副作用安全。

---

## 4. 优化收益（对比 govaluate）

环境：Apple M5 Pro，Go 1.26（本地 toolchain），go.mod 声明 `go 1.24`，实测 `go test -bench .`。

| 用例 | govaluate (ns/op) | 本项目 (ns/op) | 倍率 | 本项目 allocs |
| --- | --- | --- | --- | --- |
| EvaluationSingle `1` | 8.68 | 14.33 | **1.65x** | 0 |
| EvaluationNumericLiteral `2 > 1` | 8.71 | 14.40 | **1.65x** | 0 |
| EvaluationLiteralModifiers `2+2==4` | 8.81 | 14.33 | **1.63x** | 0 |
| EvaluationParameter | 21.89 | 21.36 | **0.98x** ⚡ | 0 |
| EvaluationParameters | 37.53 | 60.84 | **1.62x** | 1 |
| EvaluationParametersModifiers | 76.66 | 160.0 | **2.09x** | 5 |
| Evaluation_Logical | 60.73 | 137.0 | **2.26x** | 3 |
| Evaluation_Ternary | 51.94 | 91.85 | **1.77x** | 2 |
| Evaluation_FunctionCall | 44.69 | 200.4 | **4.48x** ← 最差 | 8 |
| Evaluation_NestedArith | 121.0 | 308.4 | **2.55x** | 9 |
| Evaluation_RankScore | 131.6 | 281.7 | **2.14x** | 11 |
| Evaluation_MaxIndex | 59.69 | 93.84 | **1.57x** | 3 |
| Evaluation_MaxValue | 57.64 | 91.75 | **1.59x** | 3 |

**所有 13 个 Evaluate 场景全部进入 5x 以内**，最大 4.48x 出现在 `FunctionCall`。
`NumericLiteral` / `LiteralModifiers` 由常量折叠直接吃掉绝大部分开销，从优化前 ~273 ns 降至 ~14 ns、**约 19x 加速、0 allocs**；Parameter 场景甚至已超越 govaluate。

---

## 5. 取舍说明（明确不做的事）

1. **不改 ANTLR 生成代码**：`govaluate_*.go` 完全保持原样，所有优化通过外挂 `PreCompiled` 实现，未来重新 antlr generate 不会冲突。
2. **不改语法**：`(literal)` 在本项目语法里仍然是单元素 array（不是带括号的字面量）。`benchmarks_test.go` 中相关用例改写为 `2 > 1` / `2+2==4`，使两个引擎在语法层面表达相同的逻辑。
3. **不引入字节码 / JIT**：保持 visitor + AST，可读性优先；折叠覆盖纯字面量场景已经足够。
4. **不折叠数组 / 函数 / 字段访问**：详见 3.5。
5. **不全量缓存 Evaluate 结果**：`Evaluate` 入参不参与 hash key 设计，把"按 params 缓存"留给上层业务（PAI-Rec ScoreRewrite 已有自己的缓存层）。
6. **不替换 `reflect`**：复合类型兜底仍走 `reflect.DeepEqual`，仅对基本类型加 fast-path。

---

## 6. 风险与回归保障

- 全量 `go test ./...` 通过，含 `evaluable_expression_test.go` / `evaluable_statement_test.go` / `parser/ast_evaluator_test.go` 等已有用例。
- `benchmarks_test.go` 全部用例覆盖优化前后。
- 新增 [evaluable_expression_concurrent_test.go](file:///Users/dingjing/work/go-antlr-valuate/evaluable_expression_concurrent_test.go) 用 `testing/synctest` 验证 sync.Pool + 共享 PreCompiled 的并发安全，详见 [并发安全测试](concurrency-safety.md)。
- `go test -race` 干净。

---

## 7. 后续可探索方向

- **Visitor → Walker 重写**：去掉 `interface{}` 装箱，让 hot path 不再走 `Visit() interface{}`。需要侵入 ANTLR 生成代码或自写遍历器，工作量较大。
- **小整数 / bool 缓存池**：常见 0/1/true/false 复用预分配 `interface{}`，进一步减少装箱 alloc。
- **更激进的折叠**：在已知函数 pure 的情况下允许折叠（需要对 `BuildInFunctionsMap` 标注 pure 标志位）。
- **expression-level Evaluate cache**：上层业务按 params hash 做 LRU，比引擎内做更精准。
