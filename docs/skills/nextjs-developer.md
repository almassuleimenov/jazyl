# nextjs-developer

> Источник: https://github.com/jeffallan/claude-skills/tree/main/skills/nextjs-developer (MIT). Рабочая копия скилла: `.claude/skills/nextjs-developer/`.

---
name: nextjs-developer
description: "Use when building Next.js 14+ applications with App Router, server components, or server actions. Invoke to configure route handlers, implement middleware, set up API routes, add streaming SSR, write generateMetadata for SEO, scaffold loading.tsx/error.tsx boundaries, or deploy to Vercel. Triggers on: Next.js, Next.js 14, App Router, RSC, use server, Server Components, Server Actions, React Server Components, generateMetadata, loading.tsx, Next.js deployment, Vercel, Next.js performance."
license: MIT
metadata:
  author: https://github.com/Jeffallan
  company: https://synergetic.solutions
  version: "1.1.0"
  domain: frontend
  triggers: Next.js, Next.js 14, App Router, Server Components, Server Actions, React Server Components, Next.js deployment, Vercel, Next.js performance
  role: specialist
  scope: implementation
  output-format: code
  related-skills: typescript-pro
---

# Next.js Developer

Senior Next.js developer with expertise in Next.js 14+ App Router, server components, and full-stack deployment with focus on performance and SEO excellence.

## Core Workflow

1. **Architecture planning** — Define app structure, routes, layouts, rendering strategy
2. **Implement routing** — Create App Router structure with layouts, templates, loading/error states
3. **Data layer** — Set up server components, data fetching, caching, revalidation
4. **Optimize** — Images, fonts, bundles, streaming, edge runtime
5. **Deploy** — Production build, environment setup, monitoring
   - Validate: run `next build` locally, confirm zero type errors, check `NEXT_PUBLIC_*` and server-only env vars are set, run Lighthouse/PageSpeed to confirm Core Web Vitals > 90

## Reference Guide

Load detailed guidance based on context:

| Topic | Reference | Load When |
|-------|-----------|-----------|
| App Router | `references/app-router.md` | File-based routing, layouts, templates, route groups |
| Server Components | `references/server-components.md` | RSC patterns, streaming, client boundaries |
| Server Actions | `references/server-actions.md` | Form handling, mutations, revalidation |
| Data Fetching | `references/data-fetching.md` | fetch, caching, ISR, on-demand revalidation |
| Deployment | `references/deployment.md` | Vercel, self-hosting, Docker, optimization |

## Constraints

### MUST DO (Next.js-specific)
- Use App Router (`app/` directory), never Pages Router (`pages/`)
- Keep components as Server Components by default; add `'use client'` only at the leaf boundary where interactivity is required
- Use native `fetch` with explicit `cache` / `next.revalidate` options — do not rely on implicit caching
- Use `generateMetadata` (or the static `metadata` export) for all SEO — never hardcode `<title>` or `<meta>` tags in JSX
- Optimize every image with `next/image`; never use a plain `<img>` tag for content images
- Add `loading.tsx` and `error.tsx` at every route segment that performs async data fetching

### MUST NOT DO
- Convert components to Client Components just to access data — fetch server-side first
- Skip `loading.tsx`/`error.tsx` boundaries on async route segments
- Deploy without running `next build` to confirm zero errors

## Code Examples

### Server Component with data fetching and caching
```tsx
// app/products/page.tsx
import { Suspense } from 'react'

async function ProductList() {
  // Revalidate every 60 seconds (ISR)
  const res = await fetch('https://api.example.com/products', {
    next: { revalidate: 60 },
  })
  if (!res.ok) throw new Error('Failed to fetch products')
  const products: Product[] = await res.json()

  return (
    <ul>
      {products.map((p) => (
        <li key={p.id}>{p.name}</li>
      ))}
    </ul>
  )
}

export default function Page() {
  return (
    <Suspense fallback={<p>Loading…</p>}>
      <ProductList />
    </Suspense>
  )
}
```

### Server Action with form handling and revalidation
```tsx
// app/products/actions.ts
'use server'

import { revalidatePath } from 'next/cache'

export async function createProduct(formData: FormData) {
  const name = formData.get('name') as string
  await db.product.create({ data: { name } })
  revalidatePath('/products')
}

// app/products/new/page.tsx
import { createProduct } from '../actions'

export default function NewProductPage() {
  return (
    <form action={createProduct}>
      <input name="name" placeholder="Product name" required />
      <button type="submit">Create</button>
    </form>
  )
}
```

### generateMetadata for dynamic SEO
```tsx
// app/products/[id]/page.tsx
import type { Metadata } from 'next'

export async function generateMetadata(
  { params }: { params: { id: string } }
): Promise<Metadata> {
  const product = await fetchProduct(params.id)
  return {
    title: product.name,
    description: product.description,
    openGraph: { title: product.name, images: [product.imageUrl] },
  }
}
```

## Output Templates

When implementing Next.js features, provide:
1. App structure (route organization)
2. Layout/page components with proper data fetching
3. Server actions if mutations needed
4. Configuration (`next.config.js`, TypeScript)
5. Brief explanation of rendering strategy chosen

## Knowledge Reference

Next.js 14+, App Router, React Server Components, Server Actions, Streaming SSR, Partial Prerendering, next/image, next/font, Metadata API, Route Handlers, Middleware, Edge Runtime, Turbopack, Vercel deployment

