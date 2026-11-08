# fullstack-guardian

> Источник: https://github.com/jeffallan/claude-skills/tree/main/skills/fullstack-guardian (MIT). Рабочая копия скилла: `.claude/skills/fullstack-guardian/`.

---
name: fullstack-guardian
description: Builds security-focused full-stack web applications by implementing integrated frontend and backend components with layered security at every level. Covers the complete stack from database to UI, enforcing auth, input validation, output encoding, and parameterized queries across all layers. Use when implementing features across frontend and backend, building REST APIs with corresponding UI, connecting frontend components to backend endpoints, creating end-to-end data flows from database to UI, or implementing CRUD operations with UI forms. Distinct from frontend-only, backend-only, or API-only skills in that it simultaneously addresses all three perspectives—Frontend, Backend, and Security—within a single implementation workflow. Invoke for full-stack feature work, web app development, authenticated API routes with views, microservices, real-time features, monorepo architecture, or technology selection decisions.
license: MIT
metadata:
  author: https://github.com/Jeffallan
  company: https://synergetic.solutions
  version: "1.1.1"
  domain: security
  triggers: fullstack, implement feature, build feature, create API, frontend and backend, full stack, new feature, implement, microservices, websocket, real-time, deployment pipeline, monorepo, architecture decision, technology selection, end-to-end
  role: expert
  scope: implementation
  output-format: code
  related-skills: feature-forge, test-master, devops-engineer, secure-code-guardian, architecture-designer, react-expert, typescript-pro
---

# Fullstack Guardian

Security-focused full-stack developer implementing features across the entire application stack.

## Core Workflow

1. **Gather requirements** - Understand feature scope and acceptance criteria
2. **Design solution** - Consider all three perspectives (Frontend/Backend/Security)
3. **Write technical design** - Document approach in `specs/{feature}_design.md`
4. **Security checkpoint** - Run through `references/security-checklist.md` before writing any code; confirm auth, authz, validation, and output encoding are addressed
5. **Implement** - Build incrementally, testing each component as you go
6. **Hand off** - Pass to Test Master for QA, DevOps for deployment

## Reference Guide

Load detailed guidance based on context:

| Topic | Reference | Load When |
|-------|-----------|-----------|
| Design Template | `references/design-template.md` | Starting feature, three-perspective design |
| Security Checklist | `references/security-checklist.md` | Every feature - auth, authz, validation |
| Error Handling | `references/error-handling.md` | Implementing error flows |
| Common Patterns | `references/common-patterns.md` | CRUD, forms, API flows |
| Backend Patterns | `references/backend-patterns.md` | Microservices, queues, observability, Docker |
| Frontend Patterns | `references/frontend-patterns.md` | Real-time, optimization, accessibility, testing |
| Integration Patterns | `references/integration-patterns.md` | Type sharing, deployment, architecture decisions |
| API Design | `references/api-design-standards.md` | REST/GraphQL APIs, versioning, CORS, validation |
| Architecture Decisions | `references/architecture-decisions.md` | Tech selection, monolith vs microservices |
| Deliverables Checklist | `references/deliverables-checklist.md` | Completing features, preparing handoff |

## Constraints

### MUST DO
- Address all three perspectives (Frontend, Backend, Security)
- Validate input on both client and server
- Use parameterized queries (prevent SQL injection)
- Sanitize output (prevent XSS)
- Implement proper error handling at every layer
- Log security-relevant events
- Write the implementation plan before coding
- Test each component as you build

### MUST NOT DO
- Skip security considerations
- Trust client-side validation alone
- Expose sensitive data in API responses
- Hardcode credentials or secrets
- Implement features without acceptance criteria
- Skip error handling for "happy path only"

## Three-Perspective Example

A minimal authenticated endpoint illustrating all three layers:

**[Backend]** — Authenticated route with parameterized query and scoped response:
```python
@router.get("/users/{user_id}/profile", dependencies=[Depends(require_auth)])
async def get_profile(user_id: int, current_user: User = Depends(get_current_user)):
    if current_user.id != user_id:
        raise HTTPException(status_code=403, detail="Forbidden")
    # Parameterized query — no raw string interpolation
    row = await db.fetchone("SELECT id, name, email FROM users WHERE id = ?", (user_id,))
    if not row:
        raise HTTPException(status_code=404, detail="Not found")
    return ProfileResponse(**row)   # explicit schema — no password/token leakage
```

**[Frontend]** — Component calls the endpoint and handles errors gracefully:
```typescript
async function fetchProfile(userId: number): Promise<Profile> {
  const res = await apiFetch(`/users/${userId}/profile`);   // apiFetch attaches auth header
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}
// Client-side input guard (never the only guard)
if (!Number.isInteger(userId) || userId <= 0) throw new Error("Invalid user ID");
```

**[Security]**
- Auth enforced server-side via `require_auth` dependency; client header is a convenience, not the gate.
- Response schema (`ProfileResponse`) explicitly excludes sensitive fields.
- 403 returned before any DB access when IDs don't match — no timing leak via 404.

## Output Templates

When implementing features, provide:
1. Technical design document (if non-trivial)
2. Backend code (models, schemas, endpoints)
3. Frontend code (components, hooks, API calls)
4. Brief security notes

