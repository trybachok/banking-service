-- migrations/000002_core_schema.up.sql

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$;

CREATE TABLE users (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email               text NOT NULL,
  email_normalized    text GENERATED ALWAYS AS (lower(email)) STORED,
  username            varchar(64) NOT NULL,
  username_normalized text GENERATED ALWAYS AS (lower(username)) STORED,
  password_hash       text NOT NULL,
  first_name          varchar(100),
  last_name           varchar(100),
  role                varchar(20) NOT NULL DEFAULT 'customer'
                        CHECK (role IN ('customer', 'admin')),
  status              varchar(20) NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'blocked', 'deleted')),
  last_login_at       timestamptz,
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_users_email_basic CHECK (position('@' in email) > 1),
  CONSTRAINT ck_users_username_format CHECK (username ~ '^[A-Za-z0-9_.-]{3,64}$')
);

CREATE UNIQUE INDEX uq_users_email_normalized
  ON users(email_normalized);

CREATE UNIQUE INDEX uq_users_username_normalized
  ON users(username_normalized);

CREATE TABLE accounts (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id        uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  account_no     varchar(20) NOT NULL,
  account_type   varchar(20) NOT NULL DEFAULT 'current'
                   CHECK (account_type IN ('current', 'credit_repayment')),
  currency       char(3) NOT NULL DEFAULT 'RUB'
                   CHECK (currency = 'RUB'),
  balance        numeric(19,2) NOT NULL DEFAULT 0
                   CHECK (balance >= 0),
  status         varchar(20) NOT NULL DEFAULT 'active'
                   CHECK (status IN ('active', 'blocked', 'closed')),
  blocked_reason text,
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),
  UNIQUE (account_no)
);

CREATE INDEX ix_accounts_user_id ON accounts(user_id);
CREATE INDEX ix_accounts_user_status ON accounts(user_id, status);

CREATE TABLE cards (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id          uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  account_id       uuid NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
  card_alias       varchar(100),
  pan_masked       varchar(19) NOT NULL,
  pan_last4        char(4) NOT NULL,
  pan_encrypted    bytea NOT NULL,
  expiry_encrypted bytea NOT NULL,
  pan_hmac         char(64) NOT NULL,
  integrity_hmac   char(64) NOT NULL,
  cvv_hash         text NOT NULL,
  status           varchar(20) NOT NULL DEFAULT 'active'
                     CHECK (status IN ('active', 'blocked', 'expired', 'closed')),
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  UNIQUE (pan_hmac)
);

CREATE INDEX ix_cards_user_id ON cards(user_id);
CREATE INDEX ix_cards_account_id ON cards(account_id);
CREATE INDEX ix_cards_status ON cards(status);

CREATE TABLE credits (
  id                      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id                 uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  disbursement_account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
  repayment_account_id    uuid NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
  principal_amount        numeric(19,2) NOT NULL CHECK (principal_amount > 0),
  outstanding_principal   numeric(19,2) NOT NULL CHECK (outstanding_principal >= 0),
  annual_interest_rate    numeric(7,4)  NOT NULL CHECK (annual_interest_rate >= 0),
  cbr_key_rate            numeric(7,4)  NOT NULL CHECK (cbr_key_rate >= 0),
  bank_margin             numeric(7,4)  NOT NULL DEFAULT 0 CHECK (bank_margin >= 0),
  term_months             integer       NOT NULL CHECK (term_months > 0),
  annuity_payment         numeric(19,2) NOT NULL CHECK (annuity_payment > 0),
  penalty_rate            numeric(5,4)  NOT NULL DEFAULT 0.1000 CHECK (penalty_rate >= 0),
  status                  varchar(20)   NOT NULL DEFAULT 'active'
                           CHECK (status IN ('active', 'closed', 'overdue', 'defaulted')),
  issued_at               timestamptz   NOT NULL DEFAULT now(),
  closed_at               timestamptz,
  created_at              timestamptz   NOT NULL DEFAULT now(),
  updated_at              timestamptz   NOT NULL DEFAULT now()
);

CREATE INDEX ix_credits_user_id ON credits(user_id);
CREATE INDEX ix_credits_repayment_account_id ON credits(repayment_account_id);
CREATE INDEX ix_credits_status ON credits(status);

