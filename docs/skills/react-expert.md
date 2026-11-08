# react-expert

> Источник: https://github.com/jeffallan/claude-skills/tree/main/skills/react-expert (MIT). Рабочая копия скилла: `.claude/skills/react-expert/`.

---
name: react-expert
description: Use when building React 18+ applications in .jsx or .tsx files, Next.js App Router projects, or create-react-app setups. Creates components, implements custom hooks, debugs rendering issues, migrates class components to functional, and implements state management. Invoke for Server Components, Suspense boundaries, useActionState forms, performance optimization, or React 19 features.
license: MIT
metadata:
  author: https://github.com/Jeffallan
  company: https://synergetic.solutions
  version: "1.1.0"
  domain: frontend
  triggers: React, JSX, hooks, useState, useEffect, useContext, Server Components, React 19, Suspense, TanStack Query, Redux, Zustand, component, frontend
  role: specialist
  scope: implementation
  output-format: code
  related-skills: fullstack-guardian, playwright-expert, test-master
---

# React Expert

Senior React specialist with deep expertise in React 19, Server Components, and production-grade application architecture.

## When to Use This Skill

- Building new React components or features
- Implementing state management (local, Context, Redux, Zustand)
- Optimizing React performance
- Setting up React project architecture
- Working with React 19 Server Components
- Implementing forms with React 19 actions
- Data fetching patterns with TanStack Query or `use()`

## Core Workflow

1. **Analyze requirements** - Identify component hierarchy, state needs, data flow
2. **Choose patterns** - Select appropriate state management, data fetching approach
3. **Implement** - Write TypeScript components with proper types
4. **Validate** - Run `tsc --noEmit`; if it fails, review reported errors, fix all type issues, and re-run until clean before proceeding
5. **Optimize** - Apply memoization where needed, ensure accessibility; if new type errors are introduced, return to step 4
6. **Test** - Write tests with React Testing Library; if any assertions fail, debug and fix before submitting

## Reference Guide

Load detailed guidance based on context:

| Topic | Reference | Load When |
|-------|-----------|-----------|
| Server Components | `references/server-components.md` | RSC patterns, Next.js App Router |
| React 19 | `references/react-19-features.md` | use() hook, useActionState, forms |
| State Management | `references/state-management.md` | Context, Zustand, Redux, TanStack |
| Hooks | `references/hooks-patterns.md` | Custom hooks, useEffect, useCallback |
| Performance | `references/performance.md` | memo, lazy, virtualization |
| Testing | `references/testing-react.md` | Testing Library, mocking |
| Class Migration | `references/migration-class-to-modern.md` | Converting class components to hooks/RSC |

## Key Patterns

### Server Component (Next.js App Router)
```tsx
// app/users/page.tsx — Server Component, no "use client"
import { db } from '@/lib/db';

interface User {
  id: string;
  name: string;
}

export default async function UsersPage() {
  const users: User[] = await db.user.findMany();

  return (
    <ul>
      {users.map((user) => (
        <li key={user.id}>{user.name}</li>
      ))}
    </ul>
  );
}
```

### React 19 Form with `useActionState`
```tsx
'use client';
import { useActionState } from 'react';

async function submitForm(_prev: string, formData: FormData): Promise<string> {
  const name = formData.get('name') as string;
  // perform server action or fetch
  return `Hello, ${name}!`;
}

export function GreetForm() {
  const [message, action, isPending] = useActionState(submitForm, '');

  return (
    <form action={action}>
      <input name="name" required />
      <button type="submit" disabled={isPending}>
        {isPending ? 'Submitting…' : 'Submit'}
      </button>
      {message && <p>{message}</p>}
    </form>
  );
}
```

### Custom Hook with Cleanup
```tsx
import { useState, useEffect } from 'react';

function useWindowWidth(): number {
  const [width, setWidth] = useState(() => window.innerWidth);

  useEffect(() => {
    const handler = () => setWidth(window.innerWidth);
    window.addEventListener('resize', handler);
    return () => window.removeEventListener('resize', handler); // cleanup
  }, []);

  return width;
}
```

## Constraints

### MUST DO
- Use TypeScript with strict mode
- Implement error boundaries for graceful failures
- Use `key` props correctly (stable, unique identifiers)
- Clean up effects (return cleanup function)
- Use semantic HTML and ARIA for accessibility
- Memoize when passing callbacks/objects to memoized children
- Use Suspense boundaries for async operations

### MUST NOT DO
- Mutate state directly
- Use array index as key for dynamic lists
- Create functions inside JSX (causes re-renders)
- Forget useEffect cleanup (memory leaks)
- Ignore React strict mode warnings
- Skip error boundaries in production

## Output Templates

When implementing React features, provide:
1. Component file with TypeScript types
2. Test file if non-trivial logic
3. Brief explanation of key decisions

## Knowledge Reference

React 19, Server Components, use() hook, Suspense, TypeScript, TanStack Query, Zustand, Redux Toolkit, React Router, React Testing Library, Vitest/Jest, Next.js App Router, accessibility (WCAG)

