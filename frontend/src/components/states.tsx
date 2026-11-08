"use client";

import { useEffect } from "react";
import { RotateCw } from "lucide-react";

import { Button } from "./ui";

/** Shared error boundary body: plain language, the request id when the API gave one, and a retry. */
export function ErrorView({ error, retry }: { error: unknown; retry: () => void }) {
  useEffect(() => {
    console.error(error);
  }, [error]);
  const digest = (error as { digest?: string } | null)?.digest;
  return (
    <div role="alert" className="rounded-panel border border-line bg-surface px-6 py-10 shadow-panel">
      <h2 className="font-display text-lg font-semibold text-ink">Не удалось загрузить данные</h2>
      <p className="mt-1.5 max-w-lg text-sm text-ink-2">
        Сервер Jazyl не ответил или ответил ошибкой. Ваши данные не потеряны. Попробуйте ещё раз; если повторяется, напишите
        нам и назовите код ниже.
      </p>
      {digest ? <p className="mt-3 font-mono text-xs text-ink-3">Код: {digest}</p> : null}
      <Button className="mt-5" onClick={() => retry()}>
        <RotateCw className="size-4" aria-hidden />
        Повторить
      </Button>
    </div>
  );
}
