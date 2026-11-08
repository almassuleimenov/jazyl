import type { Metadata } from "next";

import { createClinicAction } from "@/app/actions/clinic";
import { ActionForm, SubmitButton } from "@/components/form";
import { ButtonLink, Field, Input, PageHeader, Panel, Select, Textarea } from "@/components/ui";
import { kzt } from "@/lib/format";
import { plans } from "@/lib/labels";

export const metadata: Metadata = { title: "Новая клиника" };

export default function NewClinicPage() {
  return (
    <>
      <PageHeader
        title="Новая клиника"
        description="После сохранения выдадите доступ владельцу и подключите WhatsApp клиники."
        actions={<ButtonLink href="/admin">Отмена</ButtonLink>}
      />
      <ActionForm action={createClinicAction} className="max-w-3xl">
        <Panel title="Клиника">
          <div className="grid gap-4 sm:grid-cols-2">
            <Field label="Название" htmlFor="name" className="sm:col-span-2">
              <Input id="name" name="name" required placeholder="Стоматология «Ақ Тіс»" />
            </Field>
            <Field label="Короткое имя" htmlFor="slug" hint="Латиница и дефис, навсегда. Используется в адресах.">
              <Input id="slug" name="slug" required pattern="[a-z0-9][a-z0-9\-]{1,62}" placeholder="aq-tis" autoCapitalize="none" spellCheck={false} />
            </Field>
            <Field label="Город" htmlFor="city">
              <Input id="city" name="city" defaultValue="Туркестан" />
            </Field>
            <Field label="Адрес" htmlFor="address" className="sm:col-span-2" hint="Его увидит пациент в напоминании.">
              <Input id="address" name="address" placeholder="ул. Тауке хана, 12" />
            </Field>
          </div>
        </Panel>

        <Panel title="Тариф">
          <div className="grid gap-4 sm:grid-cols-3">
            <Field label="Тариф" htmlFor="plan">
              <Select id="plan" name="plan" defaultValue="no_show">
                {Object.entries(plans).map(([k, p]) => (
                  <option key={k} value={k}>
                    {p.label} · {kzt(p.price)}
                  </option>
                ))}
              </Select>
            </Field>
            <Field label="Статус" htmlFor="status">
              <Select id="status" name="status" defaultValue="pilot">
                <option value="pilot">Пилот (бесплатно)</option>
                <option value="active">Платит</option>
              </Select>
            </Field>
            <Field label="Средний чек, ₸" htmlFor="avg_check_kzt" hint="Для расчёта сохранённой выручки.">
              <Input id="avg_check_kzt" name="avg_check_kzt" type="number" min={0} step={500} defaultValue={15000} className="num" />
            </Field>
          </div>
        </Panel>

        <Panel title="WhatsApp и отчёты">
          <div className="grid gap-4 sm:grid-cols-2">
            <Field label="Телефон владельца" htmlFor="owner_phone" hint="Сюда приходят отчёты и жалобы пациентов.">
              <Input id="owner_phone" name="owner_phone" type="tel" inputMode="tel" placeholder="+7 701 123 45 67" />
            </Field>
            <Field label="Ссылка на отзывы в 2ГИС" htmlFor="gis_review_url">
              <Input id="gis_review_url" name="gis_review_url" type="url" placeholder="https://2gis.kz/…" />
            </Field>
            <Field
              label="Телефоны администраторов"
              htmlFor="admin_phones"
              className="sm:col-span-2"
              hint="С этих номеров администратор отправляет боту список записей. По одному в строке."
            >
              <Textarea id="admin_phones" name="admin_phones" rows={2} placeholder={"+7 702 000 00 01"} />
            </Field>
            <Field label="Отправка сообщений" htmlFor="wa_provider" hint="«Тест» ничего не отправляет пациентам, только пишет в журнал.">
              <Select id="wa_provider" name="wa_provider" defaultValue="openwa">
                <option value="openwa">WhatsApp (OpenWA)</option>
                <option value="mock">Тест, без отправки</option>
              </Select>
            </Field>
          </div>
        </Panel>

        <div className="flex gap-2">
          <SubmitButton variant="primary" pendingText="Создаём…">
            Создать клинику
          </SubmitButton>
        </div>
      </ActionForm>
    </>
  );
}
