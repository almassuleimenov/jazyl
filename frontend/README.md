# Jazyl — кабинет

Next.js 16 (App Router) + Tailwind v4. Кабинет команды Jazyl и клиник: записи, заявки, переписка, команда и доступ, настройки, WhatsApp, ключи сайта.

## Как устроена безопасность

- **BFF.** Браузер общается только с этим сервером, Go API он не видит. Токены хранятся в `httpOnly` cookies (`__Host-jz_at` — 15 минут, `__Host-jz_rt` — 30 дней, `SameSite=Lax`, `Secure` в production), JavaScript страницы их не прочитает. `JAZYL_API_URL` — серверная переменная.
- **`src/proxy.ts`** до рендера обновляет истёкший access-токен. Параллельные запросы делят один refresh; в API есть 30-секундное окно, поэтому гонка не принимается за кражу токена.
- **Права проверяет API** при каждом вызове: изоляция клиник, роли, `403 password_change_required`, `403 mfa_required`. Кабинет только перенаправляет: временный пароль → `/setup/password`, суперадмин без 2FA → `/setup/2fa`.
- **Доступ выдаёт основатель:** «Команда и доступ» → почта и временный пароль. При первом входе сотрудник обязан сменить пароль. Там же: новый временный пароль, отключение доступа, снятие 2FA (только Jazyl).
- **2FA (TOTP, Google Authenticator):** обязательна для команды Jazyl, по желанию для клиник. QR рисуется на нашем сервере, секрет не уходит сторонним сервисам.

## Запуск

```bash
cp .env.example .env.local     # JAZYL_API_URL=http://localhost:8080
npm ci
npm run dev                    # http://localhost:3000
```

Бэкенд запускается из `../backend` (`go run ./cmd/jazyl`). Первый суперадмин: `go run ./cmd/jazyl create-superadmin you@mail.kz 'Пароль'`. При первом входе система попросит подключить 2FA (в production это обязательно; для разработки включается через `REQUIRE_SUPERADMIN_MFA=true`).

## Проверки

```bash
npx next typegen && npx tsc --noEmit && npx eslint src && npm run build
```

## Деплой

`Dockerfile` собирает standalone-сервер Node. В `../backend/deploy/docker-compose.yml` это сервис `dashboard`. Его обслуживает Caddy на `app.jazyl.kz`, а с API он общается по внутренней сети (`http://jazyl:8080`). У API должен быть включён `TRUSTED_PROXY=true`, чтобы лимит попыток входа считался по IP клиента.