Maintained by [@jeffallan](https://github.com/jeffallan), Principal Consultant at [Synergetic Solutions](https://synergetic.solutions)

[Documentation](https://jeffallan.github.io/claude-skills/skills/frontend/react-expert/)

---

## Reference: hooks-patterns

# Hooks Patterns

## Custom Hook Pattern

```tsx
// useApi - Data fetching hook
function useApi<T>(url: string) {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const controller = new AbortController();

    fetch(url, { signal: controller.signal })
      .then(res => {
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        return res.json();
      })
      .then(setData)
      .catch(err => {
        if (err.name !== 'AbortError') setError(err);
      })
      .finally(() => setLoading(false));

    return () => controller.abort();
  }, [url]);

  return { data, error, loading };
}
```

## useDebounce

```tsx
function useDebounce<T>(value: T, delay: number): T {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);

  return debounced;
}

// Usage
function Search() {
  const [query, setQuery] = useState('');
  const debouncedQuery = useDebounce(query, 300);

  useEffect(() => {
    if (debouncedQuery) search(debouncedQuery);
  }, [debouncedQuery]);
}
```

## useLocalStorage

```tsx
function useLocalStorage<T>(key: string, initialValue: T) {
  const [value, setValue] = useState<T>(() => {
    if (typeof window === 'undefined') return initialValue;
    const stored = localStorage.getItem(key);
    return stored ? JSON.parse(stored) : initialValue;
  });

  useEffect(() => {
    localStorage.setItem(key, JSON.stringify(value));
  }, [key, value]);

  return [value, setValue] as const;
}
```

## useMediaQuery

```tsx
function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState(() =>
    typeof window !== 'undefined' && window.matchMedia(query).matches
  );

  useEffect(() => {
    const media = window.matchMedia(query);
    const listener = (e: MediaQueryListEvent) => setMatches(e.matches);

    media.addEventListener('change', listener);
    return () => media.removeEventListener('change', listener);
  }, [query]);

  return matches;
}

// Usage
function Layout() {
  const isMobile = useMediaQuery('(max-width: 768px)');
  return isMobile ? <MobileNav /> : <DesktopNav />;
}
```

## useCallback & useMemo

```tsx
// useCallback: Memoize functions (for child dependencies)
const handleClick = useCallback((id: string) => {
  setSelected(id);
}, []); // Empty deps = stable reference

// useMemo: Memoize expensive calculations
const sortedItems = useMemo(() =>
  [...items].sort((a, b) => a.name.localeCompare(b.name)),
  [items]
);

// When to use:
// - useCallback: When passing to memoized children
// - useMemo: When calculation is expensive AND deps rarely change
```

## Effect Cleanup

```tsx
useEffect(() => {
  const subscription = api.subscribe(handler);

  // Cleanup function
  return () => subscription.unsubscribe();
}, []);

// Async effect pattern
useEffect(() => {
  let cancelled = false;

  async function fetchData() {
    const data = await api.getData();
    if (!cancelled) setData(data);
  }

  fetchData();
  return () => { cancelled = true };
}, []);
```

## Quick Reference

| Hook | Purpose |
|------|---------|
| useState | Component state |
| useEffect | Side effects, subscriptions |
| useCallback | Memoize functions |
| useMemo | Memoize values |
| useRef | Mutable ref, DOM access |
| useContext | Read context |
| useReducer | Complex state logic |

| Custom Hook | Use Case |
|-------------|----------|
| useDebounce | Input delay |
| useLocalStorage | Persistent state |
| useMediaQuery | Responsive logic |
| useApi | Data fetching |

---

## Reference: migration-class-to-modern

# Class to Modern React Migration Guide

---

## When to Use This Guide

**Migrate when:**
- Adopting React 18+ features (concurrent rendering, Suspense)
- Improving code reusability and composition
- Reducing bundle size (hooks generally smaller)
- Enabling Server Components in Next.js 13+
- Team standardizing on modern patterns
- Performance optimization opportunities exist
- Testing complexity needs reduction

**Do NOT migrate when:**
- Error boundaries (still require class components)
- Legacy codebase with no maintenance budget
- Component works perfectly and isn't changing
- Team lacks hooks expertise
- Third-party library requires class inheritance
- Migration risk exceeds benefit

**Migration Priority:**
1. New features (write with hooks)
2. Frequently modified components
3. Components with reusable logic
4. Performance bottlenecks
5. Stable, working components (lowest priority)

---

## Lifecycle to Hooks Concept Map

| Class Component | Modern React Equivalent | Notes |
|----------------|------------------------|-------|
| `constructor` | `useState` initialization | No separate constructor needed |
| `componentDidMount` | `useEffect(() => {}, [])` | Empty dependency array |
| `componentDidUpdate` | `useEffect(() => {})` | Runs after every render |
| `componentWillUnmount` | `useEffect` cleanup | Return cleanup function |
| `shouldComponentUpdate` | `React.memo` | Wrap component, custom comparator |
| `getDerivedStateFromProps` | Avoid or use render-time calculation | Usually an anti-pattern |
| `getSnapshotBeforeUpdate` | `useLayoutEffect` | Rarely needed |
| `componentDidCatch` | No hook equivalent | Keep class component |
| `this.forceUpdate()` | `useState` + setter toggle | Avoid, fix architecture |
| `this.state` | `useState` or `useReducer` | Multiple state slices |
| `this.setState` callback | `useEffect` watching state | Separate effect |

---

## Pattern 1: Constructor and State → useState

### Class Component

```tsx
interface Props {
  initialCount: number;
  userId: string;
}

interface State {
  count: number;
  user: User | null;
  isLoading: boolean;
}

class Counter extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      count: props.initialCount,
      user: null,
      isLoading: false,
    };
  }

  increment = () => {
    this.setState({ count: this.state.count + 1 });
  };

  render() {
    return (
      <div>
        <p>Count: {this.state.count}</p>
        <button onClick={this.increment}>Increment</button>
      </div>
    );
  }
}
```

### Modern React

```tsx
interface Props {
  initialCount: number;
  userId: string;
}

interface User {
  id: string;
  name: string;
}

function Counter({ initialCount, userId }: Props) {
  // Separate state slices for better granularity
  const [count, setCount] = useState(initialCount);
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  // Arrow functions no longer need binding
  const increment = () => {
    setCount(prev => prev + 1); // Functional update for safety
  };

  return (
    <div>
      <p>Count: {count}</p>
      <button onClick={increment}>Increment</button>
    </div>
  );
}
```

**Key Differences:**
- No constructor needed
- Lazy initialization: `useState(() => expensiveComputation())`
- Functional updates prevent stale closure bugs
- Separate `useState` calls improve re-render optimization

---

## Pattern 2: Lifecycle Methods → useEffect

### Class Component

```tsx
class UserProfile extends React.Component<{ userId: string }, State> {
  state = {
    user: null as User | null,
    posts: [] as Post[],
  };

  async componentDidMount() {
    await this.fetchUser();
    await this.fetchPosts();
    window.addEventListener('resize', this.handleResize);
  }

  async componentDidUpdate(prevProps: Props) {
    if (prevProps.userId !== this.props.userId) {
      await this.fetchUser();
      await this.fetchPosts();
    }
  }

  componentWillUnmount() {
    window.removeEventListener('resize', this.handleResize);
  }

  fetchUser = async () => {
    const user = await api.getUser(this.props.userId);
    this.setState({ user });
  };

  fetchPosts = async () => {
    const posts = await api.getPosts(this.props.userId);
    this.setState({ posts });
  };

  handleResize = () => {
    // Handle resize
  };

  render() {
    return <div>{this.state.user?.name}</div>;
  }
}
```

### Modern React

```tsx
interface Props {
  userId: string;
}

interface User {
  id: string;
  name: string;
}

interface Post {
  id: string;
  title: string;
}

function UserProfile({ userId }: Props) {
  const [user, setUser] = useState<User | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);

  // Fetch user when userId changes
  useEffect(() => {
    let cancelled = false;

    async function fetchUser() {
      const userData = await api.getUser(userId);
      if (!cancelled) {
        setUser(userData);
      }
    }

    fetchUser();

    // Cleanup to prevent state updates after unmount
    return () => {
      cancelled = true;
    };
  }, [userId]); // Re-run when userId changes

  // Fetch posts when userId changes
  useEffect(() => {
    let cancelled = false;

    async function fetchPosts() {
      const postsData = await api.getPosts(userId);
      if (!cancelled) {
        setPosts(postsData);
      }
    }

    fetchPosts();

    return () => {
      cancelled = true;
    };
  }, [userId]);

  // Event listener with cleanup
  useEffect(() => {
    function handleResize() {
      // Handle resize
    }

    window.addEventListener('resize', handleResize);

    // Cleanup removes listener
    return () => {
      window.removeEventListener('resize', handleResize);
    };
  }, []); // Empty array = mount/unmount only

  return <div>{user?.name}</div>;
}
```

**Critical Points:**
- Separate effects for separate concerns
- Always include cleanup for subscriptions
- Cancellation flags prevent memory leaks
- Dependencies array must include all used values
- Empty array `[]` = mount/unmount only
- No array = after every render (rarely needed)

---

## Pattern 3: shouldComponentUpdate → React.memo

### Class Component

```tsx
class ExpensiveList extends React.Component<Props> {
  shouldComponentUpdate(nextProps: Props) {
    return (
      nextProps.items !== this.props.items ||
      nextProps.filter !== this.props.filter
    );
  }

  render() {
    const { items, filter } = this.props;
    const filtered = items.filter(item => item.includes(filter));
    return (
      <ul>
        {filtered.map(item => (
          <li key={item}>{item}</li>
        ))}
      </ul>
    );
  }
}
```

### Modern React

```tsx
interface Props {
  items: string[];
  filter: string;
  onItemClick?: (item: string) => void;
}

// React.memo with custom comparison
const ExpensiveList = React.memo<Props>(
  ({ items, filter, onItemClick }) => {
    // useMemo for expensive calculations
    const filtered = useMemo(
      () => items.filter(item => item.includes(filter)),
      [items, filter]
    );

    return (
      <ul>
        {filtered.map(item => (
          <li key={item} onClick={() => onItemClick?.(item)}>
            {item}
          </li>
        ))}
      </ul>
    );
  },
  // Custom comparison function (optional)
  (prevProps, nextProps) => {
    return (
      prevProps.items === nextProps.items &&
      prevProps.filter === nextProps.filter &&
      prevProps.onItemClick === nextProps.onItemClick
    );
  }
);

ExpensiveList.displayName = 'ExpensiveList';
```

**Optimization Checklist:**
- `React.memo` prevents re-renders when props unchanged
- `useMemo` caches expensive calculations
- `useCallback` stabilizes function references
- Custom comparator for complex props
- Shallow comparison is default

---

## Pattern 4: Complex State → useReducer

### Class Component

```tsx
class TodoManager extends React.Component<{}, State> {
  state = {
    todos: [] as Todo[],
    filter: 'all' as Filter,
    editingId: null as string | null,
  };

  addTodo = (text: string) => {
    this.setState(prev => ({
      todos: [...prev.todos, { id: uuid(), text, completed: false }],
    }));
  };

  toggleTodo = (id: string) => {
    this.setState(prev => ({
      todos: prev.todos.map(todo =>
        todo.id === id ? { ...todo, completed: !todo.completed } : todo
      ),
    }));
  };

  deleteTodo = (id: string) => {
    this.setState(prev => ({
      todos: prev.todos.filter(todo => todo.id !== id),
    }));
  };

  setFilter = (filter: Filter) => {
    this.setState({ filter });
  };
}
```

### Modern React

```tsx
interface Todo {
  id: string;
  text: string;
  completed: boolean;
}

type Filter = 'all' | 'active' | 'completed';

interface State {
  todos: Todo[];
  filter: Filter;
  editingId: string | null;
}

type Action =
  | { type: 'ADD_TODO'; text: string }
  | { type: 'TOGGLE_TODO'; id: string }
  | { type: 'DELETE_TODO'; id: string }
  | { type: 'SET_FILTER'; filter: Filter }
  | { type: 'START_EDITING'; id: string }
  | { type: 'STOP_EDITING' };

function todoReducer(state: State, action: Action): State {
  switch (action.type) {
    case 'ADD_TODO':
      return {
        ...state,
        todos: [
          ...state.todos,
          { id: crypto.randomUUID(), text: action.text, completed: false },
        ],
      };

    case 'TOGGLE_TODO':
      return {
        ...state,
        todos: state.todos.map(todo =>
          todo.id === action.id
            ? { ...todo, completed: !todo.completed }
            : todo
        ),
      };

    case 'DELETE_TODO':
      return {
        ...state,
        todos: state.todos.filter(todo => todo.id !== action.id),
      };

    case 'SET_FILTER':
      return { ...state, filter: action.filter };

    case 'START_EDITING':
      return { ...state, editingId: action.id };

    case 'STOP_EDITING':
      return { ...state, editingId: null };

    default:
      return state;
  }
}

function TodoManager() {
  const [state, dispatch] = useReducer(todoReducer, {
    todos: [],
    filter: 'all',
    editingId: null,
  });

  // Action creators
  const addTodo = (text: string) => {
    dispatch({ type: 'ADD_TODO', text });
  };

  const toggleTodo = (id: string) => {
    dispatch({ type: 'TOGGLE_TODO', id });
  };

  // Derived state with useMemo
  const visibleTodos = useMemo(() => {
    switch (state.filter) {
      case 'active':
        return state.todos.filter(t => !t.completed);
      case 'completed':
        return state.todos.filter(t => t.completed);
      default:
        return state.todos;
    }
  }, [state.todos, state.filter]);

  return (
    <div>
      {visibleTodos.map(todo => (
        <TodoItem
          key={todo.id}
          todo={todo}
          onToggle={() => toggleTodo(todo.id)}
        />
      ))}
    </div>
  );
}
```

**When to use useReducer:**
- Multiple related state values
- Complex state transitions
- Next state depends on previous
- Testing state logic separately
- Redux-like predictability needed

---

## Pattern 5: Refs Migration

### Class Component

```tsx
class FormWithFocus extends React.Component {
  inputRef = React.createRef<HTMLInputElement>();
  timeoutId: number | null = null;

  componentDidMount() {
    this.inputRef.current?.focus();
  }

  componentWillUnmount() {
    if (this.timeoutId) {
      clearTimeout(this.timeoutId);
    }
  }

  handleSubmit = () => {
    const value = this.inputRef.current?.value;
    console.log(value);
  };

  render() {
    return (
      <form onSubmit={this.handleSubmit}>
        <input ref={this.inputRef} />
      </form>
    );
  }
}
```

### Modern React

```tsx
function FormWithFocus() {
  // DOM ref
  const inputRef = useRef<HTMLInputElement>(null);

  // Mutable value ref (persists across renders)
  const timeoutIdRef = useRef<number | null>(null);

  useEffect(() => {
    // Focus on mount
    inputRef.current?.focus();

    // Cleanup timeout on unmount
    return () => {
      if (timeoutIdRef.current) {
        clearTimeout(timeoutIdRef.current);
      }
    };
  }, []);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const value = inputRef.current?.value;
    console.log(value);
  };

  const handleDelayedAction = () => {
    timeoutIdRef.current = window.setTimeout(() => {
      console.log('Delayed action');
    }, 1000);
  };

  return (
    <form onSubmit={handleSubmit}>
      <input ref={inputRef} />
      <button type="button" onClick={handleDelayedAction}>
        Delayed
      </button>
    </form>
  );
}
```

**Ref Use Cases:**
- DOM access (focus, scroll, measurements)
- Storing mutable values (timers, subscriptions)
- Previous value tracking
- Instance variables replacement

---

## Pattern 6: HOC → Custom Hooks

### Class Component with HOC

```tsx
// HOC
function withAuth<P extends object>(
  Component: React.ComponentType<P & { user: User }>
) {
  return class extends React.Component<P> {
    state = { user: null as User | null };

    componentDidMount() {
      this.fetchUser();
    }

    fetchUser = async () => {
      const user = await auth.getCurrentUser();
      this.setState({ user });
    };

    render() {
      if (!this.state.user) return <div>Loading...</div>;
      return <Component {...this.props} user={this.state.user} />;
    }
  };
}

// Usage
class Dashboard extends React.Component<{ user: User }> {
  render() {
    return <div>Welcome {this.props.user.name}</div>;
  }
}

export default withAuth(Dashboard);
```

### Modern React with Custom Hook

```tsx
// Custom hook
function useAuth() {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function fetchUser() {
      try {
        const userData = await auth.getCurrentUser();
        if (!cancelled) {
          setUser(userData);
          setLoading(false);
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err : new Error('Auth failed'));
          setLoading(false);
        }
      }
    }

    fetchUser();

    return () => {
      cancelled = true;
    };
  }, []);

  const logout = useCallback(async () => {
    await auth.logout();
    setUser(null);
  }, []);

  return { user, loading, error, logout };
}

// Usage
function Dashboard() {
  const { user, loading, error, logout } = useAuth();

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;
  if (!user) return <div>Not authenticated</div>;

  return (
    <div>
      <p>Welcome {user.name}</p>
      <button onClick={logout}>Logout</button>
    </div>
  );
}
```

**Custom Hook Benefits:**
- Easier composition (use multiple hooks)
- Better TypeScript inference
- No wrapper components (simpler tree)
- Easier testing in isolation
- More explicit dependencies

---

## Pattern 7: Render Props → Custom Hooks

### Class Component with Render Props

```tsx
interface MousePosition {
  x: number;
  y: number;
}

class Mouse extends React.Component<
  { children: (pos: MousePosition) => React.ReactNode },
  MousePosition
> {
  state = { x: 0, y: 0 };

  handleMouseMove = (e: MouseEvent) => {
    this.setState({ x: e.clientX, y: e.clientY });
  };

  componentDidMount() {
    window.addEventListener('mousemove', this.handleMouseMove);
  }

  componentWillUnmount() {
    window.removeEventListener('mousemove', this.handleMouseMove);
  }

  render() {
    return this.props.children(this.state);
  }
}

// Usage
<Mouse>
  {({ x, y }) => (
    <div>
      Mouse at {x}, {y}
    </div>
  )}
</Mouse>
```

### Modern React with Custom Hook

```tsx
interface MousePosition {
  x: number;
  y: number;
}

function useMouse(): MousePosition {
  const [position, setPosition] = useState<MousePosition>({ x: 0, y: 0 });

  useEffect(() => {
    function handleMouseMove(e: MouseEvent) {
      setPosition({ x: e.clientX, y: e.clientY });
    }

    window.addEventListener('mousemove', handleMouseMove);

    return () => {
      window.removeEventListener('mousemove', handleMouseMove);
    };
  }, []);

  return position;
}

// Usage
function MouseTracker() {
  const { x, y } = useMouse();

  return (
    <div>
      Mouse at {x}, {y}
    </div>
  );
}
```

**Hook Advantages:**
- No extra nesting
- Clearer data flow
- Combine multiple hooks easily
- Better performance (no wrapper render)

---

## Pattern 8: Context Migration

### Class Component

```tsx
const ThemeContext = React.createContext<Theme>('light');

class ThemedButton extends React.Component {
  static contextType = ThemeContext;
  declare context: React.ContextType<typeof ThemeContext>;

  render() {
    return <button className={this.context}>{this.props.children}</button>;
  }
}

// Or with Consumer
class ThemedButton2 extends React.Component {
  render() {
    return (
      <ThemeContext.Consumer>
        {theme => <button className={theme}>{this.props.children}</button>}
      </ThemeContext.Consumer>
    );
  }
}
```

### Modern React

```tsx
type Theme = 'light' | 'dark';

interface ThemeContextValue {
  theme: Theme;
  toggleTheme: () => void;
}

const ThemeContext = React.createContext<ThemeContextValue | undefined>(
  undefined
);

function useTheme() {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within ThemeProvider');
  }
  return context;
}

function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setTheme] = useState<Theme>('light');

  const toggleTheme = useCallback(() => {
    setTheme(prev => (prev === 'light' ? 'dark' : 'light'));
  }, []);

  const value = useMemo(
    () => ({ theme, toggleTheme }),
    [theme, toggleTheme]
  );

  return (
    <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
  );
}

// Usage
function ThemedButton({ children }: { children: React.ReactNode }) {
  const { theme, toggleTheme } = useTheme();

  return (
    <button className={theme} onClick={toggleTheme}>
      {children}
    </button>
  );
}
```

**Context Best Practices:**
- Custom hook for consuming context
- Memoize context value to prevent re-renders
- Split contexts by update frequency
- Provide type safety with undefined check

---

## Server Components Migration

Modern Next.js 13+ supports Server Components, which cannot use hooks.

### Client Component (Hooks)

```tsx
'use client';

import { useState, useEffect } from 'react';

export function ClientCounter() {
  const [count, setCount] = useState(0);

  useEffect(() => {
    console.log('Client-side effect');
  }, []);

  return <button onClick={() => setCount(count + 1)}>{count}</button>;
}
```

### Server Component (Async)

```tsx
// app/page.tsx - Server Component by default
interface User {
  id: string;
  name: string;
}

async function getUser(id: string): Promise<User> {
  const res = await fetch(`https://api.example.com/users/${id}`, {
    next: { revalidate: 3600 }, // Cache for 1 hour
  });
  return res.json();
}

