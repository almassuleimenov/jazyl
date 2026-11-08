# typescript-pro

> Источник: https://github.com/jeffallan/claude-skills/tree/main/skills/typescript-pro (MIT). Рабочая копия скилла: `.claude/skills/typescript-pro/`.

---
name: typescript-pro
description: Implements advanced TypeScript type systems, creates custom type guards, utility types, and branded types, and configures tRPC for end-to-end type safety. Use when building TypeScript applications requiring advanced generics, conditional or mapped types, discriminated unions, monorepo setup, or full-stack type safety with tRPC.
license: MIT
metadata:
  author: https://github.com/Jeffallan
  company: https://synergetic.solutions
  version: "1.1.0"
  domain: language
  triggers: TypeScript, generics, type safety, conditional types, mapped types, tRPC, tsconfig, type guards, discriminated unions
  role: specialist
  scope: implementation
  output-format: code
  related-skills: fullstack-guardian, api-designer
---

# TypeScript Pro

## Core Workflow

1. **Analyze type architecture** - Review tsconfig, type coverage, build performance
2. **Design type-first APIs** - Create branded types, generics, utility types
3. **Implement with type safety** - Write type guards, discriminated unions, conditional types; run `tsc --noEmit` to catch type errors before proceeding
4. **Optimize build** - Configure project references, incremental compilation, tree shaking; re-run `tsc --noEmit` to confirm zero errors after changes
5. **Test types** - Confirm type coverage with a tool like `type-coverage`; validate that all public APIs have explicit return types; iterate on steps 3–4 until all checks pass

## Reference Guide

Load detailed guidance based on context:

| Topic | Reference | Load When |
|-------|-----------|-----------|
| Advanced Types | `references/advanced-types.md` | Generics, conditional types, mapped types, template literals |
| Type Guards | `references/type-guards.md` | Type narrowing, discriminated unions, assertion functions |
| Utility Types | `references/utility-types.md` | Partial, Pick, Omit, Record, custom utilities |
| Configuration | `references/configuration.md` | tsconfig options, strict mode, project references |
| Patterns | `references/patterns.md` | Builder pattern, factory pattern, type-safe APIs |

## Code Examples

### Branded Types
```typescript
// Branded type for domain modeling
type Brand<T, B extends string> = T & { readonly __brand: B };
type UserId  = Brand<string, "UserId">;
type OrderId = Brand<number, "OrderId">;

const toUserId  = (id: string): UserId  => id as UserId;
const toOrderId = (id: number): OrderId => id as OrderId;

// Usage — prevents accidental id mix-ups at compile time
function getOrder(userId: UserId, orderId: OrderId) { /* ... */ }
```

### Discriminated Unions & Type Guards
```typescript
type LoadingState = { status: "loading" };
type SuccessState = { status: "success"; data: string[] };
type ErrorState   = { status: "error";   error: Error };
type RequestState = LoadingState | SuccessState | ErrorState;

// Type predicate guard
function isSuccess(state: RequestState): state is SuccessState {
  return state.status === "success";
}

// Exhaustive switch with discriminated union
function renderState(state: RequestState): string {
  switch (state.status) {
    case "loading": return "Loading…";
    case "success": return state.data.join(", ");
    case "error":   return state.error.message;
    default: {
      const _exhaustive: never = state;
      throw new Error(`Unhandled state: ${_exhaustive}`);
    }
  }
}
```

### Custom Utility Types
```typescript
// Deep readonly — immutable nested objects
type DeepReadonly<T> = {
  readonly [K in keyof T]: T[K] extends object ? DeepReadonly<T[K]> : T[K];
};

// Require exactly one of a set of keys
type RequireExactlyOne<T, Keys extends keyof T = keyof T> =
  Pick<T, Exclude<keyof T, Keys>> &
  { [K in Keys]-?: Required<Pick<T, K>> & Partial<Record<Exclude<Keys, K>, never>> }[Keys];
```

### Recommended tsconfig.json
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitOverride": true,
    "exactOptionalPropertyTypes": true,
    "isolatedModules": true,
    "declaration": true,
    "declarationMap": true,
    "incremental": true,
    "skipLibCheck": false
  }
}
```

## Constraints

### MUST DO
- Enable strict mode with all compiler flags
- Use type-first API design
- Implement branded types for domain modeling
- Use `satisfies` operator for type validation
- Create discriminated unions for state machines
- Use `Annotated` pattern with type predicates
- Generate declaration files for libraries
- Optimize for type inference

### MUST NOT DO
- Use explicit `any` without justification
- Skip type coverage for public APIs
- Mix type-only and value imports
- Disable strict null checks
- Use `as` assertions without necessity
- Ignore compiler performance warnings
- Skip declaration file generation
- Use enums (prefer const objects with `as const`)

## Output Templates

When implementing TypeScript features, provide:
1. Type definitions (interfaces, types, generics)
2. Implementation with type guards
3. tsconfig configuration if needed
4. Brief explanation of type design decisions

## Knowledge Reference

TypeScript 5.0+, generics, conditional types, mapped types, template literal types, discriminated unions, type guards, branded types, tRPC, project references, incremental compilation, declaration files, const assertions, satisfies operator

Maintained by [@jeffallan](https://github.com/jeffallan), Principal Consultant at [Synergetic Solutions](https://synergetic.solutions)

[Documentation](https://jeffallan.github.io/claude-skills/skills/language/typescript-pro/)

---

## Reference: advanced-types

# Advanced Types

## Generic Constraints

```typescript
// Basic constraint
function getProperty<T, K extends keyof T>(obj: T, key: K): T[K] {
  return obj[key];
}

// Multiple constraints
interface HasId { id: number; }
interface HasName { name: string; }

function merge<T extends HasId, U extends HasName>(obj1: T, obj2: U): T & U {
  return { ...obj1, ...obj2 };
}

// Generic constraint with default
type ApiResponse<T = unknown, E = Error> =
  | { success: true; data: T }
  | { success: false; error: E };

// Constraint with infer
type UnwrapPromise<T> = T extends Promise<infer U> ? U : T;
type Result = UnwrapPromise<Promise<string>>; // string
```

## Conditional Types

```typescript
// Basic conditional type
type IsString<T> = T extends string ? true : false;

