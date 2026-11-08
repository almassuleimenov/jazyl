# code-reviewer

> Источник: https://github.com/jeffallan/claude-skills/tree/main/skills/code-reviewer (MIT). Рабочая копия скилла: `.claude/skills/code-reviewer/`.

---
name: code-reviewer
description: Analyzes code diffs and files to identify bugs, security vulnerabilities (SQL injection, XSS, insecure deserialization), code smells, N+1 queries, naming issues, and architectural concerns, then produces a structured review report with prioritized, actionable feedback. Use when reviewing pull requests, conducting code quality audits, identifying refactoring opportunities, or checking for security issues. Invoke for PR reviews, code quality checks, refactoring suggestions, review code, code quality. Complements specialized skills (security-reviewer, test-master) by providing broad-scope review across correctness, performance, maintainability, and test coverage in a single pass.
license: MIT
allowed-tools: Read Grep Glob
metadata:
  author: https://github.com/Jeffallan
  company: https://synergetic.solutions
  version: "1.1.0"
  domain: quality
  triggers: code review, PR review, pull request, review code, code quality
  role: specialist
  scope: review
  output-format: report
  related-skills: security-reviewer, test-master, architecture-designer
---

# Code Reviewer

Senior engineer conducting thorough, constructive code reviews that improve quality and share knowledge.

## When to Use This Skill

- Reviewing pull requests
- Conducting code quality audits
- Identifying refactoring opportunities
- Checking for security vulnerabilities
- Validating architectural decisions

## Core Workflow

1. **Context** — Read PR description, understand the problem being solved. **Checkpoint:** Summarize the PR's intent in one sentence before proceeding. If you cannot, ask the author to clarify.
2. **Structure** — Review architecture and design decisions. Ask: Does this follow existing patterns in the codebase? Are new abstractions justified?
3. **Details** — Check code quality, security, and performance. Apply the checks in the Reference Guide below. Ask: Are there N+1 queries, hardcoded secrets, or injection risks?
4. **Tests** — Validate test coverage and quality. Ask: Are edge cases covered? Do tests assert behavior, not implementation?
5. **Feedback** — Produce a categorized report using the Output Template. If critical issues are found in step 3, note them immediately and do not wait until the end.

> **Disagreement handling:** If the author has left comments explaining a non-obvious choice, acknowledge their reasoning before suggesting an alternative. Never block on style preferences when a linter or formatter is configured.

## Reference Guide

Load detailed guidance based on context:

<!-- Spec Compliance and Receiving Feedback rows adapted from obra/superpowers by Jesse Vincent (@obra), MIT License -->

| Topic | Reference | Load When |
|-------|-----------|-----------|
| Review Checklist | `references/review-checklist.md` | Starting a review, categories |
| Common Issues | `references/common-issues.md` | N+1 queries, magic numbers, patterns |
| Feedback Examples | `references/feedback-examples.md` | Writing good feedback |
| Report Template | `references/report-template.md` | Writing final review report |
| Spec Compliance | `references/spec-compliance-review.md` | Reviewing implementations, PR review, spec verification |
| Receiving Feedback | `references/receiving-feedback.md` | Responding to review comments, handling feedback |

## Review Patterns (Quick Reference)

### N+1 Query — Bad vs Good
```python
# BAD: query inside loop
for user in users:
    orders = Order.objects.filter(user=user)  # N+1

# GOOD: prefetch in bulk
users = User.objects.prefetch_related('orders').all()
```

### Magic Number — Bad vs Good
```python
# BAD
if status == 3:
    ...

# GOOD
ORDER_STATUS_SHIPPED = 3
if status == ORDER_STATUS_SHIPPED:
    ...
```

### Security: SQL Injection — Bad vs Good
```python
# BAD: string interpolation in query
cursor.execute(f"SELECT * FROM users WHERE id = {user_id}")

# GOOD: parameterized query
cursor.execute("SELECT * FROM users WHERE id = %s", [user_id])
```

