# debugging-wizard

> Источник: https://github.com/jeffallan/claude-skills/tree/main/skills/debugging-wizard (MIT). Рабочая копия скилла: `.claude/skills/debugging-wizard/`.

---
name: debugging-wizard
description: Parses error messages, traces execution flow through stack traces, correlates log entries to identify failure points, and applies systematic hypothesis-driven methodology to isolate and resolve bugs. Use when investigating errors, analyzing stack traces, finding root causes of unexpected behavior, troubleshooting crashes, or performing log analysis, error investigation, or root cause analysis.
license: MIT
metadata:
  author: https://github.com/Jeffallan
  company: https://synergetic.solutions
  version: "1.1.0"
  domain: quality
  triggers: debug, error, bug, exception, traceback, stack trace, troubleshoot, not working, crash, fix issue
  role: specialist
  scope: analysis
  output-format: analysis
  related-skills: test-master, fullstack-guardian, monitoring-expert
---

# Debugging Wizard

Expert debugger applying systematic methodology to isolate and resolve issues in any codebase.

## Core Workflow

1. **Reproduce** - Establish consistent reproduction steps
2. **Isolate** - Narrow down to smallest failing case
3. **Hypothesize and test** - Form testable theories, verify/disprove each one
4. **Fix** - Implement and verify solution
5. **Prevent** - Add tests/safeguards against regression

## Reference Guide

Load detailed guidance based on context:

<!-- Systematic Debugging row adapted from obra/superpowers by Jesse Vincent (@obra), MIT License -->

| Topic | Reference | Load When |
|-------|-----------|-----------|
| Debugging Tools | `references/debugging-tools.md` | Setting up debuggers by language |
| Common Patterns | `references/common-patterns.md` | Recognizing bug patterns |
| Strategies | `references/strategies.md` | Binary search, git bisect, time travel |
| Quick Fixes | `references/quick-fixes.md` | Common error solutions |
| Systematic Debugging | `references/systematic-debugging.md` | Complex bugs, multiple failed fixes, root cause analysis |

## Constraints

### MUST DO
- Reproduce the issue first
- Gather complete error messages and stack traces
- Test one hypothesis at a time
- Document findings for future reference
- Add regression tests after fixing
- Remove all debug code before committing

### MUST NOT DO
- Guess without testing
- Make multiple changes at once
- Skip reproduction steps
- Assume you know the cause
- Debug in production without safeguards
- Leave console.log/debugger statements in code

## Common Debugging Commands

**Python (pdb)**
```bash
python -m pdb script.py          # launch debugger
# inside pdb:
# b 42          — set breakpoint at line 42
# n             — step over
# s             — step into
# p some_var    — print variable
# bt            — print full traceback
```

**JavaScript (Node.js)**
```bash
node --inspect-brk script.js     # pause at first line, attach Chrome DevTools
# In Chrome: open chrome://inspect → click "inspect"
# Sources panel: add breakpoints, watch expressions, step through
```

**Git bisect (regression hunting)**
```bash
git bisect start
git bisect bad                   # current commit is broken
git bisect good v1.2.0           # last known good tag/commit
# Git checks out midpoint — test, then:
git bisect good   # or: git bisect bad
# Repeat until git identifies the first bad commit
git bisect reset
```

**Go (delve)**
```bash
dlv debug ./cmd/server           # build & attach
# (dlv) break main.go:55
# (dlv) continue
# (dlv) print myVar
```

## Output Templates

When debugging, provide:
1. **Root Cause**: What specifically caused the issue
2. **Evidence**: Stack trace, logs, or test that proves it
3. **Fix**: Code change that resolves it
4. **Prevention**: Test or safeguard to prevent recurrence