// Distributive conditional types
type ToArray<T> = T extends any ? T[] : never;
type StringOrNumberArray = ToArray<string | number>; // string[] | number[]

// Non-distributive (use tuple)
type ToArrayNonDist<T> = [T] extends [any] ? T[] : never;
type BothArray = ToArrayNonDist<string | number>; // (string | number)[]

// Nested conditionals for type extraction
type Flatten<T> = T extends Array<infer U>
  ? U extends Array<infer V>
    ? Flatten<V>
    : U
  : T;

type Nested = Flatten<string[][][]>; // string

// Exclude null/undefined
type NonNullable<T> = T extends null | undefined ? never : T;
```

## Mapped Types

```typescript
// Basic mapped type
type ReadOnly<T> = {
  readonly [K in keyof T]: T[K];
};

// Optional properties
type Partial<T> = {
  [K in keyof T]?: T[K];
};

// Required properties
type Required<T> = {
  [K in keyof T]-?: T[K]; // Remove optional modifier
};

// Key remapping with 'as'
type Getters<T> = {
  [K in keyof T as `get${Capitalize<string & K>}`]: () => T[K];
};

interface Person {
  name: string;
  age: number;
}

type PersonGetters = Getters<Person>;
// { getName: () => string; getAge: () => number; }

// Filtering keys
type PickByType<T, U> = {
  [K in keyof T as T[K] extends U ? K : never]: T[K];
};

type StringFields = PickByType<Person, string>; // { name: string }
```

## Template Literal Types

```typescript
// Basic template literal
type EmailLocale = 'en' | 'es' | 'fr';
type EmailType = 'welcome' | 'reset-password';
type EmailTemplate = `${EmailLocale}_${EmailType}`;
// 'en_welcome' | 'en_reset-password' | 'es_welcome' | ...

// Intrinsic string manipulation
type Uppercase<S extends string> = intrinsic;
type Lowercase<S extends string> = intrinsic;
type Capitalize<S extends string> = intrinsic;
type Uncapitalize<S extends string> = intrinsic;

type EventName<T extends string> = `on${Capitalize<T>}`;
type ClickEvent = EventName<'click'>; // 'onClick'

// Template literal with mapped types
type CSSProperties = {
  [K in 'color' | 'background' | 'border' as `--${K}`]: string;
};
// { '--color': string; '--background': string; '--border': string }

// Pattern matching with infer
type ExtractRouteParams<T extends string> =
  T extends `${infer _Start}/:${infer Param}/${infer Rest}`
    ? Param | ExtractRouteParams<`/${Rest}`>
    : T extends `${infer _Start}/:${infer Param}`
    ? Param
    : never;

type Params = ExtractRouteParams<'/users/:id/posts/:postId'>; // 'id' | 'postId'
```

## Higher-Kinded Types (Simulation)

```typescript
// Type-level function simulation
interface TypeClass<F> {
  map: <A, B>(f: (a: A) => B, fa: any) => any;
}

// Functor pattern
type Maybe<T> = { type: 'just'; value: T } | { type: 'nothing' };

const MaybeFunctor: TypeClass<Maybe<any>> = {
  map: <A, B>(f: (a: A) => B, ma: Maybe<A>): Maybe<B> => {
    return ma.type === 'just'
      ? { type: 'just', value: f(ma.value) }
      : { type: 'nothing' };
  }
};

// Builder pattern with generics
type Builder<T, K extends keyof T = never> = {
  with<P extends Exclude<keyof T, K>>(
    key: P,
    value: T[P]
  ): Builder<T, K | P>;
  build(): K extends keyof T ? T : never;
};
```

## Recursive Types

```typescript
// JSON type
type JSONValue =
  | string
  | number
  | boolean
  | null
  | JSONValue[]
  | { [key: string]: JSONValue };

// Deep partial
type DeepPartial<T> = T extends object ? {
  [K in keyof T]?: DeepPartial<T[K]>;
} : T;

// Deep readonly
type DeepReadonly<T> = T extends object ? {
  readonly [K in keyof T]: DeepReadonly<T[K]>;
} : T;

// Path type for nested objects
type PathsToProps<T> = T extends object ? {
  [K in keyof T]: K extends string
    ? T[K] extends object
      ? K | `${K}.${PathsToProps<T[K]>}`
      : K
    : never;
}[keyof T] : never;

interface User {
  profile: {
    name: string;
    settings: {
      theme: string;
    };
  };
}

type UserPaths = PathsToProps<User>;
// 'profile' | 'profile.name' | 'profile.settings' | 'profile.settings.theme'
```

## Variance and Contravariance

```typescript
// Covariance (return types)
type Producer<T> = () => T;
let stringProducer: Producer<string> = () => 'hello';
let objectProducer: Producer<object> = stringProducer; // OK: string is object

// Contravariance (parameter types)
type Consumer<T> = (value: T) => void;
let objectConsumer: Consumer<object> = (obj) => console.log(obj);
let stringConsumer: Consumer<string> = objectConsumer; // OK in strict mode

// Invariance (mutable properties)
interface Box<T> {
  value: T;
  setValue(v: T): void;
}

let stringBox: Box<string> = { value: '', setValue: (v) => {} };
// let objectBox: Box<object> = stringBox; // Error: invariant
```

## Type-Level Programming

```typescript
// Type-level addition (limited)
type Length<T extends any[]> = T['length'];
type Concat<A extends any[], B extends any[]> = [...A, ...B];

// Type-level conditionals
type If<Condition extends boolean, Then, Else> =
  Condition extends true ? Then : Else;

// Type-level equality
type Equal<X, Y> =
  (<T>() => T extends X ? 1 : 2) extends
  (<T>() => T extends Y ? 1 : 2) ? true : false;