export default async function UserProfile({ params }: { params: { id: string } }) {
  const user = await getUser(params.id);

  return (
    <div>
      <h1>{user.name}</h1>
      {/* Client component for interactivity */}
      <ClientCounter />
    </div>
  );
}
```

**Server vs Client Decision Tree:**
- Need interactivity (onClick, state)? → Client Component
- Need browser APIs (localStorage, window)? → Client Component
- Need effects or hooks? → Client Component
- Fetching data, reading files, database? → Server Component
- SEO-critical content? → Server Component
- Large dependencies? → Server Component (smaller client bundle)

See reference: `references/server-components.md`

---

## Common Pitfalls

### 1. Stale Closures

**Problem:**
```tsx
function Counter() {
  const [count, setCount] = useState(0);

  useEffect(() => {
    const id = setInterval(() => {
      console.log(count); // Always logs 0!
      setCount(count + 1); // Always sets 1!
    }, 1000);

    return () => clearInterval(id);
  }, []); // Missing dependency

  return <div>{count}</div>;
}
```

**Solution:**
```tsx
function Counter() {
  const [count, setCount] = useState(0);

  useEffect(() => {
    const id = setInterval(() => {
      // Functional update - always has latest state
      setCount(prev => prev + 1);
    }, 1000);

    return () => clearInterval(id);
  }, []); // Now safe

  return <div>{count}</div>;
}
```

### 2. Missing Effect Dependencies

**Problem:**
```tsx
function UserSearch({ userId }: { userId: string }) {
  const [user, setUser] = useState(null);

  useEffect(() => {
    fetchUser(userId); // userId is a dependency!
  }, []); // Bug: won't refetch when userId changes

  return <div>{user?.name}</div>;
}
```

**Solution:**
```tsx
function UserSearch({ userId }: { userId: string }) {
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function fetch() {
      const data = await fetchUser(userId);
      if (!cancelled) setUser(data);
    }

    fetch();

    return () => {
      cancelled = true;
    };
  }, [userId]); // Correct dependency

  return <div>{user?.name}</div>;
}
```

### 3. Over-Memoization

**Problem:**
```tsx
function TodoList({ todos }: { todos: Todo[] }) {
  // Unnecessary - React is already fast
  const memoizedTodos = useMemo(() => todos, [todos]);

  // Unnecessary - simple function
  const handleClick = useCallback(() => {
    console.log('clicked');
  }, []);

  return (
    <ul>
      {memoizedTodos.map(todo => (
        <li key={todo.id} onClick={handleClick}>
          {todo.text}
        </li>
      ))}
    </ul>
  );
}
```

**Solution:**
```tsx
function TodoList({ todos }: { todos: Todo[] }) {
  // Only memoize expensive computations
  const completedCount = useMemo(
    () => todos.filter(t => t.completed).length,
    [todos]
  );

  // Only useCallback for props to memoized children
  return (
    <div>
      <p>Completed: {completedCount}</p>
      <ul>
        {todos.map(todo => (
          <TodoItem key={todo.id} todo={todo} />
        ))}
      </ul>
    </div>
  );
}
```

**Memoization Rules:**
- Measure before optimizing
- Memoize expensive calculations only
- Memoize callbacks passed to memoized children
- Don't memoize everything by default

---

## Migration Checklist

**Before Migration:**
- [ ] Add tests to current class component
- [ ] Identify all lifecycle methods used
- [ ] Document props, state, and behavior
- [ ] Check for error boundary requirements
- [ ] Verify no third-party class inheritance

**During Migration:**
- [ ] Convert constructor/state to useState
- [ ] Map lifecycle methods to useEffect
- [ ] Convert methods to functions or useCallback
- [ ] Replace this.setState with state setters
- [ ] Update ref usage to useRef
- [ ] Add proper effect dependencies
- [ ] Add cleanup functions where needed

**After Migration:**
- [ ] All tests pass
- [ ] No eslint-disable comments added
- [ ] Performance equivalent or better
- [ ] TypeScript types complete
- [ ] Code review completed
- [ ] Documentation updated

---

## Gradual Migration Strategy

**Phase 1: New Code**
- Write all new components with hooks
- Establish team patterns and conventions

**Phase 2: Leaf Components**
- Migrate components with no children first
- Build confidence and muscle memory

**Phase 3: Container Components**
- Migrate parent components
- Extract custom hooks for reusable logic

**Phase 4: Core Infrastructure**
- Migrate providers and contexts
- Update routing and state management

**Never:**
- Don't migrate everything at once
- Don't migrate stable code unnecessarily
- Don't break working features for purity

---

This migration guide provides practical patterns for modernizing React codebases while avoiding common pitfalls and maintaining code quality throughout the transition.

---

## Reference: performance

# Performance Optimization

## React.memo

```tsx
import { memo } from 'react';

