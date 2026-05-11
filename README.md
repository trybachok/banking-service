# Banking Service

REST API банковского сервиса на Go 1.25.

Проект реализует банковский сервис с регистрацией пользователей, JWT-аутентификацией, банковскими счетами, картами, переводами, кредитами, MFA/2FA, аналитикой, SMTP-уведомлениями и запуском в Docker.

## Возможности

- Регистрация пользователей.
- Аутентификация через JWT.
- Создание и просмотр банковских счетов.
- Пополнение и списание средств со счёта.
- Переводы между счетами.
- Выпуск виртуальных банковских карт.
- Генерация номера карты по алгоритму Луна.
- PGP-шифрование номера карты и срока действия через PostgreSQL `pgcrypto`.
- HMAC-SHA256 для проверки целостности карточных данных.
- bcrypt-хеширование паролей и CVV.
- Оплата картой.
- Оформление кредита.
- Расчёт аннуитетного платежа.
- Генерация графика платежей.
- Scheduler для автоматической обработки кредитных платежей.
- Начисление штрафов при просрочке.
- Интеграция с ЦБ РФ для получения ключевой ставки.
- SMTP-уведомления через `email_outbox`.
- MFA/2FA для критических операций.
- Финансовая аналитика.
- Прогноз баланса на N дней.
- Административные endpoints.

## Архитектура проекта

Проект построен по принципам чистой слоистой архитектуры.

```text
cmd/
  api/
    main.go
  scheduler/
    main.go
  migrate/
    main.go

internal/
  domain/
  application/
  infrastructure/
  transport/http/

migrations/
api/
deploy/
docs/
tests/
```

Основные слои:

```text
domain          — доменные модели, value objects, ошибки, интерфейсы репозиториев
application     — бизнес-сценарии/use cases
infrastructure  — PostgreSQL, JWT, bcrypt, SMTP, CBR SOAP, pgcrypto
transport/http  — HTTP handlers, router, middleware, JSON responses
cmd             — точки входа api, scheduler, migrator
```

## Технологии

- Go 1.25
- PostgreSQL 17
- Docker / Docker Compose
- gorilla/mux
- lib/pq
- golang-jwt/jwt/v5
- logrus
- bcrypt
- HMAC-SHA256
- PostgreSQL `pgcrypto`
- gomail.v2
- beevik/etree
- golang-migrate/migrate

## OpenAPI

Контракт API находится в файле:

```text
api/openapi.yaml
```

Go-код по OpenAPI не генерируется. Спецификация используется как документация и контракт API.

## Переменные окружения

Пример конфигурации находится в файле:

```text
.env.example
```

Основные переменные:

```env
APP_ENV=local
APP_PORT=8080
LOG_LEVEL=debug

DATABASE_URL=postgres://banking_user:banking_password@localhost:5432/banking_service?sslmode=disable

JWT_SECRET=change_me_please
JWT_TTL_HOURS=24

PGP_SYM_KEY=change_me_pgp_key
CARD_HMAC_KEY=change_me_hmac_key

SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM=no-reply@banking.local

CBR_SOAP_URL=https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx

SCHEDULER_INTERVAL_HOURS=12
```

## Запуск в Docker

### 1. Поднять PostgreSQL и Mailpit

```bash
docker compose -f deploy/docker-compose.yaml up -d postgres mailpit
```

### 2. Применить миграции

```bash
docker compose -f deploy/docker-compose.yaml run --rm migrator up
```

### 3. Запустить API

```bash
docker compose -f deploy/docker-compose.yaml up --build api
```

### 4. Запустить scheduler

В отдельном терминале:

```bash
docker compose -f deploy/docker-compose.yaml up --build scheduler
```

API будет доступен по адресу:

```text
http://localhost:8085
```

Mailpit будет доступен по адресу:

```text
http://localhost:8025
```

## Проверка health endpoint

```bash
curl -i http://localhost:8085/health
```

Ожидаемый результат:

```text
HTTP/1.1 200 OK
```

## Миграции

Миграции лежат в каталоге:

```text
migrations/
```

Применить миграции:

```bash
docker compose -f deploy/docker-compose.yaml run --rm migrator up
```

Откатить одну миграцию:

```bash
docker compose -f deploy/docker-compose.yaml run --rm migrator down
```

Посмотреть текущую версию:

```bash
docker compose -f deploy/docker-compose.yaml run --rm migrator version
```

Проверить таблицу миграций:

```bash
docker exec -it banking-postgres psql -U banking_user -d banking_service -c "SELECT version, dirty FROM schema_migrations;"
```

## Подключение к базе данных

Подключиться к PostgreSQL внутри Docker:

```bash
docker exec -it banking-postgres psql -U banking_user -d banking_service
```

Посмотреть таблицы:

```sql
\dt
```

Выйти из PostgreSQL:

```sql
\q
```

Одной командой:

```bash
docker exec -it banking-postgres psql -U banking_user -d banking_service -c "\dt"
```

## Unit-тесты

```bash
go test ./...
```

## Сборка проекта

```bash
go build ./...
```

## E2E smoke-test

Перед запуском e2e-теста должны быть подняты PostgreSQL, Mailpit, API и scheduler.

```bash
docker compose -f deploy/docker-compose.yaml up -d postgres mailpit
docker compose -f deploy/docker-compose.yaml run --rm migrator up
docker compose -f deploy/docker-compose.yaml up -d --build api scheduler
```

Запустить e2e:

```bash
make e2e
```

Или напрямую:

```bash
BASE_URL=http://localhost:8085 ./tests/e2e/smoke.sh
```

## Полная финальная проверка

```bash
make docker-final-check
```

## Основные API-сценарии

### Регистрация пользователя

```bash
TOKEN=$(curl -s -X POST http://localhost:8085/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "user_name",
    "password": "strong-password",
    "firstName": "Test",
    "lastName": "User"
  }' | sed -n 's/.*"accessToken":"\([^"]*\)".*/\1/p')

echo "$TOKEN"
```

### Логин

```bash
TOKEN=$(curl -s -X POST http://localhost:8085/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "strong-password"
  }' | sed -n 's/.*"accessToken":"\([^"]*\)".*/\1/p')

echo "$TOKEN"
```

### Создание счёта

```bash
ACCOUNT_ID=$(curl -s -X POST http://localhost:8085/accounts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"accountType":"current"}' | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')

echo "$ACCOUNT_ID"
```

### Просмотр счетов

```bash
curl -i http://localhost:8085/accounts \
  -H "Authorization: Bearer $TOKEN"
```

### Пополнение счёта

```bash
curl -i -X POST "http://localhost:8085/accounts/$ACCOUNT_ID/deposit" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "amount": "1000.00",
    "description": "Initial deposit",
    "idempotencyKey": "deposit-001"
  }'
```

### Списание средств

```bash
curl -i -X POST "http://localhost:8085/accounts/$ACCOUNT_ID/withdraw" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "amount": "100.00",
    "description": "Withdrawal",
    "idempotencyKey": "withdraw-001"
  }'
```

### Создание MFA challenge

```bash
MFA_RESPONSE=$(curl -s -X POST http://localhost:8085/mfa/challenges \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "purpose": "credit_issue"
  }')

echo "$MFA_RESPONSE"
```

MFA-код можно посмотреть в Mailpit:

```text
http://localhost:8025
```

Или в таблице `email_outbox`:

```bash
docker exec -it banking-postgres psql -U banking_user -d banking_service -c "SELECT payload FROM email_outbox ORDER BY created_at DESC LIMIT 1;"
```

### Подтверждение MFA challenge

```bash
MFA_CHALLENGE_ID="сюда_вставь_id_challenge"
MFA_CODE="сюда_вставь_код"

curl -i -X POST "http://localhost:8085/mfa/challenges/$MFA_CHALLENGE_ID/verify" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"code\": \"$MFA_CODE\"
  }"
```

### Оформление кредита

```bash
curl -i -X POST http://localhost:8085/credits \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-MFA-Challenge-ID: $MFA_CHALLENGE_ID" \
  -d "{
    \"disbursementAccountId\": \"$ACCOUNT_ID\",
    \"repaymentAccountId\": \"$ACCOUNT_ID\",
    \"principalAmount\": \"100000.00\",
    \"termMonths\": 12,
    \"bankMargin\": \"4.0000\"
  }"
```

### Аналитика

```bash
curl -i http://localhost:8085/analytics \
  -H "Authorization: Bearer $TOKEN"
```

### Прогноз баланса

```bash
curl -i "http://localhost:8085/accounts/$ACCOUNT_ID/predict?days=30" \
  -H "Authorization: Bearer $TOKEN"
```

## Административные функции

Для проверки admin endpoints можно вручную назначить пользователю роль администратора:

```bash
docker exec -it banking-postgres psql -U banking_user -d banking_service -c "UPDATE users SET role = 'admin' WHERE email = 'user@example.com';"
```

После этого нужно заново выполнить login, чтобы получить JWT с ролью `admin`.

### Просмотр пользователей:

```bash
curl -i "http://localhost:8085/admin/users?limit=10&offset=0" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### Блокировка счёта администратором

Блокировка счёта выполняется через административный endpoint:

```text
POST /admin/accounts/{accountId}/block
```

Endpoint защищён тремя проверками:

- JWT-аутентификация;
- роль пользователя `admin`;
- подтверждённый MFA challenge с purpose `account_block`.

### 1. Назначить пользователю роль администратора для теста

Для ручной проверки нужно назначить роль администратора через БД.

Пример для пользователя `user@example.com`:

```bash
docker exec -it banking-postgres psql -U banking_user -d banking_service -c "
UPDATE users
SET role = 'admin'
WHERE email = 'user@example.com';
"
```

Проверить, что роль изменилась:

```bash
docker exec -it banking-postgres psql -U banking_user -d banking_service -c "
SELECT id, email, username, role, status
FROM users
WHERE email = 'user@example.com';
"
```

### 2. Получить новый JWT с ролью admin

После изменения роли нужно заново выполнить login, потому что роль пользователя записывается в JWT.

```bash
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8085/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "strong-password"
  }' | sed -n 's/.*"accessToken":"\([^"]*\)".*/\1/p')

echo "$ADMIN_TOKEN"
```

### 3. Найти ID счёта, который нужно заблокировать

Посмотреть список счетов:

```bash
docker exec -it banking-postgres psql -U banking_user -d banking_service -c "
SELECT id, user_id, account_no, status, blocked_reason
FROM accounts;
"
```

Сохранить ID нужного счёта:

```bash
ACCOUNT_ID="сюда_вставь_id_счета"
```

### 4. Создать MFA challenge для блокировки счёта

Для блокировки счёта нужен purpose:

```text
account_block
```

Создать challenge:

```bash
ADMIN_MFA_RESPONSE=$(curl -s -X POST http://localhost:8085/mfa/challenges \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "purpose": "account_block",
    "context": {
      "operation": "admin account block"
    }
  }')

echo "$ADMIN_MFA_RESPONSE"
```

Сохранить ID challenge:

```bash
ADMIN_MFA_CHALLENGE_ID=$(echo "$ADMIN_MFA_RESPONSE" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')

echo "$ADMIN_MFA_CHALLENGE_ID"
```

### 5. Получить MFA-код

Если запущен Mailpit, код можно посмотреть в браузере:

```text
http://localhost:8025
```

Также код можно посмотреть напрямую в таблице `email_outbox`:

```bash
docker exec -it banking-postgres psql -U banking_user -d banking_service -c "
SELECT payload
FROM email_outbox
WHERE template_code = 'mfa_code'
ORDER BY created_at DESC
LIMIT 1;
"
```

Сохранить код:

```bash
ADMIN_MFA_CODE="сюда_вставь_код"
```

### 6. Подтвердить MFA challenge

```bash
curl -i -X POST "http://localhost:8085/mfa/challenges/$ADMIN_MFA_CHALLENGE_ID/verify" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"code\": \"$ADMIN_MFA_CODE\"
  }"
```

Ожидаемый результат:

```text
HTTP/1.1 200 OK
```

В ответе должен быть статус:

```json
{
  "status": "verified"
}
```

### 7. Заблокировать счёт

```bash
curl -i -X POST "http://localhost:8085/admin/accounts/$ACCOUNT_ID/block" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "X-MFA-Challenge-ID: $ADMIN_MFA_CHALLENGE_ID" \
  -d '{
    "reason": "Manual admin block"
  }'
```

Ожидаемый результат:

```text
HTTP/1.1 200 OK
```

Пример успешного ответа:

```json
{
  "accountId": "account-id",
  "status": "blocked",
  "reason": "Manual admin block"
}
```

### 8. Проверить статус счёта в БД

```bash
docker exec -it banking-postgres psql -U banking_user -d banking_service -c "
SELECT id, status, blocked_reason
FROM accounts
WHERE id = '$ACCOUNT_ID';
"
```

Ожидаемый результат:

```text
status = blocked
blocked_reason = Manual admin block
```

После блокировки счёта операции по нему должны быть запрещены бизнес-логикой, потому что статус счёта уже не `active`.

## Безопасность

Реализовано:

- bcrypt-хеширование паролей;
- JWT-аутентификация;
- middleware для защищённых маршрутов;
- role middleware для admin endpoints;
- MFA/2FA для критических операций;
- PGP-шифрование PAN и срока действия карты;
- HMAC-SHA256 для карточных данных;
- bcrypt-хеширование CVV;
- проверка владельца счёта, карты и кредита;
- параметризованные SQL-запросы;
- обработка ошибок в JSON-формате.

## Логирование

Логирование реализовано через `logrus`.

Логируются:

- запуск API;
- запуск scheduler;
- HTTP-запросы;
- ошибки бизнес-операций;
- ошибки миграций;
- обработка кредитных платежей;
- отправка email из outbox.

## Остановка приложения

```bash
docker compose -f deploy/docker-compose.yaml down
```

Остановка с удалением volume БД:

```bash
docker compose -f deploy/docker-compose.yaml down -v --remove-orphans
```