// Assert equal types (for testing)
type Assert<T extends true> = T;
type Test = Assert<Equal<1 | 2, 2 | 1>>; // OK
```

## Quick Reference

| Pattern | Use Case |
|---------|----------|
| `T extends U ? X : Y` | Conditional type logic |
| `infer R` | Extract types from patterns |
| `K in keyof T` | Iterate over object keys |
| `as NewKey` | Remap keys in mapped types |
| Template literals | String pattern types |
| `T extends any` | Distributive conditionals |
| `[T] extends [any]` | Non-distributive check |
| `-?` modifier | Remove optional |
| `readonly` modifier | Make immutable |

---

## Reference: configuration

# TypeScript Configuration

## Strict Mode Configuration

```json
{
  "compilerOptions": {
    // Strict type checking
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "strictFunctionTypes": true,
    "strictBindCallApply": true,
    "strictPropertyInitialization": true,
    "noImplicitThis": true,
    "alwaysStrict": true,

    // Additional checks
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitOverride": true,
    "noPropertyAccessFromIndexSignature": true,

    // Module resolution
    "module": "ESNext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "allowImportingTsExtensions": true,

    // Emit
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "removeComments": false,
    "importHelpers": true,

    // Interop
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,
    "forceConsistentCasingInFileNames": true,

    // Target
    "target": "ES2022",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],

    // Skip checking
    "skipLibCheck": true
  }
}
```

## Project References

```json
// Root tsconfig.json
{
  "files": [],
  "references": [
    { "path": "./packages/shared" },
    { "path": "./packages/frontend" },
    { "path": "./packages/backend" }
  ]
}

// packages/shared/tsconfig.json
{
  "compilerOptions": {
    "composite": true,
    "outDir": "./dist",
    "rootDir": "./src",
    "declaration": true,
    "declarationMap": true
  },
  "include": ["src/**/*"]
}

// packages/frontend/tsconfig.json
{
  "compilerOptions": {
    "composite": true,
    "outDir": "./dist",
    "rootDir": "./src"
  },
  "references": [
    { "path": "../shared" }
  ],
  "include": ["src/**/*"]
}
```

## Module Resolution Strategies

```json
// Node16/NodeNext (recommended for Node.js)
{
  "compilerOptions": {
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "esModuleInterop": true
  }
}

// Bundler (for bundlers like Vite, esbuild)
{
  "compilerOptions": {
    "module": "ESNext",
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "moduleDetection": "force"
  }
}

// Classic (legacy, avoid)
{
  "compilerOptions": {
    "module": "CommonJS",
    "moduleResolution": "node"
  }
}
```

## Path Mapping

```json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"],
      "@components/*": ["src/components/*"],
      "@utils/*": ["src/utils/*"],
      "@shared/*": ["../shared/src/*"],
      "@types": ["src/types/index.ts"]
    }
  }
}
```

```typescript
// Usage with path mapping
import { Button } from '@components/Button';
import { formatDate } from '@utils/date';
import type { User } from '@types';
```

## Incremental Compilation

```json
{
  "compilerOptions": {
    "incremental": true,
    "tsBuildInfoFile": "./dist/.tsbuildinfo",
    "composite": true
  }
}
```

## Declaration Files

```json
{
  "compilerOptions": {
    // Generate .d.ts files
    "declaration": true,
    "declarationMap": true,
    "emitDeclarationOnly": false,

    // Bundle declarations
    "declarationDir": "./types",

    // For libraries
    "stripInternal": true
  }
}
```

```typescript
// Using JSDoc for .d.ts generation
/**
 * Creates a user
 * @param name - User's name
 * @param email - User's email
 * @returns The created user
 * @example
 * ```ts
 * const user = createUser('John', 'john@example.com');
 * ```
 */
export function createUser(name: string, email: string): User {
  return { id: generateId(), name, email };
}
```

## Build Optimization

```json
{
  "compilerOptions": {
    // Performance
    "skipLibCheck": true,
    "skipDefaultLibCheck": true,

    // Faster builds
    "incremental": true,
    "assumeChangesOnlyAffectDirectDependencies": true,

    // Smaller output
    "removeComments": true,
    "importHelpers": true,

    // Tree shaking support
    "module": "ESNext",
    "target": "ES2020"
  }
}
```

## Multiple Configurations

```json
// tsconfig.json (base)
{
  "compilerOptions": {
    "strict": true,
    "target": "ES2022"
  }
}

// tsconfig.build.json (production)
{
  "extends": "./tsconfig.json",
  "compilerOptions": {
    "sourceMap": false,
    "removeComments": true,
    "declaration": true
  },
  "exclude": ["**/*.test.ts", "**/*.spec.ts"]
}

// tsconfig.test.json (testing)
{
  "extends": "./tsconfig.json",
  "compilerOptions": {
    "types": ["jest", "node"],
    "esModuleInterop": true
  },
  "include": ["src/**/*.test.ts", "src/**/*.spec.ts"]
}
```

## Framework-Specific Configs

```json
// React + Vite
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ESNext",
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "jsx": "react-jsx",
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "strict": true
  }
}

// Next.js
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["dom", "dom.iterable", "esnext"],
    "allowJs": true,
    "skipLibCheck": true,
    "strict": true,
    "forceConsistentCasingInFileNames": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "esnext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve",
    "incremental": true,
    "plugins": [{ "name": "next" }],
    "paths": { "@/*": ["./src/*"] }
  },
  "include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
  "exclude": ["node_modules"]
}

// Node.js + Express
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "lib": ["ES2022"],
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "declaration": true,
    "sourceMap": true
  }
}
```

## Custom Type Definitions

```typescript
// src/types/global.d.ts
declare global {
  interface Window {
    myApp: {
      version: string;
      config: AppConfig;
    };
  }

  namespace NodeJS {
    interface ProcessEnv {
      DATABASE_URL: string;
      API_KEY: string;
      NODE_ENV: 'development' | 'production' | 'test';
    }
  }
}

export {};

// src/types/modules.d.ts
declare module '*.svg' {
  const content: string;
  export default content;
}

declare module '*.css' {
  const classes: { [key: string]: string };
  export default classes;
}

declare module 'untyped-library' {
  export function doSomething(value: string): number;
}
```

## Compiler API Usage

```typescript
// programmatic compilation
import ts from 'typescript';

