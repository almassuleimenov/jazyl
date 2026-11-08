# golang-pro

> Источник: https://github.com/jeffallan/claude-skills/tree/main/skills/golang-pro (MIT). Рабочая копия скилла: `.claude/skills/golang-pro/`.

---
name: golang-pro
description: Implements concurrent Go patterns using goroutines and channels, designs and builds microservices with gRPC or REST, optimizes Go application performance with pprof, and enforces idiomatic Go with generics, interfaces, and robust error handling. Use when building Go applications requiring concurrent programming, microservices architecture, or high-performance systems. Invoke for goroutines, channels, Go generics, gRPC integration, CLI tools, benchmarks, or table-driven testing.
license: MIT
metadata:
  author: https://github.com/Jeffallan
  company: https://synergetic.solutions
  version: "1.1.1"
  domain: language
  triggers: Go, Golang, goroutines, channels, gRPC, microservices Go, Go generics, concurrent programming, Go interfaces
  role: specialist
  scope: implementation
  output-format: code
  related-skills: devops-engineer, microservices-architect, test-master
---

# Golang Pro

Senior Go developer with deep expertise in Go 1.21+, concurrent programming, and cloud-native microservices. Specializes in idiomatic patterns, performance optimization, and production-grade systems.

## Core Workflow

1. **Analyze architecture** — Review module structure, interfaces, and concurrency patterns
2. **Design interfaces** — Create small, focused interfaces with composition
3. **Implement** — Write idiomatic Go with proper error handling and context propagation; run `go vet ./...` before proceeding
4. **Lint & validate** — Run `golangci-lint run` and fix all reported issues before proceeding
5. **Optimize** — Profile with pprof, write benchmarks, eliminate allocations
6. **Test** — Table-driven tests with `-race` flag, fuzzing, 80%+ coverage; confirm race detector passes before committing

## Reference Guide

Load detailed guidance based on context:

| Topic | Reference | Load When |
|-------|-----------|-----------|
| Concurrency | `references/concurrency.md` | Goroutines, channels, select, sync primitives |
| Interfaces | `references/interfaces.md` | Interface design, io.Reader/Writer, composition |
| Generics | `references/generics.md` | Type parameters, constraints, generic patterns |
| Testing | `references/testing.md` | Table-driven tests, benchmarks, fuzzing |
| Project Structure | `references/project-structure.md` | Module layout, internal packages, go.mod |

## Core Pattern Example

Goroutine with proper context cancellation and error propagation:

```go
// worker runs until jobs is closed, ctx is cancelled, or a job fails.
// It returns nil only after draining jobs cleanly.
func worker(ctx context.Context, jobs <-chan Job) error {
    for {
        select {
        case <-ctx.Done():
            return fmt.Errorf("worker cancelled: %w", ctx.Err())
        case job, ok := <-jobs:
            if !ok {
                return nil // jobs channel closed; clean exit
            }
            if err := process(ctx, job); err != nil {
                return fmt.Errorf("process job %v: %w", job.ID, err)
            }
        }
    }
}

func runPipeline(ctx context.Context, jobs []Job) error {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    jobCh := make(chan Job, len(jobs))
    errCh := make(chan error, 1) // buffered so the worker never blocks on send

    go func() { errCh <- worker(ctx, jobCh) }()

    for _, j := range jobs {
        jobCh <- j
    }
    close(jobCh)

    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
        return fmt.Errorf("pipeline timed out: %w", ctx.Err())
    }
}
```

Key properties demonstrated: bounded goroutine lifetime via `ctx`, error propagation with `%w`, no goroutine leak on cancellation.

## Constraints

### MUST DO
- Use gofmt and golangci-lint on all code
- Add context.Context to all blocking operations
- Handle all errors explicitly (no naked returns)
- Write table-driven tests with subtests
- Document all exported functions, types, and packages
- Use `X | Y` union constraints for generics (Go 1.18+)
- Propagate errors with fmt.Errorf("%w", err)
- Run race detector on tests (-race flag)

### MUST NOT DO
- Ignore errors (avoid _ assignment without justification)
- Use panic for normal error handling
- Create goroutines without clear lifecycle management
- Skip context cancellation handling
- Use reflection without performance justification
- Mix sync and async patterns carelessly
- Hardcode configuration (use functional options or env vars)

## Output Templates

When implementing Go features, provide:
1. Interface definitions (contracts first)
2. Implementation files with proper package structure
3. Test file with table-driven tests
4. Brief explanation of concurrency patterns used

## Knowledge Reference

Go 1.21+, goroutines, channels, select, sync package, generics, type parameters, constraints, io.Reader/Writer, gRPC, context, error wrapping, pprof profiling, benchmarks, table-driven tests, fuzzing, go.mod, internal packages, functional options

Maintained by [@jeffallan](https://github.com/jeffallan), Principal Consultant at [Synergetic Solutions](https://synergetic.solutions)

[Documentation](https://jeffallan.github.io/claude-skills/skills/language/golang-pro/)

---

## Reference: concurrency

# Concurrency Patterns

## Goroutine Lifecycle Management

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "time"
)

// Worker pool with bounded concurrency
type WorkerPool struct {
    workers int
    tasks   chan func()
    wg      sync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
    wp := &WorkerPool{
        workers: workers,
        tasks:   make(chan func(), workers*2), // Buffered channel
    }
    wp.start()
    return wp
}

func (wp *WorkerPool) start() {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go func() {
            defer wp.wg.Done()
            for task := range wp.tasks {
                task()
            }
        }()
    }
}