## Constraints

### MUST DO
- Summarize PR intent before reviewing (see Workflow step 1)
- Provide specific, actionable feedback
- Include code examples in suggestions
- Praise good patterns
- Prioritize feedback (critical → minor)
- Review tests as thoroughly as code
- Check for security issues (OWASP Top 10 as baseline)

### MUST NOT DO
- Be condescending or rude
- Nitpick style when linters exist
- Block on personal preferences
- Demand perfection
- Review without understanding the why
- Skip praising good work

## Output Template

Code review report must include:
1. **Summary** — One-sentence intent recap + overall assessment
2. **Critical issues** — Must fix before merge (bugs, security, data loss)
3. **Major issues** — Should fix (performance, design, maintainability)
4. **Minor issues** — Nice to have (naming, readability)
5. **Positive feedback** — Specific patterns done well
6. **Questions for author** — Clarifications needed
7. **Verdict** — Approve / Request Changes / Comment

## Knowledge Reference

SOLID, DRY, KISS, YAGNI, design patterns, OWASP Top 10, language idioms, testing patterns

Maintained by [@jeffallan](https://github.com/jeffallan), Principal Consultant at [Synergetic Solutions](https://synergetic.solutions)

[Documentation](https://jeffallan.github.io/claude-skills/skills/quality/code-reviewer/)

---

## Reference: common-issues

# Common Issues

## N+1 Query Problem

```typescript
// N+1 queries - BAD
const posts = await Post.findAll();
for (const post of posts) {
  post.author = await User.findById(post.authorId); // N queries!
}

// Single query with join - GOOD
const posts = await Post.findAll({ include: [User] });

// Or batch load
const posts = await Post.findAll();
const authorIds = posts.map(p => p.authorId);
const authors = await User.findByIds(authorIds);
```

## Missing Error Handling

```typescript
// Unhandled rejection - BAD
const data = await fetch('/api/data').then(r => r.json());

// Proper error handling - GOOD
try {
  const response = await fetch('/api/data');
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`);
  }
  const data = await response.json();
} catch (error) {
  logger.error('Failed to fetch data', { error });
  throw new DataFetchError('Could not load data');
}
```

## Magic Numbers/Strings

```typescript
// Magic number - BAD
if (user.age >= 18) { ... }
setTimeout(fn, 86400000);

// Named constant - GOOD
const MINIMUM_AGE = 18;
const ONE_DAY_MS = 24 * 60 * 60 * 1000;

if (user.age >= MINIMUM_AGE) { ... }
setTimeout(fn, ONE_DAY_MS);
```

## Deep Nesting

```typescript
// Deep nesting - BAD
if (user) {
  if (user.isActive) {
    if (user.hasPermission) {
      doSomething();
    }
  }
}

// Early returns - GOOD
if (!user || !user.isActive || !user.hasPermission) {
  return;
}
doSomething();
```

## God Functions

```typescript
// Does too much - BAD
async function processOrder(order) {
  // validate
  // check inventory
  // process payment
  // send email
  // update database
  // log analytics
}

// Single responsibility - GOOD
async function processOrder(order) {
  await validateOrder(order);
  await reserveInventory(order);
  await chargePayment(order);
  await sendConfirmation(order);
}
```

## Mutable Shared State

```typescript
// Shared mutable - BAD
const config = { debug: false };
function enableDebug() {
  config.debug = true;
}

// Immutable pattern - GOOD
function createConfig(overrides = {}) {
  return Object.freeze({ debug: false, ...overrides });
}
```

## Missing Null Checks

```typescript
// Unsafe access - BAD
const name = user.profile.name;

// Safe access - GOOD
const name = user?.profile?.name ?? 'Unknown';
```

## Synchronous File Operations

```typescript
// Blocks event loop - BAD
const data = fs.readFileSync('file.txt');