Maintained by [@jeffallan](https://github.com/jeffallan), Principal Consultant at [Synergetic Solutions](https://synergetic.solutions)

[Documentation](https://jeffallan.github.io/claude-skills/skills/security/fullstack-guardian/)

---

## Reference: api-design-standards

# API Design Standards

## RESTful API Conventions

### URL Structure
```
# Collection vs Resource
GET    /api/users          # List all users
POST   /api/users          # Create user
GET    /api/users/:id      # Get single user
PUT    /api/users/:id      # Full update
PATCH  /api/users/:id      # Partial update
DELETE /api/users/:id      # Delete user

# Nested resources
GET    /api/users/:id/posts        # User's posts
POST   /api/users/:id/posts        # Create post for user
GET    /api/posts/:id/comments     # Comments on post
```

### HTTP Status Codes
```typescript
// Success codes
200 OK              // GET, PUT, PATCH successful
201 Created         // POST successful, resource created
204 No Content      // DELETE successful, no body
202 Accepted        // Async operation queued

// Client error codes
400 Bad Request     // Malformed request
401 Unauthorized    // Authentication required
403 Forbidden       // Authenticated but not authorized
404 Not Found       // Resource doesn't exist
409 Conflict        // Resource conflict (e.g., duplicate)
422 Unprocessable   // Validation failed
429 Too Many Requests // Rate limit exceeded

// Server error codes
500 Internal Server Error  // Unhandled exception
502 Bad Gateway           // Upstream service failed
503 Service Unavailable   // Temporary downtime
```

### Standardized Error Responses
```typescript
interface ApiError {
  error: {
    code: string;           // Machine-readable error code
    message: string;        // Human-readable message
    details?: {             // Field-level validation errors
      [field: string]: string[];
    };
    requestId: string;      // For support/debugging
    timestamp: string;      // ISO 8601 timestamp
  };
}

// Examples
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": {
      "email": ["Must be a valid email address"],
      "password": ["Must be at least 12 characters"]
    },
    "requestId": "req_abc123",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}

{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "User not found",
    "requestId": "req_def456",
    "timestamp": "2025-01-15T10:31:00Z"
  }
}
```

### Pagination
```typescript
// Query parameters
GET /api/users?page=1&limit=20&sort=-createdAt&filter[role]=admin

// Response format
interface PaginatedResponse<T> {
  data: T[];
  meta: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
  links?: {
    first: string;
    prev?: string;
    next?: string;
    last: string;
  };
}

// Implementation
@Get()
async findAll(
  @Query('page', new DefaultValuePipe(1), ParseIntPipe) page: number,
  @Query('limit', new DefaultValuePipe(20), ParseIntPipe) limit: number,
) {
  const [data, total] = await this.service.findAndCount({ page, limit });
  return {
    data,
    meta: {
      page,
      limit,
      total,
      totalPages: Math.ceil(total / limit),
    },
    links: {
      first: `/api/users?page=1&limit=${limit}`,
      next: page < totalPages ? `/api/users?page=${page + 1}&limit=${limit}` : undefined,
      last: `/api/users?page=${totalPages}&limit=${limit}`,
    },
  };
}
```

## API Versioning

### URL Path Versioning (Recommended)
```typescript
// Version in URL path
GET /api/v1/users
GET /api/v2/users

// Express routing
app.use('/api/v1', v1Router);
app.use('/api/v2', v2Router);

// NestJS versioning
@Controller({ version: '1', path: 'users' })
export class UsersV1Controller {}

@Controller({ version: '2', path: 'users' })
export class UsersV2Controller {}
```

### Header Versioning (Alternative)
```typescript
// Request header
GET /api/users
Accept-Version: v2

// Middleware
app.use((req, res, next) => {
  const version = req.headers['accept-version'] || 'v1';
  req.apiVersion = version;
  next();
});
```

## Rate Limiting

### Per-Endpoint Configuration
```typescript
// Express with express-rate-limit
import rateLimit from 'express-rate-limit';

const generalLimiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100,                 // 100 requests per window
  message: 'Too many requests from this IP',
});

const authLimiter = rateLimit({
  windowMs: 15 * 60 * 1000,
  max: 5,                   // Stricter for auth endpoints
  skipSuccessfulRequests: true,
});

app.use('/api/', generalLimiter);
app.use('/api/auth/', authLimiter);
```

### Redis-backed Rate Limiting
```typescript
import { RateLimiterRedis } from 'rate-limiter-flexible';

const rateLimiter = new RateLimiterRedis({
  storeClient: redisClient,
  keyPrefix: 'rate-limit',
  points: 100,              // Number of requests
  duration: 60,             // Per 60 seconds
});

app.use(async (req, res, next) => {
  try {
    await rateLimiter.consume(req.ip);
    next();
  } catch (error) {
    res.status(429).json({ error: 'Too Many Requests' });
  }
});
```

## CORS Configuration

### Production-ready CORS
```typescript
import cors from 'cors';

const corsOptions = {
  origin: (origin, callback) => {
    const allowedOrigins = [
      'https://app.example.com',
      'https://admin.example.com',
    ];

    if (!origin || allowedOrigins.includes(origin)) {
      callback(null, true);
    } else {
      callback(new Error('Not allowed by CORS'));
    }
  },
  credentials: true,                    // Allow cookies
  methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'],
  allowedHeaders: ['Content-Type', 'Authorization'],
  exposedHeaders: ['X-Total-Count'],
  maxAge: 86400,                        // 24 hours preflight cache
};

app.use(cors(corsOptions));
```

## Request/Response Validation

### Input Validation with Zod
```typescript
import { z } from 'zod';

const createUserSchema = z.object({
  email: z.string().email(),
  name: z.string().min(1).max(100),
  age: z.number().int().min(18).max(120).optional(),
  role: z.enum(['user', 'admin']).default('user'),
});

// Middleware
const validate = (schema: z.ZodSchema) => (req, res, next) => {
  try {
    req.validatedBody = schema.parse(req.body);
    next();
  } catch (error) {
    res.status(422).json({
      error: {
        code: 'VALIDATION_ERROR',
        message: 'Invalid request data',
        details: error.errors,
      },
    });
  }
};

app.post('/api/users', validate(createUserSchema), createUserHandler);
```

## API Documentation

### OpenAPI/Swagger Setup
```typescript
// NestJS with Swagger
import { DocumentBuilder, SwaggerModule } from '@nestjs/swagger';

const config = new DocumentBuilder()
  .setTitle('API Documentation')
  .setDescription('The API description')
  .setVersion('1.0')
  .addBearerAuth()
  .build();

const document = SwaggerModule.createDocument(app, config);
SwaggerModule.setup('api/docs', app, document);

// Decorate endpoints
@ApiOperation({ summary: 'Create a new user' })
@ApiResponse({ status: 201, description: 'User created successfully' })
@ApiResponse({ status: 422, description: 'Validation failed' })
@Post()
async create(@Body() dto: CreateUserDto) {
  return this.service.create(dto);
}
```

## Quick Reference

| Aspect | Standard | Example |
|--------|----------|---------|
| URL naming | Plural nouns | `/api/users` not `/api/user` |
| HTTP methods | RESTful semantics | GET (read), POST (create), PUT/PATCH (update), DELETE |
| Status codes | Semantic usage | 200 (success), 201 (created), 422 (validation) |
| Errors | Consistent format | `{ error: { code, message, details } }` |
| Pagination | Meta + links | `{ data, meta: { page, total }, links }` |
| Versioning | URL path | `/api/v1/users` |
| Rate limiting | Per-endpoint | Auth: 5/min, General: 100/15min |
| CORS | Whitelist origins | Production domains only |
| Validation | Schema-based | Zod/Pydantic with detailed errors |
| Documentation | OpenAPI | Auto-generated from decorators |

---

## Reference: architecture-decisions

# Architecture Decision Guide

## Technology Selection Matrix

### Backend Framework Selection

| Framework | Best For | Pros | Cons |
|-----------|----------|------|------|
| **NestJS** | Enterprise apps, microservices | TypeScript-first, dependency injection, excellent docs | Opinionated, steeper learning curve |
| **Express** | Simple APIs, flexibility | Minimal, huge ecosystem, well-known | Manual structure, less opinionated |
| **Fastify** | High performance APIs | Fast, schema validation, plugins | Smaller ecosystem than Express |
| **FastAPI** | Python APIs, ML integration | Auto-docs, type hints, fast | Python ecosystem only |
| **Go/Gin** | High-performance services | Compiled, concurrent, fast | Verbose, less rapid development |

**Decision criteria:**
- Team expertise: Choose familiar stack
- Performance needs: Go/Fastify for high throughput
- Type safety: NestJS/FastAPI for TypeScript/Python
- Flexibility: Express for custom architectures

### Frontend Framework Selection

| Framework | Best For | Pros | Cons |
|-----------|----------|------|------|
| **React** | Most use cases, large apps | Huge ecosystem, flexible, well-supported | Not batteries-included, decision fatigue |
| **Vue** | Progressive enhancement | Gentle learning curve, good docs, reactive | Smaller ecosystem than React |
| **Angular** | Enterprise apps | Complete framework, TypeScript native | Heavy, opinionated, steep curve |
| **Svelte** | Performance-critical apps | Compiled, no virtual DOM, small bundle | Smaller ecosystem, fewer resources |
| **Next.js** | SSR/SSG apps, SEO | React + routing + SSR, excellent DX | Vercel-centric, complexity for simple apps |

**Decision criteria:**
- SEO requirements: Next.js/Nuxt for SSR
- Team size: Angular for large teams, Vue for small
- Ecosystem: React for maximum third-party support
- Performance: Svelte for minimal bundle size

### Database Selection

| Database | Best For | Pros | Cons |
|----------|----------|------|------|
| **PostgreSQL** | Relational data, ACID | Feature-rich, reliable, JSON support | Complex queries can be slow |
| **MySQL** | Read-heavy workloads | Mature, fast reads, replication | Less feature-rich than Postgres |
| **MongoDB** | Flexible schemas, rapid dev | Schema-less, horizontal scaling | No transactions (old versions) |
| **Redis** | Caching, sessions, queues | Extremely fast, versatile | In-memory only, data structures limited |
| **DynamoDB** | AWS serverless, high scale | Managed, predictable performance | Vendor lock-in, query limitations |

**Decision criteria:**
- ACID requirements: PostgreSQL/MySQL
- Flexible schemas: MongoDB
- Caching layer: Redis (always)
- AWS serverless: DynamoDB
- Default choice: PostgreSQL (most versatile)

### State Management (Frontend)

| Solution | Best For | Complexity | Bundle Size |
|----------|----------|------------|-------------|
| **React Context** | Simple state, few updates | Low | None (built-in) |
| **Zustand** | Medium apps, simplicity | Low | 1KB |
| **Redux Toolkit** | Complex state, time-travel debug | Medium | 15KB |
| **Jotai/Recoil** | Atomic state, derived state | Medium | 3KB |
| **MobX** | Observable state, OOP style | Medium | 16KB |
| **TanStack Query** | Server state only | Low | 12KB |

**Decision criteria:**
- Simple app: Context or Zustand
- Complex state logic: Redux Toolkit
- Server state: TanStack Query (don't use global state)
- Real-time apps: Zustand + WebSocket

## Monolith vs Microservices

### Decision Matrix

| Factor | Monolith | Microservices |
|--------|----------|---------------|
| **Team size** | < 10 developers | > 10 developers |
| **System complexity** | Simple domain | Complex, bounded contexts |
| **Deployment** | Simple, all-at-once | Complex, independent services |
| **Scaling** | Vertical scaling | Horizontal per service |
| **Development speed** | Fast initially | Slower setup, faster iteration |
| **Infrastructure** | Simpler (1 app, 1 DB) | Complex (K8s, service mesh, multiple DBs) |
| **Data consistency** | ACID transactions | Eventual consistency, sagas |
| **Testing** | Easier integration tests | More complex testing |
| **Monitoring** | Single app to monitor | Distributed tracing needed |

### When to Use Monolith
```
✓ Starting new product (validate idea first)
✓ Small team (< 10 developers)
✓ Simple domain with few bounded contexts
✓ Need rapid development
✓ Limited infrastructure budget
✓ Straightforward deployment requirements
```

### When to Use Microservices
```
✓ Large team (> 10 developers)
✓ Clear bounded contexts in domain
✓ Different services have different scaling needs
✓ Need independent deployment cycles
✓ Multiple teams working independently
✓ Polyglot requirements (different languages)
✓ Have DevOps expertise and infrastructure
```

### Modular Monolith (Recommended Middle Ground)
```typescript
// Structure monolith with clear boundaries
project/
├── src/
│   ├── modules/
│   │   ├── users/
│   │   │   ├── users.module.ts
│   │   │   ├── users.service.ts
│   │   │   ├── users.controller.ts
│   │   │   └── users.repository.ts
│   │   ├── orders/
│   │   │   ├── orders.module.ts
│   │   │   └── ...
│   │   └── payments/
│   │       └── ...
│   └── shared/
│       ├── database/
│       └── auth/

// Clear module boundaries, can split later if needed
```

## API Architecture Patterns

### REST vs GraphQL

| Aspect | REST | GraphQL |
|--------|------|---------|
| **Best for** | CRUD operations, public APIs | Complex queries, mobile apps |
| **Learning curve** | Low | Medium-high |
| **Over-fetching** | Common issue | Solved by design |
| **Under-fetching** | Requires multiple requests | Single request |
| **Caching** | HTTP caching works well | More complex caching |
| **Versioning** | URL versioning (/v1, /v2) | Schema evolution |
| **Tooling** | Swagger, Postman | GraphiQL, Apollo Studio |

**Choose REST when:**
- Building simple CRUD APIs
- Need HTTP caching
- Public API with many consumers
- Team unfamiliar with GraphQL

**Choose GraphQL when:**
- Mobile apps need flexible queries
- Complex data requirements
- Rapid frontend iteration
- Real-time subscriptions needed

### BFF Pattern (Backend for Frontend)

```typescript
// Use when frontend needs differ from backend APIs
// Mobile BFF: Returns minimal data, optimized responses
@Controller('mobile-bff')
export class MobileBFFController {
  @Get('dashboard')
  async getMobileDashboard(@CurrentUser() user: User) {
    const [profile, notifications] = await Promise.all([
      this.userService.getProfile(user.id),
      this.notificationService.getUnread(user.id, 5), // Only 5 for mobile
    ]);
    return { profile, notifications }; // Minimal payload
  }
}

// Web BFF: Returns richer data
@Controller('web-bff')
export class WebBFFController {
  @Get('dashboard')
  async getWebDashboard(@CurrentUser() user: User) {
    const [profile, notifications, analytics, recentActivity] = await Promise.all([
      this.userService.getProfile(user.id),
      this.notificationService.getUnread(user.id, 20), // More for web
      this.analyticsService.getUserStats(user.id),
      this.activityService.getRecent(user.id),
    ]);
    return { profile, notifications, analytics, recentActivity };
  }
}
```

## Authentication Strategy

### JWT vs Session-based

| Aspect | JWT | Session |
|--------|-----|---------|
| **Scalability** | Stateless, horizontal scaling | Requires session store |
| **Performance** | No DB lookup per request | DB/Redis lookup needed |
| **Revocation** | Complex (requires blacklist) | Simple (delete session) |
| **Security** | Token can't be invalidated | Easy to invalidate |
| **Mobile/SPA** | Ideal for token storage | Requires cookies |
| **Microservices** | Easy to share across services | Harder to share |

**Hybrid approach (Recommended):**
```typescript
// Short-lived access token (15min) + refresh token (7 days)
interface AuthTokens {
  accessToken: string;   // JWT, 15 minutes, stored in memory
  refreshToken: string;  // Opaque token, 7 days, httpOnly cookie
}

// Access token: Stateless, fast validation
// Refresh token: Stored in DB, can be revoked
```

### SSO Integration Options

| Provider | Use Case | Complexity |
|----------|----------|------------|
| **OAuth2/OIDC** | Standard protocol, most IdPs | Medium |
| **SAML** | Enterprise customers, legacy | High |
| **Social logins** | B2C apps (Google, GitHub) | Low |
| **Auth0/Okta** | Managed solution, rapid setup | Low |

## Caching Strategy

### Layered Caching Approach

```typescript
// Layer 1: CDN caching (static assets)
// CloudFront, Cloudflare

// Layer 2: API response caching (Redis)
const cacheKey = `user:${userId}:profile`;
let profile = await redis.get(cacheKey);

if (!profile) {
  profile = await db.users.findById(userId);
  await redis.setex(cacheKey, 300, JSON.stringify(profile)); // 5 min TTL
}

// Layer 3: Database query caching
// PostgreSQL prepared statements, query plan caching

// Layer 4: Application-level caching
const userCache = new LRU({ max: 1000 });
```

### Cache Invalidation Patterns

```typescript
// Write-through: Update cache on write
async updateUser(id: string, data: UpdateUserDto) {
  const user = await db.users.update(id, data);
  await redis.set(`user:${id}`, JSON.stringify(user), 'EX', 300);
  return user;
}

// Write-behind: Invalidate cache, lazy load
async updateUser(id: string, data: UpdateUserDto) {
  const user = await db.users.update(id, data);
  await redis.del(`user:${id}`); // Delete, will reload on next read
  return user;
}

// Event-based: Invalidate related caches
eventBus.on('user.updated', async ({ userId }) => {
  await Promise.all([
    redis.del(`user:${userId}`),
    redis.del(`user:${userId}:posts`),
    redis.del(`user:${userId}:followers`),
  ]);
});
```

## Deployment Strategy

### Environment Progression

```
Development → Staging → Production

Development:
- Local dev servers
- Docker Compose for dependencies
- Hot reload enabled
- Debug logging
- Relaxed security

Staging:
- Production-like environment
- Real integrations (test mode)
- E2E tests run here
- Performance testing
- Security scanning

Production:
- High availability setup
- Blue-green deployment
- Monitoring & alerting
- Automated rollback
- Strict security
```

### Deployment Patterns

| Pattern | Downtime | Rollback | Complexity | Use When |
|---------|----------|----------|------------|----------|
| **Recreate** | Yes | Manual | Low | Dev/staging only |
| **Rolling** | No | Gradual | Medium | Standard deployments |
| **Blue-Green** | No | Instant | Medium | Zero-downtime required |
| **Canary** | No | Gradual | High | High-risk changes |
| **A/B Testing** | No | Gradual | High | Feature validation |

## Quick Decision Trees

### "Which database should I use?"
```
Need ACID transactions? → PostgreSQL
NoSQL with flexible schema? → MongoDB
Caching/sessions/queues? → Redis
AWS serverless? → DynamoDB
High read throughput? → PostgreSQL + read replicas
```

### "Monolith or microservices?"
```
New product? → Modular monolith
Team < 10 people? → Modular monolith
Clear bounded contexts? → Consider microservices
Different scaling needs? → Microservices
Limited DevOps resources? → Monolith
```

### "REST or GraphQL?"
```
Simple CRUD? → REST
Mobile app with flexible queries? → GraphQL
Public API? → REST
Complex data requirements? → GraphQL
Team knows GraphQL? → GraphQL, otherwise REST
```

### "Which state management?"
```
Simple app, few global state? → React Context
Server state (API data)? → TanStack Query
Medium complexity? → Zustand
Complex state logic? → Redux Toolkit
Real-time updates? → Zustand + WebSocket
```

---

## Reference: backend-patterns

# Backend Patterns

## Microservices Architecture

### Circuit Breaker Pattern
```typescript
class CircuitBreaker {
  private failures = 0;
  private threshold = 5;
  private state: 'CLOSED' | 'OPEN' | 'HALF_OPEN' = 'CLOSED';

  async call<T>(fn: () => Promise<T>): Promise<T> {
    if (this.state === 'OPEN') throw new Error('Circuit breaker is OPEN');
    try {
      const result = await fn();
      this.failures = 0;
      this.state = 'CLOSED';
      return result;
    } catch (error) {
      this.failures++;
      if (this.failures >= this.threshold) {
        this.state = 'OPEN';
        setTimeout(() => this.state = 'HALF_OPEN', 60000);
      }
      throw error;
    }
  }
}
```

### Saga Pattern (Distributed Transactions)
```typescript
class OrderSaga {
  async execute(order: Order) {
    const compensations: (() => Promise<void>)[] = [];
    try {
      await inventoryService.reserve(order.items);
      compensations.push(() => inventoryService.release(order.items));

      await paymentService.charge(order.amount);
      compensations.push(() => paymentService.refund(order.amount));

      return { success: true };
    } catch (error) {
      for (const compensate of compensations.reverse()) await compensate();
      throw error;
    }
  }
}
```

## Message Queue Integration

### Producer/Consumer with DLQ
```typescript
// RabbitMQ Consumer with Dead Letter Queue
class MessageConsumer {
  async consume(queue: string, handler: (msg: any) => Promise<void>) {
    const channel = await this.connection.createChannel();

    // Setup DLQ
    await channel.assertExchange('dlx', 'direct', { durable: true });
    await channel.assertQueue(`${queue}.dlq`, { durable: true });
    await channel.bindQueue(`${queue}.dlq`, 'dlx', queue);

    // Main queue
    await channel.assertQueue(queue, {
      durable: true,
      deadLetterExchange: 'dlx',
      deadLetterRoutingKey: queue,
    });

    channel.consume(queue, async (msg) => {
      if (!msg) return;
      try {
        await handler(JSON.parse(msg.content.toString()));
        channel.ack(msg);
      } catch (error) {
        const retryCount = (msg.properties.headers['x-retry-count'] || 0) + 1;
        if (retryCount >= 3) {
          channel.nack(msg, false, false); // Send to DLQ
        } else {
          setTimeout(() => channel.nack(msg, false, true), retryCount * 1000);
        }
      }
    });
  }
}
```

### Idempotency
```typescript
class IdempotentHandler {
  async handle(messageId: string, fn: () => Promise<void>) {
    const exists = await db.processedMessages.findOne({ messageId });
    if (exists) return; // Already processed

    await fn();
    await db.processedMessages.insert({ messageId, processedAt: new Date() });
  }
}
```

## Database Optimization

### Connection Pooling
```typescript
import { Pool } from 'pg';

const pool = new Pool({
  max: 20,
  min: 5,
  idleTimeoutMillis: 30000,
});

export async function query(sql: string, params: any[]) {
  const client = await pool.connect();
  try {
    return await client.query(sql, params);
  } finally {
    client.release();
  }
}
```

### Read Replica Strategy
```typescript
class DatabaseRouter {
  async query(sql: string, params: any[]) {
    const isWrite = /^(INSERT|UPDATE|DELETE)/i.test(sql);
    if (isWrite) return this.primary.query(sql, params);

    // Round-robin read replica
    const replica = this.replicas[Math.floor(Math.random() * this.replicas.length)];
    return replica.query(sql, params);
  }
}
```

## Monitoring & Observability

### Prometheus Metrics
```typescript
import { Counter, Histogram, Registry } from 'prom-client';

const register = new Registry();

const httpDuration = new Histogram({
  name: 'http_request_duration_seconds',
  labelNames: ['method', 'route', 'status_code'],
  registers: [register],
});

// Middleware
app.use((req, res, next) => {
  const start = Date.now();
  res.on('finish', () => {
    httpDuration.observe({
      method: req.method,
      route: req.route?.path,
      status_code: res.statusCode
    }, (Date.now() - start) / 1000);
  });
  next();
});

app.get('/metrics', async (req, res) => {
  res.set('Content-Type', register.contentType);
  res.end(await register.metrics());
});
```

### Distributed Tracing
```typescript
import { trace } from '@opentelemetry/api';

const tracer = trace.getTracer('my-service');

async function processOrder(orderId: string) {
  const span = tracer.startSpan('processOrder');
  span.setAttribute('orderId', orderId);

  try {
    await db.query('SELECT * FROM orders WHERE id = $1', [orderId]);
    span.addEvent('Order fetched');
    span.setStatus({ code: SpanStatusCode.OK });
  } catch (error) {
    span.recordException(error);
    throw error;
  } finally {
    span.end();
  }
}
```

## Docker & Deployment

### Multi-stage Dockerfile
```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production && npm run build

FROM node:18-alpine
WORKDIR /app
RUN adduser -S nodejs -u 1001
COPY --from=builder --chown=nodejs /app/dist ./dist
COPY --from=builder --chown=nodejs /app/node_modules ./node_modules
USER nodejs
EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=3s CMD node healthcheck.js
CMD ["node", "dist/main.js"]
```

### Graceful Shutdown
```typescript
process.on('SIGTERM', async () => {
  console.log('Shutting down gracefully');
  server.close(() => console.log('HTTP server closed'));
  await db.end();
  await messageQueue.close();
  process.exit(0);
});
```

## Quick Reference

| Pattern | Use Case | Key Benefit |
|---------|----------|-------------|
| Circuit Breaker | External service calls | Prevent cascade failures |
| Saga | Distributed transactions | Data consistency |
| Message Queue | Async processing | Decoupling & scalability |
| Connection Pool | Database access | Performance optimization |
| Read Replicas | High read load | Horizontal scaling |
| Distributed Tracing | Microservices debugging | End-to-end visibility |
| Graceful Shutdown | Container orchestration | Zero downtime deploys |

---

## Reference: common-patterns

# Common Patterns

## API + Frontend Flow

```
User Action → Frontend Validation → API Call → Backend Validation
→ Business Logic → Database → Response → UI Update
```

## CRUD Implementation

### Create

```typescript
// Frontend
const createUser = async (data: CreateUserDto) => {
  const response = await fetch('/api/users', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  });
  if (!response.ok) throw await response.json();
  return response.json();
};

// Backend (NestJS)
@Post()
async create(@Body() dto: CreateUserDto): Promise<User> {
  return this.userService.create(dto);
}
```

### Read (List with Pagination)

```typescript
// Frontend
const { data, isLoading } = useQuery({
  queryKey: ['users', page, limit],
  queryFn: () => fetch(`/api/users?page=${page}&limit=${limit}`).then(r => r.json()),
});

// Backend
@Get()
async findAll(
  @Query('page', new DefaultValuePipe(1), ParseIntPipe) page: number,
  @Query('limit', new DefaultValuePipe(20), ParseIntPipe) limit: number,
): Promise<PaginatedResponse<User>> {
  return this.userService.findAll({ page, limit });
}
```

### Update

```typescript
// Frontend with optimistic update
const updateUser = useMutation({
  mutationFn: (data: UpdateUserDto) => api.patch(`/users/${id}`, data),
  onMutate: async (newData) => {
    await queryClient.cancelQueries(['user', id]);
    const previous = queryClient.getQueryData(['user', id]);
    queryClient.setQueryData(['user', id], (old) => ({ ...old, ...newData }));
    return { previous };
  },
  onError: (err, newData, context) => {
    queryClient.setQueryData(['user', id], context.previous);
  },
});

// Backend
@Patch(':id')
async update(
  @Param('id') id: string,
  @Body() dto: UpdateUserDto,
): Promise<User> {
  return this.userService.update(id, dto);
}
```

### Delete

```typescript
// Frontend with confirmation
const handleDelete = async () => {
  if (!confirm('Are you sure?')) return;
  await api.delete(`/users/${id}`);
  router.push('/users');
};

// Backend (soft delete)
@Delete(':id')
@HttpCode(204)
async remove(@Param('id') id: string): Promise<void> {
  await this.userService.softDelete(id);
}
```

## Form Handling

```typescript
// React Hook Form + Zod
const schema = z.object({
  name: z.string().min(1, 'Required'),
  email: z.string().email('Invalid email'),
});

function UserForm({ onSubmit }: Props) {
  const { register, handleSubmit, formState: { errors } } = useForm({
    resolver: zodResolver(schema),
  });

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <input {...register('name')} />
      {errors.name && <span>{errors.name.message}</span>}

      <input {...register('email')} />
      {errors.email && <span>{errors.email.message}</span>}

      <button type="submit">Save</button>
    </form>
  );
}
```

## Quick Reference

| Pattern | Frontend | Backend |
|---------|----------|---------|
| Create | POST + form | Validate + insert |
| Read | GET + query | Paginate + filter |
| Update | PATCH + optimistic | Validate + update |
| Delete | DELETE + confirm | Soft delete |
| Auth | Token storage | JWT middleware |
| Upload | FormData | Multer/streaming |

---

## Reference: deliverables-checklist

# Deliverables Checklist

## Code Deliverables

### Backend Files
- [ ] API endpoint implementations
- [ ] Database models and schemas
- [ ] Validation schemas (Zod/Pydantic)
- [ ] Business logic services
- [ ] Middleware (auth, error handling, logging)
- [ ] Database migrations with rollback
- [ ] Environment configuration files
- [ ] Docker/container configuration

### Frontend Files
- [ ] Component files with TypeScript interfaces
- [ ] Custom hooks for data fetching
- [ ] State management setup (Redux/Zustand/Context)
- [ ] API client/service layer
- [ ] Form components with validation
- [ ] Error boundary components
- [ ] Routing configuration
- [ ] Style files (CSS/SCSS/styled-components)

### Shared/Integration Files
- [ ] Shared TypeScript types package
- [ ] Shared validation schemas
- [ ] API contract definitions
- [ ] Utility functions used across stack
- [ ] Configuration types
- [ ] Constants and enums

## Testing Deliverables

### Unit Tests
```typescript
// Backend: Service layer tests
describe('UserService', () => {
  it('should create user with hashed password', async () => {
    const user = await userService.create({
      email: 'test@example.com',
      password: '<test-password>',
    });
    expect(user.password).not.toBe('<test-password>');
    expect(user.email).toBe('test@example.com');
  });
});

// Frontend: Component tests
describe('UserForm', () => {
  it('should validate email format', async () => {
    render(<UserForm onSubmit={jest.fn()} />);
    await userEvent.type(screen.getByLabelText('Email'), 'invalid');
    await userEvent.click(screen.getByText('Submit'));
    expect(screen.getByText(/invalid email/i)).toBeInTheDocument();
  });
});
```

### Integration Tests
```typescript
// API endpoint tests
describe('POST /api/users', () => {
  it('should create user and return 201', async () => {
    const response = await request(app)
      .post('/api/users')
      .send({ email: 'new@example.com', password: '<test-password>' });

    expect(response.status).toBe(201);
    expect(response.body).toHaveProperty('id');
    expect(response.body.email).toBe('new@example.com');
  });

  it('should return 422 for duplicate email', async () => {
    await createUser({ email: 'existing@example.com' });

    const response = await request(app)
      .post('/api/users')
      .send({ email: 'existing@example.com', password: '<test-password>' });

    expect(response.status).toBe(422);
    expect(response.body.error.code).toBe('DUPLICATE_EMAIL');
  });
});
```

### E2E Tests
```typescript
// Playwright test
test('complete user registration flow', async ({ page }) => {
  await page.goto('/register');
  await page.fill('[name="email"]', 'newuser@example.com');
  await page.fill('[name="password"]', '<test-password>');
  await page.click('button[type="submit"]');

  await expect(page).toHaveURL('/dashboard');
  await expect(page.locator('[data-testid="welcome-message"]'))
    .toContainText('Welcome');
});
```

### Test Coverage Requirements
- [ ] Unit tests: >80% coverage
- [ ] Integration tests: All critical paths
- [ ] E2E tests: Main user journeys
- [ ] Performance tests: Load/stress scenarios
- [ ] Security tests: OWASP Top 10 validation

## Documentation Deliverables

### Technical Documentation
```markdown
# Feature: User Management API

## Overview
Complete CRUD API for user management with authentication and authorization.

## Endpoints

### Create User
POST /api/v1/users

Request:
{
  "email": "user@example.com",
  "name": "John Doe",
  "password": "<password>"
}

Response (201):
{
  "id": "usr_abc123",
  "email": "user@example.com",
  "name": "John Doe",
  "createdAt": "2025-01-15T10:00:00Z"
}

### Authentication
All endpoints except POST /users require Bearer token:
Authorization: Bearer <jwt_token>

### Error Responses
422 Validation Error:
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input",
    "details": { "email": ["Must be valid email"] }
  }
}
```

### Component Documentation
```typescript
/**
 * UserProfileForm - Editable user profile form with validation
 *
 * @example
 * <UserProfileForm
 *   initialData={currentUser}
 *   onSubmit={handleUpdate}
 *   onCancel={() => router.back()}
 * />
 *
 * @param initialData - User data to pre-populate form
 * @param onSubmit - Callback when form is submitted with valid data
 * @param onCancel - Optional callback when user cancels editing
 */
export function UserProfileForm({
  initialData,
  onSubmit,
  onCancel
}: UserProfileFormProps) {
  // Component implementation
}
```

### README Updates
- [ ] Installation instructions
- [ ] Environment variable configuration
- [ ] Development setup steps
- [ ] Build and deployment commands
- [ ] Testing instructions
- [ ] Troubleshooting guide

### Storybook Documentation (Frontend)
```typescript
// UserCard.stories.tsx
export default {
  title: 'Components/UserCard',
  component: UserCard,
} as Meta;

export const Default: Story = {
  args: {
    user: {
      name: 'John Doe',
      email: 'john@example.com',
      avatar: 'https://example.com/avatar.jpg',
    },
  },
};

export const Loading: Story = {
  args: { isLoading: true },
};

export const WithLongName: Story = {
  args: {
    user: {
      name: 'Johnathan Alexander Wellington III',
      email: 'johnathan@example.com',
    },
  },
};
```

## Performance Deliverables

### Metrics Report
```markdown
## Performance Metrics

### Backend API
- Average response time: 45ms
- P95 response time: 120ms
- P99 response time: 250ms
- Throughput: 1000 req/s
- Error rate: 0.02%

### Frontend Bundle
- Initial bundle size: 245 KB (gzipped)
- Largest chunk: 180 KB
- Time to Interactive: 1.2s
- Lighthouse score: 95/100

### Database Queries
- Average query time: 15ms
- Slowest query: 85ms (user search)
- Index usage: 98%
- Connection pool utilization: 60%
```

### Bundle Analysis
- [ ] Webpack/Vite bundle analysis report
- [ ] Lighthouse performance audit
- [ ] Core Web Vitals measurements
- [ ] Bundle size comparison (before/after)

## Security Deliverables

### Security Checklist
- [ ] Input validation on all endpoints
- [ ] Output sanitization (XSS prevention)
- [ ] SQL injection prevention (parameterized queries)
- [ ] CSRF protection enabled
- [ ] Rate limiting configured
- [ ] Authentication required where needed
- [ ] Authorization checks implemented
- [ ] Sensitive data excluded from responses
- [ ] Secrets in environment variables
- [ ] HTTPS enforced in production
- [ ] Security headers configured (CSP, HSTS, etc.)

### Security Audit Report
```markdown
## Security Review

### Authentication
- JWT with RS256 algorithm
- 15-minute access tokens
- 7-day refresh tokens
- Secure cookie storage

### Authorization
- Role-based access control (RBAC)
- Resource ownership validation
- Permission checks on all mutations

### Data Protection
- Passwords hashed with bcrypt (12 rounds)
- Sensitive data encrypted at rest
- PII excluded from logs
- Rate limiting: 100 req/15min per IP
```

## Deployment Deliverables

### Configuration Files
- [ ] `Dockerfile` with multi-stage build
- [ ] `docker-compose.yml` for local dev
- [ ] CI/CD pipeline configuration
- [ ] Environment-specific configs
- [ ] Database migration scripts
- [ ] Health check endpoints
- [ ] Kubernetes manifests (if applicable)

### Deployment Guide
```markdown
## Deployment Steps

### Prerequisites
- Node.js 18+
- PostgreSQL 15+
- Redis 7+

### Environment Variables
DATABASE_URL=postgresql://<db-user>:<db-password>@host:5432/dbname
REDIS_URL=redis://localhost:6379
JWT_SECRET=<generate-secure-secret>
API_PORT=3000

### Build & Deploy
npm run build
npm run migrate
npm run start:prod

### Health Check
GET /api/health
Expected: { "status": "ok", "database": "connected" }
```

## Handoff Checklist

### Before Handoff
- [ ] All tests passing
- [ ] Code reviewed and approved
- [ ] Documentation complete
- [ ] Performance validated
- [ ] Security reviewed
- [ ] Deployed to staging
- [ ] E2E tests pass in staging
- [ ] Accessibility audit complete

### Handoff Package
- [ ] Links to merged PRs
- [ ] Deployment instructions
- [ ] Database migration notes
- [ ] Known issues/limitations
- [ ] Monitoring dashboard URLs
- [ ] Rollback procedure
- [ ] Support contact information

## Quick Reference

| Category | Key Deliverables | Coverage Target |
|----------|-----------------|-----------------|
| Backend | API, models, migrations | 80% test coverage |
| Frontend | Components, hooks, routes | 85% test coverage |
| Tests | Unit, integration, E2E | All critical paths |
| Docs | API, components, setup | Complete |
| Performance | Metrics, bundle analysis | <200ms P95 API, <2s TTI |
| Security | Audit, OWASP validation | All vulnerabilities addressed |
| Deployment | Docker, CI/CD, guides | Zero-downtime capable |

---

## Reference: design-template

# Three-Perspective Design

## Design Template

For every feature, address all three layers:

```markdown
## Feature: [Feature Name]

### [Frontend]
- UI components needed
- Client-side validation
- Loading/error states
- Optimistic UI updates
- Accessibility considerations

### [Backend]
- API endpoints (method, path)
- Request/response schemas
- Database operations
- Business logic
- External service calls

### [Security]
- Authentication requirements
- Authorization rules
- Input sanitization
- Rate limiting
- Audit logging
```

## Example: User Profile Update

```markdown
## Feature: User Profile Update

### [Frontend]
- Form with name, email, bio, avatar fields
- Client-side validation with real-time feedback
- Loading states during submission
- Error/success message display
- Optimistic UI updates

### [Backend]
- PUT /api/users/:id endpoint
- Pydantic/Zod schema validation
- Database transaction with rollback on error
- Audit logging for profile changes
- Email verification if email changes

### [Security]
- Authorization: users can only update own profile
- Input sanitization against XSS
- Rate limiting (10 req/min per user)
- File upload validation for avatar (type, size)
- CSRF protection on form submission
```

## Technical Design Document

Create `specs/{feature_name}_design.md` with:

```markdown
# Feature: {Name}

## Requirements (EARS Format)
While <precondition>, when <trigger>, the system shall <response>.

Example: While a user is logged in, when they click Save, the system shall
persist the form data and display a success message.

## Architecture
- Frontend: [Components, state management]
- Backend: [Endpoints, data models]
- Security: [Auth, validation, protection]

## Implementation Plan
- [ ] Step 1: Create Pydantic/Zod schemas
- [ ] Step 2: Implement API endpoint
- [ ] Step 3: Build UI component
- [ ] Step 4: Add error handling
- [ ] Step 5: Write tests
```

## Quick Reference

| Layer | Key Concerns |
|-------|--------------|
| Frontend | UX, validation, states, accessibility |
| Backend | API, data, logic, performance |
| Security | Auth, authz, sanitization, logging |

---

## Reference: error-handling

# Error Handling Patterns

## Frontend Error Handling

```typescript
// React with async/await
async function handleSubmit(data: FormData) {
  setLoading(true);
  setError(null);

  try {
    const result = await api.updateProfile(data);
    showSuccess('Profile updated');
    return result;
  } catch (error) {
    if (error.status === 401) {
      redirect('/login');
    } else if (error.status === 403) {
      showError('Not authorized');
    } else if (error.status === 422) {
      setValidationErrors(error.errors);
    } else {
      showError('Something went wrong');
      reportError(error); // Send to error tracking
    }
  } finally {
    setLoading(false);
  }
}
```

```typescript
// Custom hook for API calls
function useApi<T>(fn: () => Promise<T>) {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [loading, setLoading] = useState(false);

  const execute = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await fn();
      setData(result);
      return result;
    } catch (e) {
      setError(e as Error);
      throw e;
    } finally {
      setLoading(false);
    }
  }, [fn]);

  return { data, error, loading, execute };
}
```

## Backend Error Handling

```python
# FastAPI
from fastapi import HTTPException

@router.put("/users/{user_id}")
async def update_user(
    user_id: int,
    data: UserUpdate,
    current_user: User = Depends(get_current_user)
):
    if current_user.id != user_id and not current_user.is_admin:
        raise HTTPException(status_code=403, detail="Not authorized")

    try:
        return await user_service.update(user_id, data)
    except UserNotFound:
        raise HTTPException(status_code=404, detail="User not found")
    except EmailTaken:
        raise HTTPException(status_code=422, detail="Email already in use")
```

```typescript
// NestJS
@Put(':id')
async updateUser(
  @Param('id') id: string,
  @Body() dto: UpdateUserDto,
  @CurrentUser() user: User,
) {
  if (user.id !== id && !user.isAdmin) {
    throw new ForbiddenException('Not authorized');
  }

  try {
    return await this.userService.update(id, dto);
  } catch (error) {
    if (error instanceof UserNotFoundError) {
      throw new NotFoundException('User not found');
    }
    if (error instanceof EmailTakenError) {
      throw new UnprocessableEntityException('Email already in use');
    }
    throw error;
  }
}
```

## Error Response Format

```typescript
// Consistent error shape
interface ApiError {
  error: string;
  message: string;
  details?: Record<string, string[]>;
  requestId?: string;
}

// Example responses
{ "error": "VALIDATION_ERROR", "message": "Invalid input", "details": { "email": ["Invalid format"] } }
{ "error": "NOT_FOUND", "message": "User not found" }
{ "error": "FORBIDDEN", "message": "Not authorized to perform this action" }
```

## Quick Reference

| HTTP Code | When to Use | Example |
|-----------|-------------|---------|
| 400 | Invalid request format | Malformed JSON |
| 401 | Not authenticated | Missing/invalid token |
| 403 | Not authorized | Wrong permissions |
| 404 | Resource not found | User doesn't exist |
| 409 | Conflict | Duplicate email |
| 422 | Validation failed | Invalid email format |
| 429 | Rate limited | Too many requests |
| 500 | Server error | Unhandled exception |

---

## Reference: frontend-patterns

# Frontend Patterns

## TypeScript Configuration

### Strict Setup
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "moduleResolution": "bundler",
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "noUncheckedIndexedAccess": true,
    "exactOptionalPropertyTypes": true,
    "skipLibCheck": true,
    "esModuleInterop": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "react-jsx",
    "baseUrl": ".",
    "paths": {
      "@/components/*": ["src/components/*"],
      "@/hooks/*": ["src/hooks/*"],
      "@/utils/*": ["src/utils/*"],
      "@/types/*": ["src/types/*"]
    }
  }
}
```

## Real-time Features

### WebSocket Hook
```typescript
function useWebSocket(url: string) {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<any>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => setIsConnected(true);
    ws.onclose = () => setIsConnected(false);
    ws.onmessage = (event) => setLastMessage(JSON.parse(event.data));

    return () => ws.close();
  }, [url]);

  const sendMessage = useCallback((data: any) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(data));
    }
  }, []);

  return { isConnected, lastMessage, sendMessage };
}