function compile(fileNames: string[], options: ts.CompilerOptions): void {
  const program = ts.createProgram(fileNames, options);
  const emitResult = program.emit();

  const allDiagnostics = ts
    .getPreEmitDiagnostics(program)
    .concat(emitResult.diagnostics);

  allDiagnostics.forEach(diagnostic => {
    if (diagnostic.file) {
      const { line, character } = ts.getLineAndCharacterOfPosition(
        diagnostic.file,
        diagnostic.start!
      );
      const message = ts.flattenDiagnosticMessageText(
        diagnostic.messageText,
        '\n'
      );
      console.log(
        `${diagnostic.file.fileName} (${line + 1},${character + 1}): ${message}`
      );
    } else {
      console.log(
        ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n')
      );
    }
  });

  const exitCode = emitResult.emitSkipped ? 1 : 0;
  console.log(`Process exiting with code '${exitCode}'.`);
  process.exit(exitCode);
}

compile(['src/index.ts'], {
  noEmitOnError: true,
  target: ts.ScriptTarget.ES2022,
  module: ts.ModuleKind.ES2022,
  strict: true
});
```

## Performance Monitoring

```json
{
  "compilerOptions": {
    "diagnostics": true,
    "extendedDiagnostics": true,
    "generateCpuProfile": "profile.cpuprofile",
    "explainFiles": true
  }
}
```

```bash
# Run with diagnostics
tsc --diagnostics

# Extended diagnostics
tsc --extendedDiagnostics

# Generate trace
tsc --generateTrace trace

# Analyze with @typescript/analyze-trace
npx @typescript/analyze-trace trace
```

## Quick Reference

| Option | Purpose |
|--------|---------|
| `strict` | Enable all strict checks |
| `composite` | Enable project references |
| `incremental` | Enable incremental compilation |
| `skipLibCheck` | Skip .d.ts checking for faster builds |
| `esModuleInterop` | Better CommonJS interop |
| `moduleResolution` | How modules are resolved |
| `paths` | Path mapping for imports |
| `declaration` | Generate .d.ts files |
| `sourceMap` | Generate source maps |
| `noEmit` | Don't emit output (type check only) |
| `isolatedModules` | Each file can be transpiled separately |
| `allowImportingTsExtensions` | Import .ts files directly |

---

## Reference: patterns

# TypeScript Patterns

## Builder Pattern

```typescript
// Type-safe builder with progressive types
class UserBuilder {
  private data: Partial<User> = {};

  setName(name: string): this {
    this.data.name = name;
    return this;
  }

  setEmail(email: string): this {
    this.data.email = email;
    return this;
  }

  setAge(age: number): this {
    this.data.age = age;
    return this;
  }

  build(): User {
    if (!this.data.name || !this.data.email) {
      throw new Error('Name and email are required');
    }
    return this.data as User;
  }
}

// Fluent API with type safety
const user = new UserBuilder()
  .setName('John')
  .setEmail('john@example.com')
  .setAge(30)
  .build();

// Advanced builder with compile-time validation
type Builder<T, K extends keyof T = never> = {
  [P in keyof T as `set${Capitalize<string & P>}`]: (
    value: T[P]
  ) => Builder<T, K | P>;
} & {
  build: K extends keyof T ? () => T : never;
};

function createBuilder<T>(): Builder<T> {
  const data = {} as T;

  return new Proxy({} as Builder<T>, {
    get(_, prop: string) {
      if (prop === 'build') {
        return () => data;
      }
      if (prop.startsWith('set')) {
        const key = prop.slice(3).toLowerCase();
        return (value: any) => {
          (data as any)[key] = value;
          return this;
        };
      }
    }
  });
}
```

## Factory Pattern

```typescript
// Abstract factory with type safety
interface Logger {
  log(message: string): void;
}

class ConsoleLogger implements Logger {
  log(message: string): void {
    console.log(message);
  }
}

class FileLogger implements Logger {
  constructor(private filename: string) {}

  log(message: string): void {
    // Write to file
  }
}

type LoggerType = 'console' | 'file';
type LoggerConfig<T extends LoggerType> = T extends 'file'
  ? { type: T; filename: string }
  : { type: T };

class LoggerFactory {
  static create<T extends LoggerType>(config: LoggerConfig<T>): Logger {
    switch (config.type) {
      case 'console':
        return new ConsoleLogger();
      case 'file':
        return new FileLogger(config.filename);
      default:
        throw new Error('Unknown logger type');
    }
  }
}

const consoleLogger = LoggerFactory.create({ type: 'console' });
const fileLogger = LoggerFactory.create({ type: 'file', filename: 'app.log' });

// Generic factory with dependency injection
type Constructor<T> = new (...args: any[]) => T;

class Container {
  private instances = new Map<Constructor<any>, any>();

  register<T>(token: Constructor<T>, instance: T): void {
    this.instances.set(token, instance);
  }

  resolve<T>(token: Constructor<T>): T {
    const instance = this.instances.get(token);
    if (!instance) {
      throw new Error(`No instance registered for ${token.name}`);
    }
    return instance;
  }
}
```

## Repository Pattern

```typescript
// Type-safe repository with generic CRUD
interface Entity {
  id: string | number;
}

interface Repository<T extends Entity> {
  find(id: T['id']): Promise<T | null>;
  findAll(): Promise<T[]>;
  create(data: Omit<T, 'id'>): Promise<T>;
  update(id: T['id'], data: Partial<Omit<T, 'id'>>): Promise<T>;
  delete(id: T['id']): Promise<void>;
}

class UserRepository implements Repository<User> {
  async find(id: User['id']): Promise<User | null> {
    // Database query
    return null;
  }

  async findAll(): Promise<User[]> {
    return [];
  }

  async create(data: Omit<User, 'id'>): Promise<User> {
    // Insert into database
    return { id: 1, ...data };
  }

  async update(id: User['id'], data: Partial<Omit<User, 'id'>>): Promise<User> {
    // Update database
    return { id, name: '', email: '', ...data };
  }

  async delete(id: User['id']): Promise<void> {
    // Delete from database
  }
}

// Query builder with type safety
class QueryBuilder<T> {
  private conditions: Array<(item: T) => boolean> = [];

  where<K extends keyof T>(key: K, value: T[K]): this {
    this.conditions.push(item => item[key] === value);
    return this;
  }