// Memoize component - only re-renders when props change
const ExpensiveList = memo(function ExpensiveList({ items }: { items: Item[] }) {
  return (
    <ul>
      {items.map(item => <li key={item.id}>{item.name}</li>)}
    </ul>
  );
});

// Custom comparison function
const UserCard = memo(
  function UserCard({ user }: { user: User }) {
    return <div>{user.name}</div>;
  },
  (prevProps, nextProps) => prevProps.user.id === nextProps.user.id
);
```

## Preventing Re-renders

```tsx
// Problem: New object/function on each render
function Parent() {
  // ❌ Creates new object every render
  return <Child style={{ color: 'red' }} onClick={() => doSomething()} />;
}

// Solution: Memoize or lift out
const style = { color: 'red' }; // Lifted out

function Parent() {
  const handleClick = useCallback(() => doSomething(), []);
  return <Child style={style} onClick={handleClick} />;
}
```

## Code Splitting with lazy()

```tsx
import { lazy, Suspense } from 'react';

// Split heavy components
const HeavyChart = lazy(() => import('./HeavyChart'));
const AdminPanel = lazy(() => import('./AdminPanel'));

function App() {
  return (
    <Suspense fallback={<Loading />}>
      {showChart && <HeavyChart data={data} />}
    </Suspense>
  );
}

