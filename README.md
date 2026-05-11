# banking-service

Учебный REST API банковского сервиса на Go 1.25.

## 1. Запуск приложения в Docker

Поднять PostgreSQL и Mailpit:
```bash
docker compose -f deploy/docker-compose.yaml up -d postgres mailpit
```

Применить миграции:
```bash
docker compose -f deploy/docker-compose.yaml run --rm migrator up
```

Запустить API:
```bash
docker compose -f deploy/docker-compose.yaml up --build api
```

Запустить scheduler:
```bash
docker compose -f deploy/docker-compose.yaml up --build scheduler
```

API доступен по адресу:

http://localhost:8085

Mailpit для просмотра тестовых email:

http://localhost:8025
---

## 2. Проверка health endpoint

```bash
curl -i http://localhost:8085/health
```

Ожидаемый результат `200 OK`
---

## 3. Регистрация пользователя

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
---

## 4. Логин

```bash
TOKEN=$(curl -s -X POST http://localhost:8085/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "strong-password"
  }' | sed -n 's/.*"accessToken":"\([^"]*\)".*/\1/p')

echo "$TOKEN"
```
---

## 5. Создание счёта
```bash
ACCOUNT_ID=$(curl -s -X POST http://localhost:8085/accounts \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $TOKEN" \
-d '{"accountType":"current"}' | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')

echo "$ACCOUNT_ID"
```
---

## 6. Пополнение счёта
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
---

## 7. Создание MFA challenge
```bash
   MFA_RESPONSE=$(curl -s -X POST http://localhost:8085/mfa/challenges \
   -H "Content-Type: application/json" \
   -H "Authorization: Bearer $TOKEN" \
   -d '{
   "purpose": "transfer",
   "context": {
   "operation": "transfer confirmation"
   }
   }')

echo "$MFA_RESPONSE"
```

MFA-код можно посмотреть в Mailpit: http://localhost:8025
---

## 8. Подтверждение MFA challenge
```bash
MFA_CHALLENGE_ID="сюда_вставь_id_challenge"
MFA_CODE="сюда_вставь_код_из_email"

curl -i -X POST "http://localhost:8085/mfa/challenges/$MFA_CHALLENGE_ID/verify" \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $TOKEN" \
-d "{
\"code\": \"$MFA_CODE\"
}"
```
---