// Non-blocking - GOOD
const data = await fs.promises.readFile('file.txt');
```

## Quick Reference

| Issue | Impact | Fix |
|-------|--------|-----|
| N+1 queries | Performance | Eager load or batch |
| Missing error handling | Reliability | Try/catch + logging |
| Magic numbers | Maintainability | Named constants |
| Deep nesting | Readability | Early returns |
| God functions | Testability | Single responsibility |
| Mutable shared state | Bugs | Immutable patterns |
| Missing null checks | Crashes | Optional chaining |
| Sync file operations | Performance | Async operations |

---

## Reference: feedback-examples

# Feedback Examples

## Good vs Bad Feedback

### Be Specific, Not Vague

```markdown
BAD: "This is confusing"

GOOD: "This function handles both validation and persistence. Consider
      splitting into `validateUser()` and `saveUser()` for single
      responsibility and easier testing."
```

### Be Actionable, Not Just Critical

```markdown
BAD: "Fix the query"

GOOD: "This will cause N+1 queries - one per post. Use `include: [Author]`
      to eager load authors in a single query. See: [link to docs]"
```

### Be Constructive, Not Demanding

```markdown
BAD: "Add tests"

GOOD: "Missing test for the case when `email` is already taken. Add a test
      that verifies 409 is returned with appropriate error message."
```

### Ask Questions, Don't Assume

```markdown
BAD: "This is wrong"

GOOD: "I notice this returns null instead of throwing. Is that intentional?
      The other methods throw on not-found. Should this be consistent?"
```

## Praise Examples

Reinforce good patterns with specific praise:

```markdown
"Great use of early returns here - much more readable than nested ifs!"

"Nice extraction of this validation logic into a reusable function."

"Excellent error messages - they'll help debugging in production."

"Good choice using a discriminated union here instead of optional fields."

"Appreciate the comprehensive test coverage, especially the edge cases."
```

## Feedback by Category

### Critical (Must Fix)

```markdown
**[CRITICAL] Security: SQL Injection**
Location: `src/users/service.ts:45`

The query uses string interpolation:
`SELECT * FROM users WHERE id = ${id}`

This is vulnerable to SQL injection. Use parameterized query:
`db.query('SELECT * FROM users WHERE id = $1', [id])`
```

### Major (Should Fix)

```markdown
**[MAJOR] Performance: N+1 Query**
Location: `src/posts/service.ts:23`

Current code fetches users in a loop (N+1 problem):
```typescript
for (const post of posts) {
  post.author = await User.findById(post.authorId);
}
```

Suggestion: Use eager loading:
```typescript
const posts = await Post.findAll({ include: [User] });
```

Impact: ~100 extra DB queries per request with current approach.
```

### Minor (Nice to Have)

```markdown
**[MINOR] Naming: Unclear variable**
Location: `src/utils/date.ts:12`

`d` is unclear. Consider `createdDate` or `timestamp` for better readability.

**[MINOR] Style: Prefer const**
Location: `src/config/index.ts:8`

`let config` is never reassigned. Use `const` for immutability.
```

## Question Format

```markdown
**[QUESTION]**
Location: `src/orders/service.ts:67`

What's the expected behavior when the user has an existing pending order?
Should this:
- Return the existing order?
- Create a new one anyway?
- Return an error?
```

## Summary Format

```markdown
## Summary

Overall this is a solid implementation of the user registration flow.
The validation logic is clean and the error handling is comprehensive.

**Blocking Issues**: 1 critical (SQL injection)
**Suggestions**: 2 major, 3 minor