Maintained by [@jeffallan](https://github.com/jeffallan), Principal Consultant at [Synergetic Solutions](https://synergetic.solutions)

[Documentation](https://jeffallan.github.io/claude-skills/skills/frontend/nextjs-developer/)

---

## Reference: app-router

# App Router Architecture

## File-Based Routing

```
app/
├── layout.tsx              # Root layout (required)
├── page.tsx               # Home page (/)
├── loading.tsx            # Loading UI
├── error.tsx              # Error boundary
├── not-found.tsx          # 404 page
├── template.tsx           # Re-mounted layout
│
├── (marketing)/           # Route group (no URL segment)
│   ├── layout.tsx
│   ├── about/
│   │   └── page.tsx      # /about
│   └── contact/
│       └── page.tsx      # /contact
│
├── dashboard/
│   ├── layout.tsx        # Shared dashboard layout
│   ├── page.tsx          # /dashboard
│   ├── settings/
│   │   └── page.tsx      # /dashboard/settings
│   └── @analytics/       # Parallel route (slot)
│       └── page.tsx
│
├── blog/
│   ├── [slug]/
│   │   └── page.tsx      # /blog/my-post (dynamic)
│   └── [...slug]/
│       └── page.tsx      # /blog/a/b/c (catch-all)
│
└── api/
    └── users/
        └── route.ts      # API route handler
```

## Root Layout (Required)

```tsx
// app/layout.tsx
import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: {
    default: 'My App',
    template: '%s | My App'
  },
  description: 'Next.js 14 application',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className={inter.className}>
        {children}
      </body>
    </html>
  )
}
```

## Nested Layouts

```tsx
// app/dashboard/layout.tsx
import { Sidebar } from '@/components/sidebar'
import { auth } from '@/lib/auth'
import { redirect } from 'next/navigation'

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const session = await auth()

  if (!session) {
    redirect('/login')
  }

  return (
    <div className="flex">
      <Sidebar />
      <main className="flex-1">{children}</main>
    </div>
  )
}
```

## Templates (Re-mount on Navigation)

```tsx
// app/template.tsx
'use client'

import { useEffect } from 'react'

export default function Template({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    // Runs on every navigation
    console.log('Template mounted')
  }, [])

  return <div>{children}</div>
}
```

## Loading States

```tsx
// app/dashboard/loading.tsx
export default function Loading() {
  return (
    <div className="flex items-center justify-center h-screen">
      <div className="animate-spin rounded-full h-32 w-32 border-b-2" />
    </div>
  )
}
```

## Error Boundaries

```tsx
// app/error.tsx
'use client'

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  return (
    <div>
      <h2>Something went wrong!</h2>
      <button onClick={() => reset()}>Try again</button>
    </div>
  )
}
```

## Route Groups

```tsx
// (marketing) and (shop) share the same URL level
app/
├── (marketing)/
│   ├── layout.tsx      # Marketing layout
│   └── about/
│       └── page.tsx    # /about
└── (shop)/
    ├── layout.tsx      # Shop layout
    └── products/
        └── page.tsx    # /products
```

## Parallel Routes

```tsx
// app/dashboard/layout.tsx
export default function Layout({
  children,
  analytics,
  team,
}: {
  children: React.ReactNode
  analytics: React.ReactNode
  team: React.ReactNode
}) {
  return (
    <>
      {children}
      {analytics}
      {team}
    </>
  )
}

// app/dashboard/@analytics/page.tsx
export default function Analytics() {
  return <div>Analytics Dashboard</div>
}
```

## Intercepting Routes

```tsx
// Show modal when navigating from same app
// but show full page on direct navigation

// app/photos/[id]/page.tsx (full page)
export default function PhotoPage({ params }: { params: { id: string } }) {
  return <div>Photo {params.id} - Full Page</div>
}

// app/@modal/(.)photos/[id]/page.tsx (modal)
export default function PhotoModal({ params }: { params: { id: string } }) {
  return <div>Photo {params.id} - Modal</div>
}
```

## Dynamic Routes

```tsx
// app/blog/[slug]/page.tsx
export default function BlogPost({ params }: { params: { slug: string } }) {
  return <h1>Post: {params.slug}</h1>
}

// Generate static params at build time
export async function generateStaticParams() {
  const posts = await fetch('https://api.example.com/posts').then(res => res.json())

  return posts.map((post: { slug: string }) => ({
    slug: post.slug,
  }))
}

// Opt out of static generation
export const dynamic = 'force-dynamic'

// Revalidate every 60 seconds
export const revalidate = 60
```

## Catch-All Routes

```tsx
// app/docs/[...slug]/page.tsx
// Matches: /docs/a, /docs/a/b, /docs/a/b/c
export default function Docs({ params }: { params: { slug: string[] } }) {
  return <div>Docs: {params.slug.join('/')}</div>
}

// Optional catch-all: [[...slug]]
// Also matches: /docs
```

## Route Handlers (API Routes)

```tsx
// app/api/users/route.ts
import { NextRequest, NextResponse } from 'next/server'

export async function GET(request: NextRequest) {
  const users = await db.user.findMany()
  return NextResponse.json(users)
}

export async function POST(request: NextRequest) {
  const body = await request.json()
  const user = await db.user.create({ data: body })
  return NextResponse.json(user, { status: 201 })
}

// Dynamic routes: app/api/users/[id]/route.ts
export async function GET(
  request: NextRequest,
  { params }: { params: { id: string } }
) {
  const user = await db.user.findUnique({ where: { id: params.id } })
  return NextResponse.json(user)
}
```

## Metadata API

```tsx
// app/blog/[slug]/page.tsx
import type { Metadata } from 'next'

export async function generateMetadata(
  { params }: { params: { slug: string } }
): Promise<Metadata> {
  const post = await fetchPost(params.slug)

  return {
    title: post.title,
    description: post.excerpt,
    openGraph: {
      title: post.title,
      description: post.excerpt,
      images: [{ url: post.coverImage }],
    },
  }
}
```

## Quick Reference

| File | Purpose | Use Case |
|------|---------|----------|
| `layout.tsx` | Persistent UI across routes | Shared navigation, auth wrapper |
| `page.tsx` | Route UI | Actual page content |
| `loading.tsx` | Loading fallback | Automatic Suspense boundary |
| `error.tsx` | Error boundary | Handle errors gracefully |
| `template.tsx` | Re-mounted layout | Analytics, animations |
| `not-found.tsx` | 404 page | Custom not found UI |
| `route.ts` | API handler | Backend API endpoints |

---

## Reference: data-fetching

# Data Fetching & Caching

## Extended fetch API

Next.js extends the native fetch with caching and revalidation options:

```tsx
// app/page.tsx
async function getData() {
  const res = await fetch('https://api.example.com/posts', {
    cache: 'force-cache', // Default: cache forever (SSG)
  })

  if (!res.ok) {
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

export default async function Page() {
  const data = await getData()
  return <div>{/* render data */}</div>
}
```

## Cache Options

```tsx
// 1. Force cache (Static Site Generation)
fetch('https://api.example.com/data', {
  cache: 'force-cache' // Default behavior
})

// 2. No cache (Server-Side Rendering)
fetch('https://api.example.com/data', {
  cache: 'no-store' // Always fetch fresh data
})

// 3. Revalidate (Incremental Static Regeneration)
fetch('https://api.example.com/data', {
  next: { revalidate: 3600 } // Revalidate every hour
})

// 4. Revalidate with tags
fetch('https://api.example.com/data', {
  next: { tags: ['posts'] }
})
```

## Revalidation Methods

### Time-based Revalidation (ISR)

```tsx
// Revalidate every 60 seconds
async function getPosts() {
  const res = await fetch('https://api.example.com/posts', {
    next: { revalidate: 60 }
  })
  return res.json()
}

// Route segment config
export const revalidate = 60 // seconds

export default async function Page() {
  const posts = await getPosts()
  return <div>{/* render */}</div>
}
```

### On-Demand Revalidation

```tsx
// app/api/revalidate/route.ts
import { revalidatePath, revalidateTag } from 'next/cache'
import { NextRequest } from 'next/server'

export async function POST(request: NextRequest) {
  const path = request.nextUrl.searchParams.get('path')

  if (path) {
    revalidatePath(path)
    return Response.json({ revalidated: true, now: Date.now() })
  }

  return Response.json({ revalidated: false })
}

// Usage in Server Action
'use server'

import { revalidatePath } from 'next/cache'

export async function createPost(data: FormData) {
  await db.post.create({ data })

  // Revalidate specific path
  revalidatePath('/posts')

  // Revalidate entire layout
  revalidatePath('/posts', 'layout')
}
```

### Tag-based Revalidation

```tsx
// Fetch with tags
async function getPosts() {
  const res = await fetch('https://api.example.com/posts', {
    next: { tags: ['posts'] }
  })
  return res.json()
}

async function getAuthors() {
  const res = await fetch('https://api.example.com/authors', {
    next: { tags: ['authors'] }
  })
  return res.json()
}

// Revalidate by tag
import { revalidateTag } from 'next/cache'

export async function createPost() {
  // Revalidate all fetches tagged with 'posts'
  revalidateTag('posts')
}
```

## Route Segment Config

```tsx
// app/posts/page.tsx

// Force dynamic rendering
export const dynamic = 'force-dynamic' // 'auto' | 'force-dynamic' | 'error' | 'force-static'

// Revalidation interval
export const revalidate = 3600 // false | 0 | number (seconds)

// Fetch cache
export const fetchCache = 'auto' // 'auto' | 'default-cache' | 'only-cache' | 'force-cache' | 'force-no-store' | 'default-no-store' | 'only-no-store'

// Runtime
export const runtime = 'nodejs' // 'nodejs' | 'edge'

// Preferred region
export const preferredRegion = 'auto' // 'auto' | 'home' | 'edge' | string | string[]

export default async function Page() {
  return <div>Posts</div>
}
```

## Parallel Data Fetching

```tsx
async function getUser() {
  return fetch('https://api.example.com/user')
}

async function getPosts() {
  return fetch('https://api.example.com/posts')
}

async function getComments() {
  return fetch('https://api.example.com/comments')
}

export default async function Page() {
  // Fetch in parallel with Promise.all
  const [user, posts, comments] = await Promise.all([
    getUser(),
    getPosts(),
    getComments(),
  ])

  return (
    <div>
      <UserInfo user={user} />
      <Posts posts={posts} />
      <Comments comments={comments} />
    </div>
  )
}
```

## Sequential Data Fetching

```tsx
// When one fetch depends on another
export default async function Page({ params }: { params: { id: string } }) {
  // First fetch
  const user = await fetch(`https://api.example.com/users/${params.id}`)
    .then(res => res.json())

  // Second fetch depends on first
  const posts = await fetch(`https://api.example.com/users/${user.id}/posts`)
    .then(res => res.json())

  return (
    <div>
      <h1>{user.name}</h1>
      <Posts posts={posts} />
    </div>
  )
}
```

## Streaming with Suspense

```tsx
// app/page.tsx
import { Suspense } from 'react'

async function Posts() {
  const posts = await fetch('https://api.example.com/posts', {
    cache: 'no-store'
  }).then(res => res.json())

  return (
    <ul>
      {posts.map((post: Post) => (
        <li key={post.id}>{post.title}</li>
      ))}
    </ul>
  )
}

export default function Page() {
  return (
    <div>
      <h1>Posts</h1>
      <Suspense fallback={<div>Loading posts...</div>}>
        <Posts />
      </Suspense>
    </div>
  )
}
```

## React cache for Deduplication

```tsx
// lib/data.ts
import { cache } from 'react'

export const getUser = cache(async (id: string) => {
  const res = await fetch(`https://api.example.com/users/${id}`)
  return res.json()
})

// components/user-profile.tsx
export async function UserProfile({ userId }: { userId: string }) {
  const user = await getUser(userId) // Cached
  return <div>{user.name}</div>
}

// components/user-posts.tsx
export async function UserPosts({ userId }: { userId: string }) {
  const user = await getUser(userId) // Uses cached result
  return <div>{user.posts.length} posts</div>
}

// app/page.tsx
export default function Page() {
  return (
    <>
      <UserProfile userId="123" />
      <UserPosts userId="123" /> {/* Same fetch, deduplicated */}
    </>
  )
}
```

## Database Queries

```tsx
// lib/db.ts
import { PrismaClient } from '@prisma/client'

const globalForPrisma = global as unknown as { prisma: PrismaClient }

export const db = globalForPrisma.prisma || new PrismaClient()

if (process.env.NODE_ENV !== 'production') globalForPrisma.prisma = db

// app/posts/page.tsx
import { db } from '@/lib/db'

export const revalidate = 60 // Revalidate every 60 seconds

export default async function PostsPage() {
  const posts = await db.post.findMany({
    include: { author: true },
    orderBy: { createdAt: 'desc' },
  })

  return (
    <div>
      {posts.map(post => (
        <article key={post.id}>
          <h2>{post.title}</h2>
          <p>By {post.author.name}</p>
        </article>
      ))}
    </div>
  )
}
```

## Error Handling

```tsx
async function getData() {
  const res = await fetch('https://api.example.com/data')

  if (!res.ok) {
    // This will activate the closest error.tsx
    throw new Error('Failed to fetch data')
  }

  return res.json()
}

export default async function Page() {
  const data = await getData()
  return <div>{data.title}</div>
}

// app/error.tsx
'use client'

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  return (
    <div>
      <h2>Something went wrong!</h2>
      <button onClick={() => reset()}>Try again</button>
    </div>
  )
}
```

## Loading States

```tsx
// app/posts/loading.tsx
export default function Loading() {
  return <div>Loading posts...</div>
}