// Route-based splitting (React Router)
const routes = [
  {
    path: '/admin',
    element: (
      <Suspense fallback={<Loading />}>
        <AdminPanel />
      </Suspense>
    ),
  },
];
```

## Virtualization

```tsx
import { useVirtualizer } from '@tanstack/react-virtual';

function VirtualList({ items }: { items: Item[] }) {
  const parentRef = useRef<HTMLDivElement>(null);

  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 50,
  });

  return (
    <div ref={parentRef} style={{ height: '400px', overflow: 'auto' }}>
      <div style={{ height: virtualizer.getTotalSize() }}>
        {virtualizer.getVirtualItems().map((virtualItem) => (
          <div
            key={virtualItem.key}
            style={{
              position: 'absolute',
              top: virtualItem.start,
              height: virtualItem.size,
            }}
          >
            {items[virtualItem.index].name}
          </div>
        ))}
      </div>
    </div>
  );
}
```

## useMemo for Expensive Calculations

```tsx
function Analytics({ data }: { data: DataPoint[] }) {
  // Only recalculate when data changes
  const stats = useMemo(() => ({
    total: data.reduce((sum, d) => sum + d.value, 0),
    average: data.reduce((sum, d) => sum + d.value, 0) / data.length,
    max: Math.max(...data.map(d => d.value)),
  }), [data]);

  return <StatsDisplay stats={stats} />;
}
```

## useTransition for Non-urgent Updates

```tsx
import { useTransition } from 'react';