  execute(items: T[]): T[] {
    return items.filter(item =>
      this.conditions.every(condition => condition(item))
    );
  }
}

const query = new QueryBuilder<User>()
  .where('email', 'john@example.com')
  .where('age', 30);
```

## Type-Safe API Client

```typescript
// REST API client with type safety
type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';

type ApiEndpoints = {
  '/users': {
    GET: { response: User[] };
    POST: { body: CreateUserDto; response: User };
  };
  '/users/:id': {
    GET: { params: { id: string }; response: User };
    PUT: { params: { id: string }; body: UpdateUserDto; response: User };
    DELETE: { params: { id: string }; response: void };
  };
  '/posts': {
    GET: { query: { userId?: string }; response: Post[] };
    POST: { body: CreatePostDto; response: Post };
  };
};

type ExtractParams<T extends string> =
  T extends `${infer _Start}/:${infer Param}/${infer Rest}`
    ? { [K in Param]: string } & ExtractParams<`/${Rest}`>
    : T extends `${infer _Start}/:${infer Param}`
    ? { [K in Param]: string }
    : {};

class ApiClient {
  async request<
    Path extends keyof ApiEndpoints,
    Method extends keyof ApiEndpoints[Path]
  >(
    method: Method,
    path: Path,
    options?: ApiEndpoints[Path][Method] extends { body: infer B }
      ? { body: B }
      : ApiEndpoints[Path][Method] extends { params: infer P }
      ? { params: P }
      : ApiEndpoints[Path][Method] extends { query: infer Q }
      ? { query: Q }
      : never
  ): Promise<
    ApiEndpoints[Path][Method] extends { response: infer R } ? R : never
  > {
    // Make HTTP request
    return null as any;
  }
}

const client = new ApiClient();

// Type-safe API calls
const users = await client.request('GET', '/users');
const user = await client.request('GET', '/users/:id', { params: { id: '1' } });
const newUser = await client.request('POST', '/users', {
  body: { name: 'John', email: 'john@example.com' }
});
```

## State Machine Pattern

```typescript
// Type-safe state machine
type State = 'idle' | 'loading' | 'success' | 'error';

type Event =
  | { type: 'FETCH' }
  | { type: 'SUCCESS'; data: any }
  | { type: 'ERROR'; error: Error }
  | { type: 'RETRY' };

type StateMachine = {
  [S in State]: {
    [E in Event['type']]?: State;
  };
};

const machine: StateMachine = {
  idle: { FETCH: 'loading' },
  loading: { SUCCESS: 'success', ERROR: 'error' },
  success: { FETCH: 'loading' },
  error: { RETRY: 'loading' }
};

class StateManager<S extends string, E extends { type: string }> {
  constructor(
    private state: S,
    private transitions: Record<S, Partial<Record<E['type'], S>>>
  ) {}

  getState(): S {
    return this.state;
  }

  dispatch(event: E): S {
    const nextState = this.transitions[this.state][event.type];
    if (nextState === undefined) {
      throw new Error(`Invalid transition from ${this.state} on ${event.type}`);
    }
    this.state = nextState;
    return this.state;
  }
}

const manager = new StateManager<State, Event>('idle', machine);
manager.dispatch({ type: 'FETCH' }); // 'loading'
manager.dispatch({ type: 'SUCCESS', data: {} }); // 'success'
```

## Decorator Pattern

```typescript
// Method decorators with type safety
function Log(
  target: any,
  propertyKey: string,
  descriptor: PropertyDescriptor
) {
  const originalMethod = descriptor.value;

  descriptor.value = function (...args: any[]) {
    console.log(`Calling ${propertyKey} with`, args);
    const result = originalMethod.apply(this, args);
    console.log(`Result:`, result);
    return result;
  };

  return descriptor;
}

function Memoize(
  target: any,
  propertyKey: string,
  descriptor: PropertyDescriptor
) {
  const originalMethod = descriptor.value;
  const cache = new Map<string, any>();

  descriptor.value = function (...args: any[]) {
    const key = JSON.stringify(args);
    if (cache.has(key)) {
      return cache.get(key);
    }
    const result = originalMethod.apply(this, args);
    cache.set(key, result);
    return result;
  };

  return descriptor;
}

class Calculator {
  @Log
  @Memoize
  fibonacci(n: number): number {
    if (n <= 1) return n;
    return this.fibonacci(n - 1) + this.fibonacci(n - 2);
  }
}
```

## Result/Either Pattern

```typescript
// Type-safe error handling
type Result<T, E = Error> =
  | { success: true; value: T }
  | { success: false; error: E };

function ok<T>(value: T): Result<T, never> {
  return { success: true, value };
}

function err<E>(error: E): Result<never, E> {
  return { success: false, error };
}

async function fetchUser(id: string): Promise<Result<User, string>> {
  try {
    const response = await fetch(`/api/users/${id}`);
    if (!response.ok) {
      return err('User not found');
    }
    const user = await response.json();
    return ok(user);
  } catch (error) {
    return err('Network error');
  }
}

// Usage with pattern matching
const result = await fetchUser('123');
if (result.success) {
  console.log(result.value.name); // Type-safe access
} else {
  console.error(result.error); // Type-safe error
}

// Either monad
class Either<L, R> {
  private constructor(
    private readonly value: L | R,
    private readonly isRight: boolean
  ) {}

  static left<L, R>(value: L): Either<L, R> {
    return new Either<L, R>(value, false);
  }

  static right<L, R>(value: R): Either<L, R> {
    return new Either<L, R>(value, true);
  }

  map<T>(fn: (value: R) => T): Either<L, T> {
    if (this.isRight) {
      return Either.right(fn(this.value as R));
    }
    return Either.left(this.value as L);
  }

  flatMap<T>(fn: (value: R) => Either<L, T>): Either<L, T> {
    if (this.isRight) {
      return fn(this.value as R);
    }
    return Either.left(this.value as L);
  }

  getOrElse(defaultValue: R): R {
    return this.isRight ? (this.value as R) : defaultValue;
  }
}
```

## Singleton Pattern

```typescript
// Type-safe singleton
class Database {
  private static instance: Database;
  private constructor() {
    // Private constructor prevents instantiation
  }