Once the SQL injection is fixed, this is ready to merge. The major
suggestions are performance improvements worth considering.
```

## Quick Reference

| Feedback Type | Tone | Required Action |
|---------------|------|-----------------|
| Critical | Firm, clear | Must fix before merge |
| Major | Suggestive | Should fix |
| Minor | Optional | Nice to have |
| Praise | Positive | None - reinforcement |
| Question | Curious | Response needed |

---

## Reference: receiving-feedback

# Receiving Feedback

---

## Core Mindset

> "Verify before implementing. Ask before assuming. Technical correctness over social comfort."

Code review feedback is a technical discussion, not a social one. Focus on the code, not on feelings.

---

## The Six-Step Process

### Step 1: Read Completely

**Without reacting.** Read the entire comment before forming any response.

```markdown
❌ BAD: Read first sentence → start typing defense
✅ GOOD: Read entire comment → understand full context → then respond
```

### Step 2: Restate Requirements

Rephrase the reviewer's feedback in your own words to confirm understanding.

```markdown
Reviewer: "This function is doing too much. It handles validation,
transformation, and persistence all in one place."

Your restatement: "You're suggesting I split this into three separate
functions: validate(), transform(), and persist()?"
```

### Step 3: Check Against Codebase

Verify the feedback against actual code conditions before responding.

```typescript
// Reviewer says: "This will throw if user is null"

// Check the code:
function getUsername(user: User): string {
  return user.name;  // No null check - reviewer is correct
}

// Or discover context:
function getUsername(user: User): string {
  return user.name;  // TypeScript enforces User, null not possible
}
```

### Step 4: Evaluate Technical Soundness

Consider whether the feedback applies to your specific stack and context.

```markdown
Reviewer: "You should use useMemo here for performance"

Evaluate:
- Is this component re-rendering frequently? → Check React DevTools
- Is the computation expensive? → Profile it
- Does React 19's compiler auto-optimize this? → Check version
```

### Step 5: Respond with Substance

Provide technical acknowledgment or reasoned objection.

```markdown
✅ GOOD: "Fixed. Split into validate(), transform(), persist()
         at lines 24, 45, 67."

✅ GOOD: "Respectfully disagree. This list has max 5 items
         (see schema.ts:12), so filter performance is O(5)."

❌ BAD: "You're absolutely right! Great catch!"
❌ BAD: "I don't think that's necessary."
```

### Step 6: Implement One at a Time

Address each piece of feedback individually with verification.

```markdown
Feedback item 1: Add null check
→ Implement → Test → Commit → Verify → Move to next

Feedback item 2: Extract helper function
→ Implement → Test → Commit → Verify → Move to next

NOT: Try to address all feedback in one massive commit
```

---

## Avoiding Agreement Theater

### The Problem

Performative agreement wastes time and provides no information. When you write "Great point!" you're adding noise, not signal.

### Forbidden Phrases

| Phrase | Why It's Wrong |
|--------|----------------|
| "You're absolutely right!" | Sycophantic, adds no information |
| "Great point!" | Empty praise, not a response |
| "Excellent feedback!" | Flattery, not engagement |
| "Thanks for catching this!" | Unnecessary, just fix it |
| "I really appreciate..." | Social fluff, not technical |

### Actions Demonstrate Understanding

```markdown
❌ "You're absolutely right! Great catch on that null check!
    Thanks so much for pointing this out!"

✅ "Fixed. Added null check at line 42."
```

The code change shows you understood. Words are redundant.

### When Acknowledgment IS Appropriate

Brief, technical acknowledgment when learning something new:

```markdown
✅ "I wasn't aware of that edge case. Added handling at line 42."
✅ "Good point about thread safety. Added mutex at line 67."
```

---

## When to Push Back

### Valid Reasons to Disagree

Push back with technical reasoning when feedback:

| Situation | How to Respond |
|-----------|----------------|
| Breaks existing functionality | "This change would break Feature X (see test at tests/feature-x.spec.ts:34)" |
| Lacks full codebase context | "This pattern exists because of Y (see architecture.md#constraints)" |
| Violates YAGNI | "This flexibility isn't needed yet - only one caller exists" |
| Is technically incorrect | "This actually works because of Z (link to docs)" |
| Conflicts with established architecture | "This conflicts with our JWT approach (see auth/README.md)" |

### Good Pushback Format

```markdown
## Template
This conflicts with [X]. [Evidence]. Was that the intent, or should we [alternative]?

