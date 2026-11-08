# api-designer

> Источник: https://github.com/jeffallan/claude-skills/tree/main/skills/api-designer (MIT). Рабочая копия скилла: `.claude/skills/api-designer/`.

---
name: api-designer
description: Use when designing REST or GraphQL APIs, creating OpenAPI specifications, or planning API architecture. Invoke for resource modeling, versioning strategies, pagination patterns, error handling standards.
license: MIT
metadata:
  author: https://github.com/Jeffallan
  company: https://synergetic.solutions
  version: "1.1.0"
  domain: api-architecture
  triggers: API design, REST API, OpenAPI, API specification, API architecture, resource modeling, API versioning, GraphQL schema, API documentation
  role: architect
  scope: design
  output-format: specification
  related-skills: graphql-architect, fastapi-expert, nestjs-expert, spring-boot-engineer, security-reviewer
---

# API Designer

Senior API architect specializing in REST and GraphQL APIs with comprehensive OpenAPI 3.1 specifications.

## Core Workflow

1. **Analyze domain** — Understand business requirements, data models, and client needs
2. **Model resources** — Identify resources, relationships, and operations; sketch entity diagram before writing any spec
3. **Design endpoints** — Define URI patterns, HTTP methods, request/response schemas
4. **Specify contract** — Create OpenAPI 3.1 spec; validate before proceeding: `npx @redocly/cli lint openapi.yaml`
5. **Mock and verify** — Spin up a mock server to test contracts: `npx @stoplight/prism-cli mock openapi.yaml`
6. **Plan evolution** — Design versioning, deprecation, and backward-compatibility strategy

## Reference Guide

Load detailed guidance based on context:

| Topic | Reference | Load When |
|-------|-----------|-----------|
| REST Patterns | `references/rest-patterns.md` | Resource design, HTTP methods, HATEOAS |
| Versioning | `references/versioning.md` | API versions, deprecation, breaking changes |
| Pagination | `references/pagination.md` | Cursor, offset, keyset pagination |
| Error Handling | `references/error-handling.md` | Error responses, RFC 7807, status codes |
| OpenAPI | `references/openapi.md` | OpenAPI 3.1, documentation, code generation |

## Constraints

### MUST DO
- Follow REST principles (resource-oriented, proper HTTP methods)
- Use consistent naming conventions (snake_case or camelCase — pick one, apply everywhere)
- Include comprehensive OpenAPI 3.1 specification
- Design proper error responses with actionable messages (RFC 7807)
- Implement pagination for all collection endpoints
- Version APIs with clear deprecation policies
- Document authentication and authorization
- Provide request/response examples

### MUST NOT DO
- Use verbs in resource URIs (use `/users/{id}`, not `/getUser/{id}`)
- Return inconsistent response structures
- Skip error code documentation
- Ignore HTTP status code semantics
- Design APIs without a versioning strategy
- Expose implementation details in the API surface
- Create breaking changes without a migration path
- Omit rate limiting considerations

## Templates

### OpenAPI 3.1 Resource Endpoint (copy-paste starter)

```yaml
openapi: "3.1.0"
info:
  title: Example API
  version: "1.1.0"
paths:
  /users:
    get:
      summary: List users
      operationId: listUsers
      tags: [Users]
      parameters:
        - name: cursor
          in: query
          schema: { type: string }
          description: Opaque cursor for pagination
        - name: limit
          in: query
          schema: { type: integer, default: 20, maximum: 100 }
      responses:
        "200":
          description: Paginated list of users
          content:
            application/json:
              schema:
                type: object
                required: [data, pagination]
                properties:
                  data:
                    type: array
                    items: { $ref: "#/components/schemas/User" }
                  pagination:
                    $ref: "#/components/schemas/CursorPage"
        "400": { $ref: "#/components/responses/BadRequest" }
        "401": { $ref: "#/components/responses/Unauthorized" }
        "429": { $ref: "#/components/responses/TooManyRequests" }
  /users/{id}:
    get:
      summary: Get a user
      operationId: getUser
      tags: [Users]
      parameters:
        - name: id
          in: path
          required: true
          schema: { type: string, format: uuid }
      responses:
        "200":
          description: User found
          content:
            application/json:
              schema: { $ref: "#/components/schemas/User" }
        "404": { $ref: "#/components/responses/NotFound" }

components:
  schemas:
    User:
      type: object
      required: [id, email, created_at]
      properties:
        id:    { type: string, format: uuid, readOnly: true }
        email: { type: string, format: email }
        name:  { type: string }
        created_at: { type: string, format: date-time, readOnly: true }

    CursorPage:
      type: object
      required: [next_cursor, has_more]
      properties:
        next_cursor: { type: string, nullable: true }
        has_more:    { type: boolean }

    Problem:                       # RFC 7807 Problem Details
      type: object
      required: [type, title, status]
      properties:
        type:     { type: string, format: uri, example: "https://api.example.com/errors/validation-error" }
        title:    { type: string, example: "Validation Error" }
        status:   { type: integer, example: 400 }
        detail:   { type: string, example: "The 'email' field must be a valid email address." }
        instance: { type: string, format: uri, example: "/users/req-abc123" }

  responses:
    BadRequest:
      description: Invalid request parameters
      content:
        application/problem+json:
          schema: { $ref: "#/components/schemas/Problem" }
    Unauthorized:
      description: Missing or invalid authentication
      content:
        application/problem+json:
          schema: { $ref: "#/components/schemas/Problem" }
    NotFound:
      description: Resource not found
      content:
        application/problem+json:
          schema: { $ref: "#/components/schemas/Problem" }
    TooManyRequests:
      description: Rate limit exceeded
      headers:
        Retry-After: { schema: { type: integer } }
      content:
        application/problem+json:
          schema: { $ref: "#/components/schemas/Problem" }

  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

security:
  - BearerAuth: []
```

### RFC 7807 Error Response (copy-paste)

```json
{
  "type": "https://api.example.com/errors/validation-error",
  "title": "Validation Error",
  "status": 422,
  "detail": "The 'email' field must be a valid email address.",
  "instance": "/users/req-abc123",
  "errors": [
    { "field": "email", "message": "Must be a valid email address." }
  ]
}
```

- Always use `Content-Type: application/problem+json` for error responses.
- `type` must be a stable, documented URI — never a generic string.
- `detail` must be human-readable and actionable.
- Extend with `errors[]` for field-level validation failures.

## Output Checklist

When delivering an API design, provide:
1. Resource model and relationships (diagram or table)
2. Endpoint specifications with URIs and HTTP methods
3. OpenAPI 3.1 specification (YAML)
4. Authentication and authorization flows
5. Error response catalog (all 4xx/5xx with `type` URIs)
6. Pagination and filtering patterns
7. Versioning and deprecation strategy
8. Validation result: `npx @redocly/cli lint openapi.yaml` passes with no errors

## Knowledge Reference

REST architecture, OpenAPI 3.1, GraphQL, HTTP semantics, JSON:API, HATEOAS, OAuth 2.0, JWT, RFC 7807 Problem Details, API versioning patterns, pagination strategies, rate limiting, webhook design, SDK generation