// app/posts/page.tsx
export default async function PostsPage() {
  const posts = await fetch('https://api.example.com/posts')
    .then(res => res.json())

  return <div>{/* render posts */}</div>
}
```

## Client-Side Data Fetching

```tsx
// When you need client-side fetching
'use client'

import useSWR from 'swr'

const fetcher = (url: string) => fetch(url).then(res => res.json())

export function Posts() {
  const { data, error, isLoading } = useSWR('/api/posts', fetcher, {
    refreshInterval: 3000, // Refresh every 3 seconds
  })

  if (error) return <div>Failed to load</div>
  if (isLoading) return <div>Loading...</div>

  return (
    <ul>
      {data.map((post: Post) => (
        <li key={post.id}>{post.title}</li>
      ))}
    </ul>
  )
}
```

## Preloading Data

```tsx
// lib/data.ts
import { cache } from 'react'

export const preload = (id: string) => {
  void getUser(id) // Trigger fetch without awaiting
}

export const getUser = cache(async (id: string) => {
  return fetch(`https://api.example.com/users/${id}`)
    .then(res => res.json())
})

// components/user.tsx
import { getUser, preload } from '@/lib/data'

export async function User({ id }: { id: string }) {
  const user = await getUser(id)
  return <div>{user.name}</div>
}