func (wp *WorkerPool) Submit(task func()) {
    wp.tasks <- task
}

func (wp *WorkerPool) Shutdown() {
    close(wp.tasks)
    wp.wg.Wait()
}
```

## Channel Patterns

```go
// Generator pattern
func generateNumbers(ctx context.Context, max int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for i := 0; i < max; i++ {
            select {
            case out <- i:
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

// Fan-out, fan-in pattern
func fanOut(ctx context.Context, input <-chan int, workers int) []<-chan int {
    channels := make([]<-chan int, workers)
    for i := 0; i < workers; i++ {
        channels[i] = process(ctx, input)
    }
    return channels
}

func process(ctx context.Context, input <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for {
            // Receive inside select so cancellation also stops an idle stage
            select {
            case val, ok := <-input:
                if !ok {
                    return
                }
                select {
                case out <- val * 2:
                case <-ctx.Done():
                    return
                }
            case <-ctx.Done():
                return
            }
        }
    }()
    return out
}

func fanIn(ctx context.Context, channels ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup

    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for {
                select {
                case val, ok := <-c:
                    if !ok {
                        return
                    }
                    select {
                    case out <- val:
                    case <-ctx.Done():
                        return
                    }
                case <-ctx.Done():
                    return
                }
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}
```

## Select Statement Patterns

```go
// Timeout pattern
func fetchWithTimeout(ctx context.Context, url string) (string, error) {
    result := make(chan string, 1)
    errCh := make(chan error, 1)

    go func() {
        // Simulate network call
        time.Sleep(100 * time.Millisecond)
        result <- "data from " + url
    }()

    select {
    case res := <-result:
        return res, nil
    case err := <-errCh:
        return "", err
    case <-time.After(50 * time.Millisecond):
        return "", fmt.Errorf("timeout")
    case <-ctx.Done():
        return "", ctx.Err()
    }
}

// Done channel pattern for graceful shutdown
type Server struct {
    done chan struct{}
}

func (s *Server) Shutdown() {
    close(s.done)
}

func (s *Server) Run(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            fmt.Println("tick")
        case <-s.done:
            fmt.Println("shutting down")
            return
        case <-ctx.Done():
            fmt.Println("context cancelled")
            return
        }
    }
}
```

## Sync Primitives

```go
import "sync"

// Mutex for protecting shared state
type Counter struct {
    mu    sync.Mutex
    count int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *Counter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count
}

// RWMutex for read-heavy workloads
type Cache struct {
    mu    sync.RWMutex
    items map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    val, ok := c.items[key]
    return val, ok
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = value
}

// sync.Once for initialization
type Service struct {
    once   sync.Once
    config *Config
}

func (s *Service) getConfig() *Config {
    s.once.Do(func() {
        s.config = loadConfig() // Only called once
    })
    return s.config
}
```

## Rate Limiting and Backpressure

```go
import "golang.org/x/time/rate"

// Token bucket rate limiter
type RateLimiter struct {
    limiter *rate.Limiter
}

func NewRateLimiter(rps int) *RateLimiter {
    return &RateLimiter{
        limiter: rate.NewLimiter(rate.Limit(rps), rps),
    }
}

func (rl *RateLimiter) Process(ctx context.Context, item string) error {
    if err := rl.limiter.Wait(ctx); err != nil {
        return err
    }
    // Process item
    return nil
}

// Semaphore pattern for limiting concurrency
type Semaphore struct {
    slots chan struct{}
}

func NewSemaphore(n int) *Semaphore {
    return &Semaphore{
        slots: make(chan struct{}, n),
    }
}

func (s *Semaphore) Acquire() {
    s.slots <- struct{}{}
}

func (s *Semaphore) Release() {
    <-s.slots
}

func (s *Semaphore) Do(fn func()) {
    s.Acquire()
    defer s.Release()
    fn()
}
```

## Pipeline Pattern

```go
// Stage-based processing pipeline
func pipeline(ctx context.Context, input <-chan int) <-chan int {
    // Stage 1: Square numbers
    stage1 := make(chan int)
    go func() {
        defer close(stage1)
        for {
            select {
            case num, ok := <-input:
                if !ok {
                    return
                }
                select {
                case stage1 <- num * num:
                case <-ctx.Done():
                    return
                }
            case <-ctx.Done():
                return
            }
        }
    }()

    // Stage 2: Filter even numbers
    stage2 := make(chan int)
    go func() {
        defer close(stage2)
        for {
            select {
            case num, ok := <-stage1:
                if !ok {
                    return
                }
                if num%2 != 0 {
                    continue
                }
                select {
                case stage2 <- num:
                case <-ctx.Done():
                    return
                }
            case <-ctx.Done():
                return
            }
        }
    }()

    return stage2
}
```

## Quick Reference

| Pattern | Use Case | Key Points |
|---------|----------|------------|
| Worker Pool | Bounded concurrency | Limit goroutines, reuse workers |
| Fan-out/Fan-in | Parallel processing | Distribute work, merge results |
| Pipeline | Stream processing | Chain transformations |
| Rate Limiter | API throttling | Control request rate |
| Semaphore | Resource limits | Cap concurrent operations |
| Done Channel | Graceful shutdown | Signal completion |

---

## Reference: generics

# Generics and Type Parameters

## Basic Type Parameters

```go
package main

import (
    "cmp" // cmp.Ordered is in the standard library since Go 1.21
    "fmt"
)

// Generic function with type parameter
func Max[T cmp.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}

// Multiple type parameters
func Map[T, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}

// Usage
func main() {
    maxInt := Max(10, 20)           // T = int
    maxFloat := Max(3.14, 2.71)     // T = float64
    maxString := Max("abc", "xyz")  // T = string

    nums := []int{1, 2, 3}
    doubled := Map(nums, func(n int) int { return n * 2 })
    strings := Map(nums, func(n int) string { return fmt.Sprintf("%d", n) })
}
```

## Type Constraints

```go
// Integer and Float are not in the standard library: go get golang.org/x/exp
import "golang.org/x/exp/constraints"

type Number interface {
    constraints.Integer | constraints.Float
}

func Sum[T Number](numbers []T) T {
    var total T
    for _, n := range numbers {
        total += n
    }
    return total
}

// Custom constraints with methods
type Stringer interface {
    String() string
}

func PrintAll[T Stringer](items []T) {
    for _, item := range items {
        fmt.Println(item.String())
    }
}

// Approximate constraint using ~
type Integer interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64
}

