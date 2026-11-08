"use client";

import { useActionState, useState } from "react";
import { Check, Copy } from "lucide-react";

import { createSiteKeyAction } from "@/app/actions/clinic";
import { SubmitButton } from "./form";
import { Button, Field, Input, Notice, Textarea } from "./ui";

export function SiteKeyCreate({ clinicId }: { clinicId: string }) {
  const [state, action] = useActionState(createSiteKeyAction, undefined);
  const [copied, setCopied] = useState(false);

  return (
    <div className="flex flex-col gap-4">
      {state?.key ? (
        <div className="flex flex-col gap-2 rounded-control border border-[#efd2a3] bg-action-soft p-4">
          <p className="text-sm font-medium text-action-ink">Скопируйте ключ сейчас — больше он не покажется.</p>
          <div className="flex items-center gap-2">
            <code className="min-w-0 flex-1 rounded-control border border-line bg-surface px-3 py-2 font-mono text-[13px] break-all text-ink">
              {state.key}
            </code>
            <Button
              type="button"
              size="sm"
              onClick={async () => {
                await navigator.clipboard.writeText(state.key!);
                setCopied(true);
              }}
            >
              {copied ? <Check className="size-4 text-ok" aria-hidden /> : <Copy className="size-4" aria-hidden />}
              {copied ? "Скопирован" : "Копировать"}
            </Button>
          </div>
          <p className="text-xs text-action-ink">
            На сайте: переменная <code className="font-mono">JAZYL_SITE_KEY</code>, запросы с сервера на <code className="font-mono">/api/public/v1/bookings</code>.
          </p>
        </div>
      ) : null}

      <form action={action} className="flex flex-col gap-4">
        <input type="hidden" name="clinic_id" value={clinicId} />
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Название" htmlFor="key-name">
            <Input id="key-name" name="name" required placeholder="Сайт aqtis.kz" />
          </Field>
          <Field
            label="Домены сайта"
            htmlFor="allowed_origins"
            hint="Нужны, только если форма шлёт заявки прямо из браузера. Через сервер сайта — оставьте пустым."
          >
            <Textarea id="allowed_origins" name="allowed_origins" rows={2} placeholder="https://aqtis.kz" />
          </Field>
        </div>
        {state?.error ? <Notice tone="red">{state.error}</Notice> : null}
        <div>
          <SubmitButton pendingText="Создаём…">Создать ключ</SubmitButton>
        </div>
      </form>
    </div>
  );
}