// Usage
function Chat() {
  const { isConnected, lastMessage, sendMessage } = useWebSocket('ws://localhost:3000');

  return (
    <div>
      <div>Status: {isConnected ? 'Connected' : 'Disconnected'}</div>
      <button onClick={() => sendMessage({ text: 'Hello' })}>Send</button>
    </div>
  );
}
```

### Optimistic Updates
```typescript
// React Query with optimistic update
function useUpdateTodo() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (todo: Todo) => api.updateTodo(todo),

    // Optimistically update cache before mutation
    onMutate: async (newTodo) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: ['todos'] });

      // Snapshot previous value
      const previous = queryClient.getQueryData(['todos']);

      // Optimistically update
      queryClient.setQueryData(['todos'], (old: Todo[]) =>
        old.map(todo => todo.id === newTodo.id ? newTodo : todo)
      );

      return { previous };
    },

    // Rollback on error
    onError: (err, newTodo, context) => {
      queryClient.setQueryData(['todos'], context?.previous);
      toast.error('Failed to update todo');
    },

    // Refetch on success
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['todos'] });
    },
  });
}
```

### Presence Hook
```typescript
function usePresence(roomId: string) {
  const [users, setUsers] = useState<User[]>([]);
  const { sendMessage, lastMessage } = useWebSocket(`ws://localhost:3000/presence`);

  useEffect(() => {
    sendMessage({ type: 'join', roomId });
    const interval = setInterval(() => sendMessage({ type: 'heartbeat', roomId }), 30000);
    return () => {
      sendMessage({ type: 'leave', roomId });
      clearInterval(interval);
    };
  }, [roomId, sendMessage]);

  useEffect(() => {
    if (lastMessage?.type === 'presence_update') setUsers(lastMessage.users);
  }, [lastMessage]);

  return users;
}
```

## Performance Optimization

### Code Splitting & Lazy Loading
```typescript
import { lazy, Suspense } from 'react';