Maintained by [@jeffallan](https://github.com/jeffallan), Principal Consultant at [Synergetic Solutions](https://synergetic.solutions)

[Documentation](https://jeffallan.github.io/claude-skills/skills/api-architecture/api-designer/)

---

## Reference: error-handling

# API Error Handling

## Error Response Design

Consistent, informative error responses are critical for API usability.

## Standard Error Format

### Basic Error Response

```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "User with ID 123 not found",
    "details": null
  }
}
```

### RFC 7807 Problem Details

Standardized error format (application/problem+json):

```http
HTTP/1.1 404 Not Found
Content-Type: application/problem+json

{
  "type": "https://api.example.com/errors/resource-not-found",
  "title": "Resource Not Found",
  "status": 404,
  "detail": "User with ID 123 does not exist",
  "instance": "/users/123"
}
```

**Fields:**
- `type` - URI reference identifying error type
- `title` - Short, human-readable summary
- `status` - HTTP status code
- `detail` - Human-readable explanation specific to this occurrence
- `instance` - URI reference for this specific occurrence

### Extended Error Response

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [
      {
        "field": "email",
        "code": "INVALID_FORMAT",
        "message": "Email must be a valid email address"
      },
      {
        "field": "age",
        "code": "OUT_OF_RANGE",
        "message": "Age must be between 18 and 120"
      }
    ],
    "request_id": "req_123456",
    "timestamp": "2024-01-15T10:30:00Z",
    "documentation_url": "https://api.example.com/docs/errors#validation-error"
  }
}
```

## Error Categories

### 1. Validation Errors (400 Bad Request)

Client sent invalid data.

```http
POST /users
Content-Type: application/json

{
  "name": "",
  "email": "invalid-email",
  "age": 15
}

Response: 400 Bad Request
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [
      {
        "field": "name",
        "code": "REQUIRED",
        "message": "Name is required"
      },
      {
        "field": "email",
        "code": "INVALID_FORMAT",
        "message": "Email must be a valid email address"
      },
      {
        "field": "age",
        "code": "OUT_OF_RANGE",
        "message": "Age must be at least 18",
        "constraints": {
          "min": 18,
          "max": 120
        }
      }
    ]
  }
}
```

### 2. Authentication Errors (401 Unauthorized)

Missing or invalid authentication credentials.

```http
GET /users/123
Authorization: Bearer <invalid-access-token>

Response: 401 Unauthorized
WWW-Authenticate: Bearer realm="api", error="invalid_token"

{
  "error": {
    "code": "INVALID_TOKEN",
    "message": "The access token is invalid or has expired",
    "details": {
      "reason": "token_expired",
      "expired_at": "2024-01-15T10:00:00Z"
    }
  }
}
```

**Common auth error codes:**
- `MISSING_TOKEN` - No auth token provided
- `INVALID_TOKEN` - Token is malformed or invalid
- `EXPIRED_TOKEN` - Token has expired
- `REVOKED_TOKEN` - Token has been revoked

### 3. Authorization Errors (403 Forbidden)

Authenticated but not authorized to perform action.

```http
DELETE /users/123
Authorization: Bearer <access-token>

Response: 403 Forbidden
{
  "error": {
    "code": "INSUFFICIENT_PERMISSIONS",
    "message": "You do not have permission to delete this user",
    "details": {
      "required_permission": "users:delete",
      "your_permissions": ["users:read", "users:update"]
    }
  }
}
```

### 4. Not Found Errors (404 Not Found)

Resource doesn't exist.

```http
GET /users/99999

Response: 404 Not Found
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "User with ID 99999 not found",
    "details": {
      "resource_type": "User",
      "resource_id": "99999"
    }
  }
}
```

### 5. Conflict Errors (409 Conflict)

Request conflicts with current state.

```http
POST /users
Content-Type: application/json

{
  "email": "existing@example.com",
  "name": "John Doe"
}

Response: 409 Conflict
{
  "error": {
    "code": "RESOURCE_ALREADY_EXISTS",
    "message": "User with email 'existing@example.com' already exists",
    "details": {
      "field": "email",
      "value": "existing@example.com",
      "existing_resource": "/users/123"
    }
  }
}
```

### 6. Rate Limiting (429 Too Many Requests)

Client exceeded rate limit.

```http
GET /users

Response: 429 Too Many Requests
Retry-After: 60
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1705320000

{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "You have exceeded the rate limit",
    "details": {
      "limit": 100,
      "window": "1 hour",
      "retry_after": 60,
      "reset_at": "2024-01-15T11:00:00Z"
    }
  }
}
```

### 7. Server Errors (500 Internal Server Error)

Unexpected server error.

```http
GET /users/123

Response: 500 Internal Server Error
{
  "error": {
    "code": "INTERNAL_SERVER_ERROR",
    "message": "An unexpected error occurred. Please try again later.",
    "request_id": "req_123456",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

**Never expose:**
- Stack traces
- Database errors
- Internal paths
- Sensitive configuration

### 8. Service Unavailable (503 Service Unavailable)

Service temporarily unavailable.

```http
GET /users

Response: 503 Service Unavailable
Retry-After: 300

{
  "error": {
    "code": "SERVICE_UNAVAILABLE",
    "message": "Service is temporarily unavailable due to maintenance",
    "details": {
      "retry_after": 300,
      "maintenance_end": "2024-01-15T12:00:00Z"
    }
  }
}
```

## Error Code Catalog

Define standard error codes for your API:

```json
{
  "VALIDATION_ERROR": {
    "status": 400,
    "description": "Request validation failed",
    "subcodes": {
      "REQUIRED": "Required field is missing",
      "INVALID_FORMAT": "Field has invalid format",
      "OUT_OF_RANGE": "Value is out of allowed range",
      "INVALID_ENUM": "Value is not in allowed set"
    }
  },
  "AUTHENTICATION_ERROR": {
    "status": 401,
    "description": "Authentication failed",
    "subcodes": {
      "MISSING_TOKEN": "No authentication token provided",
      "INVALID_TOKEN": "Token is invalid",
      "EXPIRED_TOKEN": "Token has expired"
    }
  },
  "AUTHORIZATION_ERROR": {
    "status": 403,
    "description": "Insufficient permissions",
    "subcodes": {
      "INSUFFICIENT_PERMISSIONS": "Missing required permission",
      "RESOURCE_FORBIDDEN": "Access to resource is forbidden"
    }
  },
  "RESOURCE_NOT_FOUND": {
    "status": 404,
    "description": "Resource not found"
  },
  "CONFLICT_ERROR": {
    "status": 409,
    "description": "Request conflicts with current state",
    "subcodes": {
      "RESOURCE_ALREADY_EXISTS": "Resource already exists",
      "CONCURRENT_MODIFICATION": "Resource was modified by another request"
    }
  },
  "RATE_LIMIT_EXCEEDED": {
    "status": 429,
    "description": "Rate limit exceeded"
  },
  "INTERNAL_SERVER_ERROR": {
    "status": 500,
    "description": "Internal server error"
  }
}
```

## Validation Error Details

### Field-Level Validation

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [
      {
        "field": "credit_card.number",
        "code": "INVALID_FORMAT",
        "message": "Credit card number must be 16 digits",
        "value_provided": "1234",
        "constraints": {
          "pattern": "^[0-9]{16}$"
        }
      },
      {
        "field": "items[0].quantity",
        "code": "OUT_OF_RANGE",
        "message": "Quantity must be at least 1",
        "value_provided": 0,
        "constraints": {
          "min": 1,
          "max": 1000
        }
      }
    ]
  }
}
```

### Cross-Field Validation

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [
      {
        "fields": ["start_date", "end_date"],
        "code": "INVALID_RANGE",
        "message": "End date must be after start date",
        "values_provided": {
          "start_date": "2024-01-20",
          "end_date": "2024-01-15"
        }
      }
    ]
  }
}
```