type MyInt int

func Double[T Integer](n T) T {
    return n * 2
}

// Works with both int and MyInt
func main() {
    fmt.Println(Double(5))          // int
    fmt.Println(Double(MyInt(5)))   // MyInt
}
```

## Generic Data Structures

```go
// Generic Stack
type Stack[T any] struct {
    items []T
}

func NewStack[T any]() *Stack[T] {
    return &Stack[T]{
        items: make([]T, 0),
    }
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    item := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return item, true
}

func (s *Stack[T]) IsEmpty() bool {
    return len(s.items) == 0
}

// Usage
intStack := NewStack[int]()
intStack.Push(1)
intStack.Push(2)

stringStack := NewStack[string]()
stringStack.Push("hello")
stringStack.Push("world")
```

## Generic Map Operations

```go
// Filter with generics
func Filter[T any](slice []T, predicate func(T) bool) []T {
    result := make([]T, 0, len(slice))
    for _, v := range slice {
        if predicate(v) {
            result = append(result, v)
        }
    }
    return result
}

// Reduce/Fold
func Reduce[T, U any](slice []T, initial U, fn func(U, T) U) U {
    acc := initial
    for _, v := range slice {
        acc = fn(acc, v)
    }
    return acc
}

// Keys from map
func Keys[K comparable, V any](m map[K]V) []K {
    keys := make([]K, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    return keys
}

// Values from map
func Values[K comparable, V any](m map[K]V) []V {
    values := make([]V, 0, len(m))
    for _, v := range m {
        values = append(values, v)
    }
    return values
}

// Usage
numbers := []int{1, 2, 3, 4, 5, 6}
evens := Filter(numbers, func(n int) bool { return n%2 == 0 })

sum := Reduce(numbers, 0, func(acc, n int) int { return acc + n })

m := map[string]int{"a": 1, "b": 2}
keys := Keys(m)     // []string{"a", "b"}
values := Values(m) // []int{1, 2}
```

## Generic Pairs and Tuples

```go
// Generic Pair
type Pair[T, U any] struct {
    First  T
    Second U
}

func NewPair[T, U any](first T, second U) Pair[T, U] {
    return Pair[T, U]{First: first, Second: second}
}

func (p Pair[T, U]) Swap() Pair[U, T] {
    return Pair[U, T]{First: p.Second, Second: p.First}
}

// Usage
pair := NewPair("name", 42)
swapped := pair.Swap() // Pair[int, string]

// Generic Result type (like Rust's Result<T, E>)
type Result[T any] struct {
    value T
    err   error
}

func Ok[T any](value T) Result[T] {
    return Result[T]{value: value}
}

func Err[T any](err error) Result[T] {
    return Result[T]{err: err}
}

func (r Result[T]) IsOk() bool {
    return r.err == nil
}

func (r Result[T]) Unwrap() (T, error) {
    return r.value, r.err
}

func (r Result[T]) UnwrapOr(defaultValue T) T {
    if r.err != nil {
        return defaultValue
    }
    return r.value
}
```

## Comparable Constraint

```go
// Find using comparable
func Find[T comparable](slice []T, target T) (int, bool) {
    for i, v := range slice {
        if v == target {
            return i, true
        }
    }
    return -1, false
}

// Contains
func Contains[T comparable](slice []T, target T) bool {
    _, found := Find(slice, target)
    return found
}

// Unique elements
func Unique[T comparable](slice []T) []T {
    seen := make(map[T]struct{})
    result := make([]T, 0, len(slice))

    for _, v := range slice {
        if _, exists := seen[v]; !exists {
            seen[v] = struct{}{}
            result = append(result, v)
        }
    }

    return result
}

// Usage
nums := []int{1, 2, 2, 3, 3, 4}
unique := Unique(nums) // []int{1, 2, 3, 4}

idx, found := Find([]string{"a", "b", "c"}, "b") // 1, true
```

## Generic Interfaces

```go
// Generic interface
type Container[T any] interface {
    Add(item T)
    Remove() (T, bool)
    Size() int
}

// Implementation
type Queue[T any] struct {
    items []T
}

func (q *Queue[T]) Add(item T) {
    q.items = append(q.items, item)
}

func (q *Queue[T]) Remove() (T, bool) {
    if len(q.items) == 0 {
        var zero T
        return zero, false
    }
    item := q.items[0]
    q.items = q.items[1:]
    return item, true
}

func (q *Queue[T]) Size() int {
    return len(q.items)
}

// Function accepting generic interface
func ProcessContainer[T any](c Container[T], item T) {
    c.Add(item)
    fmt.Printf("Container size: %d\n", c.Size())
}
```

## Type Inference

```go
// Type inference works in most cases
func Identity[T any](x T) T {
    return x
}

// No need to specify type
result := Identity(42)          // T inferred as int
str := Identity("hello")        // T inferred as string

// Type inference with constraints
func Min[T cmp.Ordered](a, b T) T {
    if a < b {
        return a
    }
    return b
}

// Inferred from arguments
minVal := Min(10, 20)           // T = int
minFloat := Min(1.5, 2.5)       // T = float64

// Explicit type when needed
result := Map[int, string]([]int{1, 2}, func(n int) string {
    return fmt.Sprintf("%d", n)
})
```

## Generic Channels

```go
// Generic channel operations
func Merge[T any](channels ...<-chan T) <-chan T {
    out := make(chan T)
    var wg sync.WaitGroup

    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan T) {
            defer wg.Done()
            for v := range c {
                out <- v
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}

// Generic pipeline stage
func Stage[T, U any](in <-chan T, fn func(T) U) <-chan U {
    out := make(chan U)
    go func() {
        defer close(out)
        for v := range in {
            out <- fn(v)
        }
    }()
    return out
}

// Usage
ch1 := make(chan int)
ch2 := make(chan int)

merged := Merge(ch1, ch2)

numbers := make(chan int)
doubled := Stage(numbers, func(n int) int { return n * 2 })
strings := Stage(doubled, func(n int) string { return fmt.Sprintf("%d", n) })
```

## Union Constraints

```go
// Union of types
type StringOrInt interface {
    string | int
}

func Process[T StringOrInt](val T) string {
    return fmt.Sprintf("%v", val)
}

// More complex unions
type Numeric interface {
    int | int8 | int16 | int32 | int64 |
    uint | uint8 | uint16 | uint32 | uint64 |
    float32 | float64
}

func Abs[T Numeric](n T) T {
    if n < 0 {
        return -n
    }
    return n
}

// Union with methods
type Serializable interface {
    string | []byte
}

func Serialize[T Serializable](data T) []byte {
    switch v := any(data).(type) {
    case string:
        return []byte(v)
    case []byte:
        return v
    default:
        panic("unreachable")
    }
}
```

## Quick Reference

| Feature | Syntax | Use Case |
|---------|--------|----------|
| Basic generic | `func F[T any]()` | Any type |
| Constraint | `func F[T Constraint]()` | Restricted types |
| Multiple params | `func F[T, U any]()` | Multiple type variables |
| Comparable | `func F[T comparable]()` | Types supporting == and != |
| Ordered | `func F[T cmp.Ordered]()` | Types supporting <, >, <=, >= |
| Union | `T interface{int \| string}` | Either type |
| Approximate | `~int` | Include type aliases |

---

## Reference: interfaces

# Interface Design and Composition

## Small, Focused Interfaces

```go
// Single-method interfaces (idiomatic Go)
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}

// Interface composition
type ReadCloser interface {
    Reader
    Closer
}

type WriteCloser interface {
    Writer
    Closer
}

type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}
```

## Accept Interfaces, Return Structs

```go
package storage

