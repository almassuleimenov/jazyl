"use client";

import { useActionState, useEffect, useRef } from "react";
import { ShieldCheck } from "lucide-react";

import { loginAction, mfaAction, type LoginState } from "@/app/actions/auth";
import { SubmitButton } from "@/components/form";
import { Field, Input, Notice } from "@/components/ui";

export function LoginForm({ next, notice }: { next: string; notice?: string }) {
  // One state machine for both steps: password → (optional) TOTP code.
  const [state, action] = useActionState<LoginState, FormData>(async (prev, fd) => {
    return fd.get("step") === "mfa" ? mfaAction(prev, fd) : loginAction(prev, fd);
  }, undefined);
  const codeRef = useRef<HTMLInputElement>(null);
  const step = state?.step ?? "password";

  useEffect(() => {
    if (step === "mfa") codeRef.current?.focus();
  }, [step]);

  if (step === "mfa") {
    return (
      <form action={action} className="flex flex-col gap-5">
        <input type="hidden" name="step" value="mfa" />
        <input type="hidden" name="next" value={state?.next ?? next} />
        <div className="flex items-start gap-3">
          <span className="grid size-10 shrink-0 place-items-center rounded-full bg-brand-soft text-brand">
            <ShieldCheck className="size-5" aria-hidden />
          </span>
          <div>
            <p className="font-display text-[15px] font-semibold text-ink">Код из приложения</p>
            <p className="mt-0.5 text-sm text-ink-2">
              Откройте Google Authenticator и введите 6 цифр для Jazyl{state?.email ? ` (${state.email})` : ""}.
            </p>
          </div>
        </div>
        <Field label="Код" htmlFor="code">
          <input
            ref={codeRef}
            id="code"
            name="code"
            inputMode="numeric"
            autoComplete="one-time-code"
            pattern="[0-9 ]{6,7}"
            maxLength={7}
            required
            placeholder="123 456"
            aria-invalid={state?.error ? true : undefined}
            className="num h-12 w-full rounded-control border border-line-strong bg-surface px-3 text-center font-display text-2xl tracking-[0.3em] text-ink outline-none focus:border-brand focus:ring-3 focus:ring-brand/15"
          />
        </Field>
        {state?.error ? (
          <div aria-live="polite">
            <Notice tone="red">{state.error}</Notice>
          </div>
        ) : null}
        <SubmitButton pendingText="Проверяем…" className="w-full">
          Войти
        </SubmitButton>
        <a href="/login" className="text-center text-sm text-ink-2 underline hover:text-ink">
          Войти под другим аккаунтом
        </a>
      </form>
    );
  }

  return (
    <form action={action} className="flex flex-col gap-4">
      <input type="hidden" name="step" value="password" />
      <input type="hidden" name="next" value={state?.next ?? next} />
      {notice && !state?.error ? <Notice>{notice}</Notice> : null}
      <Field label="Почта" htmlFor="email">
        <Input
          id="email"
          name="email"
          type="email"
          autoComplete="username"
          required
          defaultValue={state?.email}
          placeholder="name@clinic.kz"
          aria-invalid={state?.error ? true : undefined}
        />
      </Field>
      <Field label="Пароль" htmlFor="password">
        <Input id="password" name="password" type="password" autoComplete="current-password" required minLength={1} />
      </Field>
      {state?.error ? (
        <div aria-live="polite">
          <Notice tone="red">{state.error}</Notice>
        </div>
      ) : null}
      <SubmitButton pendingText="Входим…" className="mt-1 w-full">
        Войти
      </SubmitButton>
    </form>
  );
}
