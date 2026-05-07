## 2024-06-25 - sync.Pool and *[]byte allocation avoidance
**Learning:** Using `sync.Pool` with `*[]byte` instead of `[]byte` avoids interface conversion heap allocations in Go when caching buffers.
**Action:** When implementing buffer pooling to prevent memory exhaustion on concurrent file ops, always store the slice pointers in `sync.Pool` to gain maximum GC relief.
## 2024-06-25 - sync.Pool and *[]byte allocation avoidance
**Learning:** Using `sync.Pool` with `*[]byte` instead of `[]byte` avoids interface conversion heap allocations in Go when caching buffers.
**Action:** When implementing buffer pooling to prevent memory exhaustion on concurrent file ops, always store the slice pointers in `sync.Pool` to gain maximum GC relief.
