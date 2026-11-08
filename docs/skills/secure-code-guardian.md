# secure-code-guardian

> Источник: https://github.com/jeffallan/claude-skills/tree/main/skills/secure-code-guardian (MIT). Рабочая копия скилла: `.claude/skills/secure-code-guardian/`.

---
name: secure-code-guardian
description: Use when implementing authentication/authorization, securing user input, or preventing OWASP Top 10 vulnerabilities — including custom security implementations such as hashing passwords with bcrypt/argon2, sanitizing SQL queries with parameterized statements, configuring CORS/CSP headers, validating input with Zod, and setting up JWT tokens. Invoke for authentication, authorization, input validation, encryption, OWASP Top 10 prevention, secure session management, and security hardening. For pre-built OAuth/SSO integrations or standalone security audits, consider a more specialized skill.
license: MIT
metadata:
  author: https://github.com/Jeffallan
  company: https://synergetic.solutions
  version: "1.1.0"
  domain: security
  triggers: security, authentication, authorization, encryption, OWASP, vulnerability, secure coding, password, JWT, OAuth
  role: specialist
  scope: implementation
  output-format: code
  related-skills: fullstack-guardian, security-reviewer, architecture-designer
---

# Secure Code Guardian

## Core Workflow

1. **Threat model** — Identify attack surface and threats
2. **Design** — Plan security controls
3. **Implement** — Write secure code with defense in depth; see code examples below
4. **Validate** — Test security controls with explicit checkpoints (see below)
5. **Document** — Record security decisions

### Validation Checkpoints

After each implementation step, verify:

- **Authentication**: Test brute-force protection (lockout/rate limit triggers), session fixation resistance, token expiration, and invalid-credential error messages (must not leak user existence).
- **Authorization**: Verify horizontal and vertical privilege escalation paths are blocked; test with tokens belonging to different roles/users.
- **Input handling**: Confirm SQL injection payloads (`' OR 1=1--`) are rejected; confirm XSS payloads (`<script>alert(1)</script>`) are escaped or rejected.
- **Headers/CORS**: Validate with a security scanner (e.g., `curl -I`, Mozilla Observatory) that security headers are present and CORS origin allowlist is correct.

## Reference Guide

Load detailed guidance based on context:

| Topic | Reference | Load When |
|-------|-----------|-----------|
| OWASP | `references/owasp-prevention.md` | OWASP Top 10 patterns |
| Authentication | `references/authentication.md` | Password hashing, JWT |
| Input Validation | `references/input-validation.md` | Zod, SQL injection |
| XSS/CSRF | `references/xss-csrf.md` | XSS prevention, CSRF |
| Headers | `references/security-headers.md` | Helmet, rate limiting |

## Constraints

### MUST DO
- Hash passwords with bcrypt/argon2 (never MD5/SHA-1/unsalted hashes)
- Use parameterized queries (never string-interpolated SQL)
- Validate and sanitize all user input before use
- Implement rate limiting on auth endpoints
- Set security headers (CSP, HSTS, X-Frame-Options)
- Log security events (failed auth, privilege escalation attempts)
- Store secrets in environment variables or secret managers (never in source code)

### MUST NOT DO
- Store passwords in plaintext or reversibly encrypted form
- Trust user input without validation
- Expose sensitive data in logs or error responses
- Use weak or deprecated algorithms (MD5, SHA-1, DES, ECB mode)
- Hardcode secrets or credentials in code

## Code Examples

### Password Hashing (bcrypt)

```typescript
import bcrypt from 'bcrypt';

const SALT_ROUNDS = 12; // minimum 10; 12 balances security and performance

export async function hashPassword(plaintext: string): Promise<string> {
  return bcrypt.hash(plaintext, SALT_ROUNDS);
}

export async function verifyPassword(plaintext: string, hash: string): Promise<boolean> {
  return bcrypt.compare(plaintext, hash);
}
```

### Parameterized SQL Query (Node.js / pg)

```typescript
// NEVER: `SELECT * FROM users WHERE email = '${email}'`
// ALWAYS: use positional parameters
import { Pool } from 'pg';
const pool = new Pool();

export async function getUserByEmail(email: string) {
  const { rows } = await pool.query(
    'SELECT id, email, role FROM users WHERE email = $1',
    [email]  // value passed separately — never interpolated
  );
  return rows[0] ?? null;
}
```