import "io"

// Storage is the concrete type (struct)
type Storage struct {
    baseDir string
}

// NewStorage returns a concrete type
func NewStorage(baseDir string) *Storage {
    return &Storage{baseDir: baseDir}
}

// SaveFile accepts an interface for flexibility
func (s *Storage) SaveFile(filename string, data io.Reader) error {
    // Implementation can work with any Reader
    // (file, network, buffer, etc.)
    return nil
}

// Usage allows dependency injection
type Uploader interface {
    SaveFile(filename string, data io.Reader) error
}

type Service struct {
    uploader Uploader // Accept interface
}

// NewService accepts interface for testing flexibility
func NewService(uploader Uploader) *Service {
    return &Service{uploader: uploader}
}
```

## io.Reader and io.Writer Patterns

```go
import (
    "io"
    "strings"
)

// Chain readers with io.MultiReader
func combineReaders() io.Reader {
    r1 := strings.NewReader("Hello ")
    r2 := strings.NewReader("World")
    return io.MultiReader(r1, r2)
}

// Tee reader for duplicating reads
func duplicateRead(r io.Reader, w io.Writer) io.Reader {
    return io.TeeReader(r, w) // Writes to w while reading from r
}

// Limit reader to prevent reading too much
func limitedRead(r io.Reader, n int64) io.Reader {
    return io.LimitReader(r, n)
}

// Custom Reader implementation
type UppercaseReader struct {
    src io.Reader
}

func (u *UppercaseReader) Read(p []byte) (n int, err error) {
    n, err = u.src.Read(p)
    for i := 0; i < n; i++ {
        if p[i] >= 'a' && p[i] <= 'z' {
            p[i] = p[i] - 32
        }
    }
    return n, err
}

// Custom Writer implementation
type CountingWriter struct {
    w     io.Writer
    count int64
}

func (cw *CountingWriter) Write(p []byte) (n int, err error) {
    n, err = cw.w.Write(p)
    cw.count += int64(n)
    return n, err
}

func (cw *CountingWriter) BytesWritten() int64 {
    return cw.count
}
```

## Embedding for Composition

```go
import "sync"

// Embed to extend behavior
type SafeCounter struct {
    mu sync.Mutex
    m  map[string]int
}

func (sc *SafeCounter) Inc(key string) {
    sc.mu.Lock()
    defer sc.mu.Unlock()
    sc.m[key]++
}

// Embed interface to add default behavior
type Logger interface {
    Log(msg string)
}

type NoOpLogger struct{}