  static getInstance(): Database {
    if (!Database.instance) {
      Database.instance = new Database();
    }
    return Database.instance;
  }

  query<T>(sql: string): Promise<T[]> {
    // Execute query
    return Promise.resolve([]);
  }
}

const db = Database.getInstance();

// Generic singleton factory
function singleton<T>(factory: () => T): () => T {
  let instance: T | undefined;
  return () => {
    if (!instance) {
      instance = factory();
    }
    return instance;
  };
}

const getConfig = singleton(() => ({
  apiUrl: process.env.API_URL,
  apiKey: process.env.API_KEY
}));
```

## Quick Reference

| Pattern | Use Case |
|---------|----------|
| Builder | Construct complex objects step by step |
| Factory | Create objects without specifying exact class |
| Repository | Abstract data access layer |
| API Client | Type-safe HTTP requests |
| State Machine | Manage state transitions |
| Decorator | Add behavior to methods |
| Result/Either | Type-safe error handling |
| Singleton | Ensure single instance |
| Query Builder | Type-safe database queries |
| Container | Dependency injection |

---

## Reference: type-guards

# Type Guards and Narrowing

## Type Predicates

```typescript
// Basic type predicate
function isString(value: unknown): value is string {
  return typeof value === 'string';
}

function processValue(value: string | number) {
  if (isString(value)) {
    console.log(value.toUpperCase()); // value is string
  } else {
    console.log(value.toFixed(2)); // value is number
  }
}

// Generic type predicate
function isArray<T>(value: T | T[]): value is T[] {
  return Array.isArray(value);
}

// Narrowing to specific interface
interface User {
  type: 'user';
  name: string;
  email: string;
}

interface Admin {
  type: 'admin';
  name: string;
  permissions: string[];
}

function isAdmin(account: User | Admin): account is Admin {
  return account.type === 'admin';
}
```

## Discriminated Unions

```typescript
// Tagged union pattern
type Result<T, E = Error> =
  | { status: 'success'; data: T }
  | { status: 'error'; error: E }
  | { status: 'loading' };

function handleResult<T>(result: Result<T>) {
  switch (result.status) {
    case 'success':
      console.log(result.data); // Narrowed to success
      break;
    case 'error':
      console.error(result.error); // Narrowed to error
      break;
    case 'loading':
      console.log('Loading...'); // Narrowed to loading
      break;
  }
}

// Complex discriminated union
type Shape =
  | { kind: 'circle'; radius: number }
  | { kind: 'rectangle'; width: number; height: number }
  | { kind: 'triangle'; base: number; height: number };

function getArea(shape: Shape): number {
  switch (shape.kind) {
    case 'circle':
      return Math.PI * shape.radius ** 2;
    case 'rectangle':
      return shape.width * shape.height;
    case 'triangle':
      return (shape.base * shape.height) / 2;
  }
}

// Exhaustive checking
function assertNever(x: never): never {
  throw new Error('Unexpected value: ' + x);
}

function processShape(shape: Shape): number {
  switch (shape.kind) {
    case 'circle':
      return shape.radius;
    case 'rectangle':
      return shape.width;
    case 'triangle':
      return shape.base;
    default:
      return assertNever(shape); // Compile error if not exhaustive
  }
}
```

## Built-in Type Guards

```typescript
// typeof narrowing
function printValue(value: string | number | boolean) {
  if (typeof value === 'string') {
    console.log(value.toUpperCase());
  } else if (typeof value === 'number') {
    console.log(value.toFixed(2));
  } else {
    console.log(value ? 'yes' : 'no');
  }
}

// instanceof narrowing
class Dog {
  bark() { console.log('woof'); }
}

class Cat {
  meow() { console.log('meow'); }
}

function makeSound(animal: Dog | Cat) {
  if (animal instanceof Dog) {
    animal.bark();
  } else {
    animal.meow();
  }
}

// in operator narrowing
type Fish = { swim: () => void };
type Bird = { fly: () => void };

function move(animal: Fish | Bird) {
  if ('swim' in animal) {
    animal.swim();
  } else {
    animal.fly();
  }
}

// Truthiness narrowing
function printLength(value: string | null | undefined) {
  if (value) {
    console.log(value.length); // Narrowed to string
  }
}

// Equality narrowing
function compare(x: string | number, y: string | boolean) {
  if (x === y) {
    // x and y are both string
    console.log(x.toUpperCase(), y.toUpperCase());
  }
}
```

## Assertion Functions

```typescript
// Basic assertion function
function assert(condition: unknown, message?: string): asserts condition {
  if (!condition) {
    throw new Error(message || 'Assertion failed');
  }
}

function processUser(user: unknown) {
  assert(typeof user === 'object' && user !== null);
  assert('name' in user && typeof user.name === 'string');
  console.log(user.name.toUpperCase()); // user is narrowed
}

// Type assertion function
function assertIsString(value: unknown): asserts value is string {
  if (typeof value !== 'string') {
    throw new Error('Value is not a string');
  }
}

function greet(name: unknown) {
  assertIsString(name);
  console.log(`Hello, ${name.toUpperCase()}`); // name is string
}

// Generic assertion function
function assertIsDefined<T>(value: T): asserts value is NonNullable<T> {
  if (value === null || value === undefined) {
    throw new Error('Value is null or undefined');
  }
}

function processValue(value: string | null) {
  assertIsDefined(value);
  console.log(value.length); // value is string
}

// Assert with type predicate
function assertIsUser(value: unknown): asserts value is User {
  if (
    typeof value !== 'object' ||
    value === null ||
    !('type' in value) ||
    value.type !== 'user'
  ) {
    throw new Error('Not a user');
  }
}
```

## Control Flow Analysis

```typescript
// Assignment narrowing
let x: string | number = Math.random() > 0.5 ? 'hello' : 42;

if (typeof x === 'string') {
  x; // string
} else {
  x; // number
}

// Return statement narrowing
function getValue(flag: boolean): string | number {
  if (flag) {
    return 'hello';
  }
  return 42; // TypeScript knows this must be number
}

