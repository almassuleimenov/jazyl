# microservices-architect

> Источник: https://github.com/jeffallan/claude-skills/tree/main/skills/microservices-architect (MIT). Рабочая копия скилла: `.claude/skills/microservices-architect/`.

---
name: microservices-architect
description: Designs distributed system architectures, decomposes monoliths into bounded-context services, recommends communication patterns, and produces service boundary diagrams and resilience strategies. Use when designing distributed systems, decomposing monoliths, or implementing microservices patterns — including service boundaries, DDD, saga patterns, event sourcing, CQRS, service mesh, or distributed tracing.
license: MIT
metadata:
  author: https://github.com/Jeffallan
  company: https://synergetic.solutions
  version: "1.1.0"
  domain: api-architecture
  triggers: microservices, service mesh, distributed systems, service boundaries, domain-driven design, event sourcing, CQRS, saga pattern, Kubernetes microservices, Istio, distributed tracing
  role: architect
  scope: system-design
  output-format: architecture
  related-skills: devops-engineer, kubernetes-specialist, graphql-architect, architecture-designer, monitoring-expert
---

# Microservices Architect

Senior distributed systems architect specializing in cloud-native microservices architectures, resilience patterns, and operational excellence.

## Core Workflow

1. **Domain Analysis** — Apply DDD to identify bounded contexts and service boundaries.
   - *Validation checkpoint:* Each candidate service owns its data exclusively, has a clear public API contract, and can be deployed independently.
2. **Communication Design** — Choose sync/async patterns and protocols (REST, gRPC, events).
   - *Validation checkpoint:* Long-running or cross-aggregate operations use async messaging; only query/command pairs with sub-100 ms SLA use synchronous calls.
3. **Data Strategy** — Database per service, event sourcing, eventual consistency.
   - *Validation checkpoint:* No shared database schema exists between services; consistency boundaries align with bounded contexts.
4. **Resilience** — Circuit breakers, retries, timeouts, bulkheads, fallbacks.
   - *Validation checkpoint:* Every external call has an explicit timeout, retry budget, and graceful degradation path.
5. **Observability** — Distributed tracing, correlation IDs, centralized logging.
   - *Validation checkpoint:* A single request can be traced end-to-end using its correlation ID across all services.
6. **Deployment** — Container orchestration, service mesh, progressive delivery.
   - *Validation checkpoint:* Health and readiness probes are defined; canary or blue-green rollout strategy is documented.

## Reference Guide

Load detailed guidance based on context:

| Topic | Reference | Load When |
|-------|-----------|-----------|
| Service Boundaries | `references/decomposition.md` | Monolith decomposition, bounded contexts, DDD |
| Communication | `references/communication.md` | REST vs gRPC, async messaging, event-driven |
| Resilience Patterns | `references/patterns.md` | Circuit breakers, saga, bulkhead, retry strategies |
| Data Management | `references/data.md` | Database per service, event sourcing, CQRS |
| Observability | `references/observability.md` | Distributed tracing, correlation IDs, metrics |

## Implementation Examples

### Correlation ID Middleware (Node.js / Express)
```js
const { v4: uuidv4 } = require('uuid');

function correlationMiddleware(req, res, next) {
  req.correlationId = req.headers['x-correlation-id'] || uuidv4();
  res.setHeader('x-correlation-id', req.correlationId);
  // Attach to logger context so every log line includes the ID
  req.log = logger.child({ correlationId: req.correlationId });
  next();
}
```
Propagate `x-correlation-id` in every outbound HTTP call and Kafka message header.

### Circuit Breaker (Python / `pybreaker`)
```python
import pybreaker

# Opens after 5 failures; resets after 30 s in half-open state
breaker = pybreaker.CircuitBreaker(fail_max=5, reset_timeout=30)

@breaker
def call_inventory_service(order_id: str):
    response = requests.get(f"{INVENTORY_URL}/stock/{order_id}", timeout=2)
    response.raise_for_status()
    return response.json()

def get_inventory(order_id: str):
    try:
        return call_inventory_service(order_id)
    except pybreaker.CircuitBreakerError:
        return {"status": "unavailable", "fallback": True}
```

### Saga Orchestration Skeleton (TypeScript)
```ts
// Each step defines execute() and compensate() so rollback is automatic.
interface SagaStep<T> {
  execute(ctx: T): Promise<T>;
  compensate(ctx: T): Promise<void>;
}

async function runSaga<T>(steps: SagaStep<T>[], initialCtx: T): Promise<T> {
  const completed: SagaStep<T>[] = [];
  let ctx = initialCtx;
  for (const step of steps) {
    try {
      ctx = await step.execute(ctx);
      completed.push(step);
    } catch (err) {
      for (const done of completed.reverse()) {
        await done.compensate(ctx).catch(console.error);
      }
      throw err;
    }
  }
  return ctx;
}

// Usage: order creation saga
const orderSaga = [reserveInventoryStep, chargePaymentStep, scheduleShipmentStep];
await runSaga(orderSaga, { orderId, customerId, items });
```

### Health & Readiness Probe (Kubernetes)
```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 15
readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
```
`/health/live` — returns 200 if the process is running.  
`/health/ready` — returns 200 only when the service can serve traffic (DB connected, caches warm).

## Constraints

### MUST DO
- Apply domain-driven design for service boundaries
- Use database per service pattern
- Implement circuit breakers for external calls
- Add correlation IDs to all requests
- Use async communication for cross-aggregate operations
- Design for failure and graceful degradation
- Implement health checks and readiness probes
- Use API versioning strategies

### MUST NOT DO
- Create distributed monoliths
- Share databases between services
- Use synchronous calls for long-running operations
- Skip distributed tracing implementation
- Ignore network latency and partial failures
- Create chatty service interfaces
- Store shared state without proper patterns
- Deploy without observability

## Output Templates

When designing microservices architecture, provide:
1. Service boundary diagram with bounded contexts
2. Communication patterns (sync/async, protocols)
3. Data ownership and consistency model
4. Resilience patterns for each integration point
5. Deployment and infrastructure requirements

## Knowledge Reference

Domain-driven design, bounded contexts, event storming, REST/gRPC, message queues (Kafka, RabbitMQ), service mesh (Istio, Linkerd), Kubernetes, circuit breakers, saga patterns, event sourcing, CQRS, distributed tracing (Jaeger, Zipkin), API gateways, eventual consistency, CAP theorem