## Request ID Tracking

Always include request ID for debugging:

```http
Response Headers:
X-Request-ID: req_abc123

Response Body:
{
  "error": {
    "code": "INTERNAL_SERVER_ERROR",
    "message": "An unexpected error occurred",
    "request_id": "req_abc123"
  }
}
```

Clients can reference request ID in support tickets.

## Error Documentation

Document all possible errors for each endpoint:

```yaml
/users/{id}:
  get:
    responses:
      '200':
        description: Success
      '401':
        description: Authentication failed
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Error'
            examples:
              missing_token:
                value:
                  error:
                    code: MISSING_TOKEN
                    message: No authentication token provided
              invalid_token:
                value:
                  error:
                    code: INVALID_TOKEN
                    message: Token is invalid or expired
      '404':
        description: User not found
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Error'
            examples:
              not_found:
                value:
                  error:
                    code: RESOURCE_NOT_FOUND
                    message: User with ID 123 not found
```

## Retry Guidance

Help clients understand if they should retry:

```json
{
  "error": {
    "code": "SERVICE_UNAVAILABLE",
    "message": "Service temporarily unavailable",
    "retry": {
      "retryable": true,
      "retry_after": 60,
      "max_retries": 3,
      "backoff": "exponential"
    }
  }
}
```

### Retryable Errors

- 408 Request Timeout
- 429 Too Many Requests (with Retry-After)
- 500 Internal Server Error (sometimes)
- 502 Bad Gateway
- 503 Service Unavailable
- 504 Gateway Timeout

### Non-Retryable Errors

- 400 Bad Request
- 401 Unauthorized
- 403 Forbidden
- 404 Not Found
- 409 Conflict
- 422 Unprocessable Entity

## Multi-Language Support

Support error messages in multiple languages:

```http
GET /users/invalid
Accept-Language: es

Response: 404 Not Found
Content-Language: es
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Usuario con ID 'invalid' no encontrado"
  }
}
```

Always include `code` so clients can implement their own translations.

## Best Practices

1. **Use standard HTTP status codes** - Don't return 200 for errors
2. **Include machine-readable codes** - Error codes for client logic
3. **Provide human-readable messages** - Clear explanations
4. **Be specific but safe** - Don't expose sensitive information
5. **Include request ID** - For tracking and debugging
6. **Document all errors** - Every possible error for each endpoint
7. **Be consistent** - Same format across all endpoints
8. **Help clients retry** - Indicate if error is retryable
9. **Validate early** - Return validation errors immediately
10. **Log errors server-side** - Track errors for monitoring

## Anti-Patterns

Avoid these mistakes:

- **Generic error messages** - "Error occurred" without details
- **Exposing stack traces** - Security risk
- **Inconsistent error format** - Different structure per endpoint
- **Missing error codes** - Only human-readable messages
- **Wrong status codes** - Returning 200 with error in body
- **No request ID** - Makes debugging impossible
- **Undocumented errors** - Clients don't know what to expect
- **Too much information** - Exposing internal implementation

---

## Reference: openapi

# OpenAPI 3.1 Specification

## What is OpenAPI?

OpenAPI (formerly Swagger) is a standard for describing REST APIs. It enables:
- Interactive documentation
- Code generation (SDKs, clients, servers)
- API testing tools
- Contract validation
- Mock servers

## Basic Structure

### Minimal OpenAPI 3.1 Spec

```yaml
openapi: 3.1.0
info:
  title: My API
  version: 1.0.0
  description: A sample API
  contact:
    name: API Support
    email: support@example.com
    url: https://example.com/support
  license:
    name: Apache 2.0
    url: https://www.apache.org/licenses/LICENSE-2.0.html

servers:
  - url: https://api.example.com/v1
    description: Production server
  - url: https://staging-api.example.com/v1
    description: Staging server
  - url: http://localhost:3000/v1
    description: Local development

paths:
  /users:
    get:
      summary: List users
      description: Retrieve a paginated list of users
      operationId: listUsers
      tags:
        - Users
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/User'

components:
  schemas:
    User:
      type: object
      required:
        - id
        - email
      properties:
        id:
          type: integer
          format: int64
          example: 123
        email:
          type: string
          format: email
          example: john@example.com
        name:
          type: string
          example: John Doe
```

## Info Object

Metadata about the API:

```yaml
info:
  title: Users API
  version: 1.0.0
  description: |
    # Users API

    This API manages user accounts and profiles.

    ## Features
    - User CRUD operations
    - Authentication with JWT
    - Role-based authorization

  termsOfService: https://example.com/terms

  contact:
    name: API Support Team
    email: api-support@example.com
    url: https://example.com/support

  license:
    name: MIT
    url: https://opensource.org/licenses/MIT

  x-api-id: users-api-v1
  x-audience: external
```

## Servers

Define API base URLs:

```yaml
servers:
  - url: https://api.example.com/v1
    description: Production
    variables:
      version:
        default: v1
        enum:
          - v1
          - v2

  - url: https://{environment}.example.com/v1
    description: Dynamic environment
    variables:
      environment:
        default: api
        enum:
          - api
          - staging
          - dev
```

## Paths and Operations

### Complete Endpoint Example