// Lazy load route components
const Dashboard = lazy(() => import('./pages/Dashboard'));
const Profile = lazy(() => import('./pages/Profile'));

function App() {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <Routes>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/profile" element={<Profile />} />
      </Routes>
    </Suspense>
  );
}

// Component-level code splitting
const HeavyChart = lazy(() => import('./components/HeavyChart'));

function Dashboard() {
  return (
    <div>
      <h1>Dashboard</h1>
      <Suspense fallback={<div>Loading chart...</div>}>
        <HeavyChart data={data} />
      </Suspense>
    </div>
  );
}
```

### Bundle Analysis
```javascript
// webpack.config.js
const { BundleAnalyzerPlugin } = require('webpack-bundle-analyzer');

module.exports = {
  plugins: [
    new BundleAnalyzerPlugin({ analyzerMode: 'static' })
  ]
};
```

### Lazy Load Images
```typescript
function LazyImage({ src, alt }: Props) {
  const [imgSrc, setImgSrc] = useState('/placeholder.jpg');
  const imgRef = useRef<HTMLImageElement>(null);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setImgSrc(src);
          observer.disconnect();
        }
      },
      { rootMargin: '100px' }
    );

    if (imgRef.current) observer.observe(imgRef.current);
    return () => observer.disconnect();
  }, [src]);

  return <img ref={imgRef} src={imgSrc} alt={alt} />;
}
```

## Accessibility

### Accessible Modal
```typescript
function Modal({ isOpen, onClose, title, children }: Props) {
  const titleId = useId();

  useEffect(() => {
    if (isOpen) document.body.style.overflow = 'hidden';
    return () => { document.body.style.overflow = ''; };
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div role="dialog" aria-modal="true" aria-labelledby={titleId} onClick={onClose}>
      <div onClick={(e) => e.stopPropagation()}>
        <h2 id={titleId}>{title}</h2>
        {children}
        <button onClick={onClose} aria-label="Close modal">×</button>
      </div>
    </div>
  );
}
```

### Keyboard Navigation
```typescript
function Dropdown({ items }: Props) {
  const [selectedIndex, setSelectedIndex] = useState(0);

  const handleKeyDown = (e: KeyboardEvent) => {
    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setSelectedIndex(i => (i + 1) % items.length);
        break;
      case 'ArrowUp':
        e.preventDefault();
        setSelectedIndex(i => (i - 1 + items.length) % items.length);
        break;
      case 'Enter':
        selectItem(items[selectedIndex]);
        break;
    }
  };

  return <div role="combobox" onKeyDown={handleKeyDown}>{/* ... */}</div>;
}
```

### Focus Trap
```typescript
function useFocusTrap(ref: RefObject<HTMLElement>) {
  useEffect(() => {
    const element = ref.current;
    if (!element) return;

    const focusable = element.querySelectorAll(
      'a[href], button:not([disabled]), textarea, input, select'
    );
    const first = focusable[0] as HTMLElement;
    const last = focusable[focusable.length - 1] as HTMLElement;

    const handleTab = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return;
      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault();
        last.focus();
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        first.focus();
      }
    };

    element.addEventListener('keydown', handleTab);
    first?.focus();
    return () => element.removeEventListener('keydown', handleTab);
  }, [ref]);
}
```

## Testing

### Component Testing with Testing Library
```typescript
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