// app/page.tsx
import { User } from '@/components/user'
import { preload } from '@/lib/data'

export default async function Page() {
  preload('123') // Start loading immediately
  return <User id="123" />
}
```

## Static Generation with Dynamic Routes

```tsx
// app/posts/[slug]/page.tsx
type Post = {
  slug: string
  title: string
  content: string
}

export async function generateStaticParams() {
  const posts = await fetch('https://api.example.com/posts')
    .then(res => res.json())

  return posts.map((post: Post) => ({
    slug: post.slug,
  }))
}

export default async function Post({ params }: { params: { slug: string } }) {
  const post = await fetch(`https://api.example.com/posts/${params.slug}`)
    .then(res => res.json())

  return (
    <article>
      <h1>{post.title}</h1>
      <div>{post.content}</div>
    </article>
  )
}
```

## Quick Reference

| Strategy | Config | Use Case |
|----------|--------|----------|
| **SSG** | `cache: 'force-cache'` | Static content |
| **SSR** | `cache: 'no-store'` | Always fresh data |
| **ISR** | `next: { revalidate: 60 }` | Periodic updates |
| **Tag-based** | `next: { tags: ['posts'] }` | On-demand revalidation |
| **Dynamic** | `export const dynamic = 'force-dynamic'` | Per-request data |

## Best Practices

1. **Default to caching** - Use force-cache for static content
2. **Use ISR** - Revalidate periodically for semi-dynamic content
3. **Parallel fetching** - Use Promise.all for independent requests
4. **Deduplicate** - Use React cache() for repeated calls
5. **Stream with Suspense** - Show content progressively
6. **Tag your fetches** - Enable granular revalidation
7. **Handle errors** - Use error.tsx for graceful degradation

---

## Reference: deployment

# Deployment & Production

## Vercel Deployment (Recommended)

### Quick Deploy

```bash
# Install Vercel CLI
npm i -g vercel

# Deploy
vercel

# Production deployment
vercel --prod
```

### vercel.json Configuration

```json
{
  "buildCommand": "next build",
  "devCommand": "next dev",
  "installCommand": "npm install",
  "framework": "nextjs",
  "regions": ["iad1"],
  "env": {
    "DATABASE_URL": "@database-url",
    "NEXT_PUBLIC_API_URL": "https://api.example.com"
  },
  "headers": [
    {
      "source": "/api/(.*)",
      "headers": [
        { "key": "Access-Control-Allow-Origin", "value": "*" },
        { "key": "Access-Control-Allow-Methods", "value": "GET,POST,PUT,DELETE" }
      ]
    }
  ],
  "redirects": [
    {
      "source": "/old-blog/:slug",
      "destination": "/blog/:slug",
      "permanent": true
    }
  ],
  "rewrites": [
    {
      "source": "/api/:path*",
      "destination": "https://api.example.com/:path*"
    }
  ]
}
```

### Environment Variables

```bash
# .env.local (not committed)
DATABASE_URL="postgresql://<db-user>:<db-password>@localhost:5432/db"
NEXTAUTH_SECRET="<nextauth-secret>"

# .env.production (committed, public vars only)
NEXT_PUBLIC_API_URL="https://api.example.com"
```

```tsx
// Access in Server Components
const dbUrl = process.env.DATABASE_URL

// Access in Client Components (must be prefixed with NEXT_PUBLIC_)
const apiUrl = process.env.NEXT_PUBLIC_API_URL
```

## Self-Hosting

### Standalone Output

```js
// next.config.js
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
}

module.exports = nextConfig
```

```bash
# Build
npm run build

# The standalone folder contains everything needed
# Copy these to your server:
# - .next/standalone/
# - .next/static/
# - public/

# Run on server
node .next/standalone/server.js
```

### Node.js Server

```bash
# Build
npm run build

# Start production server
npm start

# With PM2 for process management
pm2 start npm --name "nextjs" -- start
pm2 startup
pm2 save
```

## Docker Deployment

### Dockerfile (Multi-stage)

```dockerfile
# Stage 1: Dependencies
FROM node:20-alpine AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