```yaml
paths:
  /users:
    get:
      summary: List users
      description: Retrieve a paginated list of users with optional filtering
      operationId: listUsers
      tags:
        - Users

      parameters:
        - name: offset
          in: query
          description: Number of items to skip
          required: false
          schema:
            type: integer
            minimum: 0
            default: 0

        - name: limit
          in: query
          description: Maximum number of items to return
          required: false
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 20

        - name: status
          in: query
          description: Filter by user status
          required: false
          schema:
            type: string
            enum:
              - active
              - inactive
              - suspended

      security:
        - bearerAuth: []

      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserListResponse'
              examples:
                success:
                  $ref: '#/components/examples/UserListSuccess'

        '401':
          $ref: '#/components/responses/Unauthorized'

        '429':
          $ref: '#/components/responses/RateLimitExceeded'

    post:
      summary: Create user
      description: Create a new user account
      operationId: createUser
      tags:
        - Users

      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserRequest'
            examples:
              basic:
                $ref: '#/components/examples/CreateUserBasic'

      responses:
        '201':
          description: User created successfully
          headers:
            Location:
              description: URL of the created user
              schema:
                type: string
                format: uri
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'

        '400':
          $ref: '#/components/responses/ValidationError'

        '409':
          description: User already exists
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /users/{userId}:
    parameters:
      - name: userId
        in: path
        description: User ID
        required: true
        schema:
          type: integer
          format: int64

    get:
      summary: Get user
      description: Retrieve a specific user by ID
      operationId: getUser
      tags:
        - Users

      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'

        '404':
          $ref: '#/components/responses/NotFound'

    put:
      summary: Update user
      description: Replace user data
      operationId: updateUser
      tags:
        - Users

      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateUserRequest'

      responses:
        '200':
          description: User updated successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'

        '404':
          $ref: '#/components/responses/NotFound'

    delete:
      summary: Delete user
      description: Delete a user account
      operationId: deleteUser
      tags:
        - Users

      responses:
        '204':
          description: User deleted successfully

        '404':
          $ref: '#/components/responses/NotFound'
```

## Components

Reusable components for your API spec.

### Schemas

```yaml
components:
  schemas:
    User:
      type: object
      required:
        - id
        - email
        - name
      properties:
        id:
          type: integer
          format: int64
          readOnly: true
          example: 123
        email:
          type: string
          format: email
          example: john@example.com
        name:
          type: string
          minLength: 1
          maxLength: 100
          example: John Doe
        status:
          type: string
          enum:
            - active
            - inactive
            - suspended
          default: active
        created_at:
          type: string
          format: date-time
          readOnly: true
          example: "2024-01-15T10:30:00Z"
        metadata:
          type: object
          additionalProperties:
            type: string

    CreateUserRequest:
      type: object
      required:
        - email
        - name
      properties:
        email:
          type: string
          format: email
        name:
          type: string
          minLength: 1
          maxLength: 100
        metadata:
          type: object
          additionalProperties:
            type: string

    UserListResponse:
      type: object
      properties:
        data:
          type: array
          items:
            $ref: '#/components/schemas/User'
        pagination:
          $ref: '#/components/schemas/Pagination'

    Pagination:
      type: object
      properties:
        offset:
          type: integer
          minimum: 0
        limit:
          type: integer
          minimum: 1
        total:
          type: integer
          minimum: 0
        has_more:
          type: boolean

    Error:
      type: object
      required:
        - error
      properties:
        error:
          type: object
          required:
            - code
            - message
          properties:
            code:
              type: string
              example: RESOURCE_NOT_FOUND
            message:
              type: string
              example: User with ID 123 not found
            details:
              type: object
            request_id:
              type: string
              example: req_abc123
```

### Security Schemes

```yaml
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      description: JWT access token

    apiKey:
      type: apiKey
      in: header
      name: X-API-Key
      description: API key for authentication

    oauth2:
      type: oauth2
      flows:
        authorizationCode:
          authorizationUrl: https://auth.example.com/oauth/authorize
          tokenUrl: https://auth.example.com/oauth/token
          scopes:
            users:read: Read user data
            users:write: Create and update users
            users:delete: Delete users
```

Apply security globally or per-operation:

```yaml
# Global security
security:
  - bearerAuth: []

# Or per-operation
paths:
  /users:
    get:
      security:
        - bearerAuth: []
        - apiKey: []  # Alternative auth method
```

### Responses

Reusable response definitions:

```yaml
components:
  responses:
    NotFound:
      description: Resource not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            error:
              code: RESOURCE_NOT_FOUND
              message: The requested resource was not found

    Unauthorized:
      description: Authentication required
      headers:
        WWW-Authenticate:
          schema:
            type: string
          description: Authentication method
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'

    ValidationError:
      description: Validation failed
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
          example:
            error:
              code: VALIDATION_ERROR
              message: Request validation failed
              details:
                - field: email
                  code: INVALID_FORMAT
                  message: Email must be a valid email address

    RateLimitExceeded:
      description: Rate limit exceeded
      headers:
        X-RateLimit-Limit:
          schema:
            type: integer
          description: Request limit per hour
        X-RateLimit-Remaining:
          schema:
            type: integer
          description: Remaining requests
        X-RateLimit-Reset:
          schema:
            type: integer
            format: int64
          description: Time when limit resets (Unix timestamp)
        Retry-After:
          schema:
            type: integer
          description: Seconds to wait before retry
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
```

### Examples

```yaml
components:
  examples:
    UserListSuccess:
      summary: Successful user list response
      value:
        data:
          - id: 1
            email: john@example.com
            name: John Doe
            status: active
            created_at: "2024-01-15T10:30:00Z"
          - id: 2
            email: jane@example.com
            name: Jane Smith
            status: active
            created_at: "2024-01-16T14:20:00Z"
        pagination:
          offset: 0
          limit: 20
          total: 150
          has_more: true

    CreateUserBasic:
      summary: Create user with minimal fields
      value:
        email: newuser@example.com
        name: New User
```

## Data Types

### Primitive Types

```yaml
# String
type: string
example: "Hello World"

# String with format
type: string
format: email
example: "user@example.com"

# Integer
type: integer
format: int64
example: 123

# Number (float)
type: number
format: double
example: 99.99

# Boolean
type: boolean
example: true

# Date-time
type: string
format: date-time
example: "2024-01-15T10:30:00Z"

# Date
type: string
format: date
example: "2024-01-15"

# UUID
type: string
format: uuid
example: "550e8400-e29b-41d4-a716-446655440000"

# URI
type: string
format: uri
example: "https://example.com/users/123"
```

### Arrays

```yaml
type: array
items:
  type: string
minItems: 1
maxItems: 10
uniqueItems: true
example: ["tag1", "tag2", "tag3"]

# Array of objects
type: array
items:
  $ref: '#/components/schemas/User'
```

### Objects

```yaml
type: object
required:
  - name
  - email
properties:
  name:
    type: string
  email:
    type: string
    format: email
  age:
    type: integer
    minimum: 0
    maximum: 120

# Additional properties
additionalProperties: false  # Strict - no extra properties
additionalProperties: true   # Allow any extra properties
additionalProperties:        # Extra properties must be strings
  type: string
```

### Enums

```yaml
type: string
enum:
  - active
  - inactive
  - suspended
default: active
```

### OneOf / AnyOf / AllOf

```yaml
# OneOf - exactly one schema matches
oneOf:
  - $ref: '#/components/schemas/CreditCard'
  - $ref: '#/components/schemas/BankAccount'

# AnyOf - one or more schemas match
anyOf:
  - $ref: '#/components/schemas/User'
  - $ref: '#/components/schemas/Organization'

# AllOf - all schemas must match (inheritance)
allOf:
  - $ref: '#/components/schemas/BaseUser'
  - type: object
    properties:
      admin_level:
        type: integer
```

## Validation

### String Validation

```yaml
type: string
minLength: 1
maxLength: 100
pattern: "^[a-zA-Z0-9_-]+$"
format: email
```

### Number Validation

