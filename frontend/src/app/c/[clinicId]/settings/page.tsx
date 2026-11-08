import type { Metadata } from "next";
import { notFound } from "next/navigation";

import { updateClinicAction } from "@/app/actions/clinic";
import { ActionForm, SubmitButton } from "@/components/form";
import { Checkbox, Field, Input, PageHeader, Panel, Select, Textarea } from "@/components/ui";
import { WhatsAppConnect } from "@/components/whatsapp";
import { kzt, phone } from "@/lib/format";
import { clinicStatus, plans } from "@/lib/labels";
import { getMe } from "@/lib/me";
import { getClinic } from "../data";

export const metadata: Metadata = { title: "Настройки" };

const hours = Array.from({ length: 24 }, (_, h) => h);

export default async function SettingsPage({ params }: PageProps<"/c/[clinicId]/settings">) {
  const { clinicId } = await params;
  const me = await getMe();
  if (me.user.role === "admin") notFound();
  const isPlatform = me.user.role === "superadmin";
  const c = await getClinic(clinicId);
  const st = c.settings;

  return (
    <>
      <PageHeader title="Настройки" description="Что и когда бот пишет пациентам и владельцу." />
      <div className="flex max-w-3xl flex-col gap-6">
        {isPlatform ? (
          <Panel title="WhatsApp клиники" description="Видно только команде Jazyl." className="scroll-mt-20">
            <div id="whatsapp" className="scroll-mt-24">
              <WhatsAppConnect clinicId={clinicId} status={c.wa_session_status} mock={c.wa_provider === "mock"} />
            </div>
          </Panel>
        ) : null}

        <ActionForm action={updateClinicAction}>
          <input type="hidden" name="clinic_id" value={clinicId} />

          <Panel title="Клиника">
            <div className="grid gap-4 sm:grid-cols-2">
              <Field label="Название" htmlFor="name" className="sm:col-span-2">
                <Input id="name" name="name" required defaultValue={c.name} />
              </Field>
              <Field label="Город" htmlFor="city">
                <Input id="city" name="city" defaultValue={c.city} />
              </Field>
              <Field label="Адрес" htmlFor="address" hint="Указывается в напоминании пациенту.">
                <Input id="address" name="address" defaultValue={c.address} />
              </Field>
              <Field label="Средний чек, ₸" htmlFor="avg_check_kzt" hint="Для расчёта сохранённой выручки.">
                <Input id="avg_check_kzt" name="avg_check_kzt" type="number" min={0} step={500} defaultValue={c.avg_check_kzt} className="num" />
              </Field>
              <Field label="Язык сообщений по умолчанию" htmlFor="language">
                <Select id="language" name="language" defaultValue={c.language}>
                  <option value="ru">Русский</option>
                  <option value="kk">Қазақша</option>
                </Select>
              </Field>
              <Field label="Ссылка на отзывы в 2ГИС" htmlFor="gis_review_url" className="sm:col-span-2" hint="Пациенты с оценкой 4–5 получат эту ссылку.">
                <Input id="gis_review_url" name="gis_review_url" type="url" defaultValue={c.gis_review_url} placeholder="https://2gis.kz/…" />
              </Field>
            </div>
          </Panel>

          <Panel title="Номера">
            <div className="grid gap-4 sm:grid-cols-2">
              <Field label="Телефон владельца" htmlFor="owner_phone" hint="Отчёты и жалобы приходят сюда.">
                <Input id="owner_phone" name="owner_phone" type="tel" inputMode="tel" defaultValue={c.owner_phone ? phone(c.owner_phone) : ""} />
              </Field>
              <Field label="Телефоны администраторов" htmlFor="admin_phones" hint="С них можно слать боту список записей. По одному в строке.">
                <Textarea id="admin_phones" name="admin_phones" rows={3} defaultValue={c.admin_phones.map(phone).join("\n")} />
              </Field>
            </div>
          </Panel>

          <Panel title="Напоминания">
            <div className="grid gap-4 sm:grid-cols-3">
              <Field label="Вечером накануне, в" htmlFor="reminder_evening_hour">
                <Select id="reminder_evening_hour" name="reminder_evening_hour" defaultValue={st.reminder_evening_hour}>
                  {hours.filter((h) => h >= 8 && h <= 21).map((h) => (
                    <option key={h} value={h}>
                      {String(h).padStart(2, "0")}:00
                    </option>
                  ))}
                </Select>
              </Field>
              <Field label="Не писать с" htmlFor="quiet_start_hour" hint="Тихие часы">
                <Select id="quiet_start_hour" name="quiet_start_hour" defaultValue={st.quiet_start_hour}>
                  {hours.filter((h) => h >= 18).map((h) => (
                    <option key={h} value={h}>
                      {String(h).padStart(2, "0")}:00
                    </option>
                  ))}
                </Select>
              </Field>
              <Field label="до" htmlFor="quiet_end_hour">
                <Select id="quiet_end_hour" name="quiet_end_hour" defaultValue={st.quiet_end_hour}>
                  {hours.filter((h) => h <= 12).map((h) => (
                    <option key={h} value={h}>
                      {String(h).padStart(2, "0")}:00
                    </option>
                  ))}
                </Select>
              </Field>
            </div>
            <div className="mt-3 flex flex-col">
              <Checkbox name="two_hour_reminder" defaultChecked={st.two_hour_reminder} label="Второе напоминание за 2 часа до приёма" />
              <Checkbox
                name="auto_complete_confirmed"
                defaultChecked={st.auto_complete_confirmed}
                label="Считать подтвердивших пришедшими, если администратор не отметил иначе"
                hint="Нужно для запроса оценки после визита."
              />
              <Checkbox name="review_requests" defaultChecked={st.review_requests} label="Просить оценку после визита" hint="4–5 → ссылка на 2ГИС, 1–3 → жалоба лично владельцу." />
            </div>
          </Panel>

          <Panel title="Приём и запись с сайта">
            <div className="grid gap-4 sm:grid-cols-2">
              <Field label="Кресел" htmlFor="chairs">
                <Input id="chairs" name="chairs" type="number" min={1} max={50} defaultValue={st.chairs} className="num" />
              </Field>
              <Field label="Шаг записи" htmlFor="slot_minutes">
                <Select id="slot_minutes" name="slot_minutes" defaultValue={st.slot_minutes}>
                  {[15, 20, 30, 45, 60].map((m) => (
                    <option key={m} value={m}>
                      {m} мин
                    </option>
                  ))}
                </Select>
              </Field>
            </div>
            <div className="mt-3">
              <Checkbox
                name="auto_book_from_site"
                defaultChecked={st.auto_book_from_site}
                label="Сразу записывать в свободное время с сайта"
                hint="Иначе заявка ждёт звонка администратора."
              />
            </div>
          </Panel>

          <Panel title="Отчёты владельцу в WhatsApp">
            <Checkbox name="weekly_report" defaultChecked={st.weekly_report} label="Каждую пятницу в 18:00, за неделю" />
            <Checkbox name="monthly_report" defaultChecked={st.monthly_report} label="1-го числа, за месяц" />
          </Panel>

          {isPlatform ? (
            <Panel title="Тариф и режим" description="Меняет только команда Jazyl.">
              <div className="grid gap-4 sm:grid-cols-3">
                <Field label="Тариф" htmlFor="plan">
                  <Select id="plan" name="plan" defaultValue={c.plan}>
                    {Object.entries(plans).map(([k, p]) => (
                      <option key={k} value={k}>
                        {p.label} · {kzt(p.price)}
                      </option>
                    ))}
                  </Select>
                </Field>
                <Field label="Статус" htmlFor="status">
                  <Select id="status" name="status" defaultValue={c.status}>
                    {Object.entries(clinicStatus).map(([k, s]) => (
                      <option key={k} value={k}>
                        {s.label}
                      </option>
                    ))}
                  </Select>
                </Field>
                <Field label="Отправка" htmlFor="wa_provider">
                  <Select id="wa_provider" name="wa_provider" defaultValue={c.wa_provider}>
                    <option value="openwa">WhatsApp (OpenWA)</option>
                    <option value="mock">Тест, без отправки</option>
                  </Select>
                </Field>
              </div>
            </Panel>
          ) : null}

          <div className="sticky bottom-0 flex items-center gap-3 border-t border-line bg-milk py-3">
            <SubmitButton variant="primary" pendingText="Сохраняем…">
              Сохранить настройки
            </SubmitButton>
          </div>
        </ActionForm>
      </div>
    </>
  );
}
