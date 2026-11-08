# feature-forge

> Источник: https://github.com/jeffallan/claude-skills/tree/main/skills/feature-forge (MIT). Рабочая копия скилла: `.claude/skills/feature-forge/`.

---
name: feature-forge
description: Conducts structured requirements workshops to produce feature specifications, user stories, EARS-format functional requirements, acceptance criteria, and implementation checklists. Use when defining new features, gathering requirements, or writing specifications. Invoke for feature definition, requirements gathering, user stories, EARS format specs, PRDs, acceptance criteria, or requirement matrices.
license: MIT
metadata:
  author: https://github.com/Jeffallan
  company: https://synergetic.solutions
  version: "1.1.0"
  domain: workflow
  triggers: requirements, specification, feature definition, user stories, EARS, planning
  role: specialist
  scope: design
  output-format: document
  related-skills: fullstack-guardian, spec-miner, test-master
---

# Feature Forge

Requirements specialist conducting structured workshops to define comprehensive feature specifications.

## Role Definition

Operate with two perspectives:
- **PM Hat**: Focused on user value, business goals, success metrics
- **Dev Hat**: Focused on technical feasibility, security, performance, edge cases

## When to Use This Skill

- Defining new features from scratch
- Gathering comprehensive requirements
- Writing specifications in EARS format
- Creating acceptance criteria
- Planning implementation TODO lists

## Core Workflow

1. **Discover** - Use `AskUserQuestions` to understand the feature goal, target users, and user value. Present structured choices where possible (e.g., user types, priority level).
2. **Interview** - Systematic questioning from both PM and Dev perspectives using `AskUserQuestions` for structured choices and open-ended follow-ups. Use multi-agent discovery with Task subagents when the feature spans multiple domains (see interview-questions.md for guidance).
3. **Document** - Write EARS-format requirements
4. **Validate** - Use `AskUserQuestions` to review acceptance criteria with stakeholder, presenting key trade-offs as structured choices
5. **Plan** - Create implementation checklist

## Reference Guide

Load detailed guidance based on context:

| Topic | Reference | Load When |
|-------|-----------|-----------|
| EARS Syntax | `references/ears-syntax.md` | Writing functional requirements |
| Interview Questions | `references/interview-questions.md` | Gathering requirements |
| Specification Template | `references/specification-template.md` | Writing final spec document |
| Acceptance Criteria | `references/acceptance-criteria.md` | Given/When/Then format |
| Pre-Discovery Subagents | `references/pre-discovery-subagents.md` | Multi-domain features needing front-loaded context |

## Constraints

### MUST DO
- Use `AskUserQuestions` tool for structured elicitation (priority, scope, format choices)
- Use open-ended questions only when choices cannot be predetermined
- Conduct thorough interview before writing spec
- Use EARS format for all functional requirements
- Include non-functional requirements (performance, security)
- Provide testable acceptance criteria
- Include implementation TODO checklist
- Ask for clarification on ambiguous requirements

### MUST NOT DO
- Output interview questions as plain text when `AskUserQuestions` can provide structured options
- Generate spec without conducting interview
- Accept vague requirements ("make it fast")
- Skip security considerations
- Forget error handling requirements
- Write untestable acceptance criteria

## Output Templates

The final specification must include:
1. Overview and user value
2. Functional requirements (EARS format)
3. Non-functional requirements
4. Acceptance criteria (Given/When/Then)
5. Error handling table
6. Implementation TODO checklist

**Inline EARS format examples** (load `references/ears-syntax.md` for full syntax):
```
When <trigger>, the <system> shall <response>.
Where <feature> is active, the <system> shall <behaviour>.
The <system> shall <action> within <measure>.
```

**Inline acceptance criteria example** (load `references/acceptance-criteria.md` for full format):
```
Given a registered user is on the login page,
When they submit valid credentials,
Then they are redirected to the dashboard within 2 seconds.
```

Save as: `specs/{feature_name}.spec.md`