# Stage 2: Builder
FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

ENV NEXT_TELEMETRY_DISABLED 1

RUN npm run build

# Stage 3: Runner
FROM node:20-alpine AS runner
WORKDIR /app

ENV NODE_ENV production
ENV NEXT_TELEMETRY_DISABLED 1

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT 3000
ENV HOSTNAME "0.0.0.0"

CMD ["node", "server.js"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  nextjs:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - DATABASE_URL=postgresql://postgres:<db-password>@db:5432/myapp
      - NEXTAUTH_URL=http://localhost:3000
      - NEXTAUTH_SECRET=<nextauth-secret>
    depends_on:
      - db
    restart: unless-stopped

  db:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=<db-password>
      - POSTGRES_DB=myapp
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

volumes:
  postgres_data:
```

```bash
# Build and run
docker-compose up -d

# View logs
docker-compose logs -f nextjs

# Rebuild
docker-compose up -d --build
```

## Production Optimization

### next.config.js

```js
/** @type {import('next').NextConfig} */
const nextConfig = {
  // Standalone for self-hosting
  output: 'standalone',

  // Image optimization
  images: {
    formats: ['image/avif', 'image/webp'],
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'cdn.example.com',
        pathname: '/images/**',
      },
    ],
    deviceSizes: [640, 750, 828, 1080, 1200, 1920, 2048, 3840],
    imageSizes: [16, 32, 48, 64, 96, 128, 256, 384],
  },

  // Compression
  compress: true,

  // Security headers
  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          {
            key: 'X-DNS-Prefetch-Control',
            value: 'on'
          },
          {
            key: 'Strict-Transport-Security',
            value: 'max-age=63072000; includeSubDomains; preload'
          },
          {
            key: 'X-Frame-Options',
            value: 'SAMEORIGIN'
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff'
          },
          {
            key: 'X-XSS-Protection',
            value: '1; mode=block'
          },
          {
            key: 'Referrer-Policy',
            value: 'origin-when-cross-origin'
          },
        ],
      },
    ]
  },

  // Experimental features
  experimental: {
    optimizePackageImports: ['@mui/material', 'lodash'],
  },

  // Bundle analyzer
  webpack: (config, { isServer }) => {
    if (process.env.ANALYZE === 'true') {
      const { BundleAnalyzerPlugin } = require('webpack-bundle-analyzer')
      config.plugins.push(
        new BundleAnalyzerPlugin({
          analyzerMode: 'static',
          reportFilename: isServer
            ? '../analyze/server.html'
            : './analyze/client.html',
        })
      )
    }
    return config
  },
}

module.exports = nextConfig
```

### Bundle Analysis

```bash
# Install analyzer
npm install -D @next/bundle-analyzer

# Analyze
ANALYZE=true npm run build

# Or use built-in
npm run build -- --experimental-build-mode=compile
```

### Performance Monitoring

```tsx
// app/layout.tsx
import { SpeedInsights } from '@vercel/speed-insights/next'
import { Analytics } from '@vercel/analytics/react'

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html>
      <body>
        {children}
        <SpeedInsights />
        <Analytics />
      </body>
    </html>
  )
}
```

## CDN & Edge

### Static Asset CDN

```js
// next.config.js
const nextConfig = {
  assetPrefix: process.env.NODE_ENV === 'production'
    ? 'https://cdn.example.com'
    : '',
}
```

### Edge Runtime

```tsx
// app/api/edge/route.ts
export const runtime = 'edge'

export async function GET(request: Request) {
  return new Response('Hello from Edge!', {
    status: 200,
    headers: {
      'content-type': 'text/plain',
    },
  })
}

// app/page.tsx
export const runtime = 'edge'

export default async function Page() {
  return <div>Edge-rendered page</div>
}
```

## Caching Strategy

### ISR (Incremental Static Regeneration)

```tsx
// app/blog/[slug]/page.tsx
export const revalidate = 3600 // Revalidate every hour

export default async function BlogPost({ params }: { params: { slug: string } }) {
  const post = await fetchPost(params.slug)
  return <article>{post.content}</article>
}
```

### On-Demand Revalidation

```tsx
// app/api/revalidate/route.ts
import { revalidatePath } from 'next/cache'
import { NextRequest } from 'next/server'

export async function POST(request: NextRequest) {
  const secret = request.nextUrl.searchParams.get('secret')

  if (secret !== process.env.REVALIDATE_SECRET) {
    return Response.json({ message: 'Invalid secret' }, { status: 401 })
  }

  const path = request.nextUrl.searchParams.get('path') || '/'

  revalidatePath(path)

  return Response.json({ revalidated: true, now: Date.now() })
}
```

## Database Connection Pooling

```ts
// lib/db.ts
import { PrismaClient } from '@prisma/client'

const globalForPrisma = global as unknown as {
  prisma: PrismaClient | undefined
}

export const db =
  globalForPrisma.prisma ??
  new PrismaClient({
    log: process.env.NODE_ENV === 'development' ? ['query', 'error', 'warn'] : ['error'],
  })

if (process.env.NODE_ENV !== 'production') globalForPrisma.prisma = db
```

## Health Check Endpoint

```tsx
// app/api/health/route.ts
import { db } from '@/lib/db'

export async function GET() {
  try {
    // Check database connection
    await db.$queryRaw`SELECT 1`

    return Response.json({
      status: 'ok',
      timestamp: new Date().toISOString(),
      uptime: process.uptime(),
    })
  } catch (error) {
    return Response.json(
      {
        status: 'error',
        message: 'Database connection failed',
      },
      { status: 503 }
    )
  }
}
```

## CI/CD with GitHub Actions

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '20'
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Run tests
        run: npm test

      - name: Build
        run: npm run build
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
          NEXTAUTH_SECRET: ${{ secrets.NEXTAUTH_SECRET }}

      - name: Deploy to Vercel
        uses: amondnet/vercel-action@v25
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
          vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
          vercel-args: '--prod'
```

## Monitoring & Logging