```yaml
type: integer
minimum: 0
maximum: 100
exclusiveMinimum: true  # > 0 instead of >= 0
multipleOf: 5
```

### Array Validation

```yaml
type: array
minItems: 1
maxItems: 10
uniqueItems: true
```

## Tags

Organize endpoints into logical groups:

```yaml
tags:
  - name: Users
    description: User management operations
  - name: Orders
    description: Order management
  - name: Products
    description: Product catalog

paths:
  /users:
    get:
      tags:
        - Users
```

## Documentation

### Markdown Support

```yaml
description: |
  # User Management

  This endpoint allows you to manage users.

  ## Features
  - Create users
  - Update profiles
  - Delete accounts

  ## Authentication
  Requires JWT bearer token.

  ## Example
  ```json
  {
    "name": "John Doe",
    "email": "john@example.com"
  }
  ```
```

## Code Generation

Generate SDKs from OpenAPI spec:

```bash
# Generate TypeScript client
openapi-generator-cli generate \
  -i openapi.yaml \
  -g typescript-axios \
  -o ./client

# Generate Python client
openapi-generator-cli generate \
  -i openapi.yaml \
  -g python \
  -o ./python-client

# Generate server stub
openapi-generator-cli generate \
  -i openapi.yaml \
  -g nodejs-express-server \
  -o ./server
```

## Validation Tools

Validate OpenAPI spec:

```bash
# Using Swagger CLI
swagger-cli validate openapi.yaml

# Using Spectral (advanced linting)
spectral lint openapi.yaml
```

## Best Practices

1. **Use components** - Reuse schemas, responses, parameters
2. **Add examples** - Include realistic examples for all schemas
3. **Document thoroughly** - Every endpoint, parameter, response
4. **Version your spec** - Track changes to the specification
5. **Validate regularly** - Use tools to catch errors
6. **Use $ref** - Reference components instead of duplicating
7. **Include error responses** - Document all possible errors
8. **Add operationId** - Unique ID for each operation (for code gen)
9. **Tag endpoints** - Organize into logical groups
10. **Provide security schemes** - Document authentication clearly

---

## Reference: pagination

# Pagination Patterns

## Why Paginate?

Large collections can't be returned all at once due to:
- Performance (slow queries, large payloads)
- Memory constraints (server and client)
- Network timeouts
- Poor user experience

Always paginate collection endpoints.

## Pagination Strategies

### 1. Offset-Based Pagination

Most common and intuitive. Uses `offset` (skip) and `limit` (page size).

**Request:**
```http
GET /users?offset=20&limit=10
```

**Response:**
```json
{
  "data": [
    {"id": 21, "name": "User 21"},
    {"id": 22, "name": "User 22"}
  ],
  "pagination": {
    "offset": 20,
    "limit": 10,
    "total": 150,
    "has_more": true
  },
  "links": {
    "first": "/users?offset=0&limit=10",
    "prev": "/users?offset=10&limit=10",
    "next": "/users?offset=30&limit=10",
    "last": "/users?offset=140&limit=10"
  }
}
```

**Advantages:**
- Simple to implement
- Easy to understand
- Random access (jump to any page)
- Shows total count

**Disadvantages:**
- Performance degrades with large offsets (database scans many rows)
- Inconsistent results if data changes during pagination
- Inefficient for real-time data
- Database must count total rows (expensive)

**Use when:**
- Small to medium datasets
- Data doesn't change frequently
- Need random page access
- Need total count

### 2. Page-Based Pagination

Simplified offset pagination using page numbers.

**Request:**
```http
GET /users?page=3&per_page=10
```

**Response:**
```json
{
  "data": [...],
  "pagination": {
    "page": 3,
    "per_page": 10,
    "total_pages": 15,
    "total_count": 150
  },
  "links": {
    "first": "/users?page=1&per_page=10",
    "prev": "/users?page=2&per_page=10",
    "next": "/users?page=4&per_page=10",
    "last": "/users?page=15&per_page=10"
  }
}
```

**Calculation:**
- `offset = (page - 1) * per_page`
- `total_pages = ceil(total_count / per_page)`

**Same pros/cons as offset-based, but:**
- More intuitive for users (page 1, page 2)
- Common in web applications

### 3. Cursor-Based Pagination

Uses an opaque cursor (pointer) to the next set of results.

**Request:**
```http
GET /users?limit=10
GET /users?cursor=<cursor>&limit=10
```

**Response:**
```json
{
  "data": [
    {"id": 21, "name": "User 21"},
    {"id": 22, "name": "User 22"}
  ],
  "pagination": {
    "next_cursor": "<next-cursor>",
    "prev_cursor": "<prev-cursor>",
    "has_more": true
  },
  "links": {
    "next": "/users?cursor=<next-cursor>&limit=10",
    "prev": "/users?cursor=<prev-cursor>&limit=10"
  }
}
```

**Cursor structure (base64 encoded):**
```json
{"id": 30, "sort": "created_at"}
```

**Implementation:**
```sql
-- First page
SELECT * FROM users ORDER BY created_at DESC LIMIT 10;

-- Next page (cursor points to last item)
SELECT * FROM users
WHERE created_at < '2024-01-15T10:30:00Z'
ORDER BY created_at DESC
LIMIT 10;
```

**Advantages:**
- Consistent results (no skipped/duplicate items)
- Efficient for large datasets
- Works well with real-time data
- No expensive COUNT query
- Better database performance

