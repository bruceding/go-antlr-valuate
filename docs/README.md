# 设计文档索引

本目录收录 go-antlr-valuate 性能优化与相关工程实践的中文设计文档。

| 文档 | 内容 |
| --- | --- |
| [performance-optimization.md](performance-optimization.md) | Evaluate 阶段性能优化的起点分析、10 项优化措施、收益对比与取舍 |
| [concurrency-safety.md](concurrency-safety.md) | `sync.Pool` + 共享 `PreCompiled` 的并发模型，及基于 `testing/synctest` 的验证 |

阅读建议：先看 `performance-optimization.md` 了解整体优化思路，再看 `concurrency-safety.md` 理解为何这些优化不会破坏并发安全。