func (NoOpLogger) Log(msg string) {}

type Service struct {
    Logger // Embedded interface (default implementation can be provided)
}

func NewService(logger Logger) *Service {
    if logger == nil {
        logger = NoOpLogger{} // Provide default
    }
    return &Service{Logger: logger}
}

// Now Service.Log() is available
```

## Interface Satisfaction Verification

```go
import "io"

// Compile-time interface verification
var _ io.Reader = (*MyReader)(nil)
var _ io.Writer = (*MyWriter)(nil)
var _ io.Closer = (*MyCloser)(nil)

type MyReader struct{}

func (m *MyReader) Read(p []byte) (n int, err error) {
    return 0, nil
}

type MyWriter struct{}

func (m *MyWriter) Write(p []byte) (n int, err error) {
    return len(p), nil
}

type MyCloser struct{}

func (m *MyCloser) Close() error {
    return nil
}
```

## Functional Options Pattern

```go
package server

import "time"

type Server struct {
    host         string
    port         int
    timeout      time.Duration
    maxConns     int
    enableLogger bool
}

// Option is a functional option for configuring Server
type Option func(*Server)

func WithHost(host string) Option {
    return func(s *Server) {
        s.host = host
    }
}

func WithPort(port int) Option {
    return func(s *Server) {
        s.port = port
    }
}

func WithTimeout(timeout time.Duration) Option {
    return func(s *Server) {
        s.timeout = timeout
    }
}

func WithMaxConnections(max int) Option {
    return func(s *Server) {
        s.maxConns = max
    }
}

func WithLogger(enabled bool) Option {
    return func(s *Server) {
        s.enableLogger = enabled
    }
}

// NewServer creates a server with functional options
func NewServer(opts ...Option) *Server {
    // Defaults
    s := &Server{
        host:     "localhost",
        port:     8080,
        timeout:  30 * time.Second,
        maxConns: 100,
    }

    // Apply options
    for _, opt := range opts {
        opt(s)
    }

    return s
}

// Usage:
// server := NewServer(
//     WithHost("0.0.0.0"),
//     WithPort(9000),
//     WithTimeout(60 * time.Second),
//     WithLogger(true),
// )
```

## Interface Segregation

```go
// Bad: Fat interface
type BadRepository interface {
    Create(item Item) error
    Read(id string) (Item, error)
    Update(item Item) error
    Delete(id string) error
    List() ([]Item, error)
    Search(query string) ([]Item, error)
    Count() (int, error)
}

// Good: Segregated interfaces
type Creator interface {
    Create(item Item) error
}

type Reader interface {
    Read(id string) (Item, error)
}

type Updater interface {
    Update(item Item) error
}

type Deleter interface {
    Delete(id string) error
}

type Lister interface {
    List() ([]Item, error)
}

// Compose only what you need
type ReadWriter interface {
    Reader
    Creator
}

type FullRepository interface {
    Creator
    Reader
    Updater
    Deleter
    Lister
}
```

## Type Assertions and Type Switches

```go
import "fmt"

// Safe type assertion
func processValue(v interface{}) {
    // Two-value assertion (safe)
    if str, ok := v.(string); ok {
        fmt.Println("String:", str)
        return
    }

    // Type switch
    switch val := v.(type) {
    case int:
        fmt.Println("Int:", val)
    case string:
        fmt.Println("String:", val)
    case bool:
        fmt.Println("Bool:", val)
    default:
        fmt.Println("Unknown type")
    }
}

// Check for optional interface methods
type Flusher interface {
    Flush() error
}

func writeAndFlush(w io.Writer, data []byte) error {
    if _, err := w.Write(data); err != nil {
        return err
    }

    // Check if Writer also implements Flusher
    if flusher, ok := w.(Flusher); ok {
        return flusher.Flush()
    }

    return nil
}
```

## Dependency Injection via Interfaces

```go
package app

import "context"

// Define interfaces for dependencies
type UserRepository interface {
    GetUser(ctx context.Context, id string) (*User, error)
    SaveUser(ctx context.Context, user *User) error
}

type EmailSender interface {
    SendEmail(ctx context.Context, to, subject, body string) error
}

// Service depends on interfaces
type UserService struct {
    repo   UserRepository
    mailer EmailSender
}

func NewUserService(repo UserRepository, mailer EmailSender) *UserService {
    return &UserService{
        repo:   repo,
        mailer: mailer,
    }
}

func (s *UserService) RegisterUser(ctx context.Context, email string) error {
    user := &User{Email: email}
    if err := s.repo.SaveUser(ctx, user); err != nil {
        return err
    }
    return s.mailer.SendEmail(ctx, email, "Welcome", "Thanks for registering!")
}

// Easy to mock in tests
type MockUserRepository struct{}

func (m *MockUserRepository) GetUser(ctx context.Context, id string) (*User, error) {
    return &User{ID: id}, nil
}