CREATE TABLE transactions (
  id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id               uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  source_account_id     uuid REFERENCES accounts(id) ON DELETE RESTRICT,
  destination_account_id uuid REFERENCES accounts(id) ON DELETE RESTRICT,
  card_id               uuid REFERENCES cards(id) ON DELETE SET NULL,
  credit_id             uuid REFERENCES credits(id) ON DELETE SET NULL,
  operation_type        varchar(30) NOT NULL
                          CHECK (operation_type IN (
                            'deposit',
                            'withdrawal',
                            'transfer',
                            'card_payment',
                            'credit_disbursement',
                            'credit_payment',
                            'penalty',
                            'refund'
                          )),
  status                varchar(20) NOT NULL DEFAULT 'completed'
                          CHECK (status IN ('pending', 'completed', 'failed', 'reversed')),
  amount                numeric(19,2) NOT NULL CHECK (amount > 0),
  currency              char(3) NOT NULL DEFAULT 'RUB'
                          CHECK (currency = 'RUB'),
  description           text,
  idempotency_key       varchar(128),
  external_reference    varchar(128),
  created_at            timestamptz NOT NULL DEFAULT now(),
  completed_at          timestamptz
);

CREATE INDEX ix_transactions_user_created
  ON transactions(user_id, created_at DESC);

CREATE INDEX ix_transactions_source_created
  ON transactions(source_account_id, created_at DESC);

CREATE INDEX ix_transactions_destination_created
  ON transactions(destination_account_id, created_at DESC);

CREATE INDEX ix_transactions_credit_id
  ON transactions(credit_id);

CREATE UNIQUE INDEX uq_transactions_idempotency_key
  ON transactions(idempotency_key)
  WHERE idempotency_key IS NOT NULL;

CREATE TABLE payment_schedules (
  id                      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  credit_id               uuid NOT NULL REFERENCES credits(id) ON DELETE CASCADE,
  installment_no          integer NOT NULL CHECK (installment_no > 0),
  due_date                date NOT NULL,
  principal_amount        numeric(19,2) NOT NULL CHECK (principal_amount >= 0),
  interest_amount         numeric(19,2) NOT NULL CHECK (interest_amount >= 0),
  penalty_amount          numeric(19,2) NOT NULL DEFAULT 0 CHECK (penalty_amount >= 0),
  total_amount            numeric(19,2) NOT NULL CHECK (total_amount >= 0),
  paid_amount             numeric(19,2) NOT NULL DEFAULT 0 CHECK (paid_amount >= 0),
  status                  varchar(20) NOT NULL DEFAULT 'pending'
                            CHECK (status IN ('pending', 'paid', 'overdue', 'partially_paid', 'cancelled')),
  last_penalty_applied_at timestamptz,
  paid_transaction_id     uuid REFERENCES transactions(id) ON DELETE SET NULL,
  created_at              timestamptz NOT NULL DEFAULT now(),
  updated_at              timestamptz NOT NULL DEFAULT now(),
  UNIQUE (credit_id, installment_no)
);

CREATE INDEX ix_payment_schedules_credit_due
  ON payment_schedules(credit_id, due_date);

CREATE INDEX ix_payment_schedules_status_due
  ON payment_schedules(status, due_date);

CREATE TABLE email_outbox (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         uuid REFERENCES users(id) ON DELETE SET NULL,
  recipient_email text NOT NULL,
  subject         text NOT NULL,
  template_code   varchar(64) NOT NULL,
  payload         jsonb NOT NULL DEFAULT '{}'::jsonb,
  status          varchar(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'sent', 'failed')),
  attempts        integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
  next_attempt_at timestamptz NOT NULL DEFAULT now(),
  last_error      text,
  created_at      timestamptz NOT NULL DEFAULT now(),
  sent_at         timestamptz
);

CREATE INDEX ix_email_outbox_dispatch
  ON email_outbox(status, next_attempt_at);

CREATE TRIGGER trg_users_set_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_accounts_set_updated_at
BEFORE UPDATE ON accounts
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_cards_set_updated_at
BEFORE UPDATE ON cards
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_credits_set_updated_at
BEFORE UPDATE ON credits
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_payment_schedules_set_updated_at
BEFORE UPDATE ON payment_schedules
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