## Example
This conflicts with our JWT authentication architecture (see auth/token.js:45).
Switching to sessions would require restructuring the API middleware.
Was that the intent, or should we keep JWT?
```

### Bad Pushback

```markdown
❌ "I don't think that's right."
❌ "That won't work."
❌ "We've always done it this way."
❌ "That's too much work."
```

---

## Verification Before Claiming Fixed

### The Checklist

Before writing "Fixed" or "Done":

- [ ] Change is implemented
- [ ] Tests pass (full suite, not just changed files)
- [ ] Specific behavior mentioned in feedback is verified
- [ ] Edge cases are tested
- [ ] No unintended side effects introduced

### Acceptable Responses

```markdown
✅ "Fixed. Added null check. Tests pass."
✅ "Fixed at line 42. Verified with test case X."
✅ "Implemented. All 47 tests pass."
```

### Unacceptable Responses

```markdown
❌ "I think this addresses your concern."
❌ "Should be fixed now."
❌ "Done, I believe."
❌ "Fixed (probably)."
```

### When You Can't Verify

If you cannot verify a fix:

```markdown
✅ "Implemented the change, but I'm unable to verify because
    [specific reason]. Can you confirm on your end?"
```

---

## Quick Reference

| Situation | Response |
|-----------|----------|
| Reviewer is correct | "Fixed. [What you changed]." |
| You need clarification | "To confirm: you're suggesting [restatement]?" |
| Reviewer is incorrect | "This works because [evidence]. [Link to proof]." |
| You disagree on approach | "This conflicts with [X]. Should we [alternative]?" |
| You learned something | "I wasn't aware of [X]. Fixed at line [N]." |
| You can't verify | "Implemented. Unable to verify because [reason]." |

---

## Anti-Patterns

| Pattern | Problem | Fix |
|---------|---------|-----|
| Defensive responses | Creates conflict, wastes time | Assume good faith, respond technically |
| Apologetic responses | Unprofessional, adds noise | Just fix it |
| Delayed responses | Blocks review cycle | Respond within hours, not days |
| Vague responses | Leaves reviewer uncertain | Be specific about changes |
| Ignoring feedback | Disrespectful, creates friction | Address every point |

---

*Content adapted from [obra/superpowers](https://github.com/obra/superpowers) by Jesse Vincent (@obra), MIT License.*

---

## Reference: report-template

# Report Template

## Full Review Report Template

```markdown
# Code Review: [PR Title]

## Summary
[1-2 sentence overview of the changes and overall assessment]

**Verdict**: [ ] Approve | [x] Request Changes | [ ] Comment

## Critical Issues (Must Fix)

### 1. [File:Line] Security: SQL Injection Risk
- **Current**: String interpolation in query
- **Suggested**: Use parameterized query
- **Impact**: Potential data breach

```typescript
// Current (vulnerable)
const query = `SELECT * FROM users WHERE id = ${id}`;

// Suggested (secure)
const query = 'SELECT * FROM users WHERE id = $1';
db.query(query, [id]);
```

## Major Issues (Should Fix)

### 1. [File:Line] Performance: N+1 Query
- **Current**: Fetching users in loop
- **Suggested**: Use eager loading with include
- **Impact**: ~100 extra DB queries per request

### 2. [File:Line] Logic: Missing edge case
- **Current**: No handling for empty array
- **Suggested**: Add guard clause
- **Impact**: Potential runtime error

## Minor Issues (Nice to Have)

### 1. [File:Line] Naming: Unclear variable name
- **Current**: `d`
- **Suggested**: `createdDate`

### 2. [File:Line] Style: Inconsistent formatting
- **Current**: Mixed quotes
- **Suggested**: Use single quotes consistently