### Input Validation with Zod

```typescript
import { z } from 'zod';

const LoginSchema = z.object({
  email: z.string().email().max(254),
  password: z.string().min(8).max(128),
});

export function validateLoginInput(raw: unknown) {
  const result = LoginSchema.safeParse(raw);
  if (!result.success) {
    // Return generic error — never echo raw input back
    throw new Error('Invalid credentials format');
  }
  return result.data;
}
```

### JWT Validation

```typescript
import jwt from 'jsonwebtoken';

const JWT_SECRET = process.env.JWT_SECRET!; // never hardcode

export function verifyToken(token: string): jwt.JwtPayload {
  // Throws if expired, tampered, or wrong algorithm
  const payload = jwt.verify(token, JWT_SECRET, {
    algorithms: ['HS256'],   // explicitly allowlist algorithm
    issuer: 'your-app',
    audience: 'your-app',
  });
  if (typeof payload === 'string') throw new Error('Invalid token payload');
  return payload;
}
```

### Securing an Endpoint — Full Flow

```typescript
import express from 'express';
import rateLimit from 'express-rate-limit';
import helmet from 'helmet';

const app = express();
app.use(helmet()); // sets CSP, HSTS, X-Frame-Options, etc.
app.use(express.json({ limit: '10kb' })); // limit payload size

const authLimiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 10,                   // 10 attempts per window per IP
  standardHeaders: true,
  legacyHeaders: false,
});

app.post('/api/login', authLimiter, async (req, res) => {
  // 1. Validate input
  const { email, password } = validateLoginInput(req.body);

  // 2. Authenticate — parameterized query, constant-time compare
  const user = await getUserByEmail(email);
  if (!user || !(await verifyPassword(password, user.passwordHash))) {
    // Generic message — do not reveal whether email exists
    return res.status(401).json({ error: 'Invalid credentials' });
  }

  // 3. Authorize — issue scoped, short-lived token
  const token = jwt.sign(
    { sub: user.id, role: user.role },
    JWT_SECRET,
    { algorithm: 'HS256', expiresIn: '15m', issuer: 'your-app', audience: 'your-app' }
  );

  // 4. Secure response — token in httpOnly cookie, not body
  res.cookie('token', token, { httpOnly: true, secure: true, sameSite: 'strict' });
  return res.json({ message: 'Authenticated' });
});
```

## Output Templates

When implementing security features, provide:
1. Secure implementation code
2. Security considerations noted
3. Configuration requirements (env vars, headers)
4. Testing recommendations

## Knowledge Reference

OWASP Top 10, bcrypt/argon2, JWT, OAuth 2.0, OIDC, CSP, CORS, rate limiting, input validation, output encoding, encryption (AES, RSA), TLS, security headers