function Search() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<Item[]>([]);
  const [isPending, startTransition] = useTransition();

  function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    setQuery(e.target.value); // Urgent: update input immediately

    startTransition(() => {
      // Non-urgent: can be interrupted
      setResults(filterItems(e.target.value));
    });
  }

  return (
    <>
      <input value={query} onChange={handleChange} />
      {isPending ? <Spinner /> : <Results items={results} />}
    </>
  );
}
```

## Quick Reference

| Technique | When to Use |
|-----------|-------------|
| `memo()` | Prevent re-renders from unchanged props |
| `useMemo()` | Cache expensive calculations |
| `useCallback()` | Stable function references |
| `lazy()` | Code split heavy components |
| `useTransition()` | Keep UI responsive during updates |
| Virtualization | Large lists (1000+ items) |

| Anti-pattern | Fix |
|--------------|-----|
| Inline objects | Lift out or useMemo |
| Inline functions | useCallback |
| Large bundle | lazy() + Suspense |
| Long lists | Virtualization |

---

## Reference: react-19-features

# React 19 Features

## use() Hook

```tsx
import { use, Suspense } from 'react';

// Read promises in render
function Comments({ commentsPromise }: { commentsPromise: Promise<Comment[]> }) {
  const comments = use(commentsPromise);
  return (
    <ul>
      {comments.map(c => <li key={c.id}>{c.text}</li>)}
    </ul>
  );
}

// Parent creates promise, child reads it
function Post({ postId }: { postId: string }) {
  const commentsPromise = fetchComments(postId);

  return (
    <article>
      <PostContent id={postId} />
      <Suspense fallback={<CommentsSkeleton />}>
        <Comments commentsPromise={commentsPromise} />
      </Suspense>
    </article>
  );
}

// Read context conditionally
function Theme({ children }: { children: React.ReactNode }) {
  if (someCondition) {
    const theme = use(ThemeContext);
    return <div className={theme}>{children}</div>;
  }
  return children;
}
```

## useActionState

```tsx
'use client';
import { useActionState } from 'react';

interface FormState {
  error?: string;
  success?: boolean;
}

async function submitAction(prevState: FormState, formData: FormData): Promise<FormState> {
  'use server';
  const email = formData.get('email') as string;

  try {
    await subscribe(email);
    return { success: true };
  } catch {
    return { error: 'Failed to subscribe' };
  }
}

function NewsletterForm() {
  const [state, formAction, isPending] = useActionState(submitAction, {});

  return (
    <form action={formAction}>
      <input name="email" type="email" required disabled={isPending} />
      <button type="submit" disabled={isPending}>
        {isPending ? 'Subscribing...' : 'Subscribe'}
      </button>
      {state.error && <p className="error">{state.error}</p>}
      {state.success && <p className="success">Subscribed!</p>}
    </form>
  );
}
```

## useFormStatus

```tsx
'use client';
import { useFormStatus } from 'react-dom';

function SubmitButton() {
  const { pending, data, method, action } = useFormStatus();

  return (
    <button type="submit" disabled={pending}>
      {pending ? 'Submitting...' : 'Submit'}
    </button>
  );
}