## Positive Feedback
- Clean separation of concerns in service layer
- Comprehensive input validation on DTOs
- Good test coverage for edge cases
- Excellent error messages

## Questions for Author
- What's the expected behavior when X happens?
- Should this support pagination for large datasets?
- Is the retry logic intentional or accidental?

## Test Coverage Assessment
- [ ] Happy path tested
- [x] Error cases tested
- [ ] Edge cases tested (missing empty array test)
- [x] Integration tests present

## Checklist
- [x] No security vulnerabilities
- [ ] Performance is acceptable (N+1 issue)
- [x] Code is readable
- [x] Tests are adequate
- [x] Documentation is present
```

## Verdict Guidelines

| Verdict | When to Use |
|---------|-------------|
| **Approve** | No blocking issues, minor suggestions only |
| **Request Changes** | Critical or major issues must be fixed |
| **Comment** | Questions need answers, no blocking issues |

## Severity Definitions

| Severity | Definition | Examples |
|----------|------------|----------|
| **Critical** | Security risk, data loss, crashes | SQL injection, auth bypass |
| **Major** | Significant performance, maintainability | N+1 queries, god functions |
| **Minor** | Style, naming, small improvements | Variable names, formatting |

## Time Boxing

| Section | Suggested Time |
|---------|----------------|
| Context & understanding | 5 minutes |
| Critical/security review | 10 minutes |
| Logic & performance | 15 minutes |
| Tests review | 10 minutes |
| Writing report | 10 minutes |
| **Total** | ~50 minutes |

## Quick Checks Before Submitting

- [ ] All critical issues have clear remediation
- [ ] Major issues explain the impact
- [ ] At least one positive comment included
- [ ] Questions are specific and answerable
- [ ] Verdict matches the issues found

---

## Reference: review-checklist

# Review Checklist

## Comprehensive Review Checklist

| Category | Key Questions |
|----------|---------------|
| **Design** | Does it fit existing patterns? Right abstraction level? |
| **Logic** | Edge cases handled? Race conditions? Null checks? |
| **Security** | Input validated? Auth checked? Secrets safe? |
| **Performance** | N+1 queries? Memory leaks? Caching needed? |
| **Tests** | Adequate coverage? Edge cases tested? Mocks appropriate? |
| **Naming** | Clear, consistent, intention-revealing? |
| **Error Handling** | Errors caught? Meaningful messages? Logged? |
| **Documentation** | Public APIs documented? Complex logic explained? |

## Review Process

### 1. Context (5 min)
- [ ] Read PR description
- [ ] Understand the problem being solved
- [ ] Check linked issues/tickets
- [ ] Note expected changes

### 2. Structure (10 min)
- [ ] Review file organization
- [ ] Check architectural fit
- [ ] Verify design patterns used
- [ ] Note any breaking changes

### 3. Code Details (20 min)
- [ ] Review logic correctness
- [ ] Check edge cases
- [ ] Verify error handling
- [ ] Look for security issues
- [ ] Check performance concerns
- [ ] Review naming clarity

### 4. Tests (10 min)
- [ ] Verify test coverage
- [ ] Check test quality
- [ ] Look for edge case tests
- [ ] Ensure mocks are appropriate

### 5. Final Pass (5 min)
- [ ] Note positive patterns
- [ ] Prioritize feedback
- [ ] Write summary

## Category Deep Dive

### Design Questions
- Does this change belong in this file/module?
- Is the abstraction level appropriate?
- Could this be simpler?
- Does it follow existing patterns?
- Is it extensible without modification?

### Logic Questions
- What happens with null/undefined inputs?
- Are boundary conditions handled?
- Could there be race conditions?
- Is the order of operations correct?
- Are all code paths tested?

### Security Questions
- Is all user input validated?
- Are SQL queries parameterized?
- Is output properly encoded?
- Are secrets handled safely?
- Is authentication checked?
- Is authorization enforced?

### Performance Questions
- Are there N+1 query patterns?
- Is data fetched efficiently?
- Are expensive operations cached?
- Could this cause memory leaks?
- Is pagination implemented?

## Quick Reference

| Review Focus | Time % |
|--------------|--------|
| Context & PR description | 10% |
| Architecture & design | 20% |
| Code logic & details | 40% |
| Tests & coverage | 20% |
| Final review & summary | 10% |

---

## Reference: spec-compliance-review

# Spec Compliance Review

---

## Two-Stage Review Architecture

```
                    ┌─────────────────────┐
                    │   Implementation    │
                    └──────────┬──────────┘
                               │
                    ┌──────────▼──────────┐
                    │  STAGE 1: Spec      │
                    │  Compliance Review  │
                    └──────────┬──────────┘
                               │
              ┌────────────────┴────────────────┐
              │                                  │
      ┌───────▼───────┐                ┌────────▼────────┐
      │   ✗ Issues    │                │   ✓ Compliant   │
      │     Found     │                │                 │
      └───────┬───────┘                └────────┬────────┘
              │                                  │
              │                        ┌────────▼────────┐
              │                        │  STAGE 2: Code  │
              │                        │  Quality Review │
              │                        └────────┬────────┘
              │                                  │
              │                    ┌─────────────┴─────────────┐
              │                    │                           │
              │            ┌───────▼───────┐         ┌────────▼────────┐
              │            │   ✗ Issues    │         │   ✓ Approved    │
              │            │     Found     │         │                 │
              │            └───────┬───────┘         └─────────────────┘
              │                    │
              └────────────────────┴────────────────────┐
                                                        │
                                              ┌─────────▼─────────┐
                                              │ Return to Author  │
                                              └───────────────────┘
