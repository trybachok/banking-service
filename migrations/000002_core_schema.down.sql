-- migrations/000002_core_schema.down.sql

DROP TRIGGER IF EXISTS trg_payment_schedules_set_updated_at ON payment_schedules;
DROP TRIGGER IF EXISTS trg_credits_set_updated_at ON credits;
DROP TRIGGER IF EXISTS trg_cards_set_updated_at ON cards;
DROP TRIGGER IF EXISTS trg_accounts_set_updated_at ON accounts;
DROP TRIGGER IF EXISTS trg_users_set_updated_at ON users;

DROP TABLE IF EXISTS email_outbox;
DROP TABLE IF EXISTS payment_schedules;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS credits;
DROP TABLE IF EXISTS cards;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS users;

DROP FUNCTION IF EXISTS set_updated_at();