// Must be used inside a <form>
function ContactForm() {
  return (
    <form action={submitAction}>
      <input name="message" />
      <SubmitButton />
    </form>
  );
}
```

## useOptimistic

```tsx
'use client';
import { useOptimistic } from 'react';

function TodoList({ todos }: { todos: Todo[] }) {
  const [optimisticTodos, addOptimisticTodo] = useOptimistic(
    todos,
    (state, newTodo: Todo) => [...state, newTodo]
  );

  async function addTodo(formData: FormData) {
    const text = formData.get('text') as string;

    // Immediately update UI
    addOptimisticTodo({ id: 'temp', text, completed: false });

    // Then persist
    await createTodo(text);
  }

  return (
    <>
      <ul>
        {optimisticTodos.map(todo => (
          <li key={todo.id}>{todo.text}</li>
        ))}
      </ul>
      <form action={addTodo}>
        <input name="text" />
        <button>Add</button>
      </form>
    </>
  );
}
```

## ref as Prop (No forwardRef)

```tsx
// React 19: ref is just a prop
function Input({ ref, ...props }: { ref?: React.Ref<HTMLInputElement> }) {
  return <input ref={ref} {...props} />;
}

// No need for forwardRef anymore
function Form() {
  const inputRef = useRef<HTMLInputElement>(null);
  return <Input ref={inputRef} placeholder="Enter text" />;
}
```

## Quick Reference

| Hook | Purpose |
|------|---------|
| `use()` | Read promise/context in render |
| `useActionState()` | Form action state + pending |
| `useFormStatus()` | Form pending state (child) |
| `useOptimistic()` | Optimistic UI updates |

| Pattern | When |
|---------|------|
| `use(promise)` | Suspense data fetching |
| `use(context)` | Conditional context read |
| `useActionState` | Server Actions with state |

---

## Reference: server-components

# Server Components

## Server vs Client Components

```tsx
// Server Component (default in App Router)
// Can: fetch data, access backend, use async/await
// Cannot: use hooks, browser APIs, event handlers
async function ProductList() {
  const products = await db.products.findMany();
  return (
    <ul>
      {products.map(p => <ProductCard key={p.id} product={p} />)}
    </ul>
  );
}

// Client Component (explicit)
'use client';
import { useState } from 'react';

function AddToCartButton({ productId }: { productId: string }) {
  const [loading, setLoading] = useState(false);
  return (
    <button onClick={() => addToCart(productId)} disabled={loading}>
      Add to Cart
    </button>
  );
}
```

## Data Fetching Pattern

```tsx
// app/products/page.tsx
export default async function ProductsPage() {
  // Runs on server only - no client bundle impact
  const products = await fetch('https://api.example.com/products', {
    next: { revalidate: 3600 } // Cache for 1 hour
  }).then(res => res.json());

  return <ProductGrid products={products} />;
}

// Parallel data fetching
async function Dashboard() {
  const [user, orders, recommendations] = await Promise.all([
    getUser(),
    getOrders(),
    getRecommendations(),
  ]);

  return (
    <>
      <UserHeader user={user} />
      <OrderList orders={orders} />
      <Recommendations items={recommendations} />
    </>
  );
}
```

## Streaming with Suspense

```tsx
import { Suspense } from 'react';

async function SlowComponent() {
  const data = await slowFetch(); // 3 second API call
  return <div>{data}</div>;
}

export default function Page() {
  return (
    <main>
      <h1>Dashboard</h1>
      <FastComponent />

      <Suspense fallback={<Skeleton />}>
        <SlowComponent />
      </Suspense>
    </main>
  );
}
```

## Passing Data Server → Client

```tsx
// Server Component
async function ProductPage({ id }: { id: string }) {
  const product = await getProduct(id);

  // Pass serializable data to client
  return (
    <div>
      <h1>{product.name}</h1>
      {/* Client component receives serialized props */}
      <AddToCartButton productId={product.id} price={product.price} />
    </div>
  );
}
```

## Server Actions

```tsx
// actions.ts
'use server';

export async function createPost(formData: FormData) {
  const title = formData.get('title') as string;
  await db.posts.create({ data: { title } });
  revalidatePath('/posts');
}

// page.tsx (Server Component)
import { createPost } from './actions';

export default function NewPost() {
  return (
    <form action={createPost}>
      <input name="title" required />
      <button type="submit">Create</button>
    </form>
  );
}
```

## Quick Reference

| Type | Can Use | Cannot Use |
|------|---------|------------|
| Server | async/await, db, fs | useState, onClick |
| Client | hooks, events, browser APIs | async component |

| Pattern | Use Case |
|---------|----------|
| Server Component | Data fetching, heavy deps |
| Client Component | Interactivity, state |
| `'use client'` | Mark client boundary |
| `'use server'` | Server Action |
| Suspense | Streaming, loading states |

---

## Reference: state-management

# State Management

## Local State (useState)

```tsx
function Counter() {
  const [count, setCount] = useState(0);

  // Functional update for derived state
  const increment = () => setCount(prev => prev + 1);

  return <button onClick={increment}>{count}</button>;
}
```

## Context for Simple Global State

```tsx
interface ThemeContext {
  theme: 'light' | 'dark';
  toggle: () => void;
}

const ThemeContext = createContext<ThemeContext | null>(null);

function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setTheme] = useState<'light' | 'dark'>('light');

  const toggle = useCallback(() => {
    setTheme(t => t === 'light' ? 'dark' : 'light');
  }, []);

  return (
    <ThemeContext.Provider value={{ theme, toggle }}>
      {children}
    </ThemeContext.Provider>
  );
}

function useTheme() {
  const context = useContext(ThemeContext);
  if (!context) throw new Error('useTheme must be inside ThemeProvider');
  return context;
}
```

## Zustand (Recommended)

```tsx
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface CartStore {
  items: CartItem[];
  addItem: (item: CartItem) => void;
  removeItem: (id: string) => void;
  clear: () => void;
  total: () => number;
}

