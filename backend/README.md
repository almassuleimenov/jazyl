# Jazyl — бэкенд

Один сервер на Go обслуживает все клиники (мультитенантность по `clinic_id`). WhatsApp подключается через self-hosted шлюз **OpenWA**. Сайты клиник живут отдельно (например, на Vercel) и отправляют заявки в Jazyl по ключу сайта.

```
 Пациент ── WhatsApp ──► OpenWA (шлюз, сессия на каждую клинику)
                           │  webhook (HMAC-подпись)          ▲ send-text
                           ▼                                  │
 Сайт клиники ──► Jazyl API ──► PostgreSQL ◄── outbox-воркер ─┘   (паузы против бана)
 (Vercel)   X-Site-Key │            ▲
                       │            └── планировщик: напоминания, автозавершение, отчёты
 Дашборд ◄── REST + WebSocket (live-события)
```

## Что умеет

| Тариф | Функция | Где в коде |
|---|---|---|
| все | Напоминание вечером накануне (20:00) и за 2 часа, ответы «1 — приду / 2 — перенести» | `core/reminders.go`, `core/inbound.go` |
| все | Режим «тетради»: администратор шлёт боту список записей текстом, бот сам создаёт записи и отвечает сводкой | `notebook/parser.go`, `core/inbound.go` |
| все | После визита оценка 1–5: 4–5 → ссылка на 2ГИС, 1–3 → жалоба лично владельцу | `core/inbound.go` |
| все | Отчёт владельцу в WhatsApp: в пятницу 18:00 за неделю, 1-го числа за месяц | `core/reports.go` |
| все | Аналитика: кто пишет боту, подтверждения, предотвращённые неявки, сохранённая выручка | `store/analytics.go` |
| 29k+ (и пилот) | Приём заявок с сайта, свободные слоты, опционально автозапись | `core/booking.go`, `api/handlers_public.go` |
| платформа | Суперадмин: все клиники, MRR, подключение WhatsApp по QR или коду | `api/handlers_clinics.go` |

Бот **не болтает с пациентом и не использует ИИ для ответов**: он понимает только 1/2, оценку 1–5 и «СТОП». На всё остальное отвечает сама клиника со своего телефона. В сообщениях пациенту нет диагнозов, только время и адрес. Сообщения уходят от имени клиники.

## Запуск локально

```bash
cd backend
go run ./cmd/jazyl migrate                                   # применить миграции
go run ./cmd/jazyl create-superadmin you@mail.kz 'СложныйПароль123'
go run ./cmd/jazyl                                           # API :8080 + воркеры
go test -race ./...
```

Конфигурация берётся из `../.env` (корень репозитория), все переменные описаны в `../.env.example`. `POSTRE_DATABASE_API` принимается как `DATABASE_URL`.

**Без WhatsApp:** создайте клинику с `"wa_provider": "mock"`, тогда сообщения «отправляются» в лог. Так удобно разрабатывать фронтенд.

**С OpenWA:** поднимите шлюз (`deploy/docker-compose.yml` или `docker run -p 2785:2785 ghcr.io/rmyndharis/openwa`). Пропишите `OPENWA_API_KEY` и `OPENWA_WEBHOOK_SECRET`. `PUBLIC_BASE_URL` должен быть **публичным HTTPS** адресом бэкенда: OpenWA не шлёт вебхуки на приватные адреса. Для локальной разработки подойдёт туннель (`cloudflared tunnel --url http://localhost:8080`).

## Подключение новой клиники (≈15 минут)

1. `POST /api/v1/admin/clinics`: название, slug, `admin_phones` (с каких номеров админ шлёт списки), `owner_phone` (куда идут отчёты и жалобы), `gis_review_url`.
2. `POST /api/v1/admin/clinics/{id}/whatsapp/connect`: создаёт сессию в OpenWA и регистрирует вебхук.
3. `GET /api/v1/admin/clinics/{id}/whatsapp/qr`: сканировать QR телефоном клиники, либо `POST .../whatsapp/pairing-code` с номером, чтобы получить 8-значный код.
4. `POST /api/v1/clinics/{id}/users` с `role: owner`: доступ для владельца.
5. Тариф 29k+: `POST /api/v1/clinics/{id}/site-keys`, ключ вставить в сайт.

> ⚠️ Бот работает на номере клиники через неофициальный шлюз, а значит есть **риск блокировки номера**. Защита: пауза 12–22 с между сообщениями с одного номера, только ответы и напоминания, отписка «СТОП». Для осторожных клиник лучше отдельная SIM-карта «Кезекші». Номера админов из `admin_phones` должны писать **на номер бота с другого телефона**.

## API

Базовый путь `/api/v1`, JSON. Авторизация: `Authorization: Bearer <access_token>`. Токен живёт 15 минут, refresh — 30 дней и ротируется при каждом обновлении. Ошибки приходят в одном формате:

```json
{"error": {"code": "invalid", "message": "invalid input: phone", "request_id": "…"}}
```

Списки с курсором: `?limit=50&cursor=…`, в ответе `{"items": [...], "next_cursor": "…"}`.

### Авторизация
| Метод | Путь | Описание |
|---|---|---|
| POST | `/auth/login` | `{email, password}` → `{access_token, refresh_token, expires_in, user, mfa, mfa_setup_required, must_change_password}`; если включена 2FA → `{mfa_required: true, mfa_token}` |
| POST | `/auth/mfa` | `{mfa_token, code}`: второй шаг входа, 6-значный код из Google Authenticator |
| POST | `/auth/refresh` | `{refresh_token}`, ротация; повторное использование старого токена отзывает все сессии |
| POST | `/auth/logout` | `{refresh_token}` |
| POST | `/auth/logout-all` | выйти на всех устройствах |
| GET | `/auth/me` | пользователь, клиника и `session: {mfa, mfa_setup_required, must_change_password}` |
| POST | `/auth/password` | `{current_password, new_password}` → новая пара токенов, остальные сессии закрываются |
| POST | `/auth/totp/setup` | `{secret, otpauth_url}` для QR-кода; 2FA ещё не включена |
| POST | `/auth/totp/enable` | `{code}` включает 2FA, возвращает новые токены |
| POST | `/auth/totp/disable` | `{password, code}`; суперадмину в production запрещено |

**Безопасность входа.** Пароль, выданный сотруднику Jazyl'ом или владельцем, временный: пока его не сменят, все запросы кроме `/auth/*` получают `403 password_change_required`. Суперадмин без 2FA получает `403 mfa_required` (в production всегда). После 5 неверных паролей или кодов подряд аккаунт блокируется на 15 минут (`429 account_locked`, заголовок `Retry-After`). Каждый TOTP-код срабатывает один раз. Секреты 2FA зашифрованы AES-256-GCM ключом `DATA_ENCRYPTION_KEY`. Потерян телефон суперадмина: на сервере `jazyl reset-mfa EMAIL`.
| POST | `/ws/ticket` | одноразовый тикет на 60 с для WebSocket |
| GET | `/ws?ticket=…&clinic_id=…` | live-события: `appointment.created/updated`, `lead.created`, `message.received`, `review.complaint`, `whatsapp.status`, `schedule.imported` |

### Платформа (только superadmin) — `/admin`
| Метод | Путь | Описание |
|---|---|---|
| GET | `/admin/overview` | все клиники с активностью за 30 дней, MRR, разбивка по тарифам |
| GET/POST | `/admin/clinics` | список / создание клиники |
| DELETE | `/admin/clinics/{id}` | `{confirm: "<slug>"}` → удалить клинику со всеми данными и отвязать WhatsApp (необратимо). Обычно хватает статуса `paused`/`churned`: он выключает напоминания, бота, исходящие и вход сотрудников, данные остаются |
| POST | `/admin/clinics/{id}/whatsapp/connect` | создать и запустить сессию OpenWA, зарегистрировать вебхук |
| GET | `/admin/clinics/{id}/whatsapp` | статус сессии |
| GET | `/admin/clinics/{id}/whatsapp/qr` | QR (data URL PNG) |
| POST | `/admin/clinics/{id}/whatsapp/pairing-code` | `{phone}` → 8-значный код |

### Клиника — `/clinics/{clinicID}` (владелец, админ клиники, суперадмин)
| Метод | Путь | Описание |
|---|---|---|
| GET / PATCH | `/` | профиль и настройки (`settings`: часы напоминаний, тихие часы, кресла, график, `auto_book_from_site`…). Тариф и статус меняет только Jazyl |
| GET/POST, PATCH `/{userID}` | `/users` | сотрудники (`owner` создаёт только Jazyl); пароль при создании временный |
| POST | `/users/{userID}/password` | `{password}`: выдать новый временный пароль (Jazyl; владелец — только админам), сессии сотрудника закрываются, блокировка снимается |
| POST | `/users/{userID}/reset-2fa` | снять 2FA сотрудника, потерявшего телефон (только Jazyl) |
| GET/POST, PUT `/{id}` | `/doctors`, `/services` | врачи и прайс |
| GET, GET/PATCH `/{id}` | `/patients` | `?q=` поиск по имени и телефону; `opted_out` — отписка |
| GET/POST | `/appointments` | `?from&to&status&doctor_id&patient_id`; по умолчанию сегодня и завтра |
| GET/PATCH | `/appointments/{id}` | перенос времени перепланирует напоминания |
| POST | `/appointments/{id}/status` | `{status}`: `confirmed`, `cancelled`, `completed` (запускает запрос оценки), `no_show`… |
| POST | `/schedule/import` | `{text}`: тот же режим «тетради», но из дашборда |
| GET, PATCH `/{id}`, POST `/{id}/convert` | `/leads` | заявки; `convert` с `{starts_at}` превращает заявку в запись |
| GET/POST | `/messages` | журнал переписки (`?phone=`), ручная отправка `{phone, text}` |
| GET | `/chats` | кто пишет боту: по номерам, счётчики, последнее сообщение |
| GET | `/reviews` | оценки |
| GET | `/analytics/summary`, `/analytics/daily` | метрики за `?from&to` (по умолчанию 30 дней) |
| GET | `/reports`, `/reports/preview?period=week\|month` | отчёты |
| POST | `/reports/send?period=week` | отправить отчёт владельцу сейчас |
| GET/POST, DELETE `/{id}` | `/site-keys` | ключи сайтов с `allowed_origins`; сам ключ показывается один раз |