**Disadvantages:**
- No random access (can't jump to page 10)
- No total count
- More complex to implement
- Cursor is opaque (users can't modify it)

**Use when:**
- Large datasets
- Data changes frequently
- Infinite scroll UI
- Real-time feeds
- Performance is critical

### 4. Keyset Pagination

Similar to cursor but uses actual field values instead of opaque cursor.

**Request:**
```http
GET /users?after_id=20&limit=10
GET /users?after_created_at=2024-01-15T10:30:00Z&limit=10
```

**Response:**
```json
{
  "data": [
    {"id": 21, "name": "User 21", "created_at": "2024-01-15T11:00:00Z"},
    {"id": 22, "name": "User 22", "created_at": "2024-01-15T11:30:00Z"}
  ],
  "pagination": {
    "after_id": 30,
    "limit": 10,
    "has_more": true
  },
  "links": {
    "next": "/users?after_id=30&limit=10"
  }
}
```

**Implementation:**
```sql
SELECT * FROM users
WHERE id > 20
ORDER BY id ASC
LIMIT 10;
```

**Advantages:**
- Very efficient (uses index)
- Transparent cursor (human readable)
- Consistent results
- Simple implementation

**Disadvantages:**
- Requires indexed column
- No random access
- Sorting limited to cursor field
- Complex for multi-field sorting

**Use when:**
- Simple ordering (by ID, timestamp)
- Need efficient pagination
- Want transparent cursor
- Have proper indexes

### 5. Seek Pagination (Time-Based)

Specialized keyset pagination for time-series data.

**Request:**
```http
GET /events?since=2024-01-15T10:00:00Z&until=2024-01-15T11:00:00Z&limit=100
```

**Response:**
```json
{
  "data": [...],
  "pagination": {
    "since": "2024-01-15T10:00:00Z",
    "until": "2024-01-15T11:00:00Z",
    "limit": 100,
    "has_more": true
  },
  "links": {
    "next": "/events?since=2024-01-15T11:00:00Z&until=2024-01-15T12:00:00Z&limit=100"
  }
}
```

**Use for:**
- Time-series data
- Logs and events
- Activity streams
- Analytics data

## Default Limits

Always set reasonable defaults and maximum limits:

```json
{
  "default_limit": 20,
  "max_limit": 100,
  "min_limit": 1
}
```

**Validation:**
```http
GET /users?limit=1000

Response: 400 Bad Request
{
  "error": {
    "code": "INVALID_LIMIT",
    "message": "Limit must be between 1 and 100. Default is 20."
  }
}
```

## Response Format

### Standard Pagination Object

```json
{
  "data": [...],
  "pagination": {
    "limit": 10,
    "offset": 20,
    "total": 150,
    "has_more": true,
    "has_previous": true
  }
}
```

### Link Header (RFC 5988)

```http
Link: </users?offset=0&limit=10>; rel="first",
      </users?offset=10&limit=10>; rel="prev",
      </users?offset=30&limit=10>; rel="next",
      </users?offset=140&limit=10>; rel="last"
```

**Used by:** GitHub API

### Embedded Links

```json
{
  "data": [...],
  "_links": {
    "self": { "href": "/users?offset=20&limit=10" },
    "first": { "href": "/users?offset=0&limit=10" },
    "prev": { "href": "/users?offset=10&limit=10" },
    "next": { "href": "/users?offset=30&limit=10" },
    "last": { "href": "/users?offset=140&limit=10" }
  }
}
```

## Sorting with Pagination

Always support sorting when paginating:

```http
GET /users?sort=created_at&order=desc&limit=10
GET /users?sort=-created_at&limit=10                    # Descending
GET /users?sort=last_name,first_name&limit=10           # Multi-field
```

**For cursor pagination, cursor must include sort fields:**
```json
{
  "cursor": {
    "id": 123,
    "created_at": "2024-01-15T10:30:00Z",
    "sort_fields": ["created_at", "id"]
  }
}
```

## Filtering with Pagination

Combine filtering with pagination:

```http
GET /users?status=active&role=admin&offset=0&limit=10
```

**Important:** Apply filters before pagination:
1. Filter records
2. Count filtered results
3. Apply pagination
4. Return paginated subset

## Total Count

### Include Total Count

```json
{
  "data": [...],
  "pagination": {
    "total": 1523,
    "limit": 10,
    "offset": 20
  }
}
```

**Pros:**
- Clients know total results
- Can calculate total pages
- Better UX (show "Page 3 of 153")

**Cons:**
- COUNT query is expensive
- Slows down response
- Inaccurate for large/changing datasets

### Omit Total Count

```json
{
  "data": [...],
  "pagination": {
    "has_more": true,
    "limit": 10
  }
}
```

**Use when:**
- Large datasets (COUNT is too slow)
- Real-time data (count changes constantly)
- Cursor pagination
- Infinite scroll UI

### Optional Total Count

Let client request total count:

```http
GET /users?limit=10&include_total=true
```

## Edge Cases

### Empty Results

```json
{
  "data": [],
  "pagination": {
    "offset": 0,
    "limit": 10,
    "total": 0,
    "has_more": false
  }
}
```

### Last Page

```json
{
  "data": [{"id": 150, "name": "Last User"}],
  "pagination": {
    "offset": 140,
    "limit": 10,
    "total": 150,
    "has_more": false
  },
  "links": {
    "first": "/users?offset=0&limit=10",
    "prev": "/users?offset=130&limit=10",
    "next": null
  }
}
```

### Out of Range

```http
GET /users?offset=10000&limit=10

Response: 200 OK (empty results)
{
  "data": [],
  "pagination": {
    "offset": 10000,
    "limit": 10,
    "total": 150,
    "has_more": false
  }
}
```

Or return 404 for pages that don't exist:
```http
GET /users?page=1000&per_page=10

Response: 404 Not Found
{
  "error": {
    "code": "PAGE_NOT_FOUND",
    "message": "Page 1000 does not exist. Total pages: 15"
  }
}
```

## Best Practices

1. **Always paginate collections** - Never return unbounded lists
2. **Set reasonable defaults** - Default limit of 20-50 items
3. **Enforce maximum limits** - Prevent excessive loads (max 100-1000)
4. **Include has_more flag** - Tell clients if more results exist
5. **Provide navigation links** - Make it easy to get next/prev pages
6. **Document pagination** - Explain cursor format, limits, defaults
7. **Be consistent** - Use same pagination pattern across all endpoints
8. **Consider performance** - Choose strategy based on data size/type
9. **Support sorting** - Let clients control result order
10. **Handle edge cases** - Empty results, last page, invalid cursors

## Comparison Matrix

| Feature | Offset | Page | Cursor | Keyset |
|---------|--------|------|--------|--------|
| Performance | Poor for large offsets | Poor | Excellent | Excellent |
| Random access | Yes | Yes | No | No |
| Total count | Yes | Yes | No | Optional |
| Consistency | Poor | Poor | Excellent | Excellent |
| Complexity | Simple | Simple | Medium | Medium |
| Real-time data | Poor | Poor | Excellent | Excellent |
| Database load | High | High | Low | Low |
| Use case | Small datasets | Web UIs | Feeds/streams | Large datasets |

---

## Reference: rest-patterns

# REST Design Patterns

## Resource-Oriented Architecture

REST APIs are built around resources, not actions. Resources are the nouns of your API.

### Resource Identification

**Good Resource URIs:**
```
GET    /users                  # Collection
GET    /users/{id}             # Individual resource
GET    /users/{id}/orders      # Nested collection
POST   /users                  # Create resource
PUT    /users/{id}             # Replace resource
PATCH  /users/{id}             # Update resource
DELETE /users/{id}             # Delete resource
```

**Bad Resource URIs:**
```
POST   /getUser                # Verb in URI
POST   /createUser             # Verb in URI
GET    /user?action=delete     # Action as query param
```

### Resource Naming Conventions

- Use plural nouns for collections: `/users`, `/orders`, `/products`
- Use lowercase and hyphens for readability: `/shipping-addresses`
- Avoid deep nesting (max 2-3 levels): `/users/{id}/orders/{orderId}`
- Use query parameters for filtering: `/users?status=active&role=admin`

## HTTP Method Semantics

### Safe and Idempotent Methods

| Method | Safe | Idempotent | Use Case |
|--------|------|------------|----------|
| GET | Yes | Yes | Retrieve resource(s) |
| POST | No | No | Create resource, non-idempotent operations |
| PUT | No | Yes | Replace entire resource |
| PATCH | No | No | Partial update |
| DELETE | No | Yes | Remove resource |
| HEAD | Yes | Yes | Get metadata only |
| OPTIONS | Yes | Yes | Get allowed methods |

### Method Usage

**GET - Retrieve Resources**
```http
GET /users/123
Accept: application/json

Response: 200 OK
{
  "id": 123,
  "name": "John Doe",
  "email": "john@example.com",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**POST - Create Resources**
```http
POST /users
Content-Type: application/json

{
  "name": "Jane Smith",
  "email": "jane@example.com"
}

Response: 201 Created
Location: /users/124
{
  "id": 124,
  "name": "Jane Smith",
  "email": "jane@example.com",
  "created_at": "2024-01-16T14:20:00Z"
}
```

**PUT - Replace Resource**
```http
PUT /users/123
Content-Type: application/json

{
  "name": "John Doe Updated",
  "email": "john.new@example.com"
}

Response: 200 OK
{
  "id": 123,
  "name": "John Doe Updated",
  "email": "john.new@example.com",
  "updated_at": "2024-01-17T09:15:00Z"
}
```

**PATCH - Partial Update**
```http
PATCH /users/123
Content-Type: application/json

{
  "email": "john.updated@example.com"
}

Response: 200 OK
{
  "id": 123,
  "name": "John Doe",
  "email": "john.updated@example.com",
  "updated_at": "2024-01-17T10:00:00Z"
}
```

**DELETE - Remove Resource**
```http
DELETE /users/123

Response: 204 No Content
```

## HTTP Status Codes

### Success Codes (2xx)

- **200 OK** - Request succeeded (GET, PUT, PATCH)
- **201 Created** - Resource created (POST), include Location header
- **202 Accepted** - Request accepted for async processing
- **204 No Content** - Success with no response body (DELETE)

### Redirection (3xx)

- **301 Moved Permanently** - Resource permanently moved
- **302 Found** - Temporary redirect
- **304 Not Modified** - Cached version is still valid

### Client Errors (4xx)

- **400 Bad Request** - Invalid request syntax or validation error
- **401 Unauthorized** - Authentication required or failed
- **403 Forbidden** - Authenticated but not authorized
- **404 Not Found** - Resource doesn't exist
- **405 Method Not Allowed** - HTTP method not supported for resource
- **409 Conflict** - Request conflicts with current state (e.g., duplicate)
- **422 Unprocessable Entity** - Valid syntax but semantic errors
- **429 Too Many Requests** - Rate limit exceeded

### Server Errors (5xx)

- **500 Internal Server Error** - Unexpected server error
- **502 Bad Gateway** - Invalid response from upstream server
- **503 Service Unavailable** - Server temporarily unavailable
- **504 Gateway Timeout** - Upstream server timeout

## HATEOAS (Hypermedia)

### Hypermedia-Driven APIs

Include links to related resources and available actions:

```json
{
  "id": 123,
  "name": "John Doe",
  "email": "john@example.com",
  "_links": {
    "self": { "href": "/users/123" },
    "orders": { "href": "/users/123/orders" },
    "update": { "href": "/users/123", "method": "PATCH" },
    "delete": { "href": "/users/123", "method": "DELETE" }
  }
}
```

### HAL (Hypertext Application Language)

```json
{
  "id": 123,
  "name": "John Doe",
  "_links": {
    "self": { "href": "/users/123" }
  },
  "_embedded": {
    "orders": [
      {
        "id": 456,
        "total": 99.99,
        "_links": {
          "self": { "href": "/orders/456" }
        }
      }
    ]
  }
}
```

## Content Negotiation

### Accept Headers

```http
GET /users/123
Accept: application/json

GET /users/123
Accept: application/xml

GET /users/123
Accept: application/hal+json
```

### Response Content-Type

```http
Content-Type: application/json; charset=utf-8
Content-Type: application/problem+json
Content-Type: application/hal+json
```

## Idempotency

### Idempotent Operations

**PUT - Always idempotent:**
Multiple identical PUT requests produce the same result as a single request.

**DELETE - Idempotent:**
First DELETE returns 204, subsequent DELETEs return 404 (same end state).

**POST - Not idempotent by default:**
Use `Idempotency-Key` header for idempotent POST:

```http
POST /payments
Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000
Content-Type: application/json

{
  "amount": 100.00,
  "currency": "USD"
}
```

Server stores idempotency key and returns same response for duplicate requests.

## Cache Control

### Cache Headers

```http
Cache-Control: public, max-age=3600
Cache-Control: private, no-cache
Cache-Control: no-store
ETag: "33a64df551425fcc55e4d42a148795d9f25f89d4"
Last-Modified: Wed, 15 Jan 2024 10:30:00 GMT
```

### Conditional Requests

```http
GET /users/123
If-None-Match: "33a64df551425fcc55e4d42a148795d9f25f89d4"

Response: 304 Not Modified
```

```http
PUT /users/123
If-Match: "33a64df551425fcc55e4d42a148795d9f25f89d4"
Content-Type: application/json

{
  "name": "Updated Name"
}

Response: 412 Precondition Failed (if ETag doesn't match)
```

## URI Patterns

### Consistent URI Structure

```
/{version}/{resource}
/{version}/{resource}/{id}
/{version}/{resource}/{id}/{sub-resource}
/{version}/{resource}/{id}/{sub-resource}/{sub-id}
```

### Query Parameters

**Filtering:**
```
GET /users?status=active&role=admin
GET /products?category=electronics&price_min=100&price_max=500
```

**Sorting:**
```
GET /users?sort=created_at
GET /users?sort=-created_at          # Descending
GET /users?sort=name,created_at      # Multiple fields
```

**Field Selection:**
```
GET /users?fields=id,name,email
GET /users?exclude=password,social_security_number
```

**Search:**
```
GET /users?q=john
GET /products?search=laptop
```

## Best Practices

1. **Use nouns, not verbs** - Resources are nouns, methods are verbs
2. **Plural collections** - Use `/users` not `/user`
3. **Consistent naming** - Choose snake_case or camelCase and stick to it
4. **Proper status codes** - Use appropriate HTTP status codes
5. **Include metadata** - Pagination, filtering, sorting info in responses
6. **Version your API** - Plan for evolution from day one
7. **Document everything** - OpenAPI specs, examples, error codes
8. **Security by default** - HTTPS, authentication, rate limiting
9. **Support filtering** - Enable clients to get exactly what they need
10. **Implement HATEOAS** - Make APIs self-documenting and discoverable

---

## Reference: versioning

# API Versioning Strategies

## Why Version APIs?

API versioning allows you to evolve your API while maintaining backward compatibility for existing clients. Breaking changes require a new version.

### Breaking Changes

Changes that require a new version:
- Removing or renaming fields
- Changing field types (string to integer)
- Adding required fields to requests
- Changing response structure
- Removing endpoints
- Changing HTTP status codes for same scenario
- Changing authentication mechanisms

### Non-Breaking Changes

Safe changes that don't require a new version:
- Adding new endpoints
- Adding optional request fields
- Adding new fields to responses (clients should ignore unknown fields)
- Fixing bugs
- Performance improvements
- Adding new HTTP methods to existing resources

## Versioning Strategies

### 1. URI Versioning

Most common and visible approach. Version is part of the URL path.

```http
GET /v1/users/123
GET /v2/users/123
```

**Advantages:**
- Clear and visible in URLs
- Easy to understand and implement
- Simple routing and caching
- Can run multiple versions simultaneously

**Disadvantages:**
- Violates REST principle (same resource, different URIs)
- Requires updating client code to change version
- Can lead to URI proliferation

**Implementation:**
```
/v1/users
/v1/products
/v2/users      # New version with breaking changes
/v2/products
```

### 2. Header Versioning

Version specified in HTTP headers (Accept header or custom header).

**Accept Header:**
```http
GET /users/123
Accept: application/vnd.myapi.v1+json

GET /users/123
Accept: application/vnd.myapi.v2+json
```

**Custom Header:**
```http
GET /users/123
API-Version: 1

GET /users/123
API-Version: 2
```

**Advantages:**
- URIs remain stable
- More RESTful (same resource, same URI)
- Separates versioning from resource identification

**Disadvantages:**
- Less visible (harder to debug)
- More complex routing
- Difficult to test in browser
- Cache complexity

### 3. Query Parameter Versioning

Version specified as query parameter.

```http
GET /users/123?version=1
GET /users/123?version=2

# or
GET /users/123?api-version=1
GET /users/123?api-version=2
```

**Advantages:**
- Simple to implement
- Easy to test
- Visible in URLs

**Disadvantages:**
- Pollutes query string
- Not semantic (version not a filter)
- Can interfere with other query params

### 4. Content Negotiation

Client specifies desired version through content negotiation.

```http
GET /users/123
Accept: application/vnd.myapi+json; version=1

GET /users/123
Accept: application/vnd.myapi+json; version=2
```

**Advantages:**
- Very RESTful
- Flexible content type negotiation
- Stable URIs

**Disadvantages:**
- Complex implementation
- Less intuitive for developers
- Harder to test

## Recommended Approach

**URI versioning is recommended for most APIs** because:
- It's the most explicit and discoverable
- Easy to understand and debug
- Simple to implement and maintain
- Clear separation between versions

```
/v1/users
/v2/users
/v3/users
```

## Version Format

### Major Versions Only

Use simple major versions (v1, v2, v3) for public APIs:
```
/v1/users
/v2/users
```

**Advantages:**
- Simple and clear
- Easy to communicate
- Forces thoughtful breaking changes

### Date-Based Versions

Some APIs use dates for versions:
```
/2024-01-01/users
/2024-06-15/users
```

**Used by:** Stripe, GitHub API

**Advantages:**
- Clear when version was released
- Easy to understand timeline
- No confusion about major/minor

**Disadvantages:**
- Less intuitive for clients
- Harder to understand what changed

## Version Lifecycle

### 1. Introduction Phase

New version is released alongside existing version:
```
/v1/users  # Still supported
/v2/users  # New version available
```

Announce new version:
- Blog post explaining changes
- Migration guide
- Breaking changes list
- Timeline for v1 deprecation

### 2. Deprecation Phase

Mark old version as deprecated but keep it running:

```http
GET /v1/users/123

Response:
Deprecation: true
Sunset: Wed, 15 Jan 2025 00:00:00 GMT
Link: </v2/users/123>; rel="successor-version"

{
  "id": 123,
  "name": "John Doe"
}
```

**Deprecation Headers:**
- `Deprecation: true` - Indicates version is deprecated
- `Sunset: <date>` - When version will be removed (RFC 8594)
- `Link: <url>; rel="successor-version"` - Points to new version

### 3. Sunset Phase

Old version is shut down on announced date.

Return 410 Gone for deprecated endpoints:
```http
GET /v1/users/123

Response: 410 Gone
{
  "error": {
    "code": "VERSION_SUNSET",
    "message": "API v1 was sunset on 2025-01-15. Please use v2.",
    "documentation_url": "https://api.example.com/docs/migration-v1-to-v2"
  }
}
```

## Deprecation Policy

### Recommended Timeline

1. **Announce deprecation** - At least 6 months before sunset
2. **Support period** - Run both versions for 6-12 months
3. **Sunset date** - Clear date communicated in advance
4. **Grace period** - 30 days of 410 Gone responses before complete shutdown

### Communication Channels

- API response headers
- Email to registered developers
- Blog posts and changelog
- Dashboard notifications
- Documentation updates
- Status page announcements

## Migration Strategy

### Provide Migration Guide

```markdown
# Migrating from v1 to v2

## Breaking Changes

### User Resource Changes

**v1:**
```json
{
  "id": 123,
  "name": "John Doe",
  "email": "john@example.com"
}
```

**v2:**
```json
{
  "id": 123,
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com"
}
```

**Migration:**
- Split `name` field into `first_name` and `last_name`
- Update client code to use new fields
```

### Offer Tools

- Migration scripts
- SDK updates
- API diff viewer
- Compatibility layer (temporary)

## Version Discovery

### Root Endpoint

```http
GET /

Response:
{
  "versions": {
    "v1": {
      "status": "deprecated",
      "sunset_date": "2025-01-15",
      "documentation_url": "https://api.example.com/docs/v1"
    },
    "v2": {
      "status": "current",
      "documentation_url": "https://api.example.com/docs/v2"
    },
    "v3": {
      "status": "beta",
      "documentation_url": "https://api.example.com/docs/v3"
    }
  }
}
```

### Version Info Endpoint

```http
GET /v2/version

Response:
{
  "version": "v2",
  "released": "2024-01-15",
  "status": "stable",
  "sunset_date": null
}
```

## OpenAPI Versioning

### Separate Specs per Version

```
openapi-v1.yaml
openapi-v2.yaml
openapi-v3.yaml
```

Each spec is complete and independent.

### Single Spec with Servers

```yaml
openapi: 3.1.0
info:
  title: My API
  version: 2.0.0
servers:
  - url: https://api.example.com/v1
    description: Version 1 (deprecated)
  - url: https://api.example.com/v2
    description: Version 2 (current)
```

## Best Practices

1. **Version from day one** - Start with /v1, not /api
2. **Major versions only** - Use v1, v2, v3 (not v1.1, v1.2)
3. **Long deprecation periods** - Give clients time to migrate (6-12 months)
4. **Clear communication** - Use headers, docs, emails
5. **Maintain old versions** - Support at least 2 versions simultaneously
6. **Document changes** - Provide detailed migration guides
7. **Use semantic versioning** - For internal/SDK versioning
8. **Never break without warning** - Always announce breaking changes
9. **Provide tools** - Migration scripts, updated SDKs
10. **Monitor usage** - Track which versions are being used

## Anti-Patterns

Avoid these mistakes:

- **Breaking changes without version bump** - Breaks existing clients
- **Too many versions** - Maintenance nightmare (max 2-3 active versions)
- **Short deprecation periods** - Frustrates developers
- **No migration path** - Makes upgrades painful
- **Surprise sunsets** - Breaks production apps without warning
- **Inconsistent versioning** - Different strategies for different endpoints
- **Versioning individual endpoints** - Use consistent version across API