Maintained by [@jeffallan](https://github.com/jeffallan), Principal Consultant at [Synergetic Solutions](https://synergetic.solutions)

[Documentation](https://jeffallan.github.io/claude-skills/skills/security/secure-code-guardian/)

---

## Reference: authentication

# Authentication

## Password Hashing

```typescript
import bcrypt from 'bcrypt';

const SALT_ROUNDS = 12;

async function hashPassword(password: string): Promise<string> {
  return bcrypt.hash(password, SALT_ROUNDS);
}

async function verifyPassword(password: string, hash: string): Promise<boolean> {
  return bcrypt.compare(password, hash);
}

// Password requirements
const PASSWORD_REGEX = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{12,}$/;

function validatePassword(password: string): { valid: boolean; errors: string[] } {
  const errors: string[] = [];

  if (password.length < 12) errors.push('Minimum 12 characters');
  if (!/[a-z]/.test(password)) errors.push('Requires lowercase');
  if (!/[A-Z]/.test(password)) errors.push('Requires uppercase');
  if (!/\d/.test(password)) errors.push('Requires digit');
  if (!/[@$!%*?&]/.test(password)) errors.push('Requires special character');

  return { valid: errors.length === 0, errors };
}
```

## JWT Implementation

```typescript
import jwt from 'jsonwebtoken';

const JWT_SECRET = process.env.JWT_SECRET!;
const ACCESS_TOKEN_EXPIRY = '15m';
const REFRESH_TOKEN_EXPIRY = '7d';

interface TokenPayload {
  sub: string;
  type: 'access' | 'refresh';
}

function generateAccessToken(userId: string): string {
  return jwt.sign(
    { sub: userId, type: 'access' },
    JWT_SECRET,
    { expiresIn: ACCESS_TOKEN_EXPIRY }
  );
}

function generateRefreshToken(userId: string): string {
  return jwt.sign(
    { sub: userId, type: 'refresh' },
    JWT_SECRET,
    { expiresIn: REFRESH_TOKEN_EXPIRY }
  );
}

function verifyToken(token: string): TokenPayload {
  return jwt.verify(token, JWT_SECRET) as TokenPayload;
}
```

## Auth Middleware

```typescript
function authMiddleware(req: Request, res: Response, next: NextFunction) {
  const header = req.headers.authorization;

  if (!header?.startsWith('Bearer ')) {
    return res.status(401).json({ error: 'Missing token' });
  }

  try {
    const token = header.slice(7);
    const payload = verifyToken(token);

    if (payload.type !== 'access') {
      return res.status(401).json({ error: 'Invalid token type' });
    }

    req.userId = payload.sub;
    next();
  } catch (error) {
    if (error instanceof jwt.TokenExpiredError) {
      return res.status(401).json({ error: 'Token expired' });
    }
    return res.status(401).json({ error: 'Invalid token' });
  }
}
```

## Account Lockout

```typescript
const MAX_ATTEMPTS = 5;
const LOCKOUT_DURATION = 15 * 60 * 1000; // 15 minutes

async function handleLoginAttempt(email: string, success: boolean) {
  const key = `login:attempts:${email}`;

  if (success) {
    await redis.del(key);
    return;
  }

  const attempts = await redis.incr(key);
  await redis.expire(key, LOCKOUT_DURATION / 1000);

  if (attempts >= MAX_ATTEMPTS) {
    await redis.set(`login:locked:${email}`, '1', 'PX', LOCKOUT_DURATION);
    throw new Error('Account locked. Try again later.');
  }
}
```

## Quick Reference

| Practice | Implementation |
|----------|----------------|
| Password hash | bcrypt (12+ rounds) |
| Token expiry | Access: 15m, Refresh: 7d |
| Lockout | 5 attempts, 15min lockout |
| MFA | TOTP (authenticator apps) |

| JWT Claim | Purpose |
|-----------|---------|
| `sub` | User ID |
| `exp` | Expiration |
| `iat` | Issued at |
| `type` | access/refresh |

---

## Reference: input-validation

# Input Validation

## Zod Validation

```typescript
import { z } from 'zod';

const UserSchema = z.object({
  email: z.string().email().max(255),
  name: z.string().min(1).max(100).regex(/^[\w\s-]+$/),
  age: z.number().int().min(0).max(150).optional(),
  role: z.enum(['user', 'admin']).default('user'),
});

function validateUser(data: unknown) {
  return UserSchema.parse(data); // Throws on invalid
}

// Safe parse (no throw)
const result = UserSchema.safeParse(data);
if (!result.success) {
  console.error(result.error.issues);
}
```

## SQL Injection Prevention

```typescript
// ❌ NEVER do this
const bad = `SELECT * FROM users WHERE id = ${userId}`;
const bad2 = `SELECT * FROM users WHERE name = '${name}'`;

// ✅ Parameterized queries
const good = await db.query(
  'SELECT * FROM users WHERE id = $1 AND name = $2',
  [userId, name]
);

// ✅ Use ORM
const user = await prisma.user.findFirst({
  where: { id: userId, name: name }
});

// ✅ Query builder
const user = await knex('users')
  .where({ id: userId, name: name })
  .first();
```

## Path Traversal Prevention

```typescript
import path from 'path';

// ❌ Vulnerable
const vulnerable = path.join('/uploads', userInput);

// ✅ Safe - validate and sanitize
function getSecurePath(baseDir: string, userInput: string): string {
  // Remove any path traversal attempts
  const sanitized = path.basename(userInput);

  // Resolve and verify it's within base directory
  const fullPath = path.resolve(baseDir, sanitized);

  if (!fullPath.startsWith(path.resolve(baseDir))) {
    throw new Error('Invalid path');
  }

  return fullPath;
}
```

## Command Injection Prevention

```typescript
import { execFile } from 'child_process';

// ❌ Never use exec with user input
exec(`convert ${userInput}`); // Vulnerable!

// ✅ Use execFile with arguments array
execFile('convert', ['-resize', '100x100', safeFilename], (error, stdout) => {
  // ...
});

// ✅ Better: Use library functions instead of shell
import sharp from 'sharp';
await sharp(inputPath).resize(100, 100).toFile(outputPath);
```

## URL Validation

```typescript
function validateUrl(input: string, allowedHosts: string[]): URL {
  const url = new URL(input);

  // Check protocol
  if (!['http:', 'https:'].includes(url.protocol)) {
    throw new Error('Invalid protocol');
  }

  // Check host allowlist
  if (!allowedHosts.includes(url.hostname)) {
    throw new Error('Host not allowed');
  }

  return url;
}
```

## File Upload Validation

```typescript
const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/gif'];
const MAX_SIZE = 5 * 1024 * 1024; // 5MB

function validateUpload(file: Express.Multer.File) {
  if (!ALLOWED_TYPES.includes(file.mimetype)) {
    throw new Error('Invalid file type');
  }

  if (file.size > MAX_SIZE) {
    throw new Error('File too large');
  }

  // Verify magic bytes (not just extension)
  const buffer = fs.readFileSync(file.path);
  const type = fileType.fromBuffer(buffer);

  if (!type || !ALLOWED_TYPES.includes(type.mime)) {
    throw new Error('Invalid file content');
  }
}
```

## Quick Reference

| Input Type | Validation |
|------------|------------|
| Email | Regex + max length |
| URL | Protocol + host allowlist |
| File path | basename + resolve check |
| SQL | Parameterized queries |
| Command | execFile + no shell |
| File upload | Type + size + magic bytes |

---

## Reference: owasp-prevention

# OWASP Top 10 Prevention

## OWASP Top 10 Quick Reference

| # | Vulnerability | Prevention |
|---|---------------|------------|
| 1 | Injection | Parameterized queries, ORMs |
| 2 | Broken Auth | Strong passwords, MFA, secure sessions |
| 3 | Sensitive Data | Encryption at rest/transit |
| 4 | XXE | Disable DTDs, use JSON |
| 5 | Broken Access | Deny by default, server-side validation |
| 6 | Misconfig | Security headers, disable defaults |
| 7 | XSS | Output encoding, CSP |
| 8 | Insecure Deserialization | Schema validation, allowlists |
| 9 | Known Vulnerabilities | Dependency scanning |
| 10 | Insufficient Logging | Log security events |

## A01: Injection Prevention

```typescript
// SQL Injection - Use parameterized queries
// ❌ Bad
const bad = `SELECT * FROM users WHERE id = ${userId}`;

// ✅ Good
const good = await db.query('SELECT * FROM users WHERE id = $1', [userId]);

// ✅ Good - Use ORM
const user = await prisma.user.findUnique({ where: { id: userId } });

// Command Injection - Avoid shell execution
// ❌ Bad
exec(`ls ${userInput}`);

// ✅ Good - Use library functions
const files = fs.readdirSync(safeDirectory);
```

## A02: Broken Authentication

```typescript
// Use bcrypt for passwords
const hash = await bcrypt.hash(password, 12);
const isValid = await bcrypt.compare(password, hash);

// Implement account lockout
if (failedAttempts >= 5) {
  await lockAccount(userId, 15 * 60 * 1000); // 15 min
}

// Use secure session configuration
app.use(session({
  secret: process.env.SESSION_SECRET,
  cookie: {
    httpOnly: true,
    secure: true,
    sameSite: 'strict',
    maxAge: 15 * 60 * 1000, // 15 minutes
  },
}));
```

## A03: Sensitive Data Exposure

```typescript
// Encrypt sensitive data at rest
import crypto from 'crypto';

function encrypt(text: string, key: Buffer): string {
  const iv = crypto.randomBytes(16);
  const cipher = crypto.createCipheriv('aes-256-gcm', key, iv);
  // ... encryption logic
}

// Use HTTPS only
app.use((req, res, next) => {
  if (!req.secure) {
    return res.redirect(`https://${req.hostname}${req.url}`);
  }
  next();
});
```

## A05: Broken Access Control

```typescript
// Always validate on server side
async function getResource(userId: string, resourceId: string) {
  const resource = await db.resource.findUnique({ where: { id: resourceId } });

  // Verify ownership
  if (resource.ownerId !== userId) {
    throw new ForbiddenError('Access denied');
  }

  return resource;
}

