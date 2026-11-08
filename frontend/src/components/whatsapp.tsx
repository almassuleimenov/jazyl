"use client";

import { useActionState, useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import { connectWhatsAppAction, pairingCodeAction, whatsappQrAction, type QrState } from "@/app/actions/clinic";
import { ActionForm, SubmitButton } from "./form";
import { Button, Chip, Field, Input, Notice } from "./ui";
import { waStatus } from "@/lib/labels";

/** Founder-only: start the clinic's OpenWA session and link the clinic phone by QR or 8-digit code. */
export function WhatsAppConnect({ clinicId, status, mock }: { clinicId: string; status: string; mock: boolean }) {
  const router = useRouter();
  const [qr, setQr] = useState<QrState>(undefined);
  const [polling, setPolling] = useState(false);
  const [pair, pairAction] = useActionState(pairingCodeAction, undefined);
  const wa = waStatus(qr?.status ?? status);
  const connected = wa.tone === "green";

  useEffect(() => {
    if (!polling) return;
    let stop = false;
    const load = async () => {
      const r = await whatsappQrAction(clinicId);
      if (stop) return;
      setQr(r);
      if (waStatus(r?.status ?? "").tone === "green") {
        setPolling(false);
        router.refresh();
      }
    };
    load();
    // WhatsApp rotates the QR roughly every 20 seconds.
    const t = setInterval(load, 8000);
    return () => {
      stop = true;
      clearInterval(t);
    };
  }, [polling, clinicId, router]);

  if (mock) {
    return <Notice>Тестовый режим: сообщения не отправляются, подключать телефон не нужно.</Notice>;
  }

  return (
    <div className="flex flex-col gap-5">
      <div className="flex flex-wrap items-center gap-3">
        <span className="text-sm text-ink-2">Состояние:</span>
        <Chip tone={wa.tone}>{wa.label}</Chip>
      </div>

      {!connected ? (
        <>
          <ol className="flex list-decimal flex-col gap-1.5 pl-5 text-sm text-ink-2 marker:text-ink-3">
            <li>Запустите сессию (один раз для клиники).</li>
            <li>
              На телефоне клиники: WhatsApp → Настройки → Связанные устройства → Привязка устройства.
            </li>
            <li>Отсканируйте QR-код или введите код привязки по номеру.</li>
          </ol>
          <div className="flex flex-wrap gap-2">
            <ActionForm action={connectWhatsAppAction}>
              <input type="hidden" name="clinic_id" value={clinicId} />
              <SubmitButton variant="brand" pendingText="Запускаем…">
                Запустить сессию
              </SubmitButton>
            </ActionForm>
            <Button type="button" onClick={() => setPolling((p) => !p)}>
              {polling ? "Скрыть QR-код" : "Показать QR-код"}
            </Button>
          </div>

          {polling ? (
            <div className="flex flex-col gap-2">
              {qr?.error ? <Notice tone="red">{qr.error}</Notice> : null}
              {qr?.qr ? (
                // eslint-disable-next-line @next/next/no-img-element -- data URL from the gateway, validated server-side
                <img src={qr.qr} alt="QR-код для привязки WhatsApp" className="size-56 rounded-panel border border-line bg-white p-3" />
              ) : !qr?.error ? (
                <p className="text-sm text-ink-2">Получаем QR-код от шлюза…</p>
              ) : null}
              <p className="text-xs text-ink-3">Код обновляется сам. Страница поймёт, когда телефон подключится.</p>
            </div>
          ) : null}

          <form action={pairAction} className="flex flex-col gap-3 border-t border-line pt-5">
            <input type="hidden" name="clinic_id" value={clinicId} />
            <Field label="Или код привязки по номеру" htmlFor="pair-phone" hint="Удобно, если QR не сканируется: WhatsApp попросит ввести 8 символов.">
              <div className="flex flex-wrap gap-2">
                <Input id="pair-phone" name="phone" type="tel" inputMode="tel" required placeholder="+7 701 123 45 67" className="max-w-60" />
                <SubmitButton variant="secondary" pendingText="Запрашиваем…">Получить код</SubmitButton>
              </div>
            </Field>
            {pair?.error ? <Notice tone="red">{pair.error}</Notice> : null}
            {pair?.code ? (
              <p className="num font-display text-3xl font-bold tracking-[0.2em] text-ink" aria-live="polite">
                {pair.code}
              </p>
            ) : null}
          </form>
        </>
      ) : (
        <p className="text-sm text-ink-2">Номер клиники подключён, напоминания уходят от имени клиники.</p>
      )}
    </div>
  );
}