```tsx
// app/error.tsx
'use client'

import * as Sentry from '@sentry/nextjs'
import { useEffect } from 'react'

export default function Error({
  error,
}: {
  error: Error & { digest?: string }
}) {
  useEffect(() => {
    Sentry.captureException(error)
  }, [error])

  return <div>Something went wrong!</div>
}
```

## Quick Reference

| Platform | Best For | Effort |
|----------|----------|--------|
| **Vercel** | Zero-config, optimal performance | Low |
| **Netlify** | Alternative to Vercel | Low |
| **Railway** | Simple hosting with databases | Medium |
| **AWS/GCP** | Enterprise, custom needs | High |
| **Docker** | Self-hosting, full control | High |

## Production Checklist

- [ ] Enable TypeScript strict mode
- [ ] Configure CSP headers
- [ ] Setup error monitoring (Sentry)
- [ ] Configure analytics (Vercel/GA)
- [ ] Optimize images (next/image)
- [ ] Enable compression
- [ ] Setup CDN for static assets
- [ ] Configure database connection pooling
- [ ] Add health check endpoint
- [ ] Setup CI/CD pipeline
- [ ] Configure environment variables
- [ ] Enable ISR/SSG where possible
- [ ] Test Core Web Vitals
- [ ] Setup logging (Datadog/LogRocket)
- [ ] Configure backup strategy

---

## Reference: server-actions

# Server Actions

## Basic Server Action

```tsx
// app/actions.ts
'use server'

import { db } from '@/lib/db'
import { revalidatePath } from 'next/cache'

export async function createPost(formData: FormData) {
  const title = formData.get('title') as string
  const content = formData.get('content') as string

  await db.post.create({
    data: { title, content }
  })

  revalidatePath('/posts')
}
```

## Form with Server Action

```tsx
// app/posts/new/page.tsx
import { createPost } from '@/app/actions'

export default function NewPost() {
  return (
    <form action={createPost}>
      <input name="title" required />
      <textarea name="content" required />
      <button type="submit">Create Post</button>
    </form>
  )
}
```

## Server Action with Validation

```tsx
// app/actions.ts
'use server'

import { z } from 'zod'
import { revalidatePath } from 'next/cache'

const CreatePostSchema = z.object({
  title: z.string().min(3).max(100),
  content: z.string().min(10),
})

export async function createPost(formData: FormData) {
  const validatedFields = CreatePostSchema.safeParse({
    title: formData.get('title'),
    content: formData.get('content'),
  })

  if (!validatedFields.success) {
    return {
      errors: validatedFields.error.flatten().fieldErrors,
    }
  }

  const { title, content } = validatedFields.data

  await db.post.create({
    data: { title, content }
  })

  revalidatePath('/posts')
  return { success: true }
}
```

## Client Component with Server Action

```tsx
// components/create-post-form.tsx
'use client'

import { createPost } from '@/app/actions'
import { useFormState, useFormStatus } from 'react-dom'

const initialState = {
  errors: {},
}

function SubmitButton() {
  const { pending } = useFormStatus()

  return (
    <button type="submit" disabled={pending}>
      {pending ? 'Creating...' : 'Create Post'}
    </button>
  )
}

export function CreatePostForm() {
  const [state, formAction] = useFormState(createPost, initialState)

  return (
    <form action={formAction}>
      <div>
        <input name="title" />
        {state.errors?.title && <p>{state.errors.title[0]}</p>}
      </div>

      <div>
        <textarea name="content" />
        {state.errors?.content && <p>{state.errors.content[0]}</p>}
      </div>

      <SubmitButton />
    </form>
  )
}
```

## Server Action with Redirect

```tsx
// app/actions.ts
'use server'

import { redirect } from 'next/navigation'
import { revalidatePath } from 'next/cache'

export async function createPost(formData: FormData) {
  const post = await db.post.create({
    data: {
      title: formData.get('title') as string,
      content: formData.get('content') as string,
    }
  })

  revalidatePath('/posts')
  redirect(`/posts/${post.id}`)
}
```

## Optimistic Updates

```tsx
// components/todo-list.tsx
'use client'

import { experimental_useOptimistic as useOptimistic } from 'react'
import { toggleTodo } from '@/app/actions'

export function TodoList({ todos }: { todos: Todo[] }) {
  const [optimisticTodos, addOptimisticTodo] = useOptimistic(
    todos,
    (state, newTodo: Todo) => [...state, newTodo]
  )

  async function handleSubmit(formData: FormData) {
    const title = formData.get('title') as string
    const newTodo = { id: crypto.randomUUID(), title, completed: false }

    // Optimistically update UI
    addOptimisticTodo(newTodo)

    // Send to server
    await createTodo(formData)
  }

  return (
    <div>
      <ul>
        {optimisticTodos.map(todo => (
          <li key={todo.id}>{todo.title}</li>
        ))}
      </ul>

      <form action={handleSubmit}>
        <input name="title" />
        <button type="submit">Add</button>
      </form>
    </div>
  )
}
```

## Server Action with Authentication

```tsx
// app/actions.ts
'use server'

import { auth } from '@/lib/auth'
import { redirect } from 'next/navigation'

export async function createPost(formData: FormData) {
  const session = await auth()

  if (!session) {
    redirect('/login')
  }

  await db.post.create({
    data: {
      title: formData.get('title') as string,
      content: formData.get('content') as string,
      authorId: session.user.id,
    }
  })

  revalidatePath('/posts')
}
```

## Inline Server Action

```tsx
// app/posts/page.tsx
import { db } from '@/lib/db'
import { revalidatePath } from 'next/cache'

export default async function Posts() {
  const posts = await db.post.findMany()

  async function deletePost(formData: FormData) {
    'use server'

    const id = formData.get('id') as string
    await db.post.delete({ where: { id } })
    revalidatePath('/posts')
  }

  return (
    <ul>
      {posts.map(post => (
        <li key={post.id}>
          {post.title}
          <form action={deletePost}>
            <input type="hidden" name="id" value={post.id} />
            <button type="submit">Delete</button>
          </form>
        </li>
      ))}
    </ul>
  )
}
```

## Programmatic Server Action Call