Maintained by [@jeffallan](https://github.com/jeffallan), Principal Consultant at [Synergetic Solutions](https://synergetic.solutions)

[Documentation](https://jeffallan.github.io/claude-skills/skills/quality/debugging-wizard/)

---

## Reference: common-patterns

# Common Bug Patterns

## Pattern Recognition

| Pattern | Symptom | Likely Cause |
|---------|---------|--------------|
| Race condition | Intermittent failures | Missing await, async timing |
| Off-by-one | Missing first/last item | `<` vs `<=`, array bounds |
| Null reference | "undefined is not..." | Missing null check |
| Memory leak | Growing memory | Uncleaned listeners/intervals |
| N+1 queries | Slow with more data | Fetching in loop |
| Type coercion | Unexpected behavior | `==` instead of `===` |
| Closure issue | Wrong variable value | Loop variable capture |
| Stale state | Old value used | React state closure |

## Race Condition

```typescript
// BUG: Race condition
let data;
fetchData().then(result => { data = result; });
console.log(data); // undefined!

// FIX: Await the result
const data = await fetchData();
console.log(data);
```

## Off-by-One

```typescript
// BUG: Skips last element
for (let i = 0; i < array.length - 1; i++) { }

// FIX: Include last element
for (let i = 0; i < array.length; i++) { }

// BUG: Array index out of bounds
const last = array[array.length]; // undefined

// FIX: Correct index
const last = array[array.length - 1];
```

## Null Reference

```typescript
// BUG: Crashes if user is null
const name = user.profile.name;

// FIX: Optional chaining
const name = user?.profile?.name ?? 'Unknown';

// FIX: Guard clause
if (!user?.profile) {
  return 'Unknown';
}
return user.profile.name;
```

## Memory Leak

```typescript
// BUG: Listener never removed
useEffect(() => {
  window.addEventListener('resize', handleResize);
}, []);

// FIX: Cleanup function
useEffect(() => {
  window.addEventListener('resize', handleResize);
  return () => window.removeEventListener('resize', handleResize);
}, []);

// BUG: Interval never cleared
setInterval(pollData, 1000);

// FIX: Store and clear
const intervalId = setInterval(pollData, 1000);
return () => clearInterval(intervalId);
```

## Closure in Loop

```typescript
// BUG: All callbacks use i = 5
for (var i = 0; i < 5; i++) {
  setTimeout(() => console.log(i), 100);
}

// FIX: Use let (block scoped)
for (let i = 0; i < 5; i++) {
  setTimeout(() => console.log(i), 100);
}

// FIX: Capture in closure
for (var i = 0; i < 5; i++) {
  ((j) => setTimeout(() => console.log(j), 100))(i);
}
```

## React Stale State

```typescript
// BUG: count is stale in closure
const [count, setCount] = useState(0);
useEffect(() => {
  setInterval(() => {
    setCount(count + 1); // Always uses initial count
  }, 1000);
}, []);

// FIX: Use functional update
setCount(prev => prev + 1);

// FIX: Include in dependency array with cleanup
useEffect(() => {
  const id = setInterval(() => setCount(c => c + 1), 1000);
  return () => clearInterval(id);
}, []);
```

## Quick Reference

| Symptom | First Check |
|---------|-------------|
| "undefined is not..." | Null check missing |
| Works sometimes | Race condition |
| Wrong value in callback | Closure/stale state |
| Gets slower over time | Memory leak, N+1 |
| Off by one item | Loop bounds, array index |
| Type mismatch | `==` vs `===`, coercion |

---

## Reference: debugging-tools

# Debugging Tools

## Debuggers by Language

| Language | Debugger | Start Command |
|----------|----------|---------------|
| TypeScript/JS | Node Inspector | `node --inspect` |
| Python | pdb/ipdb | `python -m pdb` |
| Go | Delve | `dlv debug` |
| Rust | rust-gdb/lldb | `rust-gdb ./target/debug/app` |
| Java | JDB/IDE | IDE debugger |

## Node.js / TypeScript

```bash
# Start with inspector
node --inspect dist/main.js

# Break on first line
node --inspect-brk dist/main.js

# With ts-node
node --inspect -r ts-node/register src/main.ts
```

```typescript
// In code
debugger; // Breakpoint

// Quick print
console.log({ variable }); // Shows name and value
console.table(arrayOfObjects); // Table format
console.trace('Called from'); // Stack trace
```

## Python

```bash
# Start debugger
python -m pdb script.py

# Post-mortem on exception
python -m pdb -c continue script.py
```

```python
# In code
breakpoint()  # Python 3.7+
import pdb; pdb.set_trace()  # Older Python

# Quick print
print(f"{variable=}")  # Python 3.8+ shows name and value

# Rich debugging
from rich import inspect
inspect(object, methods=True)
```

### pdb Commands

| Command | Action |
|---------|--------|
| `n` | Next line |
| `s` | Step into |
| `c` | Continue |
| `l` | List code |
| `p expr` | Print expression |
| `pp expr` | Pretty print |
| `w` | Where (stack) |
| `q` | Quit |

## Go

```bash
# Start delve
dlv debug ./cmd/app

# Attach to running process
dlv attach <pid>

# Debug test
dlv test ./pkg/...
```

```go
// Quick print
log.Printf("%+v", variable) // With field names
fmt.Printf("%#v\n", variable) // Go syntax representation

// Spew for complex structures
import "github.com/davecgh/go-spew/spew"
spew.Dump(variable)
```

### Delve Commands

| Command | Action |
|---------|--------|
| `break main.go:42` | Set breakpoint |
| `continue` | Continue |
| `next` | Next line |
| `step` | Step into |
| `print var` | Print variable |
| `goroutines` | List goroutines |

## VS Code Debug Config

```json
// .vscode/launch.json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "node",
      "request": "launch",
      "name": "Debug TypeScript",
      "program": "${workspaceFolder}/src/main.ts",
      "preLaunchTask": "tsc: build",
      "outFiles": ["${workspaceFolder}/dist/**/*.js"]
    },
    {
      "type": "python",
      "request": "launch",
      "name": "Debug Python",
      "program": "${workspaceFolder}/main.py",
      "console": "integratedTerminal"
    }
  ]
}
```

## Quick Reference

| Need | Tool |
|------|------|
| Breakpoint in code | `debugger;` / `breakpoint()` |
| Print with name | `console.log({x})` / `print(f"{x=}")` |
| Stack trace | `console.trace()` / `traceback.print_stack()` |
| Inspect object | `console.dir(obj)` / `dir(obj)` |
| Step through | IDE debugger or CLI debugger |

---

## Reference: quick-fixes

# Quick Fixes

## TypeError: Cannot read property 'x' of undefined

```typescript
// Error
user.profile.name
// user or profile is undefined

// Fix: Optional chaining
user?.profile?.name

// Fix: Default value
user?.profile?.name ?? 'Unknown'

// Fix: Guard clause
if (!user?.profile) {
  return null;
}
return user.profile.name;
```

## Unhandled Promise Rejection

```typescript
// Error
fetchData().then(process);
// What if fetchData rejects?

// Fix: Add catch
fetchData()
  .then(process)
  .catch(error => {
    console.error('Fetch failed:', error);
  });

// Fix: try/catch with await
try {
  const data = await fetchData();
  await process(data);
} catch (error) {
  console.error('Operation failed:', error);
}
```

## React: Too Many Re-renders

```typescript
// Error: Calling setState during render
function Component() {
  const [count, setCount] = useState(0);
  setCount(count + 1); // Infinite loop!
}

// Fix: Use useEffect for side effects
function Component() {
  const [count, setCount] = useState(0);
  useEffect(() => {
    setCount(c => c + 1);
  }, []); // Only on mount
}

// Error: Object/array in dependency array
useEffect(() => {}, [{ a: 1 }]); // New object every render!

// Fix: Memoize or use primitives
const config = useMemo(() => ({ a: 1 }), []);
useEffect(() => {}, [config]);
```

## CORS Error

```typescript
// Browser blocks cross-origin request

// Fix 1: Server - Add CORS headers
app.use(cors({
  origin: 'http://localhost:3000',
  credentials: true,
}));

// Fix 2: Proxy in development (Vite)
// vite.config.ts
export default {
  server: {
    proxy: {
      '/api': 'http://localhost:8000',
    },
  },
};
```

## Maximum Call Stack Size Exceeded

```typescript
// Error: Infinite recursion
function factorial(n) {
  return n * factorial(n - 1); // No base case!
}

// Fix: Add base case
function factorial(n) {
  if (n <= 1) return 1;
  return n * factorial(n - 1);
}

// Error: Circular dependency in objects
const a = {};
const b = { ref: a };
a.ref = b;
JSON.stringify(a); // Fails!

// Fix: Break circular reference
JSON.stringify(a, (key, value) => {
  if (key === 'ref') return '[Circular]';
  return value;
});
```

## Module Not Found

```bash
# Error: Cannot find module 'x'

# Fix 1: Install the package
npm install x

# Fix 2: Check import path
import x from './x';     # Relative - needs ./
import x from 'x';       # Package - no ./

# Fix 3: Check file extension
import x from './x.js';  # ESM may need extension

# Fix 4: Clear cache
rm -rf node_modules package-lock.json
npm install
```

## Async/Await Issues

```typescript
// Error: await in non-async function
function getData() {
  const data = await fetch('/api'); // SyntaxError!
}

// Fix: Mark function as async
async function getData() {
  const data = await fetch('/api');
}

// Error: forEach doesn't await
items.forEach(async item => {
  await process(item); // Doesn't wait!
});

// Fix: Use for...of
for (const item of items) {
  await process(item);
}

// Fix: Use Promise.all for parallel
await Promise.all(items.map(item => process(item)));
```

## Quick Reference

| Error Message | Likely Fix |
|--------------|------------|
| Cannot read property of undefined | Optional chaining `?.` |
| Unhandled promise rejection | Add `.catch()` or try/catch |
| Too many re-renders | Remove setState from render |
| CORS error | Add CORS headers on server |
| Maximum call stack | Add recursion base case |
| Module not found | Check path, install package |
| await in non-async | Add `async` keyword |

---

## Reference: strategies

# Debugging Strategies

## Binary Search

Divide and conquer to find the bug location.

```markdown
1. Comment out/disable half the code
2. Test if bug still occurs
3. If yes: bug is in remaining half
4. If no: bug is in disabled half
5. Repeat until isolated
```

```typescript
// Example: Bug in data processing pipeline
async function process(data) {
  const step1 = await transform(data);
  // Bug somewhere below?

  const step2 = await validate(step1);
  console.log('After step2:', step2); // Check here

  const step3 = await enrich(step2);
  const step4 = await save(step3);
  return step4;
}
```

## Minimal Reproduction

Strip away everything until only the bug remains.

```markdown
1. Create new minimal project
2. Add only code needed to reproduce
3. Remove dependencies one by one
4. Simplify inputs to smallest failing case
5. Document exact reproduction steps
```

```typescript
// Instead of debugging full app
// Create minimal test case:
const input = { id: null }; // Minimal failing input
const result = processUser(input);
console.log(result); // Isolate the exact failure
```

## Git Bisect

Find the commit that introduced the bug.

```bash
# Start bisect
git bisect start

# Mark current commit as bad
git bisect bad

# Mark known good commit
git bisect good v1.0.0

# Git checks out middle commit
# Test and mark:
git bisect good  # or
git bisect bad

# Repeat until found
# Git will say: "abc123 is the first bad commit"

# End bisect
git bisect reset
```

```bash
# Automated bisect with test script
git bisect start HEAD v1.0.0
git bisect run npm test
```

## Time Travel Debugging

Work backwards from the failure.

```markdown
1. Start at the error/failure point
2. What value caused it? Where did that come from?
3. Trace backwards through the code
4. Find where the value diverged from expected
```

```typescript
// Error: Cannot read 'name' of undefined at line 45

// Line 45: const name = user.name;
// Q: Why is user undefined?

// Line 40: const user = users.find(u => u.id === id);
// Q: Why didn't find() return a user?

// Check: Is the id correct? Are users populated?
console.log({ id, users, user });
```

## Rubber Duck Debugging

Explain the problem step by step.

```markdown
1. State what the code should do
2. Explain what it actually does
3. Walk through the code line by line
4. Describe what each line does
5. The discrepancy often becomes obvious
```

## Delta Debugging

When something recently broke.

```bash
# Check what changed
git diff HEAD~5..HEAD

# Check specific file history
git log -p --follow -- src/problematic-file.ts

# Find when file last worked
git log --oneline -- src/problematic-file.ts
```

## Quick Reference

| Strategy | Best For |
|----------|----------|
| Binary Search | Unknown bug location |
| Minimal Repro | Complex bugs, reporting |
| Git Bisect | Regression bugs |
| Time Travel | Known error location |
| Rubber Duck | Logic errors |
| Delta Debug | Recent breakage |

---

## Reference: systematic-debugging

# Systematic Debugging

---

## Core Principle

> **NO FIXES WITHOUT ROOT CAUSE INVESTIGATION FIRST.**

Jumping to fixes without understanding causes creates more bugs. Systematic debugging prevents the "fix one thing, break two more" cycle.

---

## The Four Mandatory Phases

```
┌─────────────────────────────────────────────────────────────┐
│                    SYSTEMATIC DEBUGGING                      │
├─────────────────────────────────────────────────────────────┤
│  Phase 1: ROOT CAUSE INVESTIGATION                          │
│  ├── Read error messages thoroughly                         │
│  ├── Reproduce reliably with documented steps               │
│  ├── Examine recent changes                                 │
│  └── Trace data flow backward                               │
├─────────────────────────────────────────────────────────────┤
│  Phase 2: PATTERN ANALYSIS                                   │
│  ├── Find similar working implementations                   │
│  ├── Study reference implementations completely             │
│  └── Document all differences                               │
├─────────────────────────────────────────────────────────────┤
│  Phase 3: HYPOTHESIS TESTING                                 │
│  ├── Form specific, written hypothesis                      │
│  ├── Test with minimal, isolated changes                    │
│  └── One variable at a time                                 │
├─────────────────────────────────────────────────────────────┤
│  Phase 4: IMPLEMENTATION                                     │
│  ├── Create failing test case                               │
│  ├── Implement single fix addressing root cause             │
│  └── Verify no new breakage                                 │
└─────────────────────────────────────────────────────────────┘
```

---

## Phase 1: Root Cause Investigation

**Objective:** Understand exactly what is failing and why before attempting any fix.

### Step 1.1: Read Error Messages Thoroughly

```bash
# Don't just read the first line
TypeError: Cannot read property 'map' of undefined
    at UserList.render (UserList.tsx:24)
    at renderWithHooks (react-dom.js:14985)
    at mountIndeterminateComponent (react-dom.js:17811)
```

**Key questions:**
- What exact operation failed?
- Where in the code (file, line)?
- What was the call stack?
- Are there multiple errors or just one?

### Step 1.2: Reproduce Reliably

```markdown
## Reproduction Steps
1. Navigate to /users
2. Click "Load More" button
3. Wait for loading spinner
4. **ERROR: "Cannot read property 'map' of undefined"**

## Environment
- Browser: Chrome 120
- User: Admin role
- Data state: 50+ users in database
```

**Requirement:** Document exact steps that reproduce the bug 100% of the time.

### Step 1.3: Examine Recent Changes

```bash
# What changed recently?
git log --oneline -10

# What specifically changed in the failing file?
git log -p UserList.tsx

# When did this start failing?
git bisect start
git bisect bad HEAD
git bisect good v1.2.0
```

### Step 1.4: Trace Data Flow Backward

```typescript
// Error happens here:
users.map(u => u.name)  // users is undefined

// Trace backward:
// Where does 'users' come from?
const users = props.users;

// Where do props come from?
<UserList users={data.users} />

// Where does data come from?
const { data } = useQuery(GET_USERS);

// ROOT CAUSE: Query returns { users: null } when loading
```

### Step 1.5: Add Diagnostic Instrumentation

```typescript
// Add temporary logging at boundaries
console.log('[UserList] props:', JSON.stringify(props));
console.log('[UserList] users type:', typeof props.users);
console.log('[UserList] users value:', props.users);

// Check at data source
console.log('[API] Response:', response);
console.log('[API] Response.data:', response.data);
```

---

## Phase 2: Pattern Analysis

**Objective:** Find working examples to understand what correct behavior looks like.

### Step 2.1: Locate Similar Working Implementations

```bash
# Find similar components that work correctly
grep -r "useQuery" src/components/ --include="*.tsx"

# Find how other lists handle loading states
grep -r "loading" src/components/*List* --include="*.tsx"
```

### Step 2.2: Study Reference Implementations Completely

```typescript
// WORKING: ProductList.tsx
function ProductList({ products, loading }) {
  if (loading) return <Spinner />;
  if (!products) return null;  // ← Handles undefined case

  return products.map(p => <ProductItem key={p.id} {...p} />);
}

// BROKEN: UserList.tsx
function UserList({ users, loading }) {
  if (loading) return <Spinner />;
  // Missing: !users check

  return users.map(u => <UserItem key={u.id} {...u} />);  // 💥 Crashes
}
```

### Step 2.3: Document All Differences

| Aspect | Working (ProductList) | Broken (UserList) |
|--------|----------------------|-------------------|
| Null check | `if (!products)` | Missing |
| Default value | `products ?? []` | None |
| Loading handled | Before render | Before render |
| Error handled | Returns ErrorState | Missing |

---

## Phase 3: Hypothesis Testing

**Objective:** Verify your understanding with controlled experiments.

### Step 3.1: Form Specific, Written Hypothesis

```markdown
## Hypothesis #1
**Statement:** The crash occurs because `users` is undefined when the
query is complete but returns no data.

**Prediction:** Adding a null check before `.map()` will prevent the crash.

**Test:** Add `if (!users) return null;` before the map call.
```

### Step 3.2: Test with Minimal Changes

```typescript
// Change ONLY one thing
function UserList({ users, loading }) {
  if (loading) return <Spinner />;
  if (!users) return null;  // ← Single change

  return users.map(u => <UserItem key={u.id} {...u} />);
}
```

### Step 3.3: One Variable at a Time

```markdown
## Test Results

| Hypothesis | Change | Result | Conclusion |
|------------|--------|--------|------------|
| #1: Null check | Add `if (!users)` | ✓ Pass | Confirmed |

Do NOT test multiple hypotheses simultaneously.
```

---

## Phase 4: Implementation

**Objective:** Fix the bug permanently with proper safeguards.

### Step 4.1: Create Failing Test Case First

```typescript
describe('UserList', () => {
  it('should handle undefined users gracefully', () => {
    // This test should FAIL before the fix
    const { container } = render(<UserList users={undefined} loading={false} />);
    expect(container).not.toThrow();
    expect(screen.queryByRole('list')).not.toBeInTheDocument();
  });
});
```

### Step 4.2: Implement Single Fix

```typescript
function UserList({ users, loading }: UserListProps) {
  if (loading) return <Spinner />;
  if (!users || users.length === 0) {
    return <EmptyState message="No users found" />;
  }

  return (
    <ul role="list">
      {users.map(u => <UserItem key={u.id} {...u} />)}
    </ul>
  );
}
```

### Step 4.3: Verify No New Breakage

```bash
# Run full test suite
npm test

# Run specific component tests
npm test UserList

# Run integration tests
npm run test:integration

# Verify in browser
# 1. Normal case: 50 users
# 2. Empty case: 0 users
# 3. Loading case: spinner shows
# 4. Error case: error message shows
```

---

## The Three-Fix Threshold

> **After 3 failed fix attempts → STOP.**

Three failures in different locations signals architectural problems, not isolated bugs.

### What Three Failures Means

```
Fix Attempt 1: Added null check → New error in child component
Fix Attempt 2: Fixed child component → New error in parent
Fix Attempt 3: Fixed parent → Original error returns
                              ↓
                    STOP. QUESTION ARCHITECTURE.
```

### At the Threshold, Do This

1. **Stop fixing symptoms**
2. **Document the pattern** of failures
3. **Identify architectural assumptions** being violated
4. **Propose structural change** rather than patch
5. **Discuss with team** before proceeding

---

## Red Flags Requiring Process Reset

When you notice these, stop and restart from Phase 1:

| Red Flag | Why It's Wrong |
|----------|----------------|
| Proposing solutions before tracing data flow | Guessing, not debugging |
| Making multiple simultaneous changes | Can't identify which change worked |
| Skipping test creation | Bug will recur |
| "Let's try this and see if it works" | Shotgun debugging |
| Fixing without understanding the cause | Band-aid, not cure |

---

## Decision Flowchart

```
                    ┌──────────────────┐
                    │   Bug Reported   │
                    └────────┬─────────┘
                             │
              ┌──────────────▼──────────────┐
              │   Can you reproduce it?      │
              └──────────────┬──────────────┘
                    No       │       Yes
            ┌────────────────┴────────────────┐
            ▼                                  ▼
    ┌───────────────┐               ┌─────────────────┐
    │ Get more info │               │ Trace data flow │
    └───────────────┘               └────────┬────────┘
                                             │
                              ┌──────────────▼──────────────┐
                              │ Do you understand the cause? │
                              └──────────────┬──────────────┘
                                    No       │       Yes
                    ┌────────────────────────┴─────────┐
                    ▼                                   ▼
            ┌───────────────┐               ┌─────────────────┐
            │ Study working │               │ Write hypothesis│
            │   examples    │               └────────┬────────┘
            └───────────────┘                        │
                                             ┌───────▼───────┐
                                             │  Write test   │
                                             └───────┬───────┘
                                                     │
                                             ┌───────▼───────┐
                                             │  Implement    │
                                             └───────┬───────┘
                                                     │
                                  ┌──────────────────▼──────────────────┐
                                  │          Does test pass?            │
                                  └──────────────────┬──────────────────┘
                                            No       │       Yes
                            ┌────────────────────────┴──────────┐
                            ▼                                    ▼
                    ┌───────────────┐                  ┌─────────────────┐
                    │ Attempt < 3?  │                  │      Done       │
                    └───────┬───────┘                  └─────────────────┘
                    No      │      Yes
            ┌───────────────┴─────────────────┐
            ▼                                  ▼
    ┌───────────────────┐          ┌─────────────────────┐
    │ Question          │          │ Return to Phase 1   │
    │ architecture      │          └─────────────────────┘
    └───────────────────┘
```

---

*Content adapted from [obra/superpowers](https://github.com/obra/superpowers) by Jesse Vincent (@obra), MIT License.*