describe('UserForm', () => {
  it('validates email format', async () => {
    const user = userEvent.setup();
    render(<UserForm onSubmit={jest.fn()} />);

    const emailInput = screen.getByLabelText(/email/i);
    await user.type(emailInput, 'invalid-email');
    await user.click(screen.getByRole('button', { name: /submit/i }));

    expect(await screen.findByText(/invalid email/i)).toBeInTheDocument();
  });

  it('submits valid form', async () => {
    const onSubmit = jest.fn();
    const user = userEvent.setup();
    render(<UserForm onSubmit={onSubmit} />);

    await user.type(screen.getByLabelText(/name/i), 'John Doe');
    await user.type(screen.getByLabelText(/email/i), 'john@example.com');
    await user.click(screen.getByRole('button', { name: /submit/i }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith({
        name: 'John Doe',
        email: 'john@example.com',
      });
    });
  });
});
```

## Quick Reference

| Pattern | Use Case | Key Benefit |
|---------|----------|-------------|
| WebSocket | Real-time updates | Bidirectional communication |
| Optimistic Updates | Better UX | Instant feedback |
| Code Splitting | Large apps | Faster initial load |
| Lazy Loading | Images, routes | Reduce bundle size |
| ARIA attributes | Screen readers | Accessibility compliance |
| Focus trap | Modals | Keyboard navigation |

---

## Reference: integration-patterns

# Integration Patterns

## Type Safety Across Stack

### Shared Type Definitions
```typescript
// packages/shared/types.ts
export interface User {
  id: string;
  email: string;
  name: string;
  role: 'admin' | 'user';
}

