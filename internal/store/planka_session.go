package store

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

const plankaMiniAppUserAgent = "planka-monitor-mini-app"

func (s *Store) CreatePlankaSessionToken(ctx context.Context, workspaceUsername, secret string, ttl time.Duration) (string, error) {
	workspaceUsername = normalize(workspaceUsername)
	if workspaceUsername == "" {
		return "", errors.New("workspace username is required")
	}
	if secret == "" {
		return "", errors.New("planka secret key is required")
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}

	var userID string
	if err := s.pool.QueryRow(ctx,
		`SELECT id::text FROM user_account WHERE username = $1`, workspaceUsername,
	).Scan(&userID); err != nil {
		return "", err
	}

	token, err := signPlankaAccessToken(userID, secret, time.Now(), ttl)
	if err != nil {
		return "", err
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO session (
			user_id,
			access_token,
			remote_address,
			user_agent,
			created_at,
			updated_at
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			now(),
			now()
		)`,
		userID,
		token,
		plankaMiniAppUserAgent,
		plankaMiniAppUserAgent,
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

func signPlankaAccessToken(subject, secret string, issuedAt time.Time, ttl time.Duration) (string, error) {
	issuedAtUnix := issuedAt.Unix()
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
		"kid": randomTokenID(),
	}
	payload := map[string]any{
		"iat": issuedAtUnix,
		"exp": issuedAtUnix + int64(ttl.Seconds()),
		"sub": subject,
	}

	encodedHeader, err := encodeJWTPart(header)
	if err != nil {
		return "", err
	}
	encodedPayload, err := encodeJWTPart(payload)
	if err != nil {
		return "", err
	}

	unsigned := encodedHeader + "." + encodedPayload
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(unsigned))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return unsigned + "." + signature, nil
}

func encodeJWTPart(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func randomTokenID() string {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(data[:])
}
