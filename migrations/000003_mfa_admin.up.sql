-- migrations/000003_mfa_admin.up.sql

CREATE TABLE mfa_challenges (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id          uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  purpose          varchar(32) NOT NULL
                     CHECK (purpose IN ('transfer', 'card_payment', 'credit_issue', 'account_block')),
  delivery_channel varchar(16) NOT NULL DEFAULT 'email'
                     CHECK (delivery_channel IN ('email')),
  destination      text NOT NULL,
  code_hash        text NOT NULL,
  context          jsonb NOT NULL DEFAULT '{}'::jsonb,
  status           varchar(20) NOT NULL DEFAULT 'pending'
                     CHECK (status IN ('pending', 'verified', 'expired', 'failed')),
  attempts         integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
  expires_at       timestamptz NOT NULL,
  verified_at      timestamptz,
  created_at       timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ix_mfa_challenges_lookup
  ON mfa_challenges(user_id, purpose, status, expires_at);