// Use role-based access
function requireRole(...roles: string[]) {
  return (req: Request, res: Response, next: NextFunction) => {
    if (!roles.includes(req.user.role)) {
      return res.status(403).json({ error: 'Forbidden' });
    }
    next();
  };
}
```

## A07: XSS Prevention

```typescript
// Use Content Security Policy
app.use(helmet.contentSecurityPolicy({
  directives: {
    defaultSrc: ["'self'"],
    scriptSrc: ["'self'"],
    styleSrc: ["'self'", "'unsafe-inline'"],
  },
}));

// Sanitize user input for HTML
import DOMPurify from 'dompurify';
const clean = DOMPurify.sanitize(userInput);
```

## Quick Reference

| Attack | Defense |
|--------|---------|
| SQL Injection | Parameterized queries |
| XSS | Output encoding, CSP |
| CSRF | CSRF tokens |
| IDOR | Authorization checks |
| Command Injection | Avoid exec(), validate input |

---

## Reference: security-headers

# Security Headers

## Helmet (Express)

```typescript
import helmet from 'helmet';

app.use(helmet()); // Enable all defaults

// Or configure individually
app.use(helmet({
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      scriptSrc: ["'self'"],
      styleSrc: ["'self'", "'unsafe-inline'"],
    },
  },
  hsts: {
    maxAge: 31536000,
    includeSubDomains: true,
    preload: true,
  },
}));
```

## Manual Headers

```typescript
app.use((req, res, next) => {
  // Prevent clickjacking
  res.setHeader('X-Frame-Options', 'DENY');

  // Prevent MIME sniffing
  res.setHeader('X-Content-Type-Options', 'nosniff');

  // HSTS (HTTPS only)
  res.setHeader('Strict-Transport-Security', 'max-age=31536000; includeSubDomains');

  // Referrer policy
  res.setHeader('Referrer-Policy', 'strict-origin-when-cross-origin');

  // Permissions policy
  res.setHeader('Permissions-Policy', 'geolocation=(), microphone=(), camera=()');

  next();
});
```

## Rate Limiting

```typescript
import rateLimit from 'express-rate-limit';

