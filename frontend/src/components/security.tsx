"use client";

import { useState, useTransition } from "react";
import { Copy, Check, Smartphone } from "lucide-react";

import {
  changePasswordAction,
  disableTotpAction,
  enableTotpAction,
  startTotpAction,
  type TotpSetup,
} from "@/app/actions/auth";
import { ActionForm, SubmitButton } from "./form";
import { Button, Field, Input, Notice } from "./ui";

export function ChangePasswordForm({ forced = false }: { forced?: boolean }) {
  return (
    <ActionForm action={changePasswordAction} resetOnSuccess>
      {forced ? <input type="hidden" name="forced" value="1" /> : null}
      {/* Lets password managers attach the new password to the right account. */}
      <Field label={forced ? "Временный пароль" : "Текущий пароль"} htmlFor="current_password">
        <Input id="current_password" name="current_password" type="password" autoComplete="current-password" required />
      </Field>
      <div className="grid gap-4 sm:grid-cols-2">
        <Field label="Новый пароль" htmlFor="new_password" hint="Не короче 10 символов. Лучше фраза из 3–4 слов.">
          <Input id="new_password" name="new_password" type="password" autoComplete="new-password" minLength={10} maxLength={72} required />
        </Field>
        <Field label="Повторите пароль" htmlFor="confirm_password">
          <Input id="confirm_password" name="confirm_password" type="password" autoComplete="new-password" minLength={10} maxLength={72} required />
        </Field>
      </div>
      <div>
        <SubmitButton pendingText="Сохраняем…">{forced ? "Сохранить и продолжить" : "Сменить пароль"}</SubmitButton>
      </div>
    </ActionForm>
  );
}

function CopySecret({ secret }: { secret: string }) {
  const [copied, setCopied] = useState(false);
  const grouped = secret.match(/.{1,4}/g)?.join(" ") ?? secret;
  return (
    <div className="flex items-center gap-2">
      <code className="num min-w-0 flex-1 rounded-control border border-line bg-sunk px-3 py-2 font-mono text-[13px] break-all text-ink">
        {grouped}
      </code>
      <Button
        type="button"
        size="sm"
        onClick={async () => {
          await navigator.clipboard.writeText(secret);
          setCopied(true);
          setTimeout(() => setCopied(false), 1800);
        }}
        aria-label="Скопировать ключ"
      >
        {copied ? <Check className="size-4 text-ok" aria-hidden /> : <Copy className="size-4" aria-hidden />}
        {copied ? "Скопировано" : "Копировать"}
      </Button>
    </div>
  );
}

/** Enrolment: get a secret → scan QR → confirm with the first code. */
export function TotpEnroll({ forced = false }: { forced?: boolean }) {
  const [setup, setSetup] = useState<TotpSetup | null>(null);
  const [pending, start] = useTransition();

  if (!setup || "error" in setup) {
    return (
      <div className="flex flex-col items-start gap-4">
        <ol className="flex list-decimal flex-col gap-1.5 pl-5 text-sm text-ink-2 marker:text-ink-3">
          <li>Установите на телефон Google Authenticator (или 1Password, Aegis).</li>
          <li>Нажмите кнопку ниже и отсканируйте QR-код приложением.</li>
          <li>Введите 6 цифр, которые покажет приложение.</li>
        </ol>
        {setup && "error" in setup ? <Notice tone="red">{setup.error}</Notice> : null}
        <Button variant="brand" disabled={pending} onClick={() => start(async () => setSetup(await startTotpAction()))}>
          <Smartphone className="size-4" aria-hidden />
          {pending ? "Готовим QR-код…" : "Подключить приложение"}
        </Button>
      </div>
    );
  }

  return (
    <div className="grid gap-6 md:grid-cols-[auto_1fr] md:items-start">
      <div className="w-fit rounded-panel border border-line bg-white p-4">
        {/* SVG generated on our server from the otpauth URL; no third-party QR service sees the secret. */}
        <div className="size-44 [&>svg]:size-full" role="img" aria-label="QR-код для приложения" dangerouslySetInnerHTML={{ __html: setup.qr }} />
      </div>
      <div className="flex min-w-0 flex-col gap-4">
        <div className="flex flex-col gap-1.5">
          <p className="text-sm text-ink-2">Не получается отсканировать? Введите ключ вручную:</p>
          <CopySecret secret={setup.secret} />
        </div>
        <ActionForm action={enableTotpAction}>
          {forced ? <input type="hidden" name="forced" value="1" /> : null}
          <Field label="Код из приложения" htmlFor="totp-code">
            <Input
              id="totp-code"
              name="code"
              inputMode="numeric"
              autoComplete="one-time-code"
              placeholder="123 456"
              maxLength={7}
              required
              className="num max-w-48 text-center tracking-[0.2em]"
            />
          </Field>
          <div>
            <SubmitButton variant="brand" pendingText="Проверяем…">Включить защиту</SubmitButton>
          </div>
        </ActionForm>
      </div>
    </div>
  );
}

export function TotpDisable() {
  return (
    <ActionForm action={disableTotpAction}>
      <div className="grid gap-4 sm:grid-cols-2">
        <Field label="Пароль" htmlFor="disable-password">
          <Input id="disable-password" name="password" type="password" autoComplete="current-password" required />
        </Field>
        <Field label="Код из приложения" htmlFor="disable-code">
          <Input id="disable-code" name="code" inputMode="numeric" autoComplete="one-time-code" maxLength={7} required className="num" />
        </Field>
      </div>
      <div>
        <SubmitButton variant="danger" pendingText="Выключаем…">
          Выключить двухфакторную защиту
        </SubmitButton>
      </div>
    </ActionForm>
  );
}