Maintained by [@jeffallan](https://github.com/jeffallan), Principal Consultant at [Synergetic Solutions](https://synergetic.solutions)

[Documentation](https://jeffallan.github.io/claude-skills/skills/api-architecture/microservices-architect/)

---

## Reference: communication

# Inter-Service Communication Patterns

Comprehensive guide for designing communication between microservices.

## Communication Styles

### Synchronous Communication

**REST APIs:**
```
When to Use:
- Request/response pattern needed
- Client needs immediate result
- Simple CRUD operations
- Public-facing APIs

Design Principles:
- Resource-oriented URLs
- HTTP verbs (GET, POST, PUT, DELETE, PATCH)
- Stateless operations
- Idempotent operations where possible
- Proper status codes (200, 201, 400, 404, 500)

Example:
GET    /api/v1/orders/{orderId}
POST   /api/v1/orders
PUT    /api/v1/orders/{orderId}
DELETE /api/v1/orders/{orderId}
PATCH  /api/v1/orders/{orderId}/status
```

**gRPC:**
```
When to Use:
- Low-latency requirements
- Strong typing needed
- Streaming data
- Internal service-to-service calls
- Polyglot environments

Advantages:
- Binary protocol (faster than JSON)
- Built-in code generation
- Bi-directional streaming
- HTTP/2 multiplexing
- Strong schema enforcement via Protobuf

Example Proto:
service OrderService {
  rpc GetOrder(OrderRequest) returns (OrderResponse);
  rpc CreateOrder(CreateOrderRequest) returns (OrderResponse);
  rpc StreamOrders(StreamRequest) returns (stream OrderResponse);
}

message OrderRequest {
  string order_id = 1;
}

message OrderResponse {
  string order_id = 1;
  string status = 2;
  repeated OrderItem items = 3;
}
```

**GraphQL:**
```
When to Use:
- Frontend-driven data requirements
- Aggregating data from multiple services
- Flexible query requirements
- Reducing over-fetching/under-fetching

Federation Pattern:
- Each service owns its subdomain schema
- Gateway stitches schemas together
- Clients query unified API
- Services resolve their own fields

Example:
# User Service Schema
type User @key(fields: "id") {
  id: ID!
  name: String!
  email: String!
}

# Order Service Schema
extend type User @key(fields: "id") {
  id: ID! @external
  orders: [Order!]!
}
```

### Asynchronous Communication

**Message Queues (Point-to-Point):**
```
When to Use:
- Task distribution
- Load leveling
- Guaranteed delivery needed
- Single consumer per message

Examples:
- RabbitMQ with work queues
- AWS SQS
- Azure Service Bus Queues

Pattern:
Producer → Queue → Consumer
- Consumer acknowledges message
- Unacknowledged messages redelivered
- Dead letter queue for failures

Use Cases:
- Background job processing
- Email/SMS sending
- Image processing
- Report generation
```

**Event Streaming (Pub/Sub):**
```
When to Use:
- Multiple consumers need same event
- Event sourcing
- Real-time data pipelines
- Audit logging
- CQRS read model updates

Kafka Example:
Topics:
- order.created
- order.updated
- order.cancelled

Producers:
- OrderService publishes events

Consumers:
- NotificationService (send confirmation email)
- InventoryService (reserve stock)
- AnalyticsService (track metrics)
- WarehouseService (prepare shipment)

Each consumer processes independently
```

**Event-Driven Architecture:**
```
Event Types:

1. Domain Events:
   - order.placed
   - payment.completed
   - shipment.dispatched

   Characteristics:
   - Represent something that happened
   - Immutable
   - Past tense naming
   - Contain minimal necessary data

2. Integration Events:
   - Published across bounded contexts
   - Designed for external consumption
   - Schema versioned
   - Backward compatible

3. Command Events:
   - Imperative (do something)
   - Example: process.order, send.notification
   - Use sparingly (prefer domain events)

Event Schema Example:
{
  "eventId": "uuid",
  "eventType": "order.placed",
  "eventVersion": "1.0",
  "timestamp": "2025-12-14T10:00:00Z",
  "aggregateId": "order-12345",
  "correlationId": "request-uuid",
  "payload": {
    "orderId": "12345",
    "customerId": "67890",
    "totalAmount": 99.99,
    "currency": "USD"
  }
}
```

## Communication Patterns

### Request/Response

**Synchronous Request/Response:**
```
Pattern:
Client → Service A → Service B → Response

Pros:
- Simple to implement
- Immediate feedback
- Easy to debug

Cons:
- Tight temporal coupling
- Cascading failures
- Higher latency
- Blocking operations

Use When:
- Real-time user interaction
- Small number of hops (max 2-3)
- Low latency requirements
- Failure of dependency should fail request
```

**Asynchronous Request/Response:**
```
Pattern:
1. Client sends request to Service A
2. Service A returns request ID immediately
3. Service A processes asynchronously
4. Client polls or receives webhook when complete

Implementation:
POST /api/v1/orders
Response: 202 Accepted
{
  "requestId": "req-12345",
  "statusUrl": "/api/v1/requests/req-12345"
}

GET /api/v1/requests/req-12345
Response: 200 OK
{
  "status": "completed",
  "result": { ... }
}

Alternative: WebSocket notification when ready
```

### Fire and Forget

**Pattern:**
```
Client → Message Queue → Consumer

Characteristics:
- Client doesn't wait for response
- Eventual consistency
- High throughput
- Loose coupling

Example:
User uploads image:
1. API returns 202 Accepted immediately
2. Message queued: image.uploaded
3. Worker processes asynchronously:
   - Generate thumbnails
   - Optimize image
   - Update database
4. User notified via WebSocket/SSE when ready

Pros:
- Non-blocking
- Resilient (retry on failure)
- Scalable (multiple workers)

Cons:
- No immediate feedback
- Requires status tracking
- Complex error handling
```

### Event Choreography

**Pattern:**
```
Distributed workflow via events (no central orchestrator)

Example: Order Placement
1. OrderService publishes: order.created
2. PaymentService listens, processes payment, publishes: payment.completed
3. InventoryService listens, reserves stock, publishes: inventory.reserved
4. ShippingService listens, creates shipment, publishes: shipment.created
5. NotificationService listens to all, sends appropriate notifications

Pros:
- No single point of failure
- Services highly decoupled
- Scales independently

Cons:
- Difficult to understand full workflow
- Hard to debug
- No central monitoring
- Eventual consistency challenges
```

### Saga Orchestration

**Pattern:**
```
Central orchestrator manages distributed transaction

Example: Order Saga
Orchestrator: OrderSagaService

Steps:
1. Create Order (OrderService)
2. Process Payment (PaymentService)
3. Reserve Inventory (InventoryService)
4. Create Shipment (ShippingService)

If step 3 fails:
- Compensate step 2: Refund payment
- Compensate step 1: Cancel order

Implementation:
- State machine tracks progress
- Stores saga state persistently
- Handles retries and compensations
- Sends commands to services

Pros:
- Clear workflow visibility
- Easier debugging
- Centralized monitoring

Cons:
- Orchestrator can become bottleneck
- Single point of failure (mitigate with HA)
- More complex implementation
```

## Protocol Selection Guide

### Decision Matrix

**REST vs gRPC:**
```
Use REST when:
- Public API (external clients)
- Browser-based clients
- Human-readable debugging needed
- Wide tooling support required
- Caching at HTTP layer

Use gRPC when:
- Internal service-to-service
- Low latency critical
- Strong typing needed
- Bi-directional streaming
- Polyglot teams (code generation)
```

**Synchronous vs Asynchronous:**
```
Use Synchronous when:
- User waiting for response
- Strong consistency required
- Simple request/response
- Low latency possible (<100ms)
- Few service hops (1-2)

Use Asynchronous when:
- Long-running operations (>5s)
- Multiple consumers need same data
- Decoupling services
- High throughput required
- Eventual consistency acceptable
```

**Message Queue vs Event Stream:**
```
Use Message Queue (RabbitMQ, SQS) when:
- Single consumer per message
- Task distribution
- Guaranteed processing
- Simpler model sufficient

Use Event Stream (Kafka) when:
- Multiple consumers per event
- Event replay needed
- High throughput (millions/sec)
- Event sourcing
- Long retention required
```

## API Design Best Practices

### RESTful API Design

**URL Structure:**
```
Good:
GET    /api/v1/customers/{customerId}/orders
POST   /api/v1/orders
GET    /api/v1/orders/{orderId}/items

Avoid:
GET    /api/v1/getCustomerOrders?customerId=123
POST   /api/v1/createOrder
```

**Versioning Strategies:**
```
1. URL Versioning:
   /api/v1/orders
   /api/v2/orders
   Pros: Clear, easy to route
   Cons: URL pollution

2. Header Versioning:
   Accept: application/vnd.company.v1+json
   Pros: Clean URLs
   Cons: Harder to debug

3. Query Parameter:
   /api/orders?version=1
   Pros: Flexible
   Cons: Easy to miss

Recommendation: URL versioning for simplicity
```

**Pagination:**
```
Cursor-Based (Recommended):
GET /api/v1/orders?cursor=abc123&limit=20
Response:
{
  "data": [...],
  "nextCursor": "xyz789",
  "hasMore": true
}

Offset-Based (Simple but problematic):
GET /api/v1/orders?page=2&pageSize=20
Problem: Results change if data inserted
```

### gRPC Best Practices

**Error Handling:**
```
Use standard gRPC status codes:
- OK (0)
- INVALID_ARGUMENT (3)
- NOT_FOUND (5)
- ALREADY_EXISTS (6)
- PERMISSION_DENIED (7)
- RESOURCE_EXHAUSTED (8)
- FAILED_PRECONDITION (9)
- UNAVAILABLE (14)

Include error details:
rpc CreateOrder(CreateOrderRequest) returns (OrderResponse) {
  // On error, return status with details
}

Error details in metadata for rich context
```

**Streaming Patterns:**
```
1. Server Streaming:
   rpc ListOrders(ListRequest) returns (stream Order);
   Use: Large result sets

2. Client Streaming:
   rpc UploadImages(stream Image) returns (UploadResponse);
   Use: Bulk uploads

3. Bidirectional Streaming:
   rpc Chat(stream Message) returns (stream Message);
   Use: Real-time communication
```

## Summary

Choose communication patterns based on:
- Consistency requirements (strong vs eventual)
- Latency tolerance
- Coupling tolerance
- Complexity budget
- Team expertise

**Rule of Thumb:**
- Synchronous for reads and simple writes
- Asynchronous for complex workflows
- Events for cross-aggregate updates
- Sagas for distributed transactions

Always implement timeouts, retries, and circuit breakers regardless of pattern chosen.

---

## Reference: data

# Data Management in Microservices

Comprehensive guide for managing data across distributed services.

## Fundamental Principles

### Database per Service

**Core Principle:** Each microservice owns its data exclusively.

**Rules:**
```
✓ DO:
- Each service has its own database/schema
- Service owns all CRUD operations on its data
- Other services access data via APIs only
- Service can choose its own database technology

✗ DON'T:
- Share database between services
- Direct database queries across services
- Shared tables or schemas
- Database-level joins across services
```

**Implementation Options:**

**1. Separate Database Instances:**
```
UserService → PostgreSQL instance 1
OrderService → PostgreSQL instance 2
InventoryService → PostgreSQL instance 3

Pros:
- Complete isolation
- Independent scaling
- No shared resource contention

Cons:
- Higher infrastructure cost
- More operational overhead
```

**2. Separate Schemas:**
```
Same PostgreSQL instance:
- Schema: user_service
- Schema: order_service
- Schema: inventory_service

Pros:
- Lower cost
- Easier local development

Cons:
- Shared resource (CPU, memory)
- Not true isolation
- Scaling limitations

Recommendation: Use separate schemas for dev/test, separate instances for production
```

**3. Polyglot Persistence:**
```
Each service chooses optimal database:

UserService → PostgreSQL
  (Relational data, ACID transactions)

ProductCatalog → Elasticsearch
  (Full-text search, faceted navigation)

SessionStore → Redis
  (Fast key-value, TTL support)

EventLog → Kafka
  (Event streaming, replay)

RecommendationEngine → MongoDB
  (Flexible schema, denormalized data)

Benefits: Right tool for the job
Challenges: Multiple technologies to manage
```

## Data Consistency Patterns

### Strong Consistency vs Eventual Consistency

**Strong Consistency:**
```
Definition: Read after write returns latest value

Requires:
- Distributed transaction (2PC, 3PC)
- Coordination across services
- Blocking operations

Cost:
- Higher latency
- Reduced availability (CAP theorem)
- Complexity

When to Use:
- Financial transactions
- Inventory reservations
- Critical business operations
- Regulatory requirements
```

**Eventual Consistency:**
```
Definition: System converges to consistent state over time

Characteristics:
- Temporary inconsistencies acceptable
- Non-blocking operations
- Higher availability
- Lower latency

Example:
1. Order placed (OrderService)
2. Immediately return success to user
3. Event published: order.created
4. InventoryService eventually processes event
5. Stock count updated (few milliseconds later)

When to Use:
- Social media feeds
- Analytics dashboards
- Recommendation systems
- Non-critical updates
```

### Managing Cross-Service Data

**Problem:** Order service needs customer data owned by User service.

**Anti-Pattern Solutions:**
```
✗ Direct database access
✗ Shared database
✗ Database replication between services
```

**Proper Solutions:**

**1. API Composition:**
```
Client Query: Get order with customer details

API Gateway:
1. GET /orders/123 from OrderService
   Response: { orderId: 123, customerId: 456, items: [...] }

2. GET /customers/456 from UserService
   Response: { customerId: 456, name: "John", email: "john@example.com" }

3. Combine responses and return to client

Pros:
- Maintains service boundaries
- Real-time data

Cons:
- Multiple network calls (latency)
- Partial failure handling complex
- N+1 query problem
```

**2. Data Replication via Events:**
```
OrderService maintains denormalized customer data:

CREATE TABLE orders (
    order_id UUID PRIMARY KEY,
    customer_id UUID,
    customer_name VARCHAR(255),  -- Denormalized
    customer_email VARCHAR(255), -- Denormalized
    order_total DECIMAL,
    created_at TIMESTAMP
);

UserService publishes events:
- customer.created
- customer.updated
- customer.deleted

OrderService subscribes and updates local copy:

async def on_customer_updated(event):
    await db.execute(
        "UPDATE orders SET customer_name = $1, customer_email = $2 WHERE customer_id = $3",
        event.name, event.email, event.customer_id
    )

Pros:
- Fast queries (no joins across services)
- Resilient to UserService downtime

Cons:
- Eventual consistency
- Storage duplication
- Keeping data in sync
```

**3. CQRS with Shared Read Model:**
```
Write Models (Command Side):
- UserService writes to user_db
- OrderService writes to order_db

Read Model (Query Side):
- Dedicated database for queries
- Subscribes to events from both services
- Denormalized view for efficient queries

Example Read Model:
CREATE TABLE order_details_view (
    order_id UUID,
    customer_id UUID,
    customer_name VARCHAR(255),
    customer_email VARCHAR(255),
    items JSONB,
    order_total DECIMAL,
    order_status VARCHAR(50)
);

Pros:
- Optimized for queries
- No cross-service calls
- Can rebuild from events

Cons:
- Eventual consistency
- Additional infrastructure
- Event replay mechanism needed
```

## Distributed Transactions

### Two-Phase Commit (2PC)

**How It Works:**
```
Phase 1: Prepare
Coordinator asks all participants: "Can you commit?"
- Service A: YES
- Service B: YES
- Service C: YES

Phase 2: Commit
If all YES:
  Coordinator tells all: "Commit"
If any NO:
  Coordinator tells all: "Rollback"

Example:
Transfer $100 from Account A to Account B

Prepare:
- AccountService A: Can deduct $100? YES (balance sufficient)
- AccountService B: Can add $100? YES (account active)

Commit:
- AccountService A: Deduct $100 (committed)
- AccountService B: Add $100 (committed)
```

**Problems with 2PC:**
```
✗ Blocking protocol (participants wait for coordinator)
✗ Single point of failure (coordinator down = all blocked)
✗ Reduced availability
✗ Poor performance (synchronous coordination)
✗ Doesn't scale well

Recommendation: Avoid 2PC in microservices, use Saga pattern instead
```

### Saga Pattern (Recommended)

**Orchestration-Based Saga:**
```
Transfer Money Saga:

Steps:
1. Debit Account A
2. Credit Account B

Compensations:
1. Credit Account A (reverse debit)

Saga Orchestrator:
saga_state = {
    "saga_id": "saga-123",
    "status": "in_progress",
    "steps_completed": []
}

# Step 1
result1 = await account_service.debit(account_a, 100)
if not result1.success:
    return fail_saga("Insufficient funds")

saga_state["steps_completed"].append("debit_a")

# Step 2
result2 = await account_service.credit(account_b, 100)
if not result2.success:
    # Compensate step 1
    await account_service.credit(account_a, 100)
    return fail_saga("Account B invalid")

saga_state["status"] = "completed"
return success_saga()
```

**Saga State Persistence:**
```
CREATE TABLE saga_state (
    saga_id UUID PRIMARY KEY,
    saga_type VARCHAR(50),
    current_step INTEGER,
    max_steps INTEGER,
    status VARCHAR(20),
    payload JSONB,
    steps_completed JSONB,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

After each step:
UPDATE saga_state
SET
    current_step = current_step + 1,
    steps_completed = jsonb_array_append(steps_completed, 'step_name'),
    updated_at = NOW()
WHERE saga_id = $1;

On failure, load saga state and execute compensations
```

**Idempotency for Saga Steps:**
```
Each saga step must be idempotent:

Debit Operation:
async def debit_account(account_id, amount, saga_id):
    # Check if already processed
    existing = await db.fetchone(
        "SELECT * FROM transactions WHERE saga_id = $1 AND operation = 'debit'",
        saga_id
    )
    if existing:
        return {"success": True, "transaction_id": existing.id}

    # Process debit
    result = await db.execute(
        "UPDATE accounts SET balance = balance - $1 WHERE id = $2 AND balance >= $1",
        amount, account_id
    )

    if result.rowcount == 0:
        return {"success": False, "error": "Insufficient funds"}

    # Record transaction
    await db.execute(
        "INSERT INTO transactions (saga_id, account_id, amount, operation) VALUES ($1, $2, $3, 'debit')",
        saga_id, account_id, amount
    )

    return {"success": True}

Compensating Operation:
async def compensate_debit(account_id, amount, saga_id):
    await credit_account(account_id, amount, saga_id)
```

## Event Sourcing

### Core Concepts

**Event Store:**
```
All state changes stored as immutable events

Example: Bank Account

Events:
1. AccountOpened { accountId: "acc-123", customerId: "cust-456", initialBalance: 0 }
2. MoneyDeposited { accountId: "acc-123", amount: 1000, timestamp: "2025-01-15T10:00:00Z" }
3. MoneyWithdrawn { accountId: "acc-123", amount: 200, timestamp: "2025-01-16T14:30:00Z" }
4. MoneyDeposited { accountId: "acc-123", amount: 500, timestamp: "2025-01-17T09:15:00Z" }

Current Balance = 0 + 1000 - 200 + 500 = 1300

Replay all events to reconstruct current state
```

**Event Schema:**
```json
{
  "eventId": "evt-789",
  "aggregateId": "acc-123",
  "aggregateType": "BankAccount",
  "eventType": "MoneyDeposited",
  "eventVersion": "1.0",
  "timestamp": "2025-01-15T10:00:00Z",
  "correlationId": "corr-456",
  "causationId": "cmd-123",
  "payload": {
    "amount": 1000,
    "currency": "USD",
    "source": "wire_transfer"
  },
  "metadata": {
    "userId": "user-789",
    "ipAddress": "192.168.1.1"
  }
}
```

### Snapshots

**Problem:** Replaying thousands of events is slow.

**Solution:** Periodic snapshots.

```
Event Stream:
1. AccountOpened (version 1)
2. MoneyDeposited (version 2)
...
1000. MoneyDeposited (version 1000)
[SNAPSHOT at version 1000: balance = $50,000]
1001. MoneyWithdrawn (version 1001)
...
1500. MoneyDeposited (version 1500)

To get current state:
1. Load snapshot at version 1000 (balance = $50,000)
2. Replay events 1001-1500 (only 500 events)

Much faster than replaying all 1500 events

Snapshot Strategy:
- Every 100 events
- Or every 24 hours
- Async background process
```

**Snapshot Table:**
```sql
CREATE TABLE snapshots (
    aggregate_id UUID,
    aggregate_type VARCHAR(50),
    version INTEGER,
    state JSONB,
    created_at TIMESTAMP,
    PRIMARY KEY (aggregate_id, version)
);

CREATE INDEX idx_latest_snapshot ON snapshots(aggregate_id, version DESC);
```

### Event Schema Evolution

**Challenge:** Events are immutable, but requirements change.

**Strategies:**

**1. Event Versioning:**
```
Version 1:
{
  "eventType": "OrderPlaced",
  "eventVersion": "1.0",
  "payload": {
    "orderId": "123",
    "amount": 99.99
  }
}

Version 2 (added customer email):
{
  "eventType": "OrderPlaced",
  "eventVersion": "2.0",
  "payload": {
    "orderId": "123",
    "amount": 99.99,
    "customerEmail": "customer@example.com"
  }
}

Event Handler:
def handle_order_placed(event):
    if event.eventVersion == "1.0":
        # Handle old format
        process_order_v1(event.payload)
    elif event.eventVersion == "2.0":
        # Handle new format
        process_order_v2(event.payload)
```

**2. Event Upcasting:**
```
Transform old events to new format during replay:

def upcast_event(event):
    if event.eventType == "OrderPlaced" and event.eventVersion == "1.0":
        # Transform to v2.0
        return {
            "eventType": "OrderPlaced",
            "eventVersion": "2.0",
            "payload": {
                **event.payload,
                "customerEmail": "unknown@example.com"  # Default value
            }
        }
    return event
```

**3. Event Transformation:**
```
Create new event types, keep old ones for historical accuracy:

Old: OrderPlaced
New: OrderPlacedV2

Projections handle both:
- Old events for historical data
- New events for current processing
```

## Data Synchronization

### Change Data Capture (CDC)

**Purpose:** Capture database changes and publish as events.

**How It Works:**
```
Database transaction log → CDC Tool → Event Stream

Example with Debezium:

PostgreSQL:
INSERT INTO orders (id, customer_id, total) VALUES (123, 456, 99.99);

Debezium captures:
{
  "before": null,
  "after": {
    "id": 123,
    "customer_id": 456,
    "total": 99.99,
    "created_at": "2025-01-15T10:00:00Z"
  },
  "op": "c",  // create
  "ts_ms": 1705314000000
}

Published to Kafka topic: postgres.public.orders

Other services subscribe and update their read models
```

**Benefits:**
```
✓ No application code changes
✓ Guaranteed delivery (based on database transaction log)
✓ Captures all changes (even from direct DB access)
✓ Low latency
✓ Ordering preserved

Use Cases:
- Keep search index synchronized
- Update cache automatically
- Replicate to data warehouse
- Trigger workflows on database changes
```

### Materialized Views

**Purpose:** Pre-computed denormalized views for fast queries.

**Pattern:**
```
Event-Driven Materialized View:

1. Services publish domain events
2. View service subscribes to events
3. Updates materialized view in real-time

Example: Order Summary View

Events:
- order.created
- order.payment_received
- order.shipped
- order.delivered

Materialized View:
CREATE TABLE order_summary (
    order_id UUID PRIMARY KEY,
    customer_id UUID,
    customer_name VARCHAR(255),
    order_date TIMESTAMP,
    total_amount DECIMAL,
    status VARCHAR(50),
    items_count INTEGER,
    last_updated TIMESTAMP
);

View Service:
async def on_order_created(event):
    await db.execute(
        "INSERT INTO order_summary (order_id, customer_id, status, ...) VALUES (...)",
        event.data
    )

async def on_order_shipped(event):
    await db.execute(
        "UPDATE order_summary SET status = 'shipped', last_updated = NOW() WHERE order_id = $1",
        event.order_id
    )
```

## Data Partitioning

### Horizontal Partitioning (Sharding)

**When to Use:**
```
- Single database can't handle load
- Data size exceeds single server capacity
- Want to distribute geographically
```

**Sharding Strategies:**

**1. Hash-Based Sharding:**
```
Shard = hash(customer_id) % num_shards

customer_id: cust-123 → hash → 7234 → mod 4 → Shard 2
customer_id: cust-456 → hash → 9812 → mod 4 → Shard 0

Pros:
- Even distribution
- Simple to implement

Cons:
- Adding shards requires re-sharding
- Range queries difficult
```

**2. Range-Based Sharding:**
```
Shard 0: customer_id 0-999
Shard 1: customer_id 1000-1999
Shard 2: customer_id 2000-2999

Pros:
- Range queries efficient
- Easy to add shards

Cons:
- Uneven distribution (hotspots)
- Requires shard map
```

**3. Geography-Based Sharding:**
```
Shard US: customers in USA
Shard EU: customers in Europe
Shard APAC: customers in Asia-Pacific

Pros:
- Data locality (GDPR compliance)
- Lower latency

Cons:
- Uneven distribution
- Cross-shard queries complex
```

**Shard Management:**
```
Shard Map Service:

GET /shard-location?customer_id=cust-123
Response: { "shard": "shard-2", "endpoint": "db2.example.com" }

Application logic:
customer_id = request.customer_id
shard_info = await shard_map.get_shard(customer_id)
db_connection = connection_pool.get(shard_info.endpoint)
result = await db_connection.query("SELECT * FROM customers WHERE id = $1", customer_id)
```

## Summary

Data management in microservices requires careful design:

**Key Principles:**
- Database per service (non-negotiable)
- Embrace eventual consistency where possible
- Use Saga pattern for distributed transactions
- Event sourcing for audit trail and temporal queries
- CQRS for read/write optimization
- CDC for data synchronization

**Decision Framework:**
- Strong consistency → Saga with careful compensation logic
- Audit trail → Event sourcing
- Complex queries → CQRS with read models
- Large scale → Sharding with appropriate strategy

Always design for failure: compensating transactions, idempotent operations, and proper monitoring are essential.

---

## Reference: decomposition

# Service Decomposition and Boundaries

Guide for identifying service boundaries using domain-driven design principles.

## Domain-Driven Design Foundation

### Bounded Context Identification

**Strategic Patterns:**
- **Ubiquitous Language** - Each bounded context has its own domain language
- **Context Mapping** - Define relationships between bounded contexts
- **Subdomain Classification** - Core, supporting, generic domains

**Bounded Context Indicators:**
```
Strong Indicators:
- Different teams own different parts
- Different release cadences needed
- Different scalability requirements
- Different technology stacks optimal
- Clear domain model boundaries

Warning Signs:
- Entities mean different things in different contexts
- Same term used with different meanings
- Workflows cross multiple aggregates
- Teams communicate frequently about shared data
```

### Service Boundary Patterns

**Database-Driven Decomposition:**
```
1. Identify aggregates (entities with invariants)
2. Each aggregate becomes a service candidate
3. Group related aggregates by transaction boundaries
4. Services own their data (no shared databases)
```

**Business Capability Decomposition:**
```
Services organized by:
- User Management (authentication, profiles, permissions)
- Order Management (cart, checkout, fulfillment)
- Inventory Management (stock, warehousing, allocation)
- Payment Processing (transactions, refunds, reconciliation)
- Notification Service (email, SMS, push notifications)
```

**Strangler Fig Pattern:**
```
Monolith Decomposition Strategy:
1. Identify seams in existing codebase
2. Extract one service at a time
3. Route traffic through facade/proxy
4. Gradually migrate functionality
5. Decommission old code when safe

Order of Extraction:
1. Start with leaf dependencies (no downstream calls)
2. Extract supporting services first
3. Core business logic last
4. Data migration strategy per service
```

## Service Sizing Guidelines

### Microservice Characteristics

**Right-Sized Service:**
```
Team Metrics:
- 2-pizza team can own it (5-9 people)
- Single team has full ownership
- Can be rewritten in 2-4 weeks if needed
- Independent deployment pipeline

Technical Metrics:
- 100-1000 lines of business logic
- 5-15 API endpoints
- 1-5 database tables
- Startup time < 30 seconds
- Single responsibility focus
```

**Too Small (Nano-service):**
```
Warning Signs:
- Services with 1-2 endpoints
- Excessive network overhead
- More infrastructure than business logic
- Difficult to trace requests
- Version coupling between services
```

**Too Large (Distributed Monolith):**
```
Warning Signs:
- Multiple teams working on same service
- Conflicting scalability requirements
- Difficult to understand in one sitting
- Long deployment times
- Tight coupling with other services
```

## Conway's Law Alignment

### Team Structure and Service Design

**Team Topologies:**
```
Stream-Aligned Teams:
- Own end-to-end service lifecycle
- Aligned to business capabilities
- Full-stack ownership (frontend to database)

Platform Teams:
- Provide self-service capabilities
- Enable stream-aligned teams
- Kubernetes, CI/CD, observability

Enabling Teams:
- Help with complex implementations
- Service mesh setup, security patterns
- Temporary coaching role

Complicated Subsystem Teams:
- Specialized domains (ML, search, payments)
- Heavy technical expertise required
- Clear interfaces to other teams
```

## Decomposition Checklist

### Pre-Decomposition Analysis

**Business Justification:**
```
Check:
- Independent scalability needed?
- Different teams responsible?
- Isolated failure acceptable?
- Frequent independent deployments?
- Technology diversity required?

If mostly "no" → Consider modular monolith first
```

**Technical Readiness:**
```
Prerequisites:
✓ CI/CD pipelines automated
✓ Monitoring and alerting in place
✓ Distributed tracing capability
✓ Container orchestration ready
✓ Team has microservices experience
✓ Clear service ownership model
```

### Decomposition Steps

**1. Identify Bounded Contexts:**
```
Activities:
- Event storming workshop
- Identify aggregates and entities
- Map business workflows
- Document ubiquitous language
- Draw context boundaries
```

**2. Define Service Contracts:**
```
For each service:
- REST/gRPC API specification
- Event schema definitions
- Data ownership boundaries
- SLA commitments (latency, availability)
- Versioning strategy
```

**3. Plan Data Migration:**
```
Data Strategy:
- Identify shared data
- Choose consistency model (eventual vs strong)
- Design data synchronization mechanism
- Plan schema evolution
- Test rollback scenarios
```

**4. Extract Service:**
```
Implementation Order:
1. Create new service skeleton
2. Implement business logic
3. Set up database (if needed)
4. Add observability (logs, metrics, traces)
5. Deploy to staging
6. Dual-write from monolith (if applicable)
7. Switch reads to new service
8. Remove from monolith
9. Production deployment
```

## Anti-Patterns to Avoid

### Common Mistakes

**Distributed Monolith:**
```
Symptoms:
- Services must deploy together
- Shared database between services
- Synchronous coupling everywhere
- Version lock-step required
- Cascading failures common

Solution:
- Enforce database per service
- Use async communication
- Version APIs independently
- Add circuit breakers
```

**Entity Services:**
```
Anti-Pattern:
UserService (CRUD on User entity)
OrderService (CRUD on Order entity)
ProductService (CRUD on Product entity)

Problem: Anemic domain model, no business logic

Better Approach:
AccountManagement (authentication, authorization, profiles)
OrderFulfillment (workflow: cart → payment → shipping)
ProductCatalog (search, recommendations, inventory)
```

**Shared Libraries with Business Logic:**
```
Anti-Pattern:
common-lib (shared across all services with domain logic)

Problem:
- Tight coupling via dependency
- Forces synchronized deployments
- Violates service autonomy

Better:
- Shared libraries for technical concerns only
- Duplicate business logic per service
- Use events to keep data synchronized
```

## Service Boundary Validation

### Design Review Checklist

**Service Independence:**
```
Questions:
- Can this service be deployed independently?
- Does it own its data completely?
- Can it function if dependencies are down?
- Is the team autonomous to make changes?
- Are integration points well-defined?
```

**Data Ownership:**
```
Verify:
- No shared database tables
- Clear data ownership boundaries
- Event-driven synchronization for shared concepts
- API provides all necessary data
- No direct database access from other services
```

**Operational Readiness:**
```
Check:
- Health check endpoint implemented
- Readiness probe configured
- Circuit breakers for external calls
- Distributed tracing instrumented
- Logs structured with correlation IDs
- Metrics exposed (Prometheus format)
- Documentation up to date
```

## Migration Strategies

### Monolith to Microservices

**Gradual Extraction:**
```
Phase 1: Prepare
- Add seams to monolith
- Implement API layer
- Set up monitoring

Phase 2: Extract Leaf Services
- Start with services that have no dependencies
- Examples: notification service, reporting

Phase 3: Extract Supporting Services
- Authentication/authorization
- User management
- File storage

Phase 4: Extract Core Services
- Order processing
- Payment handling
- Inventory management

Phase 5: Decompose Remaining Monolith
- Gradual extraction
- Eventual retirement
```

**Parallel Run Pattern:**
```
Strategy:
1. Build new microservice
2. Run both systems simultaneously
3. Compare outputs (shadow mode)
4. Gradually shift traffic
5. Decommission old system

Use for: High-risk migrations, critical paths
```

## Summary

Service decomposition is both art and science. Start with domain-driven design to identify natural boundaries, align with team structure, and extract incrementally. Avoid the temptation to over-decompose. A modular monolith is better than a poorly designed distributed system.

**Key Takeaways:**
- Bounded contexts define service boundaries
- Database per service is non-negotiable
- Team autonomy drives service design
- Extract incrementally, not all at once
- Observability is prerequisite for microservices

---

## Reference: observability

# Observability in Microservices

Comprehensive guide for monitoring, tracing, and debugging distributed systems.

## The Three Pillars

### 1. Metrics

**Purpose:** Quantitative measurements of system behavior over time.

**Categories:**

**Business Metrics:**
```
Examples:
- Orders per minute
- Revenue per hour
- Active users
- Conversion rate
- Cart abandonment rate

Why Important:
- Align with business goals
- Detect business anomalies
- Inform scaling decisions

Implementation:
from prometheus_client import Counter, Histogram

orders_total = Counter(
    'orders_total',
    'Total number of orders',
    ['status', 'payment_method']
)

order_value = Histogram(
    'order_value_dollars',
    'Order value in dollars',
    buckets=[10, 50, 100, 500, 1000, 5000]
)

# In code
orders_total.labels(status='completed', payment_method='credit_card').inc()
order_value.observe(order.total_amount)
```

**System Metrics:**
```
Infrastructure:
- CPU usage
- Memory usage
- Disk I/O
- Network throughput

Application:
- Request rate
- Error rate
- Request duration (latency)
- Active connections
- Thread pool utilization

Database:
- Query duration
- Connection pool usage
- Slow queries
- Deadlocks

Message Queue:
- Queue depth
- Message processing rate
- Consumer lag
- Dead letter queue size
```

**The Four Golden Signals (Google SRE):**
```
1. Latency:
   - Time to serve requests
   - Track p50, p95, p99, p99.9
   - Separate success vs error latency

   request_duration = Histogram(
       'http_request_duration_seconds',
       'HTTP request duration',
       ['method', 'endpoint', 'status']
   )

2. Traffic:
   - Requests per second
   - Transactions per second
   - Concurrent users

   requests_total = Counter(
       'http_requests_total',
       'Total HTTP requests',
       ['method', 'endpoint', 'status']
   )

3. Errors:
   - Rate of failed requests
   - 4xx vs 5xx errors
   - Exception types

   errors_total = Counter(
       'errors_total',
       'Total errors',
       ['service', 'error_type']
   )

4. Saturation:
   - Resource utilization
   - Queue depth
   - Thread pool usage

   connection_pool_usage = Gauge(
       'db_connection_pool_active',
       'Active database connections'
   )
```

**RED Method (for services):**
```
- Rate: Requests per second
- Errors: Failed requests per second
- Duration: Request latency distribution

Perfect for microservices dashboards
```

**USE Method (for resources):**
```
- Utilization: Percentage of time resource busy
- Saturation: Queue depth or waiting threads
- Errors: Error count

Perfect for infrastructure monitoring
```

### 2. Logs

**Purpose:** Discrete event records with context.

**Structured Logging:**
```json
{
  "timestamp": "2025-12-14T15:30:45.123Z",
  "level": "INFO",
  "service": "order-service",
  "version": "1.2.3",
  "traceId": "abc123def456",
  "spanId": "span789",
  "userId": "user-123",
  "message": "Order created successfully",
  "orderId": "order-456",
  "totalAmount": 99.99,
  "currency": "USD",
  "duration_ms": 45,
  "endpoint": "/api/v1/orders",
  "method": "POST",
  "statusCode": 201
}
```

**Log Levels:**
```
ERROR:
- Application errors
- Failed operations
- Exceptions
Use: Alerts, immediate attention

WARN:
- Degraded functionality
- Retry attempts
- Deprecated API usage
Use: Investigation, potential issues

INFO:
- Business events (order created, user logged in)
- System events (service started, configuration loaded)
Use: Audit trail, business analytics

DEBUG:
- Detailed execution flow
- Variable values
- Function entry/exit
Use: Development, troubleshooting

TRACE:
- Very detailed debugging
Use: Deep troubleshooting (disabled in production usually)
```

**Correlation IDs:**
```
Request flow across services:

Client Request → API Gateway
                 ↓ (correlationId: corr-123)
                 Order Service
                 ↓ (correlationId: corr-123)
                 Payment Service
                 ↓ (correlationId: corr-123)
                 Notification Service

All logs include correlationId: corr-123
Easy to trace entire request flow

Implementation:
import logging
from contextvars import ContextVar

correlation_id_var = ContextVar('correlation_id', default=None)

class CorrelationIdFilter(logging.Filter):
    def filter(self, record):
        record.correlation_id = correlation_id_var.get()
        return True

# Middleware
async def correlation_middleware(request, call_next):
    correlation_id = request.headers.get('X-Correlation-ID', str(uuid4()))
    correlation_id_var.set(correlation_id)
    response = await call_next(request)
    response.headers['X-Correlation-ID'] = correlation_id
    return response
```

**Log Aggregation:**
```
Services → Log Shipper → Centralized Log Storage → Visualization

Tools:
- ELK Stack (Elasticsearch, Logstash, Kibana)
- EFK Stack (Elasticsearch, Fluentd, Kibana)
- Loki (from Grafana)
- CloudWatch Logs (AWS)
- Stackdriver (GCP)

Query Examples:
# Find all errors for specific user
service:"order-service" AND level:"ERROR" AND userId:"user-123"

# Find slow requests
service:"payment-service" AND duration_ms:>5000

# Find requests with specific correlation ID
correlationId:"corr-123"
```

### 3. Distributed Tracing

**Purpose:** Visualize request flow across services, identify bottlenecks.

**Concepts:**

**Trace:**
```
Entire request journey across all services

Example: User places order
Trace ID: trace-abc123

Spans in trace:
1. api-gateway: /checkout (200ms)
2. order-service: createOrder (150ms)
3. payment-service: processPayment (80ms)
4. inventory-service: reserveItems (40ms)
5. notification-service: sendEmail (30ms)

Total: 200ms (some parallel execution)
```

**Span:**
```
Single operation within a trace

Span attributes:
{
  "traceId": "trace-abc123",
  "spanId": "span-456",
  "parentSpanId": "span-123",
  "name": "POST /api/v1/orders",
  "startTime": "2025-12-14T15:30:45.000Z",
  "endTime": "2025-12-14T15:30:45.150Z",
  "duration": 150,
  "status": "OK",
  "attributes": {
    "http.method": "POST",
    "http.url": "/api/v1/orders",
    "http.status_code": 201,
    "user.id": "user-123",
    "order.id": "order-456",
    "order.total": 99.99
  },
  "events": [
    {
      "timestamp": "2025-12-14T15:30:45.050Z",
      "name": "Validating order items"
    },
    {
      "timestamp": "2025-12-14T15:30:45.100Z",
      "name": "Calling payment service"
    }
  ]
}
```

**Implementation (OpenTelemetry):**
```python
from opentelemetry import trace
from opentelemetry.exporter.jaeger.thrift import JaegerExporter
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor

# Setup tracing
provider = TracerProvider()
jaeger_exporter = JaegerExporter(
    agent_host_name="jaeger",
    agent_port=6831
)
provider.add_span_processor(BatchSpanProcessor(jaeger_exporter))
trace.set_tracer_provider(provider)

# Instrument FastAPI
app = FastAPI()
FastAPIInstrumentor.instrument_app(app)

# Manual span creation
tracer = trace.get_tracer(__name__)

async def create_order(order_data):
    with tracer.start_as_current_span("create_order") as span:
        span.set_attribute("order.items_count", len(order_data.items))
        span.set_attribute("order.total", order_data.total)

        # Database operation
        with tracer.start_as_current_span("db.insert_order"):
            order_id = await db.insert_order(order_data)

        # Call payment service
        with tracer.start_as_current_span("http.payment_service") as payment_span:
            payment_span.set_attribute("http.url", f"{PAYMENT_URL}/payments")
            result = await payment_service.charge(order_id, order_data.total)

        return order_id
```

**Trace Visualization:**
```
Jaeger UI shows:

Timeline view:
|-- api-gateway (200ms) ----------------------------------|
    |-- order-service (150ms) ------------------------|
        |-- db.insert_order (30ms) --|
        |-- payment-service (80ms) -----------------|
            |-- db.create_transaction (20ms) ----|
        |-- notification-service (30ms) ----------|

Critical path highlighted
Bottlenecks identified (payment-service taking 80ms)
Parallel operations visible
```

**Sampling Strategies:**
```
Problem: Tracing every request is expensive

Solutions:

1. Probabilistic Sampling:
   - Trace 1% of requests
   - Good for high-volume services

2. Rate Limiting Sampling:
   - Max 100 traces per second
   - Prevents overwhelming trace backend

3. Tail-Based Sampling:
   - Trace all errors
   - Trace slow requests (>5s)
   - Sample 1% of fast successful requests

4. Priority Sampling:
   - Always trace premium users
   - Always trace critical endpoints
   - Sample others

Implementation:
from opentelemetry.sdk.trace.sampling import (
    ParentBasedTraceIdRatioBased,
    ALWAYS_ON,
    ALWAYS_OFF
)

# Sample 1% of traces
sampler = ParentBasedTraceIdRatioBased(0.01)

# Or custom sampler
class CustomSampler:
    def should_sample(self, context, trace_id, name, attributes):
        # Always sample errors
        if attributes.get("http.status_code", 0) >= 500:
            return ALWAYS_ON

        # Always sample slow requests
        if attributes.get("duration_ms", 0) > 5000:
            return ALWAYS_ON

        # Sample 1% of others
        return ParentBasedTraceIdRatioBased(0.01).should_sample(...)
```

## Service Level Objectives (SLOs)

### Defining SLOs

**SLI (Service Level Indicator):**
```
Quantitative measure of service level

Examples:
- Request latency: p99 < 200ms
- Availability: 99.9% of requests succeed
- Throughput: Handle 10,000 requests/sec
```

**SLO (Service Level Objective):**
```
Target value for SLI

Examples:
- 99.9% of requests complete in < 200ms
- 99.95% availability over 30 days
- Zero data loss

SLO Components:
- Metric: What you measure (latency, availability)
- Target: Threshold (99.9%, 200ms)
- Time window: Evaluation period (30 days, weekly)
```

**SLA (Service Level Agreement):**
```
Contract with consequences if SLO not met

Example:
- SLO: 99.9% availability
- SLA: If availability < 99.9%, customers get 10% credit

SLA ≤ SLO (leave buffer for incidents)
```

**Error Budget:**
```
Allowed failure to meet SLO = (100% - SLO target)

Example:
SLO: 99.9% availability
Error budget: 0.1% = 43.8 minutes downtime per month

Error budget consumed:
- Outages
- Slow responses
- Failed requests

When error budget exhausted:
- Freeze feature deployments
- Focus on reliability
- Only critical fixes deployed

Benefits:
- Balances innovation vs stability
- Data-driven deployment decisions
- Aligns engineering priorities
```

### Implementing SLO Monitoring

**Prometheus + Grafana:**
```
# SLI: Availability
availability_sli = (
    sum(rate(http_requests_total{status!~"5.."}[30d]))
    /
    sum(rate(http_requests_total[30d]))
) * 100

# SLI: Latency
latency_sli = histogram_quantile(
    0.99,
    rate(http_request_duration_seconds_bucket[30d])
)

# Error Budget
error_budget_remaining = (
    1 - (target_slo / 100)
) - (
    1 - (availability_sli / 100)
)

Alert when error budget < 10%:
alert: ErrorBudgetCritical
expr: error_budget_remaining < 0.1
annotations:
  summary: "Error budget critically low"
  description: "Only 10% error budget remaining. Freeze deployments."
```

## Alerting Strategies

### Alert Levels

**Critical (Page immediately):**
```
Conditions:
- Service completely down
- Error rate > 50%
- Data loss occurring
- SLO burn rate critical

Actions:
- Page on-call engineer
- Incident created automatically
- Escalate if not acknowledged in 5 min

Example:
alert: ServiceDown
expr: up{service="payment-service"} == 0
for: 1m
severity: critical
```

**Warning (Investigate soon):**
```
Conditions:
- Elevated error rate (5-10%)
- Latency degraded (p99 > 500ms)
- Queue depth increasing
- Error budget < 25%

Actions:
- Slack notification
- Create ticket
- Investigate during business hours

Example:
alert: HighErrorRate
expr: rate(http_requests_total{status="500"}[5m]) > 0.05
for: 10m
severity: warning
```

**Info (Awareness):**
```
Conditions:
- Deployment completed
- Scaling event
- Configuration changed
- Capacity threshold reached

Actions:
- Log to monitoring system
- Dashboard annotation
- Optional Slack notification
```

### Alert Best Practices

**Actionable Alerts:**
```
Bad Alert:
"High CPU usage"

Good Alert:
"CPU usage > 80% on order-service-pod-abc for 10 minutes
Runbook: https://wiki.company.com/runbooks/high-cpu
Likely cause: Memory leak or infinite loop
Actions: 1) Check recent deployments 2) Review logs for exceptions 3) Consider rolling back"

Include:
✓ What is wrong
✓ Why it matters
✓ How to investigate
✓ Runbook link
✓ Suggested actions
```

**Avoid Alert Fatigue:**
```
Problems:
- Too many alerts
- False positives
- Non-actionable alerts
- Duplicate alerts

Solutions:
- Alert on symptoms, not causes
- Proper thresholds and durations
- Alert aggregation (don't alert per pod, alert per service)
- Regular alert review and tuning
- Auto-resolve alerts
- Silence during maintenance

Good Practice:
for: 5m  # Don't alert on transient spikes
group_by: [service]  # Aggregate per service
group_wait: 30s  # Wait before sending
group_interval: 5m  # Batch notifications
```

## Observability Stack

### Recommended Tools

**Metrics:**
```
Collection: Prometheus
- Pull-based metrics
- Time-series database
- Powerful query language (PromQL)
- Service discovery

Visualization: Grafana
- Beautiful dashboards
- Alerting integration
- Multiple data sources
- Template variables

Alternative: Datadog, New Relic, CloudWatch
```

**Logs:**
```
Aggregation: ELK Stack
- Elasticsearch (storage & search)
- Logstash / Fluentd (collection)
- Kibana (visualization)

Or: Loki (lightweight alternative)
- Integrates with Grafana
- Labels instead of full-text indexing
- Lower resource usage

Alternative: Splunk, Datadog, CloudWatch Logs
```

**Tracing:**
```
Backend: Jaeger or Zipkin
- Trace storage
- Trace visualization
- Dependency graphs
- Performance analysis

Instrumentation: OpenTelemetry
- Vendor-neutral standard
- Auto-instrumentation for common frameworks
- Manual instrumentation API
- Export to any backend

Alternative: Datadog APM, New Relic, Lightstep
```

**All-in-One:**
```
Observability platforms:
- Datadog (metrics, logs, traces, RUM)
- New Relic (APM, logs, infrastructure)
- Dynatrace (auto-instrumentation, AI)

Pros:
- Unified experience
- Correlated data
- Easier setup

Cons:
- Vendor lock-in
- Higher cost
- Less flexibility
```

### Implementation Checklist

**For Each Service:**
```
✓ Structured logging with correlation IDs
✓ Metrics exported (Prometheus format)
✓ Distributed tracing instrumented
✓ Health check endpoints (/health/live, /health/ready)
✓ Graceful shutdown handling
✓ Resource limits set (CPU, memory)
✓ Alerts configured for critical paths
✓ Dashboards created
✓ Runbooks documented
✓ On-call rotation established
```

**For System-Wide:**
```
✓ Centralized log aggregation
✓ Distributed tracing backend
✓ Metrics aggregation and storage
✓ Unified dashboards (service overview)
✓ Alert routing configured
✓ Incident management process
✓ Post-mortem template
✓ SLO definitions and tracking
✓ Dependency mapping
✓ Chaos engineering experiments
```

## Troubleshooting Workflow

**Incident Response:**
```
1. Detect (Alert fires)
   - Check dashboard
   - Verify alert is valid
   - Assess impact

2. Triage (Determine severity)
   - Critical: Page on-call
   - Warning: Create ticket
   - How many users affected?
   - What functionality broken?

3. Investigate (Find root cause)
   - Check recent deployments
   - Review logs (search by correlation ID)
   - Analyze traces (slow operations)
   - Check metrics (resource saturation)
   - Examine dependencies

4. Mitigate (Stop the bleeding)
   - Rollback deployment
   - Scale up resources
   - Failover to backup
   - Enable circuit breakers
   - Rate limit traffic

5. Resolve (Fix root cause)
   - Deploy fix
   - Verify resolution
   - Monitor for recurrence

6. Post-mortem (Learn and improve)
   - Timeline of events
   - Root cause analysis
   - Action items
   - Update runbooks
```

**Using Traces to Debug:**
```
Scenario: API returning 500 errors

1. Find failing trace:
   - Filter: status = error, service = api-gateway
   - Sort by timestamp (most recent)

2. Analyze span waterfall:
   - Identify which service failed (order-service returned 500)
   - Check error message in span
   - Review span attributes

3. Correlate with logs:
   - Extract trace ID from failed trace
   - Search logs: traceId:"trace-abc123"
   - Find exception stack trace

4. Check related metrics:
   - order-service error rate spiked 10 min ago
   - Corresponds with deployment
   - Likely cause: Bad deployment

5. Remediate:
   - Rollback order-service
   - Verify errors stopped
   - Create ticket for bug fix
```

## Summary

Observability is non-negotiable in microservices:

**Must-Haves:**
- Structured logging with correlation IDs
- Metrics (RED/USE methodology)
- Distributed tracing (OpenTelemetry)
- Centralized log aggregation
- SLO tracking with error budgets
- Actionable alerts with runbooks

**Best Practices:**
- Correlate metrics, logs, and traces
- Define SLOs based on user experience
- Alert on symptoms, not causes
- Maintain runbooks for common issues
- Regular post-mortems and learning
- Practice incident response with game days

Without observability, you're flying blind in production.

---

## Reference: patterns

# Resilience and Reliability Patterns

Essential patterns for building fault-tolerant distributed systems.

## Resilience Patterns

### Circuit Breaker

**Purpose:** Prevent cascading failures by failing fast when a dependency is unhealthy.

**How It Works:**
```
States:
1. CLOSED (normal operation)
   - Requests pass through
   - Track failure rate
   - If failures exceed threshold → OPEN

2. OPEN (failing fast)
   - Immediately reject requests
   - Return fallback response
   - After timeout period → HALF_OPEN

3. HALF_OPEN (testing recovery)
   - Allow limited test requests
   - If successful → CLOSED
   - If failed → OPEN

Configuration:
- Failure threshold: 50% failures in 10 requests
- Timeout: 30 seconds in OPEN state
- Success threshold: 2 consecutive successes in HALF_OPEN
```

**Implementation Example:**
```python
# Using resilience4j-like pattern
@CircuitBreaker(
    name="payment-service",
    fallbackMethod="paymentFallback",
    failureThreshold=50,
    waitDurationInOpenState=30000,  # 30s
    permittedNumberOfCallsInHalfOpenState=3
)
async def process_payment(order_id: str, amount: float):
    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{PAYMENT_SERVICE_URL}/payments",
            json={"orderId": order_id, "amount": amount},
            timeout=5.0
        )
        return response.json()

async def paymentFallback(order_id: str, amount: float, exception):
    # Log the failure
    logger.error(f"Payment service unavailable: {exception}")
    # Return graceful degradation
    return {
        "status": "pending",
        "message": "Payment processing delayed, will retry"
    }
```

**When to Use:**
```
Apply circuit breakers to:
✓ External service calls
✓ Database queries
✓ Third-party APIs
✓ Microservice-to-microservice calls

Configuration Guidelines:
- Fast services (p99 < 100ms): 5s timeout, 10s circuit open
- Medium services (p99 < 1s): 10s timeout, 30s circuit open
- Slow services (p99 > 1s): 30s timeout, 60s circuit open
```

### Retry Pattern

**Purpose:** Handle transient failures by retrying operations.

**Strategies:**

**1. Exponential Backoff:**
```
Retry delays: 100ms, 200ms, 400ms, 800ms, 1600ms

Benefits:
- Reduces load during incidents
- Gives service time to recover
- Prevents thundering herd

Implementation:
attempts = 0
max_attempts = 5
base_delay = 0.1  # 100ms

while attempts < max_attempts:
    try:
        return await make_request()
    except TransientError as e:
        attempts += 1
        if attempts == max_attempts:
            raise
        delay = base_delay * (2 ** attempts) + random.uniform(0, 0.1)
        await asyncio.sleep(delay)
```

**2. Retry with Jitter:**
```
Why: Prevents synchronized retries (thundering herd)

Full Jitter:
delay = random.uniform(0, base_delay * (2 ** attempt))

Decorrelated Jitter:
delay = min(cap, random.uniform(base, previous_delay * 3))

Recommended: Decorrelated jitter for production systems
```

**3. Idempotency Keys:**
```
Problem: Retries can cause duplicate operations

Solution: Idempotency keys
POST /api/v1/payments
Headers:
  Idempotency-Key: uuid-12345

Server Logic:
1. Check if operation with this key already processed
2. If yes, return cached response
3. If no, process and cache result
4. Cache for 24 hours

Ensures safe retries even for non-idempotent operations
```

**Retry Best Practices:**
```
DO:
✓ Only retry transient errors (timeout, 503, 429)
✓ Use exponential backoff with jitter
✓ Set maximum retry attempts (3-5)
✓ Implement overall timeout
✓ Use idempotency keys for writes
✓ Log each retry attempt

DON'T:
✗ Retry client errors (400, 401, 404)
✗ Retry without backoff (causes load spikes)
✗ Infinite retries
✗ Retry non-idempotent operations without safeguards
```

### Bulkhead Pattern

**Purpose:** Isolate resources to prevent total system failure.

**Thread Pool Isolation:**
```
Concept: Separate thread pools for different operations

Example:
- Payment Service Thread Pool: 20 threads
- Inventory Service Thread Pool: 20 threads
- Notification Service Thread Pool: 10 threads

If payment service becomes slow:
- Only payment thread pool exhausted
- Inventory and notification still work
- System partially degraded, not completely down
```

**Connection Pool Isolation:**
```
Database Connection Pools:
- Read-only queries: 50 connections
- Write queries: 20 connections
- Reporting queries: 10 connections

Heavy reporting query won't starve transactional operations
```

**Rate Limiting per Tenant:**
```
Multi-tenant SaaS application:

tenant-a: 1000 requests/minute
tenant-b: 1000 requests/minute
tenant-c: 1000 requests/minute

If tenant-a floods the system:
- Only tenant-a throttled
- tenant-b and tenant-c unaffected
```

**Implementation:**
```python
# Using semaphores for concurrency limits
class BulkheadExecutor:
    def __init__(self):
        self.payment_semaphore = asyncio.Semaphore(20)
        self.inventory_semaphore = asyncio.Semaphore(20)
        self.notification_semaphore = asyncio.Semaphore(10)

    async def call_payment_service(self, data):
        async with self.payment_semaphore:
            return await payment_service.call(data)

    async def call_inventory_service(self, data):
        async with self.inventory_semaphore:
            return await inventory_service.call(data)
```

### Timeout Pattern

**Purpose:** Prevent indefinite waiting for responses.

**Timeout Types:**

**1. Connection Timeout:**
```
Time allowed to establish connection

Recommended: 2-5 seconds
If takes longer, network likely has issues

httpx.AsyncClient(timeout=httpx.Timeout(connect=3.0))
```

**2. Read Timeout:**
```
Time allowed to receive response after connection

Varies by service:
- Fast APIs: 5 seconds
- Database queries: 10 seconds
- Complex processing: 30 seconds

httpx.AsyncClient(timeout=httpx.Timeout(read=10.0))
```

**3. Total Timeout:**
```
Overall time budget for entire operation

Example: User checkout flow
- Total budget: 30 seconds
- Payment service: 10 seconds
- Inventory check: 5 seconds
- Order creation: 5 seconds
- Buffer: 10 seconds

async with asyncio.timeout(30):
    result = await complete_checkout()
```

**Timeout Best Practices:**
```
Timeouts Hierarchy:
Parent timeout > sum of child timeouts

Request → API Gateway (30s timeout)
  → Service A (10s timeout)
    → Service B (5s timeout)
      → Database (2s timeout)

Set timeouts everywhere:
✓ HTTP clients
✓ Database connections
✓ Message consumers
✓ gRPC calls
✓ Cache operations
```

## Distributed Transaction Patterns

### Saga Pattern

**Purpose:** Manage distributed transactions across services.

**Choreography-Based Saga:**
```
Example: Order Creation Saga

Events:
1. OrderService: order.created
2. PaymentService: payment.completed OR payment.failed
3. InventoryService: inventory.reserved OR inventory.reservation.failed
4. ShippingService: shipment.created

Compensating Transactions:
If inventory.reservation.failed:
  → PaymentService listens → refund.initiated
  → OrderService listens → order.cancelled

Pros:
- Decentralized
- No single point of failure
- Services autonomous

Cons:
- Difficult to track saga state
- Complex debugging
- No saga-wide timeout
```

**Orchestration-Based Saga:**
```
Example: Order Saga Orchestrator

Saga Steps:
1. Create Order (OrderService)
2. Charge Payment (PaymentService)
3. Reserve Inventory (InventoryService)
4. Create Shipment (ShippingService)

Orchestrator Logic:
step1_result = await order_service.create_order()
if not step1_result.success:
    return failure("Order creation failed")

step2_result = await payment_service.charge(amount)
if not step2_result.success:
    await order_service.cancel_order(step1_result.order_id)
    return failure("Payment failed")

step3_result = await inventory_service.reserve(items)
if not step3_result.success:
    await payment_service.refund(step2_result.payment_id)
    await order_service.cancel_order(step1_result.order_id)
    return failure("Inventory unavailable")

# Continue saga...

Pros:
- Clear workflow
- Centralized monitoring
- Easy to understand

Cons:
- Orchestrator complexity
- Potential bottleneck
- Coupling to orchestrator
```

**Saga State Management:**
```
Persist saga state to handle failures:

CREATE TABLE saga_instances (
    saga_id UUID PRIMARY KEY,
    saga_type VARCHAR(50),
    current_step VARCHAR(50),
    status VARCHAR(20),
    payload JSONB,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

On orchestrator restart:
- Load incomplete sagas
- Resume from last completed step
- Execute remaining steps or compensations
```

### Event Sourcing

**Purpose:** Store all state changes as events, derive current state by replaying.

**Implementation:**
```
Traditional Approach:
UPDATE orders SET status = 'shipped' WHERE id = 123;
(Lost: when shipped, by whom, from where)

Event Sourcing Approach:
Events:
1. OrderPlaced { orderId, customerId, items, timestamp }
2. PaymentReceived { orderId, amount, paymentId, timestamp }
3. OrderShipped { orderId, trackingNumber, carrier, timestamp }

Current state = replay all events
```

**Event Store:**
```
CREATE TABLE events (
    event_id UUID PRIMARY KEY,
    aggregate_id UUID,
    aggregate_type VARCHAR(50),
    event_type VARCHAR(100),
    event_data JSONB,
    version INTEGER,
    timestamp TIMESTAMP,
    correlation_id UUID
);

CREATE INDEX idx_aggregate ON events(aggregate_id, version);

Guarantees:
- Events immutable
- Events ordered by version
- Optimistic locking prevents conflicts
```

**Benefits:**
```
✓ Full audit trail
✓ Time travel (replay to any point)
✓ Event replay for debugging
✓ Multiple read models from same events
✓ Temporal queries ("show orders as of yesterday")

Challenges:
✗ Eventual consistency
✗ Event schema evolution
✗ Snapshot strategy needed
✗ Increased storage
```

### CQRS (Command Query Responsibility Segregation)

**Purpose:** Separate read and write models for different optimization strategies.

**Architecture:**
```
Write Side (Command):
- Receives commands (CreateOrder, UpdateInventory)
- Validates business rules
- Stores events in event store
- Optimized for consistency and writes

Read Side (Query):
- Listens to events
- Updates denormalized read models
- Optimized for queries
- Eventual consistency

Example:
Command: CreateOrder
  → Order aggregate validates
  → Publishes OrderCreated event
  → Event stored in event store

Query Side:
  → Listens to OrderCreated
  → Updates order_summary table (denormalized)
  → Updates customer_order_history (different view)
  → Updates order_analytics (aggregated metrics)
```

**Read Models:**
```
Multiple specialized views from same events:

1. Order Detail View (for customer):
   { orderId, items, status, total, estimatedDelivery }

2. Order List View (for admin):
   { orderId, customerName, orderDate, status, total }

3. Analytics View:
   { date, totalOrders, totalRevenue, averageOrderValue }

Each optimized for specific query patterns
```

## Fault Tolerance Patterns

### Health Checks

**Types:**

**1. Liveness Probe:**
```
Purpose: Is the service alive?

Endpoint: GET /health/live

Returns 200 if:
- Application process running
- Not deadlocked

Kubernetes Action:
- If fails: Restart container
```

**2. Readiness Probe:**
```
Purpose: Is the service ready to receive traffic?

Endpoint: GET /health/ready

Returns 200 if:
- Database connection pool healthy
- Cache accessible
- Downstream dependencies responsive

Kubernetes Action:
- If fails: Remove from load balancer
- Don't send traffic until ready
```

**3. Startup Probe:**
```
Purpose: Has the service finished initialization?

Endpoint: GET /health/startup

For slow-starting applications:
- Prevents premature liveness checks
- Allows longer startup time
```

**Implementation:**
```python
@app.get("/health/live")
async def liveness():
    return {"status": "alive"}

@app.get("/health/ready")
async def readiness():
    checks = {
        "database": await check_database(),
        "cache": await check_cache(),
        "payment_service": await check_payment_service()
    }

    all_healthy = all(checks.values())
    status_code = 200 if all_healthy else 503

    return JSONResponse(
        status_code=status_code,
        content={"status": "ready" if all_healthy else "not ready", "checks": checks}
    )
```

### Graceful Degradation

**Purpose:** Provide reduced functionality when dependencies fail.

**Strategies:**

**1. Cached Responses:**
```
async def get_product_recommendations(user_id):
    try:
        async with circuit_breaker:
            return await ml_service.get_recommendations(user_id)
    except ServiceUnavailable:
        # Fallback to cached popular products
        return await cache.get_popular_products()
```

**2. Default Values:**
```
async def get_user_preferences(user_id):
    try:
        return await preferences_service.get(user_id)
    except ServiceUnavailable:
        # Return sensible defaults
        return {
            "language": "en",
            "currency": "USD",
            "theme": "light"
        }
```

**3. Feature Toggles:**
```
if feature_flags.is_enabled("personalized_recommendations"):
    recommendations = await ml_service.get_recommendations()
else:
    # Fallback to simple algorithm
    recommendations = await get_popular_products()
```

## Summary

Resilience patterns are mandatory in distributed systems. Layer multiple patterns for defense in depth:

**Essential Stack:**
1. Timeouts (prevent hanging)
2. Retries with backoff (handle transient errors)
3. Circuit breakers (prevent cascading failures)
4. Bulkheads (isolate failures)
5. Health checks (enable auto-healing)
6. Graceful degradation (maintain partial functionality)

**Choose Saga Pattern When:**
- Distributed transaction needed
- Strong consistency not required
- Compensating transactions possible

**Choose Event Sourcing When:**
- Full audit trail required
- Temporal queries needed
- Multiple read models beneficial

Always test failure scenarios. Use chaos engineering to validate resilience.