// Throw statement narrowing
function processValue(value: string | null) {
  if (!value) {
    throw new Error('Value is required');
  }
  console.log(value.length); // value is string (null thrown above)
}

// Type guards in array methods
const mixed: (string | number)[] = ['a', 1, 'b', 2];
const strings = mixed.filter((x): x is string => typeof x === 'string');
// strings is string[]
```

## Branded Types

```typescript
// Nominal typing with branded types
type Brand<K, T> = K & { __brand: T };

type UserId = Brand<string, 'UserId'>;
type Email = Brand<string, 'Email'>;
type Url = Brand<string, 'Url'>;

// Constructor functions
function createUserId(id: string): UserId {
  return id as UserId;
}

function createEmail(email: string): Email {
  if (!email.includes('@')) {
    throw new Error('Invalid email');
  }
  return email as Email;
}

// Usage prevents mixing
const userId: UserId = createUserId('user-123');
const email: Email = createEmail('user@example.com');

// const wrongAssignment: UserId = email; // Error!

// Type guard for branded types
function isUserId(value: string): value is UserId {
  return /^user-\d+$/.test(value);
}

// Branded numbers
type Positive = Brand<number, 'Positive'>;
type Integer = Brand<number, 'Integer'>;

function createPositive(n: number): Positive {
  if (n <= 0) throw new Error('Must be positive');
  return n as Positive;
}

function createInteger(n: number): Integer {
  if (!Number.isInteger(n)) throw new Error('Must be integer');
  return n as Integer;
}
```

## Advanced Narrowing Patterns

```typescript
// Array.isArray with generics
function processInput<T>(input: T | T[]): T[] {
  return Array.isArray(input) ? input : [input];
}

// Object key narrowing
function getProperty<T extends object, K extends keyof T>(
  obj: T,
  key: K
): T[K] {
  return obj[key];
}

// Mapped type narrowing
type Nullable<T> = { [K in keyof T]: T[K] | null };

function isComplete<T extends object>(
  obj: Nullable<T>
): obj is T {
  return Object.values(obj).every((v) => v !== null);
}

// Custom narrowing with type maps
type TypeMap = {
  string: string;
  number: number;
  boolean: boolean;
};

function is<K extends keyof TypeMap>(
  type: K,
  value: unknown
): value is TypeMap[K] {
  return typeof value === type;
}

if (is('string', someValue)) {
  someValue.toUpperCase(); // someValue is string
}
```

## Quick Reference

| Pattern | Use Case |
|---------|----------|
| `value is Type` | Type predicate function |
| `asserts condition` | Assertion function |
| `asserts value is Type` | Type assertion function |
| Discriminated union | Tagged union with literal type |
| `typeof` guard | Primitive type checking |
| `instanceof` guard | Class instance checking |
| `in` operator | Property existence check |
| `assertNever` | Exhaustive switch checking |
| Branded types | Nominal typing simulation |
| `NonNullable<T>` | Remove null/undefined |

---

## Reference: utility-types

# Utility Types

## Built-in Utility Types

```typescript
// Partial - All properties optional
interface User {
  id: number;
  name: string;
  email: string;
}

type PartialUser = Partial<User>;
// { id?: number; name?: string; email?: string; }

function updateUser(id: number, updates: Partial<User>) {
  // Only pass fields to update
}

// Required - All properties required
type RequiredUser = Required<PartialUser>;
// { id: number; name: string; email: string; }

// Readonly - All properties readonly
type ReadonlyUser = Readonly<User>;
// { readonly id: number; readonly name: string; readonly email: string; }

// Pick - Select specific properties
type UserSummary = Pick<User, 'id' | 'name'>;
// { id: number; name: string; }

// Omit - Exclude specific properties
type UserWithoutEmail = Omit<User, 'email'>;
// { id: number; name: string; }

// Record - Create object type with specific keys
type UserRoles = Record<string, 'admin' | 'user' | 'guest'>;
// { [key: string]: 'admin' | 'user' | 'guest' }

type PageInfo = Record<'home' | 'about' | 'contact', { title: string }>;
// { home: { title: string }, about: { title: string }, contact: { title: string } }
```

## Type Extraction Utilities

```typescript
// Extract - Extract types from union
type AllTypes = 'a' | 'b' | 'c' | 1 | 2 | 3;
type StringTypes = Extract<AllTypes, string>; // 'a' | 'b' | 'c'
type NumberTypes = Extract<AllTypes, number>; // 1 | 2 | 3

// Exclude - Remove types from union
type WithoutNumbers = Exclude<AllTypes, number>; // 'a' | 'b' | 'c'

// NonNullable - Remove null and undefined
type MaybeString = string | null | undefined;
type DefiniteString = NonNullable<MaybeString>; // string

// ReturnType - Extract function return type
function getUser() {
  return { id: 1, name: 'John' };
}

type User = ReturnType<typeof getUser>; // { id: number; name: string }

// Parameters - Extract function parameter types
function createUser(name: string, age: number) {
  return { name, age };
}

type CreateUserParams = Parameters<typeof createUser>; // [string, number]

// ConstructorParameters - Extract constructor parameters
class Point {
  constructor(public x: number, public y: number) {}
}

type PointParams = ConstructorParameters<typeof Point>; // [number, number]

// InstanceType - Extract instance type from constructor
type PointInstance = InstanceType<typeof Point>; // Point
```

## Custom Utility Types

```typescript
// DeepPartial - Recursive partial
type DeepPartial<T> = T extends object ? {
  [K in keyof T]?: DeepPartial<T[K]>;
} : T;

interface Config {
  database: {
    host: string;
    port: number;
    credentials: {
      username: string;
      password: string;
    };
  };
}

type PartialConfig = DeepPartial<Config>;
// All nested properties are optional

// DeepReadonly - Recursive readonly
type DeepReadonly<T> = T extends object ? {
  readonly [K in keyof T]: DeepReadonly<T[K]>;
} : T;

// Mutable - Remove readonly
type Mutable<T> = {
  -readonly [K in keyof T]: T[K];
};

type MutableUser = Mutable<ReadonlyUser>;