func (m *MockUserRepository) SaveUser(ctx context.Context, user *User) error {
    return nil
}
```

## Quick Reference

| Pattern | Use Case | Key Principle |
|---------|----------|---------------|
| Small interfaces | Flexibility | Single-method interfaces |
| Accept interfaces | Testability | Depend on abstractions |
| Return structs | Clarity | Concrete return types |
| io.Reader/Writer | I/O operations | Standard library integration |
| Embedding | Composition | Extend behavior without inheritance |
| Functional options | Configuration | Flexible constructors |
| Type assertions | Runtime checks | Safe downcasting |

---

## Reference: project-structure

# Project Structure and Module Management

## Standard Project Layout

```
myproject/
├── cmd/                    # Main applications
│   ├── server/
│   │   └── main.go        # Entry point for server
│   └── cli/
│       └── main.go        # Entry point for CLI tool
├── internal/              # Private application code
│   ├── api/              # API handlers
│   ├── service/          # Business logic
│   └── repository/       # Data access layer
├── pkg/                   # Public library code
│   └── models/           # Shared models
├── api/                   # API definitions
│   ├── openapi.yaml      # OpenAPI spec
│   └── proto/            # Protocol buffers
├── web/                   # Web assets
│   ├── static/
│   └── templates/
├── scripts/               # Build and install scripts
├── configs/              # Configuration files
├── deployments/          # Docker, K8s configs
├── test/                 # Additional test data
├── docs/                 # Documentation
├── go.mod               # Module definition
├── go.sum               # Dependency checksums
├── Makefile             # Build automation
└── README.md
```

## go.mod Basics

```go
// Initialize module
// go mod init github.com/user/project

module github.com/user/myproject

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/lib/pq v1.10.9
    go.uber.org/zap v1.26.0
)

require (
    // Indirect dependencies (automatically managed)
    github.com/bytedance/sonic v1.9.1 // indirect
    github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
)

// Replace directive for local development
replace github.com/user/mylib => ../mylib

// Retract directive to mark bad versions
retract v1.0.1 // Contains critical bug
```

## Module Commands

```bash
# Initialize module
go mod init github.com/user/project

# Add missing dependencies
go mod tidy

# Download dependencies
go mod download

# Verify dependencies
go mod verify

# Show module graph
go mod graph

# Show why package is needed
go mod why github.com/user/package

# Vendor dependencies (copy to vendor/)
go mod vendor

# Update dependency
go get -u github.com/user/package

# Update to specific version
go get github.com/user/package@v1.2.3

# Update all dependencies
go get -u ./...

# Remove unused dependencies
go mod tidy
```

## Internal Packages

```go
// internal/ packages can only be imported by code in the parent tree

myproject/
├── internal/
│   ├── auth/           # Can only be imported by myproject
│   │   └── jwt.go
│   └── database/
│       └── postgres.go
└── pkg/
    └── models/         # Can be imported by anyone
        └── user.go

// This works (same project):
import "github.com/user/myproject/internal/auth"

// This fails (different project):
import "github.com/other/project/internal/auth" // Error!

// Internal subdirectories
myproject/
└── api/
    └── internal/       # Can only be imported by code in api/
        └── helpers.go
```

## Package Organization

```go
// user/user.go - Domain package
package user

import (
    "context"
    "time"
)

// User represents a user entity
type User struct {
    ID        string
    Email     string
    CreatedAt time.Time
}

// Repository defines data access interface
type Repository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}

// Service handles business logic
type Service struct {
    repo Repository
}

// NewService creates a new user service
func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) RegisterUser(ctx context.Context, email string) (*User, error) {
    user := &User{
        ID:        generateID(),
        Email:     email,
        CreatedAt: time.Now(),
    }
    return user, s.repo.Create(ctx, user)
}
```

## Multi-Module Repository (Monorepo)

```
monorepo/
├── go.work              # Workspace file
├── services/
│   ├── api/
│   │   ├── go.mod
│   │   └── main.go
│   └── worker/
│       ├── go.mod
│       └── main.go
└── shared/
    └── models/
        ├── go.mod
        └── user.go

// go.work
go 1.21

use (
    ./services/api
    ./services/worker
    ./shared/models
)

// Commands:
// go work init ./services/api ./services/worker
// go work use ./shared/models
// go work sync
```

## Build Tags and Constraints

```go
// +build integration
// integration_test.go

package myapp

import "testing"

func TestIntegration(t *testing.T) {
    // Integration test code
}

// Build: go test -tags=integration

// File-level build constraints (Go 1.17+)
//go:build linux && amd64

package myapp

// Multiple constraints
//go:build linux || darwin
//go:build amd64

// Negation
//go:build !windows

// Common tags:
// linux, darwin, windows, freebsd
// amd64, arm64, 386, arm
// cgo, !cgo
```

## Makefile Example

```makefile
# Makefile
.PHONY: build test lint clean run

# Variables
BINARY_NAME=myapp
BUILD_DIR=bin
GO=go
GOFLAGS=-v

# Build the application
build:
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

# Run tests
test:
	$(GO) test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	$(GO) tool cover -html=coverage.out

# Run linters
lint:
	golangci-lint run ./...

# Format code
fmt:
	$(GO) fmt ./...
	goimports -w .

# Run the application
run:
	$(GO) run ./cmd/server

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out

# Install dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/server
	GOOS=darwin GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/server
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/server

# Run with race detector
run-race:
	$(GO) run -race ./cmd/server

# Generate code
generate:
	$(GO) generate ./...

# Docker build
docker-build:
	docker build -t $(BINARY_NAME):latest .

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run linters"
	@echo "  fmt           - Format code"
	@echo "  run           - Run the application"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Install dependencies"
```

## Dockerfile Multi-Stage Build

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/server .

# Copy config files if needed
COPY --from=builder /app/configs ./configs

EXPOSE 8080

CMD ["./server"]
```

## Version Information