// General API rate limit
const apiLimiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100,
  message: { error: 'Too many requests' },
  standardHeaders: true,
  legacyHeaders: false,
});

app.use('/api/', apiLimiter);

// Strict limit for auth endpoints
const authLimiter = rateLimit({
  windowMs: 15 * 60 * 1000,
  max: 5,
  message: { error: 'Too many login attempts' },
  skipSuccessfulRequests: true,
});

app.post('/api/login', authLimiter, loginHandler);
app.post('/api/register', authLimiter, registerHandler);
```

## CORS Configuration

```typescript
import cors from 'cors';

// Strict CORS
app.use(cors({
  origin: ['https://example.com', 'https://app.example.com'],
  methods: ['GET', 'POST', 'PUT', 'DELETE'],
  allowedHeaders: ['Content-Type', 'Authorization'],
  credentials: true,
  maxAge: 86400, // Cache preflight for 24 hours
}));

// Dynamic origin validation
app.use(cors({
  origin: (origin, callback) => {
    const allowedOrigins = ['https://example.com'];
    if (!origin || allowedOrigins.includes(origin)) {
      callback(null, true);
    } else {
      callback(new Error('Not allowed by CORS'));
    }
  },
}));
```

## Cookie Security

```typescript
res.cookie('session', token, {
  httpOnly: true,      // No JavaScript access
  secure: true,        // HTTPS only
  sameSite: 'strict',  // CSRF protection
  maxAge: 900000,      // 15 minutes
  path: '/',
  domain: '.example.com',
});
```

## Quick Reference

| Header | Value | Purpose |
|--------|-------|---------|
| X-Frame-Options | DENY | Clickjacking |
| X-Content-Type-Options | nosniff | MIME sniffing |
| Strict-Transport-Security | max-age=31536000 | Force HTTPS |
| Content-Security-Policy | default-src 'self' | XSS |
| Referrer-Policy | strict-origin-when-cross-origin | Privacy |

| Cookie Flag | Purpose |
|-------------|---------|
| httpOnly | No JS access |
| secure | HTTPS only |
| sameSite=strict | CSRF protection |
| maxAge | Expiration |

---

## Reference: xss-csrf

# XSS & CSRF Prevention

## XSS Prevention

### Output Encoding

```typescript
// React automatically escapes by default
function SafeComponent({ userInput }: { userInput: string }) {
  return <div>{userInput}</div>; // Safe - auto-escaped
}