```

**Critical:** Complete Stage 1 (spec compliance) BEFORE Stage 2 (code quality). Never review code quality for functionality that doesn't meet the specification.

---

## Stage 1: Spec Compliance Review

### Core Directive

> "The implementer finished suspiciously quickly. Their report may be incomplete, inaccurate, or optimistic."

Approach every review with professional skepticism. Verify claims independently.

### The Three Verification Categories

#### Category 1: Missing Requirements

**Check for features that were requested but not implemented.**

| Question | How to Verify |
|----------|---------------|
| Did they skip requested features? | Compare PR to original requirements line by line |
| Are edge cases handled? | Check error paths, empty states, boundaries |
| Were error scenarios addressed? | Look for try/catch, error boundaries, validation |
| Is the happy path complete? | Trace through primary use case manually |

```markdown
## Example Review Finding

**Missing Requirement:** Issue #42 requested "password must be at least 8 characters"

**Found in code:**
```typescript
// No length validation present
function validatePassword(password: string) {
  return password.length > 0;  // Only checks non-empty
}
```

**Status:** ❌ Incomplete - minimum length validation missing
```

#### Category 2: Unnecessary Additions

**Check for scope creep and over-engineering.**

| Question | How to Verify |
|----------|---------------|
| Features beyond specification? | Compare to original requirements |
| Over-engineering? | Is complexity justified by requirements? |
| Premature optimization? | Is performance cited without measurements? |
| Unrequested abstractions? | Are there helpers/utils for one-time use? |

```markdown
## Example Review Finding

**Unnecessary Addition:** Added caching layer not in requirements

**Found in code:**
```typescript
// Original requirement: "Fetch user by ID"
// Actual implementation:
class CachedUserRepository {  // Not requested
  private cache = new Map();
  private ttl = 60000;

  async getUser(id: string) {
    if (this.cache.has(id)) { ... }
    // 50 lines of cache logic
  }
}
```