```tsx
// components/delete-button.tsx
'use client'

import { deletePost } from '@/app/actions'

export function DeleteButton({ postId }: { postId: string }) {
  async function handleDelete() {
    if (confirm('Are you sure?')) {
      await deletePost(postId)
    }
  }

  return (
    <button onClick={handleDelete}>
      Delete
    </button>
  )
}

// app/actions.ts
'use server'

export async function deletePost(postId: string) {
  await db.post.delete({ where: { id: postId } })
  revalidatePath('/posts')
}
```

## Revalidation Strategies

```tsx
// app/actions.ts
'use server'

import { revalidatePath, revalidateTag } from 'next/cache'

export async function updatePost(id: string, data: UpdatePostData) {
  await db.post.update({ where: { id }, data })

  // Revalidate specific path
  revalidatePath('/posts')
  revalidatePath(`/posts/${id}`)

  // Revalidate all paths in a layout
  revalidatePath('/posts', 'layout')

  // Revalidate by cache tag
  revalidateTag('posts')
}
```

## Server Action with File Upload

```tsx
// app/actions.ts
'use server'

import { writeFile } from 'fs/promises'
import { join } from 'path'

export async function uploadAvatar(formData: FormData) {
  const file = formData.get('avatar') as File

  if (!file) {
    return { error: 'No file uploaded' }
  }

  const bytes = await file.arrayBuffer()
  const buffer = Buffer.from(bytes)

  const path = join(process.cwd(), 'public', 'uploads', file.name)
  await writeFile(path, buffer)

  return { success: true, path: `/uploads/${file.name}` }
}

// components/upload-form.tsx
'use client'

import { uploadAvatar } from '@/app/actions'

export function UploadForm() {
  async function handleSubmit(formData: FormData) {
    const result = await uploadAvatar(formData)
    if (result.success) {
      console.log('Uploaded to:', result.path)
    }
  }

  return (
    <form action={handleSubmit}>
      <input type="file" name="avatar" accept="image/*" />
      <button type="submit">Upload</button>
    </form>
  )
}
```

## Error Handling

```tsx
// app/actions.ts
'use server'

export async function createPost(formData: FormData) {
  try {
    await db.post.create({
      data: {
        title: formData.get('title') as string,
        content: formData.get('content') as string,
      }
    })

    revalidatePath('/posts')
    return { success: true }
  } catch (error) {
    console.error('Failed to create post:', error)
    return { error: 'Failed to create post' }
  }
}

// components/form.tsx
'use client'

export function CreatePostForm() {
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(formData: FormData) {
    const result = await createPost(formData)

    if (result.error) {
      setError(result.error)
    } else {
      // Success
      router.push('/posts')
    }
  }

  return (
    <form action={handleSubmit}>
      {error && <div className="error">{error}</div>}
      {/* form fields */}
    </form>
  )
}
```

## Server Action with Cookies

```tsx
// app/actions.ts
'use server'

import { cookies } from 'next/headers'

export async function setTheme(theme: 'light' | 'dark') {
  cookies().set('theme', theme, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    maxAge: 60 * 60 * 24 * 365, // 1 year
    path: '/',
  })
}

export async function getTheme() {
  return cookies().get('theme')?.value ?? 'light'
}
```

## Rate Limiting

```tsx
// app/actions.ts
'use server'

import { ratelimit } from '@/lib/redis'

export async function createPost(formData: FormData) {
  const session = await auth()
  const { success } = await ratelimit.limit(session.user.id)

  if (!success) {
    return { error: 'Rate limit exceeded' }
  }

  // Create post...
}
```

## Quick Reference

| Capability | Usage |
|------------|-------|
| **Define** | Add 'use server' at top of file or function |
| **Form** | Pass action to `<form action={serverAction}>` |
| **Programmatic** | Call directly: `await serverAction(data)` |
| **Validation** | Use Zod/TypeBox before mutations |
| **Revalidate** | `revalidatePath()` or `revalidateTag()` |
| **Redirect** | `redirect()` after mutation |
| **Errors** | Return error objects, handle in client |
| **Files** | Access via `formData.get()` as File |

## Best Practices

1. **Always validate** - Use Zod/TypeBox for type-safe validation
2. **Revalidate** - Call revalidatePath() after mutations
3. **Handle errors** - Return error objects instead of throwing
4. **Auth checks** - Verify session before mutations
5. **Rate limiting** - Protect against abuse
6. **Type safety** - Define input/output types
7. **Optimistic updates** - Use useOptimistic for better UX

---

## Reference: server-components

# React Server Components

## Server Components (Default)

```tsx
// app/page.tsx - Server Component by default
import { db } from '@/lib/db'

export default async function Page() {
  // Data fetching in Server Component
  const users = await db.user.findMany()

  return (
    <div>
      <h1>Users</h1>
      <ul>
        {users.map(user => (
          <li key={user.id}>{user.name}</li>
        ))}
      </ul>
    </div>
  )
}
```

## Benefits of Server Components

- **Zero bundle size** - Server Components don't add JavaScript to client bundle
- **Direct backend access** - Query databases, read files, use secrets
- **Automatic code splitting** - Only Client Components add to bundle
- **Streaming** - Send UI progressively as data loads
- **No client-side waterfalls** - Fetch all data in parallel on server

## Client Components

```tsx
// components/counter.tsx
'use client' // Required directive

import { useState } from 'react'

export function Counter() {
  const [count, setCount] = useState(0)

  return (
    <button onClick={() => setCount(count + 1)}>
      Count: {count}
    </button>
  )
}
```

## When to Use Client Components

Use `'use client'` when you need:
- **Interactivity** - onClick, onChange, event handlers
- **State** - useState, useReducer
- **Effects** - useEffect, useLayoutEffect
- **Browser APIs** - localStorage, window, document
- **Custom hooks** - Any hook using client-only features
- **Class components** - Component lifecycle methods

## Composition Pattern