// If you must render HTML, sanitize first
import DOMPurify from 'dompurify';

function HtmlContent({ html }: { html: string }) {
  return (
    <div
      dangerouslySetInnerHTML={{
        __html: DOMPurify.sanitize(html)
      }}
    />
  );
}
```

### Content Security Policy

```typescript
import helmet from 'helmet';

app.use(helmet.contentSecurityPolicy({
  directives: {
    defaultSrc: ["'self'"],
    scriptSrc: ["'self'"],
    styleSrc: ["'self'", "'unsafe-inline'"],
    imgSrc: ["'self'", "data:", "https:"],
    connectSrc: ["'self'", "https://api.example.com"],
    fontSrc: ["'self'"],
    objectSrc: ["'none'"],
    frameSrc: ["'none'"],
    upgradeInsecureRequests: [],
  },
}));
```

### Input Sanitization

```typescript
import DOMPurify from 'dompurify';

// Sanitize HTML
const clean = DOMPurify.sanitize(dirty);

// Sanitize with config
const cleanStrict = DOMPurify.sanitize(dirty, {
  ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'a'],
  ALLOWED_ATTR: ['href'],
});

// Strip all HTML
const textOnly = DOMPurify.sanitize(dirty, { ALLOWED_TAGS: [] });
```

## CSRF Prevention

### Synchronizer Token Pattern

```typescript
import csrf from 'csurf';

const csrfProtection = csrf({ cookie: true });

// Add to forms
app.get('/form', csrfProtection, (req, res) => {
  res.render('form', { csrfToken: req.csrfToken() });
});

// Validate on submission
app.post('/submit', csrfProtection, (req, res) => {
  // Token validated automatically
});
```

### Double Submit Cookie

```typescript
// Set CSRF cookie
res.cookie('csrf', token, {
  httpOnly: false, // Must be readable by JS
  secure: true,
  sameSite: 'strict',
});

// Client sends in header
fetch('/api/action', {
  method: 'POST',
  headers: {
    'X-CSRF-Token': getCookie('csrf'),
  },
});

// Server validates
if (req.cookies.csrf !== req.headers['x-csrf-token']) {
  return res.status(403).json({ error: 'CSRF validation failed' });
}
```

### SameSite Cookies

```typescript
// Modern CSRF protection
app.use(session({
  cookie: {
    httpOnly: true,
    secure: true,
    sameSite: 'strict', // Or 'lax' for GET requests
  },
}));
```

## HTTP Headers

```typescript
// Security headers
app.use((req, res, next) => {
  // Prevent clickjacking
  res.setHeader('X-Frame-Options', 'DENY');

  // Prevent MIME sniffing
  res.setHeader('X-Content-Type-Options', 'nosniff');

  // XSS filter (legacy)
  res.setHeader('X-XSS-Protection', '1; mode=block');

  // Referrer policy
  res.setHeader('Referrer-Policy', 'strict-origin-when-cross-origin');

  next();
});
```

## Quick Reference

| Attack | Prevention |
|--------|------------|
| Reflected XSS | Output encoding |
| Stored XSS | Input sanitization + encoding |
| DOM XSS | Avoid innerHTML, use textContent |
| CSRF | Tokens + SameSite cookies |

| Header | Purpose |
|--------|---------|
| CSP | Script/resource restrictions |
| X-Frame-Options | Clickjacking |
| X-Content-Type-Options | MIME sniffing |
| SameSite | CSRF protection |