**Status:** ⚠️ Scope creep - discuss before merging
```

#### Category 3: Interpretation Gaps

**Check for misunderstandings of requirements.**

| Question | How to Verify |
|----------|---------------|
| Different understanding of requirements? | Ask author to explain their interpretation |
| Unclarified assumptions? | Look for comments like "assuming..." |
| Ambiguous specs resolved incorrectly? | Compare to similar existing features |

```markdown
## Example Review Finding

**Interpretation Gap:** "Sort by date" implemented as ascending

**Requirement stated:** "Sort by date" (ambiguous)

**Author implemented:** Oldest first (ascending)

**Expected:** Most recent first is typical UX pattern

**Status:** ❓ Clarify - which sort order was intended?
```

---

## Why Order Matters

### Stage 1 Must Come First

| Scenario | Waste from Wrong Order |
|----------|------------------------|
| Skip Stage 1 | Review 500 lines of code quality, then discover wrong feature was built |
| Stage 2 First | Suggest refactoring, then realize the code shouldn't exist |
| Combined | Mix concerns, miss systematic issues |

### Separation of Concerns

- **Stage 1 (Spec):** Does it do the right thing?
- **Stage 2 (Quality):** Does it do the thing right?

Code quality review is meaningless if the code doesn't implement the correct functionality.

---

## Spec Compliance Checklist

### Before You Start

- [ ] Read the original issue/ticket completely
- [ ] Identify all explicit requirements
- [ ] Identify implicit requirements from context
- [ ] Note any acceptance criteria listed

### During Review

**Missing Requirements:**
- [ ] All required features present
- [ ] Edge cases covered (empty, null, max values)
- [ ] Error handling as specified
- [ ] Happy path fully functional
- [ ] UI matches mockups/specs if provided

**Unnecessary Additions:**
- [ ] No unrequested features
- [ ] No speculative abstractions
- [ ] No premature optimizations
- [ ] Scope matches requirements exactly

**Interpretation Gaps:**
- [ ] Author's understanding matches spec
- [ ] Ambiguities resolved correctly
- [ ] Assumptions are documented and valid
- [ ] Behavior matches similar existing features

### After Review

- [ ] Document all findings with file:line references
- [ ] Categorize as missing/unnecessary/interpretation
- [ ] Prioritize: blocking vs. non-blocking issues

---

## Output Format

### Compliant Result

```markdown
## Spec Compliance Review: ✅ PASS

All requirements verified:
- ✅ User can upload profile image (req #1)
- ✅ Image resized to 200x200 (req #2)
- ✅ Invalid formats rejected with error message (req #3)
- ✅ Progress indicator during upload (req #4)

**Proceed to:** Code Quality Review
```

### Issues Found

```markdown
## Spec Compliance Review: ❌ ISSUES FOUND

### Missing Requirements

1. **Progress indicator not implemented** (req #4)
   - File: `ProfileUpload.tsx`
   - Expected: Progress bar during upload
   - Found: No progress indication

2. **Error messages not user-friendly** (req #3)
   - File: `ProfileUpload.tsx:45`
   - Expected: "Please upload a JPG or PNG file"
   - Found: "Error: INVALID_FORMAT"

### Unnecessary Additions

1. **Image cropping feature not requested**
   - File: `ImageCropper.tsx` (new file, 150 lines)
   - Impact: Adds complexity, delays delivery
   - Recommendation: Remove or create separate PR

**Action Required:** Address missing requirements before code quality review
```

---

## Common Mistakes to Avoid

| Mistake | Why It's Wrong |
|---------|----------------|
| Reviewing code style before spec compliance | Wasted effort if wrong thing was built |
| Assuming spec was followed | Verify independently |
| Skipping edge cases | Bugs hide in boundaries |
| Accepting "we can add it later" | Technical debt accumulates |
| Missing scope creep | Unreviewed code enters codebase |

---

*Content adapted from [obra/superpowers](https://github.com/obra/superpowers) by Jesse Vincent (@obra), MIT License.*