```tsx
// app/page.tsx - Server Component
import { ClientWrapper } from './client-wrapper'
import { db } from '@/lib/db'

export default async function Page() {
  const data = await db.query()

  return (
    <div>
      {/* Server Component content */}
      <h1>Server Content</h1>

      {/* Pass data to Client Component */}
      <ClientWrapper initialData={data}>
        {/* Server Component as children */}
        <ServerSidebar />
      </ClientWrapper>
    </div>
  )
}

// components/client-wrapper.tsx
'use client'

export function ClientWrapper({
  children,
  initialData,
}: {
  children: React.ReactNode
  initialData: Data
}) {
  const [data, setData] = useState(initialData)

  return (
    <div>
      {/* Client Component UI */}
      <button onClick={() => refresh()}>Refresh</button>
      {/* Server Component children */}
      {children}
    </div>
  )
}
```

## Streaming with Suspense

```tsx
// app/page.tsx
import { Suspense } from 'react'
import { SlowComponent } from './slow-component'
import { FastComponent } from './fast-component'

export default function Page() {
  return (
    <div>
      {/* Renders immediately */}
      <FastComponent />

      {/* Shows fallback while loading */}
      <Suspense fallback={<div>Loading...</div>}>
        <SlowComponent />
      </Suspense>
    </div>
  )
}

// components/slow-component.tsx
async function getData() {
  await new Promise(resolve => setTimeout(resolve, 3000))
  return { data: 'Loaded!' }
}

export async function SlowComponent() {
  const data = await getData()
  return <div>{data.data}</div>
}
```

## Parallel Data Fetching

```tsx
// app/dashboard/page.tsx
async function getUser() {
  return fetch('https://api.example.com/user')
}

async function getPosts() {
  return fetch('https://api.example.com/posts')
}

export default async function Dashboard() {
  // Fetch in parallel
  const [user, posts] = await Promise.all([
    getUser(),
    getPosts(),
  ])

  return (
    <div>
      <UserProfile user={user} />
      <PostsList posts={posts} />
    </div>
  )
}
```

## Sequential Data Fetching

```tsx
// app/artist/[id]/page.tsx
async function getArtist(id: string) {
  return fetch(`https://api.example.com/artists/${id}`)
}

async function getAlbums(artistId: string) {
  return fetch(`https://api.example.com/artists/${artistId}/albums`)
}

export default async function ArtistPage({ params }: { params: { id: string } }) {
  // Sequential: albums depends on artist
  const artist = await getArtist(params.id)
  const albums = await getAlbums(artist.id)

  return (
    <div>
      <h1>{artist.name}</h1>
      <Albums albums={albums} />
    </div>
  )
}
```

## Preloading Data

```tsx
// lib/data.ts
import { cache } from 'react'

export const getUser = cache(async (id: string) => {
  return db.user.findUnique({ where: { id } })
})

// components/user-profile.tsx
export async function UserProfile({ userId }: { userId: string }) {
  const user = await getUser(userId)
  return <div>{user.name}</div>
}

// app/page.tsx
import { getUser } from '@/lib/data'
import { UserProfile } from '@/components/user-profile'

export default async function Page() {
  // Preload
  getUser('123')

  return (
    <div>
      {/* This will use cached result */}
      <UserProfile userId="123" />
    </div>
  )
}
```

## Server Component Patterns

### Pattern: Layout with Data Fetching

```tsx
// app/dashboard/layout.tsx
import { auth } from '@/lib/auth'
import { db } from '@/lib/db'

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const session = await auth()
  const user = await db.user.findUnique({ where: { id: session.userId } })

  return (
    <div>
      <Sidebar user={user} />
      <main>{children}</main>
    </div>
  )
}
```

### Pattern: Conditional Client Components

```tsx
// app/page.tsx
import { ClientComponent } from './client-component'

export default async function Page() {
  const data = await fetchData()

  // Only render Client Component when needed
  if (data.requiresInteractivity) {
    return <ClientComponent data={data} />
  }

  return <div>{data.content}</div>
}
```

### Pattern: Server Component with Client Island

```tsx
// app/blog/[slug]/page.tsx
import { LikeButton } from './like-button'

export default async function BlogPost({ params }: { params: { slug: string } }) {
  const post = await getPost(params.slug)

  return (
    <article>
      {/* Server-rendered content */}
      <h1>{post.title}</h1>
      <div dangerouslySetInnerHTML={{ __html: post.content }} />

      {/* Client island for interactivity */}
      <LikeButton postId={post.id} initialLikes={post.likes} />
    </article>
  )
}
```

## Context in Server/Client Components

```tsx
// app/providers.tsx
'use client'

import { ThemeProvider } from 'next-themes'

export function Providers({ children }: { children: React.ReactNode }) {
  return <ThemeProvider>{children}</ThemeProvider>
}

// app/layout.tsx
import { Providers } from './providers'

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html>
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  )
}
```

## Third-Party Components

```tsx
// components/carousel-wrapper.tsx
'use client'

import { Carousel } from 'third-party-carousel'

export function CarouselWrapper({ items }: { items: Item[] }) {
  return <Carousel items={items} />
}

// app/page.tsx
import { CarouselWrapper } from '@/components/carousel-wrapper'

export default async function Page() {
  const items = await fetchItems()
  return <CarouselWrapper items={items} />
}
```

## Edge Runtime

```tsx
// app/api/route.ts
export const runtime = 'edge'

export async function GET() {
  return new Response('Hello from Edge!')
}

// app/page.tsx
export const runtime = 'edge'

export default async function Page() {
  return <div>Edge-rendered page</div>
}
```

## Quick Reference

| Capability | Server Component | Client Component |
|------------|------------------|------------------|
| Data fetching | ✅ Yes | ⚠️ Use SWR/React Query |
| Backend access | ✅ Yes (DB, files) | ❌ No |
| Event handlers | ❌ No | ✅ Yes |
| State/Effects | ❌ No | ✅ Yes |
| Browser APIs | ❌ No | ✅ Yes |
| Bundle size | 0 KB | Adds to bundle |
| Streaming | ✅ Yes | ❌ No |

## Best Practices

1. **Default to Server Components** - Only use 'use client' when needed
2. **Move Client Components down** - Push them to leaves of component tree
3. **Pass data down** - Fetch in Server Components, pass to Client Components
4. **Use composition** - Nest Server Components inside Client Components via children
5. **Cache expensive operations** - Use React cache() for deduplication