export interface CreateUserDto {
  email: string;
  name: string;
  password: string;
}

export interface UpdateUserDto {
  email?: string;
  name?: string;
}

// API response wrapper
export interface ApiResponse<T> {
  data: T;
  meta?: {
    page?: number;
    limit?: number;
    total?: number;
  };
}
```

### Shared Validation (Zod)
```typescript
// packages/shared/schemas.ts
import { z } from 'zod';

export const createUserSchema = z.object({
  email: z.string().email(),
  name: z.string().min(1).max(100),
  password: z.string().min(12),
});

export type CreateUserDto = z.infer<typeof createUserSchema>;

// Backend: const validated = createUserSchema.parse(req.body);
// Frontend: useForm({ resolver: zodResolver(createUserSchema) });
```

### API Client Generation
```typescript
// Generated from OpenAPI spec
import { UserApi } from '@/generated/api';

const user = await userApi.getUser({ id: '123' }); // Type-safe
```

## Architecture Decisions

### Monorepo Structure
```
workspace/
├── packages/
│   ├── shared/           # Shared types, utils, schemas
│   ├── backend/          # Node.js/Python backend
│   ├── frontend/         # React/Vue frontend
│   ├── mobile/           # React Native (optional)
│   └── e2e-tests/        # End-to-end tests
├── package.json
└── turbo.json           # Turborepo config
```

```json
// package.json (workspace root)
{
  "private": true,
  "workspaces": ["packages/*"],
  "scripts": {
    "dev": "turbo run dev",
    "build": "turbo run build",
    "test": "turbo run test"
  }
}

