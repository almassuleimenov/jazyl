"use client";

import { useActionState, useEffect, useRef, type ReactNode } from "react";
import { useFormStatus } from "react-dom";
import { LoaderCircle } from "lucide-react";

import type { ActionState } from "@/lib/api";
import { Button, Notice, cx } from "./ui";

export function SubmitButton({
  children,
  pendingText,
  variant = "primary",
  size = "md",
  className = "",
  name,
  value,
}: {
  children: ReactNode;
  pendingText?: string;
  variant?: "primary" | "brand" | "secondary" | "ghost" | "danger";
  size?: "sm" | "md";
  className?: string;
  name?: string;
  value?: string;
}) {
  const { pending } = useFormStatus();
  return (
    <Button type="submit" variant={variant} size={size} disabled={pending} aria-busy={pending} className={className} name={name} value={value}>
      {pending ? <LoaderCircle className="size-4 animate-spin" aria-hidden /> : null}
      {pending && pendingText ? pendingText : children}
    </Button>
  );
}

type Action = (state: ActionState, formData: FormData) => Promise<ActionState>;

/**
 * A form bound to a Server Action that reports its result inline.
 * `resetOnSuccess` clears the fields after a successful submit (create forms).
 */
export function ActionForm({
  action,
  children,
  className = "",
  resetOnSuccess = false,
  id,
}: {
  action: Action;
  children: ReactNode;
  className?: string;
  resetOnSuccess?: boolean;
  id?: string;
}) {
  const [state, formAction] = useActionState(action, undefined);
  const ref = useRef<HTMLFormElement>(null);
  useEffect(() => {
    if (state?.ok && resetOnSuccess) ref.current?.reset();
  }, [state, resetOnSuccess]);
  return (
    <form ref={ref} action={formAction} className={cx("flex flex-col gap-4", className)} id={id} noValidate={false}>
      {children}
      <FormResult state={state} />
    </form>
  );
}

export function FormResult({ state }: { state: ActionState }) {
  if (!state) return null;
  return (
    <div aria-live="polite">
      {state.error ? <Notice tone="red">{state.error}</Notice> : null}
      {state.ok && state.message ? <Notice tone="green">{state.message}</Notice> : null}
    </div>
  );
}

/** A one-button form for row actions (status changes, revoke), with inline error. */
export function InlineAction({
  action,
  fields,
  children,
  variant = "secondary",
  confirmText,
}: {
  action: Action;
  fields: Record<string, string>;
  children: ReactNode;
  variant?: "primary" | "brand" | "secondary" | "ghost" | "danger";
  confirmText?: string;
}) {
  const [state, formAction] = useActionState(action, undefined);
  return (
    <form
      action={formAction}
      onSubmit={(e) => {
        if (confirmText && !window.confirm(confirmText)) e.preventDefault();
      }}
      className="inline-flex flex-col items-start"
    >
      {Object.entries(fields).map(([k, v]) => (
        <input key={k} type="hidden" name={k} value={v} />
      ))}
      <SubmitButton variant={variant} size="sm">
        {children}
      </SubmitButton>
      {state?.error ? (
        <span role="alert" className="mt-1 max-w-56 text-xs text-bad">
          {state.error}
        </span>
      ) : null}
    </form>
  );
}