### Публичный API для сайтов — `/api/public/v1` (заголовок `X-Site-Key: jzs_…`)
| Метод | Путь | Описание |
|---|---|---|
| GET | `/clinic` | название, адрес, график, услуги, врачи (для формы записи) |
| GET | `/slots?date=YYYY-MM-DD&service_id=…` | свободное время |
| POST | `/bookings` | `{name, phone, service_id?, service?, preferred_at?, comment?, page_url?, utm?, consent: true, website: ""}` → `202 {lead_id, status: received\|booked}` |

`website` — поле-ловушка для ботов, его нужно оставлять пустым. `consent` обязателен: это согласие на обработку персональных данных.

**Пример для сайта на Next.js (Vercel).** Ключ лучше держать на сервере, в Route Handler:

```ts
// app/api/book/route.ts
export async function POST(req: Request) {
  const body = await req.json();
  const r = await fetch(`${process.env.JAZYL_API}/api/public/v1/bookings`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "X-Site-Key": process.env.JAZYL_SITE_KEY! },
    body: JSON.stringify({ ...body, page_url: req.headers.get("referer") ?? "" }),
  });
  return new Response(await r.text(), { status: r.status });
}
```

Если вызывать API прямо из браузера, укажите домен сайта в `allowed_origins` ключа. Запросы с других доменов получат 403.

### Системное
`GET /healthz` (жив ли процесс) · `GET /readyz` (БД и OpenWA) · `GET /metrics` (Prometheus; закрывается `METRICS_TOKEN`) · `POST /webhooks/openwa` (вебхук OpenWA, проверяется `X-OpenWA-Signature`).

## Надёжность

- **Outbox**: каждое сообщение сначала пишется в таблицу `messages`, в той же транзакции, что и изменение записи. Воркер забирает его через `FOR UPDATE SKIP LOCKED`, отправляет с `Idempotency-Key`, повторяет с экспоненциальной паузой и снимает зависшие блокировки. Если сессия WhatsApp отключена, сообщения ждут; напоминание старше 6 часов отбрасывается.
- **Вебхуки** обрабатываются ровно один раз: ключ идемпотентности OpenWA хранится в той же транзакции, что и обработка. При ошибке возвращается 5xx, и OpenWA повторит доставку.
- **Аналитика** считается по истории статусов (`appointment_events`), а не по текущему статусу, поэтому цифры в отчёте не «плывут».
- Все фоновые задачи безопасно запускать на нескольких инстансах сразу. Единственное ограничение: WebSocket-хаб живёт в памяти, так что live-события видны только клиентам того же инстанса.
- Метрики и алерты: `observability/metrics.go`, `deploy/prometheus/alerts.yml`.

## Безопасность

- Изоляция клиник: каждый SQL-запрос фильтруется по `clinic_id`, а middleware отвечает 404 на чужой `clinicID`.
- Пароли хешируются bcrypt (cost 12). Refresh-токены и ключи сайтов хранятся только как SHA-256 хеш.
- Есть rate-limit на логин и публичный API, строгий JSON (неизвестные поля отклоняются), лимит размера тела запроса и журнал действий в `audit_log`.
- В логах телефоны маскируются (`7701***4567`), тексты сообщений в лог не пишутся.

## Структура

```
cmd/jazyl            точка входа: serve | migrate | create-superadmin | reset-mfa | healthcheck
internal/api         HTTP: роутинг, middleware, обработчики
internal/core        бизнес-логика (напоминания, бот, заявки, отчёты, фоновые задачи)
internal/store       PostgreSQL (pgx), аналитические запросы
internal/messaging   шаблоны ru/kk и outbox-воркер
internal/whatsapp    клиент OpenWA, mock, проверка подписи
internal/notebook    парсер списка записей «тетрадь»
internal/realtime    WebSocket-хаб
internal/platform    пул БД, миграции, логгер
deploy/              docker-compose (jazyl + OpenWA + Caddy), алерты Prometheus
```
