package crypto

import (
	"context"
	"database/sql"
	"fmt"

	cardapp "github.com/example/banking-service/internal/application/cards"

	"golang.org/x/crypto/bcrypt"
)

type CardCrypto struct {
	db      *sql.DB
	pgpKey  string
	hmacKey string
}

func NewCardCrypto(db *sql.DB, pgpKey string, hmacKey string) *CardCrypto {
	return &CardCrypto{
		db:      db,
		pgpKey:  pgpKey,
		hmacKey: hmacKey,
	}
}

func (c *CardCrypto) Protect(ctx context.Context, pan string, expiry string, cvv string) (*cardapp.ProtectedCardData, error) {
	var panEncrypted []byte
	var expiryEncrypted []byte

	err := c.db.QueryRowContext(
		ctx,
		`
			SELECT
				pgp_sym_encrypt($1, $3, 'cipher-algo=aes256')::bytea,
				pgp_sym_encrypt($2, $3, 'cipher-algo=aes256')::bytea
		`,
		pan,
		expiry,
		c.pgpKey,
	).Scan(&panEncrypted, &expiryEncrypted)
	if err != nil {
		return nil, fmt.Errorf("encrypt card data with pgcrypto: %w", err)
	}

	cvvHash, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash cvv: %w", err)
	}

	return &cardapp.ProtectedCardData{
		PANEncrypted:    panEncrypted,
		ExpiryEncrypted: expiryEncrypted,
		PANHMAC:         HMACSHA256Hex(c.hmacKey, pan),
		IntegrityHMAC:   HMACSHA256Hex(c.hmacKey, pan, expiry),
		CVVHash:         string(cvvHash),
	}, nil
}

func (c *CardCrypto) Reveal(ctx context.Context, panEncrypted []byte, expiryEncrypted []byte) (*cardapp.CardPlainData, error) {
	var pan string
	var expiry string

	err := c.db.QueryRowContext(
		ctx,
		`
			SELECT
				pgp_sym_decrypt($1::bytea, $3),
				pgp_sym_decrypt($2::bytea, $3)
		`,
		panEncrypted,
		expiryEncrypted,
		c.pgpKey,
	).Scan(&pan, &expiry)
	if err != nil {
		return nil, fmt.Errorf("decrypt card data with pgcrypto: %w", err)
	}

	return &cardapp.CardPlainData{
		PAN:    pan,
		Expiry: expiry,
	}, nil
}

func (c *CardCrypto) VerifyIntegrity(pan string, expiry string, integrityHMAC string) bool {
	expected := HMACSHA256Hex(c.hmacKey, pan, expiry)

	return hmacEqual(expected, integrityHMAC)
}

func (c *CardCrypto) VerifyCVV(cvv string, cvvHash string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(cvvHash), []byte(cvv)); err != nil {
		return fmt.Errorf("verify cvv: %w", err)
	}

	return nil
}

func (c *CardCrypto) PANHMAC(pan string) string {
	return HMACSHA256Hex(c.hmacKey, pan)
}

func hmacEqual(left string, right string) bool {
	if len(left) != len(right) {
		return false
	}

	var result byte

	for i := range left {
		result |= left[i] ^ right[i]
	}

	return result == 0
}