const useCartStore = create<CartStore>()(
  persist(
    (set, get) => ({
      items: [],
      addItem: (item) => set((state) => ({
        items: [...state.items, item]
      })),
      removeItem: (id) => set((state) => ({
        items: state.items.filter(i => i.id !== id)
      })),
      clear: () => set({ items: [] }),
      total: () => get().items.reduce((sum, i) => sum + i.price, 0),
    }),
    { name: 'cart-storage' }
  )
);

// Component usage
function Cart() {
  const items = useCartStore((state) => state.items);
  const total = useCartStore((state) => state.total());
  const clear = useCartStore((state) => state.clear);

  return (
    <div>
      {items.map(item => <CartItem key={item.id} item={item} />)}
      <p>Total: ${total}</p>
      <button onClick={clear}>Clear Cart</button>
    </div>
  );
}
```

## Redux Toolkit

```tsx
import { createSlice, configureStore, PayloadAction } from '@reduxjs/toolkit';
import { Provider, useSelector, useDispatch } from 'react-redux';

const counterSlice = createSlice({
  name: 'counter',
  initialState: { value: 0 },
  reducers: {
    increment: (state) => { state.value += 1 },
    decrement: (state) => { state.value -= 1 },
    incrementBy: (state, action: PayloadAction<number>) => {
      state.value += action.payload;
    },
  },
});

const store = configureStore({
  reducer: { counter: counterSlice.reducer },
});

type RootState = ReturnType<typeof store.getState>;
type AppDispatch = typeof store.dispatch;

// Typed hooks
const useAppSelector = useSelector.withTypes<RootState>();
const useAppDispatch = useDispatch.withTypes<AppDispatch>();

function Counter() {
  const count = useAppSelector((state) => state.counter.value);
  const dispatch = useAppDispatch();

  return (
    <button onClick={() => dispatch(counterSlice.actions.increment())}>
      {count}
    </button>
  );
}
```

## TanStack Query (Server State)

```tsx
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

function UserProfile({ userId }: { userId: string }) {
  const queryClient = useQueryClient();

  const { data, isLoading, error } = useQuery({
    queryKey: ['user', userId],
    queryFn: () => fetchUser(userId),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  const mutation = useMutation({
    mutationFn: updateUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['user', userId] });
    },
  });

  if (isLoading) return <Skeleton />;
  if (error) return <Error error={error} />;

  return <UserCard user={data} onUpdate={mutation.mutate} />;
}
```

## Quick Reference

| Solution | Best For |
|----------|----------|
| useState | Local component state |
| Context | Theme, auth, simple globals |
| Zustand | Medium complexity, minimal boilerplate |
| Redux Toolkit | Complex state, middleware, devtools |
| TanStack Query | Server state, caching |

---

## Reference: testing-react

# Testing React

## Basic Component Test

```tsx
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

test('renders greeting', () => {
  render(<Greeting name="World" />);
  expect(screen.getByText('Hello, World!')).toBeInTheDocument();
});

test('increments counter on click', async () => {
  const user = userEvent.setup();
  render(<Counter />);

  await user.click(screen.getByRole('button', { name: /increment/i }));

  expect(screen.getByText('1')).toBeInTheDocument();
});
```

## Query Priority

```tsx
// Preferred: Accessible queries (how users find elements)
screen.getByRole('button', { name: /submit/i });
screen.getByLabelText('Email');
screen.getByPlaceholderText('Search...');
screen.getByText('Welcome');

// Fallback: Test IDs (when no accessible name)
screen.getByTestId('custom-element');

// Async queries (wait for element)
await screen.findByText('Loading complete');
```

## Testing Forms

```tsx
test('submits form with user data', async () => {
  const handleSubmit = vi.fn();
  const user = userEvent.setup();

  render(<ContactForm onSubmit={handleSubmit} />);

  await user.type(screen.getByLabelText('Name'), 'John Doe');
  await user.type(screen.getByLabelText('Email'), 'john@example.com');
  await user.selectOptions(screen.getByLabelText('Topic'), 'support');
  await user.click(screen.getByRole('button', { name: /submit/i }));

  expect(handleSubmit).toHaveBeenCalledWith({
    name: 'John Doe',
    email: 'john@example.com',
    topic: 'support',
  });
});
```

## Testing with Providers

```tsx
function renderWithProviders(
  ui: React.ReactElement,
  { initialState = {}, ...options } = {}
) {
  function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <ThemeProvider>
          {children}
        </ThemeProvider>
      </QueryClientProvider>
    );
  }

  return render(ui, { wrapper: Wrapper, ...options });
}

test('displays user data', async () => {
  renderWithProviders(<UserProfile userId="123" />);

  await screen.findByText('John Doe');
});
```

## Mocking API Calls

```tsx
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';

const server = setupServer(
  http.get('/api/users/:id', ({ params }) => {
    return HttpResponse.json({ id: params.id, name: 'John' });
  })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

test('fetches and displays user', async () => {
  render(<UserProfile userId="123" />);

  await screen.findByText('John');
});

test('handles error', async () => {
  server.use(
    http.get('/api/users/:id', () => {
      return new HttpResponse(null, { status: 500 });
    })
  );

  render(<UserProfile userId="123" />);

  await screen.findByText('Error loading user');
});
```

## Testing Hooks

```tsx
import { renderHook, act } from '@testing-library/react';

test('useCounter increments', () => {
  const { result } = renderHook(() => useCounter());

  act(() => {
    result.current.increment();
  });

  expect(result.current.count).toBe(1);
});

test('useDebounce delays value', async () => {
  vi.useFakeTimers();

  const { result, rerender } = renderHook(
    ({ value }) => useDebounce(value, 500),
    { initialProps: { value: 'initial' } }
  );

  rerender({ value: 'updated' });
  expect(result.current).toBe('initial');

  await act(async () => {
    vi.advanceTimersByTime(500);
  });

  expect(result.current).toBe('updated');
  vi.useRealTimers();
});
```

## Quick Reference

| Query | Use When |
|-------|----------|
| `getByRole` | Buttons, links, headings |
| `getByLabelText` | Form inputs |
| `getByText` | Non-interactive text |
| `findByX` | Async/loading content |
| `queryByX` | Assert NOT present |

| Pattern | Use Case |
|---------|----------|
| `userEvent.setup()` | User interactions |
| `renderHook()` | Testing custom hooks |
| `msw` | Mocking API calls |
| Custom render | Wrap with providers |