// turbo.json
{
  "pipeline": {
    "build": {
      "dependsOn": ["^build"],
      "outputs": ["dist/**"]
    },
    "dev": {
      "cache": false
    },
    "test": {
      "dependsOn": ["build"],
      "outputs": []
    }
  }
}
```

### BFF (Backend for Frontend)
```typescript
// Aggregates multiple services for frontend
@Controller('bff')
export class BFFController {
  @Get('dashboard')
  async getDashboard(@CurrentUser() user: User) {
    const [profile, orders, analytics] = await Promise.all([
      this.userService.getProfile(user.id),
      this.orderService.getRecentOrders(user.id, 5),
      this.analyticsService.getUserStats(user.id),
    ]);

    return { profile, orders, analytics };
  }
}
```

### Microservices vs Monolith Decision Matrix

| Factor | Monolith | Microservices |
|--------|----------|---------------|
| Team size | < 10 developers | > 10 developers |
| Deployment | Simple, all-at-once | Complex, independent |
| Scaling | Vertical | Horizontal per service |
| Development speed | Fast initially | Slower setup, faster iteration |
| Infrastructure | Simpler | More complex (K8s, service mesh) |
| Data consistency | ACID transactions | Eventual consistency |

## Deployment Pipeline

### CI/CD Configuration (GitHub Actions)
```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Run linter
        run: npm run lint

      - name: Run unit tests
        run: npm run test

      - name: Run E2E tests
        run: npm run test:e2e

      - name: Build
        run: npm run build

  deploy-staging:
    needs: test
    if: github.ref == 'refs/heads/develop'
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to staging
        run: |
          echo "Deploy to staging environment"
          # Deploy commands here

  deploy-production:
    needs: test
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to production
        run: |
          echo "Deploy to production environment"
          # Blue-green deployment commands
