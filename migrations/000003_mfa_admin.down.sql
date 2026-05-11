-- migrations/000003_mfa_admin.down.sql

DROP INDEX IF EXISTS ix_mfa_challenges_lookup;

DROP TABLE IF EXISTS mfa_challenges;