```go
// version/version.go
package version

import "runtime"

var (
    // Set via ldflags during build
    Version   = "dev"
    GitCommit = "none"
    BuildTime = "unknown"
)

// Info returns version information
func Info() map[string]string {
    return map[string]string{
        "version":    Version,
        "git_commit": GitCommit,
        "build_time": BuildTime,
        "go_version": runtime.Version(),
        "os":         runtime.GOOS,
        "arch":       runtime.GOARCH,
    }
}

// Build with version info:
// go build -ldflags "-X github.com/user/project/version.Version=1.0.0 \
//   -X github.com/user/project/version.GitCommit=$(git rev-parse HEAD) \
//   -X github.com/user/project/version.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

## Go Generate

```go
// models/user.go
//go:generate mockgen -source=user.go -destination=../mocks/user_mock.go -package=mocks

package models

type UserRepository interface {
    GetUser(id string) (*User, error)
    SaveUser(user *User) error
}

// tools.go - Track tool dependencies
//go:build tools

package tools

import (
    _ "github.com/golang/mock/mockgen"
    _ "golang.org/x/tools/cmd/stringer"
)

// Install tools:
// go install github.com/golang/mock/mockgen@latest

// Run generate:
// go generate ./...
```

## Configuration Management

```go
// config/config.go
package config

import (
    "os"
    "time"

    "github.com/kelseyhightower/envconfig"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
}

type ServerConfig struct {
    Host         string        `envconfig:"SERVER_HOST" default:"0.0.0.0"`
    Port         int           `envconfig:"SERVER_PORT" default:"8080"`
    ReadTimeout  time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"10s"`
    WriteTimeout time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"10s"`
}

type DatabaseConfig struct {
    URL          string `envconfig:"DATABASE_URL" required:"true"`
    MaxOpenConns int    `envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
    MaxIdleConns int    `envconfig:"DB_MAX_IDLE_CONNS" default:"5"`
}

type RedisConfig struct {
    Addr     string `envconfig:"REDIS_ADDR" default:"localhost:6379"`
    Password string `envconfig:"REDIS_PASSWORD"`
    DB       int    `envconfig:"REDIS_DB" default:"0"`
}

// Load loads configuration from environment
func Load() (*Config, error) {
    var cfg Config
    if err := envconfig.Process("", &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}
```

## Quick Reference

| Command | Description |
|---------|-------------|
| `go mod init` | Initialize module |
| `go mod tidy` | Add/remove dependencies |
| `go mod download` | Download dependencies |
| `go get package@version` | Add/update dependency |
| `go build -ldflags "-X ..."` | Set version info |
| `go generate ./...` | Run code generation |
| `GOOS=linux go build` | Cross-compile |
| `go work init` | Initialize workspace |

---

## Reference: testing

# Testing and Benchmarking

## Table-Driven Tests

```go
package math

import "testing"

func Add(a, b int) int {
    return a + b
}

func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -2, -3, -5},
        {"mixed signs", -2, 3, 1},
        {"zeros", 0, 0, 0},
        {"large numbers", 1000000, 2000000, 3000000},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

## Subtests and Parallel Execution

```go
func TestParallel(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  string
    }{
        {"lowercase", "hello", "HELLO"},
        {"uppercase", "WORLD", "WORLD"},
        {"mixed", "HeLLo", "HELLO"},
    }

    for _, tt := range tests {
        tt := tt // Capture range variable for parallel tests
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // Run subtests in parallel

            result := strings.ToUpper(tt.input)
            if result != tt.want {
                t.Errorf("got %q, want %q", result, tt.want)
            }
        })
    }
}
```

## Test Helpers and Setup/Teardown

```go
func TestWithSetup(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    tests := []struct {
        name string
        user User
    }{
        {"valid user", User{Name: "John", Email: "john@example.com"}},
        {"empty name", User{Name: "", Email: "test@example.com"}},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := db.SaveUser(tt.user)
            if err != nil {
                t.Fatalf("SaveUser failed: %v", err)
            }
        })
    }
}

// Helper function (doesn't show in stack trace)
func setupTestDB(t *testing.T) *DB {
    t.Helper()

    db, err := NewDB(":memory:")
    if err != nil {
        t.Fatalf("failed to create test DB: %v", err)
    }
    return db
}

func cleanupTestDB(t *testing.T, db *DB) {
    t.Helper()

    if err := db.Close(); err != nil {
        t.Errorf("failed to close DB: %v", err)
    }
}
```

## Mocking with Interfaces

```go
// Interface to mock
type EmailSender interface {
    Send(to, subject, body string) error
}

// Mock implementation
type MockEmailSender struct {
    SentEmails []Email
    ShouldFail bool
}

type Email struct {
    To, Subject, Body string
}

func (m *MockEmailSender) Send(to, subject, body string) error {
    if m.ShouldFail {
        return fmt.Errorf("failed to send email")
    }
    m.SentEmails = append(m.SentEmails, Email{to, subject, body})
    return nil
}

// Test using mock
func TestUserService_Register(t *testing.T) {
    mockSender := &MockEmailSender{}
    service := NewUserService(mockSender)

    err := service.Register("user@example.com")
    if err != nil {
        t.Fatalf("Register failed: %v", err)
    }

    if len(mockSender.SentEmails) != 1 {
        t.Errorf("expected 1 email sent, got %d", len(mockSender.SentEmails))
    }

    email := mockSender.SentEmails[0]
    if email.To != "user@example.com" {
        t.Errorf("expected email to user@example.com, got %s", email.To)
    }
}
```

## Benchmarking

```go
func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(100, 200)
    }
}

// Benchmark with subtests
func BenchmarkStringOperations(b *testing.B) {
    benchmarks := []struct {
        name  string
        input string
    }{
        {"short", "hello"},
        {"medium", strings.Repeat("hello", 10)},
        {"long", strings.Repeat("hello", 100)},
    }

    for _, bm := range benchmarks {
        b.Run(bm.name, func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                _ = strings.ToUpper(bm.input)
            }
        })
    }
}