Maintained by [@jeffallan](https://github.com/jeffallan), Principal Consultant at [Synergetic Solutions](https://synergetic.solutions)

[Documentation](https://jeffallan.github.io/claude-skills/skills/workflow/feature-forge/)

---

## Reference: acceptance-criteria

# Acceptance Criteria

## Given-When-Then Format

```markdown
### AC-001: [Scenario Name]
Given [context/precondition]
When [action taken]
Then [expected result]
```

## Examples by Type

### Happy Path

```markdown
### AC-001: Successful Login
Given a registered user with valid credentials
When they submit the login form
Then they are redirected to the dashboard
And a success message is displayed
And their session is created

### AC-002: Add Item to Cart
Given a logged-in user viewing a product
When they click "Add to Cart"
Then the item appears in their cart
And the cart badge updates with the count
And a confirmation toast is shown
```

### Error Cases

```markdown
### AC-003: Invalid Login
Given a user with incorrect password
When they submit the login form
Then an error message "Invalid credentials" is displayed
And the password field is cleared
And they remain on the login page

### AC-004: Duplicate Email Registration
Given an email already exists in the system
When a new user tries to register with that email
Then an error message "Email already registered" is displayed
And the form is not submitted
```

### Edge Cases

```markdown
### AC-005: Empty Cart Checkout
Given a user with an empty cart
When they navigate to checkout
Then they see "Your cart is empty" message
And a "Continue Shopping" button is displayed

### AC-006: Session Expiry
Given a user whose session has expired
When they try to perform any authenticated action
Then they are redirected to login
And a message "Session expired, please log in again" is shown
And their intended action is preserved for after login
```

### Authorization

```markdown
### AC-007: Admin-Only Access
Given a regular user (non-admin)
When they try to access /admin/users
Then they receive a 403 Forbidden response
And are redirected to the home page
And an "Access denied" message is shown

### AC-008: Own Resource Only
Given a user viewing another user's profile
When they try to edit the profile
Then the edit button is not visible
And direct URL access returns 403
```

## INVEST Criteria

Good acceptance criteria follow INVEST:

| Criterion | Description | Check |
|-----------|-------------|-------|
| **I**ndependent | Can be tested alone | No dependencies on other ACs |
| **N**egotiable | Details can be discussed | Not over-specified |
| **V**aluable | Delivers user value | Ties to requirement |
| **E**stimable | Effort can be estimated | Clear scope |
| **S**mall | Testable in one session | Not too broad |
| **T**estable | Pass/fail is clear | Objective criteria |

## Quick Reference

| Scenario Type | Given | When | Then |
|---------------|-------|------|------|
| Happy path | Valid state | Valid action | Success result |
| Error | Invalid state/input | Action | Error message |
| Edge case | Boundary condition | Action | Graceful handling |
| Authorization | User role | Protected action | Appropriate access |
| Concurrency | Multiple actors | Simultaneous action | Consistent state |

---

## Reference: ears-syntax

# EARS Syntax

## EARS Format

Easy Approach to Requirements Syntax for clear, unambiguous requirements.

### Basic Pattern

```
While <precondition>, when <trigger>, the system shall <response>.
```

### Pattern Types

**Ubiquitous (Always True)**
```
The system shall [action].
```
Example: The system shall encrypt all passwords using bcrypt.

**Event-Driven**
```
When [trigger], the system shall [action].
```
Example: When the user clicks "Submit", the system shall save the form data.

**State-Driven**
```
While [state], the system shall [action].
```
Example: While the user is logged in, the system shall display the dashboard.

**Conditional (Most Common)**
```
While [state], when [trigger], the system shall [action].
```
Example: While the cart contains items, when the user clicks "Checkout", the system shall navigate to the payment page.

**Optional**
```
Where [feature enabled], the system shall [action].
```
Example: Where two-factor authentication is enabled, the system shall require a verification code.

## Examples by Domain

### Authentication

```markdown
**FR-AUTH-001**: Login
While credentials are valid, when POST /auth/login is called,
the system shall return JWT access token (15min) and refresh token (7d).

**FR-AUTH-002**: Invalid Login
When invalid credentials are provided,
the system shall return 401 and increment failed login counter.

**FR-AUTH-003**: Account Lockout
While failed login count exceeds 5, when login is attempted,
the system shall reject the attempt and require password reset.
```

### E-commerce

```markdown
**FR-CART-001**: Add to Cart
While user is logged in, when they click "Add to Cart",
the system shall add the item and update the cart badge count.

**FR-CART-002**: Apply Coupon
While the cart contains items, when a valid coupon code is applied,
the system shall reduce the total by the discount amount.

**FR-ORDER-001**: Checkout
While payment method is valid, when user confirms order,
the system shall create order, charge payment, and send confirmation email.
```

### Data Management

```markdown
**FR-EXPORT-001**: CSV Export
While user has data access permission, when they click "Export",
the system shall generate a CSV file and initiate download.

**FR-DELETE-001**: Soft Delete
When a resource is deleted,
the system shall set deleted_at timestamp instead of removing the record.
```

## Quick Reference

| Type | Structure | Use When |
|------|-----------|----------|
| Ubiquitous | shall [action] | Always applies |
| Event | When [X], shall | On trigger |
| State | While [X], shall | Continuous state |
| Conditional | While [X], when [Y], shall | State + trigger |
| Optional | Where [X], shall | Feature flag |

---

## Reference: interview-questions

# Interview Questions

## PM Hat Questions

Focus on user value and business goals.

| Area | Questions |
|------|-----------|
| **Problem** | What problem does this solve? Who experiences this problem? How often? |
| **Users** | Who are the target users? What are their goals? Technical level? |
| **Value** | How will users benefit? What's the business value? ROI? |
| **Scope** | What's in scope? What's explicitly out of scope? MVP vs full version? |
| **Success** | How will we measure success? Key metrics? |
| **Priority** | Is this a must-have, should-have, or nice-to-have? |

### Example PM Questions

```markdown
For a "User Export" feature:
- Who needs to export data and why?
- What format do they need (CSV, JSON, Excel)?
- How much data? 100 rows or 1 million?
- Is this for compliance (GDPR) or convenience?
- How often will this be used?
- What's the deadline?
```

## Dev Hat Questions

Focus on technical feasibility and edge cases.

| Area | Questions |
|------|-----------|
| **Integration** | What systems does this touch? APIs, databases, services? |
| **Security** | Authentication required? Data sensitivity (PII, PCI)? |
| **Performance** | Expected load? Response time requirements? Async OK? |
| **Edge Cases** | What happens when X fails? Empty states? Limits? |
| **Data** | What's stored? Retention period? Backup needs? |
| **Dependencies** | External services? Rate limits? Costs? |

### Example Dev Questions

```markdown
For a "User Export" feature:
- What fields to include? Are any sensitive (passwords, tokens)?
- Max export size? Need streaming or background job?
- Should include soft-deleted records?
- What happens if export fails midway?
- File retention - how long to keep generated files?
- Need progress indicator for large exports?
```

## Tool Usage: AskUserQuestions

Use `AskUserQuestions` when questions have a finite set of likely answers. Use open-ended follow-up when answers are unbounded.

### When to Use Structured Options

| Question Pattern | Example | Options Style |
|-----------------|---------|---------------|
| Priority/ranking | "Is this must-have or nice-to-have?" | Single select: Must-have, Should-have, Nice-to-have |
| Format selection | "What export format?" | Multi-select: CSV, JSON, Excel, PDF |
| Scope decisions | "MVP or full version?" | Single select: MVP, Full, Phased |
| Yes/No with nuance | "Auth required?" | Single select: Public, Authenticated, Role-based |

### When to Use Open-Ended

- "Describe the user journey in your own words"
- "What problem does this solve?"
- "Walk me through the workflow"

### Example: Structured Elicitation

For a "User Export" feature, batch related choices:

**Question 1** (header: "Export scope"):
"What data should users be able to export?"
Options: "Own data only", "Team data", "Organization-wide", multi-select enabled

**Question 2** (header: "Format"):
"Which export formats should be supported?"
Options: "CSV", "JSON", "Excel (.xlsx)", "PDF", multi-select enabled

**Question 3** (header: "Priority"):
"How critical is this feature?"
Options: "Must-have (blocking)", "Should-have (important)", "Nice-to-have (future)"

---

## Interview Flow

### Phase 1: Discovery
Use open-ended questions to understand the problem space:
1. "Tell me about this feature in your own words"
2. "What problem are we solving?"

Then use `AskUserQuestions` to narrow down:
- Target users (single select from identified personas)
- Usage frequency (Daily, Weekly, Monthly, Rarely)
- Priority (Must-have, Should-have, Nice-to-have)

### Phase 2: Details
Use `AskUserQuestions` for scope and constraint decisions:
- Scope: MVP vs Full vs Phased (single select)
- Key capabilities (multi-select from discovered items)

Then open-ended: "Walk me through the user journey"

### Phase 3: Edge Cases
Use `AskUserQuestions` for technical trade-offs:
- Error handling approach (Retry, Fail fast, Queue, Notify)
- Data limits (multi-select thresholds)

Then open-ended: "What happens when [X] fails?"

### Phase 4: Validation
Present spec summary, then use `AskUserQuestions`:
- "Does this capture your requirements?" (Yes / Needs changes / Major gaps)
- Per-requirement priority confirmation if needed

## Multi-Agent Pre-Discovery

For features spanning multiple domains, launch Task subagents with relevant skills **before** starting the interview. This front-loads technical context so the interview focuses on decisions rather than exploration.

### Pattern: Parallel Skill-Invoked Discovery

```
User request: "I need a feature that does X"

Before interview, launch subagents in parallel:
- Task(subagent_type="general-purpose"): Invoke architecture-designer skill to assess system impact
- Task(subagent_type="general-purpose"): Invoke security-reviewer skill to identify auth/data concerns
- Task(subagent_type="Explore"): Search codebase for existing patterns related to the feature

Collect subagent findings → Use them to inform interview questions
```

This ensures the Feature Forge interview starts with concrete technical context rather than assumptions.

---

## Quick Reference

| Phase | Focus | Tool |
|-------|-------|------|
| Pre-Discovery | Technical context | Task subagents with skills |
| Discovery | Problem, users, value | Open-ended → AskUserQuestions |
| Details | Journey, scope, constraints | AskUserQuestions → Open-ended |
| Edge Cases | Failures, limits, security | AskUserQuestions → Open-ended |
| Validation | Summary, gaps | AskUserQuestions |

---

## Reference: pre-discovery-subagents

# Pre-Discovery with Subagents

For features spanning multiple domains (auth, database, UI, etc.) that need front-loaded technical context before the Feature Forge interview.

## Overview

For features spanning multiple domains, you can accelerate discovery by launching Task subagents with relevant skills BEFORE starting the Feature Forge interview. This front-loads technical context so the interview focuses on decisions rather than exploration.

## When to Use

- Feature touches 3+ distinct system layers (e.g., auth, database, UI)
- Codebase is unfamiliar or underdocumented
- You need concrete technical facts before asking requirements questions
- Stakeholder time is limited and you want to minimize back-and-forth

## When NOT to Use

- Feature is well-scoped to a single domain
- You already have deep codebase knowledge
- Requirements are purely business/UX (no technical exploration needed)

## Pattern

```
1. Identify domains the feature touches
2. Launch parallel Task subagents with relevant skills:
   - Architecture Designer → existing patterns and constraints
   - Framework Expert → current implementation details
   - Security Reviewer → security requirements and risks
3. Collect findings from all subagents
4. Begin Feature Forge interview with technical context loaded
5. Focus interview on decisions, trade-offs, and requirements
```

## Example

For a "user profile with avatar upload" feature:

```
Task subagent 1 (Architecture Designer):
  "Analyze the current user model, storage patterns, and image handling in this codebase"

Task subagent 2 (Security Reviewer):
  "What security concerns exist for file upload in this stack?"

Task subagent 3 (Framework Expert):
  "How does this project handle API endpoints and file storage?"
```

Results feed into the Feature Forge interview, so questions like "Where should we store avatars?" come with context about existing patterns.

## Integration with Interview Questions

See `interview-questions.md` for the full multi-agent discovery pattern and how subagent findings map to interview categories.

---

## Reference: specification-template

# Specification Template

## Full Template

```markdown
# Feature: [Name]

## Overview
[2-3 sentence description of the feature and its value to users]

## Functional Requirements

### FR-001: [Requirement Name]
While <precondition>, when <trigger>, the system shall <response>.

### FR-002: [Requirement Name]
While <precondition>, when <trigger>, the system shall <response>.

## Non-Functional Requirements

### Performance
- Response time: < 200ms p95
- Throughput: 1000 requests/minute
- Data volume: Up to 1M records

### Security
- Authentication: JWT required
- Authorization: Role-based (admin, user)
- Data protection: PII encrypted at rest

### Scalability
- Concurrent users: 10,000
- Peak load handling: Auto-scale to 3x
- Data retention: 90 days

## Acceptance Criteria

### AC-001: [Scenario Name]
Given [context/precondition]
When [action taken]
Then [expected result]

### AC-002: [Scenario Name]
Given [context/precondition]
When [action taken]
Then [expected result]

## Error Handling

| Error Condition | HTTP Code | User Message |
|-----------------|-----------|--------------|
| Invalid input | 400 | "Please check your input" |
| Unauthorized | 401 | "Please log in to continue" |
| Forbidden | 403 | "You don't have permission" |
| Not found | 404 | "Resource not found" |
| Conflict | 409 | "This already exists" |

## Implementation TODO

### Backend
- [ ] Create database migration for X table
- [ ] Implement X service with Y method
- [ ] Add API endpoint POST /api/x
- [ ] Add input validation schema
- [ ] Add authorization check

### Frontend
- [ ] Create X component
- [ ] Add form with validation
- [ ] Implement API integration
- [ ] Add loading/error states
- [ ] Add success feedback

### Testing
- [ ] Unit tests for X service
- [ ] Integration tests for API endpoint
- [ ] E2E test for complete user flow

## Out of Scope
- [Feature/capability explicitly not included]
- [Future enhancement to consider later]

## Open Questions
- [ ] [Question needing stakeholder input]
- [ ] [Technical decision pending]
```

## Save Location

Save as: `specs/{feature_name}.spec.md`

## Required Sections Checklist

| Section | Purpose | Required |
|---------|---------|----------|
| Overview | Quick understanding | Yes |
| Functional Requirements | What it does | Yes |
| Non-Functional Requirements | How well it does it | Yes |
| Acceptance Criteria | How to verify | Yes |
| Error Handling | Failure cases | Yes |
| Implementation TODO | Action items | Yes |
| Out of Scope | Prevent scope creep | Recommended |
| Open Questions | Track decisions | As needed |