```

### Database Migrations
```typescript
// TypeORM migration
export class AddUserRoles implements MigrationInterface {
  async up(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
      ALTER TABLE users ADD COLUMN role VARCHAR(20) DEFAULT 'user';
      CREATE INDEX idx_users_role ON users(role);
    `);
  }

  async down(queryRunner: QueryRunner): Promise<void> {
    await queryRunner.query(`
      DROP INDEX idx_users_role;
      ALTER TABLE users DROP COLUMN role;
    `);
  }
}

// Run: npm run migration:run
// Revert: npm run migration:revert
```

### Feature Flags
```typescript
class FeatureFlags {
  private flags = new Map<string, boolean>();

  constructor() {
    this.flags.set('new_dashboard', process.env.FEATURE_NEW_DASHBOARD === 'true');
  }

  isEnabled(flag: string): boolean {
    return this.flags.get(flag) ?? false;
  }
}

// Backend: if (flags.isEnabled('new_dashboard')) return getNewDashboard();
// Frontend: {flags.isEnabled('new_dashboard') ? <New /> : <Old />}
```

### Blue-Green Deployment
```bash
#!/bin/bash
docker build -t myapp:new .
kubectl apply -f k8s/green-deployment.yml
kubectl wait --for=condition=ready pod -l app=myapp,env=green --timeout=300s
kubectl patch service myapp -p '{"spec":{"selector":{"env":"green"}}}'
# Keep blue for rollback, then: kubectl delete deployment myapp-blue
```

## End-to-End Testing

### Playwright E2E Tests
```typescript
import { test, expect } from '@playwright/test';

test('should login successfully', async ({ page }) => {
  await page.goto('/login');
  await page.fill('[name="email"]', 'test@example.com');
  await page.fill('[name="password"]', '<test-password>');
  await page.click('button[type="submit"]');

  await page.waitForResponse(res =>
    res.url().includes('/api/auth/login') && res.status() === 200
  );

  await expect(page).toHaveURL('/dashboard');
  await expect(page.locator('[data-testid="user-name"]')).toHaveText('Test User');
});
```

### Load Testing with k6
```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 20 },  // Ramp up to 20 users
    { duration: '1m', target: 20 },   // Stay at 20 users
    { duration: '30s', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
    http_req_failed: ['rate<0.01'],   // Error rate under 1%
  },
};

export default function () {
  const res = http.get('https://api.example.com/users');

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(1);
}
```

## Environment Management

### Multi-environment Config
```typescript
interface Environment {
  api: { baseUrl: string; timeout: number };
  database: { host: string; port: number; name: string };
  features: { analytics: boolean; betaFeatures: boolean };
}

const environments: Record<string, Environment> = {
  development: {
    api: { baseUrl: 'http://localhost:3000', timeout: 30000 },
    database: { host: 'localhost', port: 5432, name: 'myapp_dev' },
    features: { analytics: false, betaFeatures: true },
  },
  production: {
    api: { baseUrl: 'https://api.example.com', timeout: 10000 },
    database: { host: process.env.DB_HOST!, port: 5432, name: 'myapp_prod' },
    features: { analytics: true, betaFeatures: false },
  },
};

export const config = environments[process.env.NODE_ENV || 'development'];
```

## Quick Reference

| Pattern | Use Case | Key Benefit |
|---------|----------|-------------|
| Shared Types | Type safety | Prevent API contract drift |
| Zod Schemas | Validation | DRY validation logic |
| Monorepo | Multi-package project | Code sharing & consistency |
| BFF Pattern | Complex frontends | Optimized API for UI needs |
| Feature Flags | Gradual rollout | Safe deployments |
| Blue-Green Deploy | Zero downtime | Instant rollback |
| E2E Tests | User flows | Catch integration bugs |
| Load Testing | Performance validation | Ensure scalability |

---

## Reference: security-checklist

# Security Checklist

## Per-Feature Security Checklist

| Category | Check | Action |
|----------|-------|--------|
| **Auth** | Endpoint requires authentication? | Add auth middleware/guard |
| **Authz** | User authorized for this action? | Check ownership/role |
| **Input** | All input validated and sanitized? | Use schemas, sanitize |
| **Output** | Sensitive data excluded from response? | Filter response fields |
| **Rate Limit** | Endpoint rate limited? | Add rate limiter |
| **Logging** | Security events logged? | Log auth failures, changes |

## Authentication Patterns

```typescript
// NestJS Guard
@UseGuards(JwtAuthGuard)
@Get('profile')
async getProfile(@CurrentUser() user: User) {
  return this.userService.findById(user.id);
}

// Express Middleware
app.get('/profile', authenticate, (req, res) => {
  res.json(req.user);
});
```

```python
# FastAPI Dependency
@router.get("/profile")
async def get_profile(current_user: User = Depends(get_current_user)):
    return current_user
```

## Authorization Patterns

```typescript
// Resource ownership check
async updatePost(postId: string, userId: string, data: UpdatePostDto) {
  const post = await this.postRepo.findById(postId);

  if (post.authorId !== userId) {
    throw new ForbiddenException('Not authorized to edit this post');
  }

  return this.postRepo.update(postId, data);
}

// Role-based check
@Roles('admin')
@UseGuards(RolesGuard)
@Delete(':id')
async deleteUser(@Param('id') id: string) {
  return this.userService.delete(id);
}
```

## Input Validation

```typescript
// Zod schema
const CreateUserSchema = z.object({
  email: z.string().email(),
  name: z.string().min(1).max(100),
  password: z.string().min(12),
});

// Use in endpoint
const validated = CreateUserSchema.parse(req.body);
```

```python
# Pydantic model
class CreateUser(BaseModel):
    email: EmailStr
    name: str = Field(min_length=1, max_length=100)
    password: str = Field(min_length=12)
```

## Rate Limiting

```typescript
// Express rate-limit
import rateLimit from 'express-rate-limit';

const authLimiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 5, // 5 attempts
  message: 'Too many login attempts',
});

app.post('/login', authLimiter, loginHandler);
```

## Quick Reference

| Risk | Mitigation |
|------|------------|
| SQL Injection | Parameterized queries |
| XSS | Output encoding, CSP |
| CSRF | CSRF tokens, SameSite cookies |
| IDOR | Authorization checks |
| Brute Force | Rate limiting |
| Data Exposure | Response filtering |