// PickByType - Pick properties by value type
type PickByType<T, U> = {
  [K in keyof T as T[K] extends U ? K : never]: T[K];
};

interface Mixed {
  id: number;
  name: string;
  age: number;
  email: string;
}

type StringProps = PickByType<Mixed, string>; // { name: string; email: string }
type NumberProps = PickByType<Mixed, number>; // { id: number; age: number }

// OmitByType - Omit properties by value type
type OmitByType<T, U> = {
  [K in keyof T as T[K] extends U ? never : K]: T[K];
};

type NoStrings = OmitByType<Mixed, string>; // { id: number; age: number }
```

## Function Utilities

```typescript
// Promisify - Convert sync to async
type Promisify<T extends (...args: any[]) => any> = (
  ...args: Parameters<T>
) => Promise<ReturnType<T>>;

function syncFunction(x: number): string {
  return x.toString();
}

type AsyncVersion = Promisify<typeof syncFunction>;
// (x: number) => Promise<string>

// Awaited - Unwrap promise type
type AwaitedString = Awaited<Promise<string>>; // string
type DeepAwaited = Awaited<Promise<Promise<number>>>; // number

// ThisParameterType - Extract this parameter
function greet(this: User, message: string) {
  return `${this.name}: ${message}`;
}

type ThisType = ThisParameterType<typeof greet>; // User

// OmitThisParameter - Remove this parameter
type GreetFunction = OmitThisParameter<typeof greet>;
// (message: string) => string
```

## Advanced Custom Utilities

```typescript
// Nullable - Add null and undefined
type Nullable<T> = T | null | undefined;

// ValueOf - Get union of all property values
type ValueOf<T> = T[keyof T];

interface Codes {
  success: 200;
  notFound: 404;
  error: 500;
}

type StatusCode = ValueOf<Codes>; // 200 | 404 | 500

// RequireAtLeastOne - Require at least one property
type RequireAtLeastOne<T, Keys extends keyof T = keyof T> =
  Pick<T, Exclude<keyof T, Keys>> &
  {
    [K in Keys]-?: Required<Pick<T, K>> & Partial<Pick<T, Exclude<Keys, K>>>;
  }[Keys];

interface Options {
  id?: number;
  name?: string;
  email?: string;
}

type AtLeastOne = RequireAtLeastOne<Options>;
// Must have at least one of id, name, or email

// RequireOnlyOne - Require exactly one property
type RequireOnlyOne<T, Keys extends keyof T = keyof T> =
  Pick<T, Exclude<keyof T, Keys>> &
  {
    [K in Keys]-?:
      Required<Pick<T, K>> &
      Partial<Record<Exclude<Keys, K>, undefined>>;
  }[Keys];

type OnlyOne = RequireOnlyOne<Options>;
// Must have exactly one of id, name, or email

// Merge - Deep merge two types
type Merge<T, U> = Omit<T, keyof U> & U;

interface Base {
  id: number;
  name: string;
}

interface Extension {
  name: string; // Override
  email: string; // Add
}

type Combined = Merge<Base, Extension>;
// { id: number; name: string; email: string }

// ConditionalKeys - Get keys matching condition
type ConditionalKeys<T, Condition> = {
  [K in keyof T]: T[K] extends Condition ? K : never;
}[keyof T];

type FunctionKeys = ConditionalKeys<typeof Math, Function>;
// 'abs' | 'acos' | 'sin' | ...
```

## Tuple Utilities

```typescript
// First - Get first element type
type First<T extends any[]> = T extends [infer F, ...any[]] ? F : never;

type FirstType = First<[string, number, boolean]>; // string

// Last - Get last element type
type Last<T extends any[]> = T extends [...any[], infer L] ? L : never;

type LastType = Last<[string, number, boolean]>; // boolean

// Tail - Remove first element
type Tail<T extends any[]> = T extends [any, ...infer Rest] ? Rest : never;

type TailTypes = Tail<[string, number, boolean]>; // [number, boolean]

// Prepend - Add element to beginning
type Prepend<T extends any[], U> = [U, ...T];

type WithString = Prepend<[number, boolean], string>; // [string, number, boolean]

// Reverse - Reverse tuple
type Reverse<T extends any[]> =
  T extends [infer First, ...infer Rest]
    ? [...Reverse<Rest>, First]
    : [];

type Reversed = Reverse<[1, 2, 3]>; // [3, 2, 1]
```

## String Utilities

```typescript
// Split - Split string into tuple
type Split<S extends string, D extends string> =
  S extends `${infer T}${D}${infer U}`
    ? [T, ...Split<U, D>]
    : [S];

type Parts = Split<'a-b-c', '-'>; // ['a', 'b', 'c']

// Join - Join tuple into string
type Join<T extends string[], D extends string> =
  T extends [infer F extends string, ...infer R extends string[]]
    ? R extends []
      ? F
      : `${F}${D}${Join<R, D>}`
    : '';

type Joined = Join<['a', 'b', 'c'], '-'>; // 'a-b-c'

// Replace - Replace substring
type Replace<
  S extends string,
  From extends string,
  To extends string
> = S extends `${infer L}${From}${infer R}`
  ? `${L}${To}${R}`
  : S;

type Replaced = Replace<'hello world', 'world', 'TypeScript'>;
// 'hello TypeScript'

// TrimLeft - Remove leading whitespace
type TrimLeft<S extends string> =
  S extends ` ${infer Rest}` ? TrimLeft<Rest> : S;

type Trimmed = TrimLeft<'  hello'>; // 'hello'
```

## Quick Reference

| Utility | Purpose |
|---------|---------|
| `Partial<T>` | Make all properties optional |
| `Required<T>` | Make all properties required |
| `Readonly<T>` | Make all properties readonly |
| `Pick<T, K>` | Select subset of properties |
| `Omit<T, K>` | Remove subset of properties |
| `Record<K, T>` | Create object type with keys K |
| `Extract<T, U>` | Extract types assignable to U |
| `Exclude<T, U>` | Remove types assignable to U |
| `NonNullable<T>` | Remove null and undefined |
| `ReturnType<T>` | Extract function return type |
| `Parameters<T>` | Extract function parameters |
| `Awaited<T>` | Unwrap Promise type |