// Benchmark with setup
func BenchmarkMapOperations(b *testing.B) {
    m := make(map[string]int)
    for i := 0; i < 1000; i++ {
        m[fmt.Sprintf("key%d", i)] = i
    }

    b.ResetTimer() // Don't count setup time

    for i := 0; i < b.N; i++ {
        _ = m["key500"]
    }
}

// Parallel benchmark
func BenchmarkConcurrentAccess(b *testing.B) {
    var counter int64

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            atomic.AddInt64(&counter, 1)
        }
    })
}

// Memory allocation benchmark
func BenchmarkAllocation(b *testing.B) {
    b.ReportAllocs() // Report allocations

    for i := 0; i < b.N; i++ {
        s := make([]int, 1000)
        _ = s
    }
}
```

## Fuzzing (Go 1.18+)

```go
func FuzzReverse(f *testing.F) {
    // Seed corpus
    testcases := []string{"hello", "world", "123", ""}
    for _, tc := range testcases {
        f.Add(tc)
    }

    f.Fuzz(func(t *testing.T, input string) {
        reversed := Reverse(input)
        doubleReversed := Reverse(reversed)

        if input != doubleReversed {
            t.Errorf("Reverse(Reverse(%q)) = %q, want %q", input, doubleReversed, input)
        }
    })
}

// Fuzz with multiple parameters
func FuzzAdd(f *testing.F) {
    f.Add(1, 2)
    f.Add(0, 0)
    f.Add(-1, 1)

    f.Fuzz(func(t *testing.T, a, b int) {
        result := Add(a, b)

        // Properties that should always hold
        if result < a && b >= 0 {
            t.Errorf("Add(%d, %d) = %d; result should be >= a when b >= 0", a, b, result)
        }
    })
}
```

## Test Coverage

```go
// Run tests with coverage:
// go test -cover
// go test -coverprofile=coverage.out
// go tool cover -html=coverage.out

func TestCalculate(t *testing.T) {
    tests := []struct {
        name     string
        input    int
        expected int
    }{
        {"zero", 0, 0},
        {"positive", 5, 25},
        {"negative", -3, 9},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Calculate(tt.input)
            if result != tt.expected {
                t.Errorf("Calculate(%d) = %d; want %d", tt.input, result, tt.expected)
            }
        })
    }
}
```

## Race Detector

```go
// Run with: go test -race

func TestConcurrentAccess(t *testing.T) {
    var counter int
    var wg sync.WaitGroup

    // This will fail with -race if not synchronized
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter++ // Data race!
        }()
    }

    wg.Wait()
}

// Fixed version with mutex
func TestConcurrentAccessSafe(t *testing.T) {
    var counter int
    var mu sync.Mutex
    var wg sync.WaitGroup

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            mu.Lock()
            counter++
            mu.Unlock()
        }()
    }

    wg.Wait()

    if counter != 10 {
        t.Errorf("expected 10, got %d", counter)
    }
}
```

## Golden Files

```go
import (
    "os"
    "path/filepath"
    "testing"
)

func TestRenderHTML(t *testing.T) {
    data := Data{Title: "Test", Content: "Hello"}
    result := RenderHTML(data)

    goldenFile := filepath.Join("testdata", "expected.html")

    if *update {
        // Update golden file: go test -update
        os.WriteFile(goldenFile, []byte(result), 0644)
    }

    expected, err := os.ReadFile(goldenFile)
    if err != nil {
        t.Fatalf("failed to read golden file: %v", err)
    }

    if result != string(expected) {
        t.Errorf("output doesn't match golden file\ngot:\n%s\nwant:\n%s", result, expected)
    }
}

var update = flag.Bool("update", false, "update golden files")
```

## Integration Tests

```go
// integration_test.go
// +build integration

package myapp

import (
    "testing"
    "time"
)

func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    // Long-running integration test
    server := startTestServer(t)
    defer server.Stop()

    time.Sleep(100 * time.Millisecond) // Wait for server

    client := NewClient(server.URL)
    resp, err := client.Get("/health")
    if err != nil {
        t.Fatalf("health check failed: %v", err)
    }

    if resp.Status != "ok" {
        t.Errorf("expected status ok, got %s", resp.Status)
    }
}

// Run: go test -tags=integration
// Run short tests only: go test -short
```

## Testable Examples

```go
// Example tests that appear in godoc
func ExampleAdd() {
    result := Add(2, 3)
    fmt.Println(result)
    // Output: 5
}

func ExampleAdd_negative() {
    result := Add(-2, -3)
    fmt.Println(result)
    // Output: -5
}

// Unordered output
func ExampleKeys() {
    m := map[string]int{"a": 1, "b": 2, "c": 3}
    keys := Keys(m)
    for _, k := range keys {
        fmt.Println(k)
    }
    // Unordered output:
    // a
    // b
    // c
}
```

## Quick Reference

| Command | Description |
|---------|-------------|
| `go test` | Run tests |
| `go test -v` | Verbose output |
| `go test -run TestName` | Run specific test |
| `go test -bench .` | Run benchmarks |
| `go test -cover` | Show coverage |
| `go test -race` | Run race detector |
| `go test -short` | Skip long tests |
| `go test -fuzz FuzzName` | Run fuzzing |
| `go test -cpuprofile cpu.prof` | CPU profiling |
| `go test -memprofile mem.prof` | Memory profiling |
