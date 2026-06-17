package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func signedInitData(t *testing.T, botToken string, values url.Values) string {
	t.Helper()
	keys := make([]string, 0, len(values))
	for key := range values {
		if key != "hash" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+values.Get(key))
	}

	secretMac := hmac.New(sha256.New, []byte("WebAppData"))
	_, _ = secretMac.Write([]byte(botToken))
	secret := secretMac.Sum(nil)

	dataMac := hmac.New(sha256.New, secret)
	_, _ = dataMac.Write([]byte(strings.Join(parts, "\n")))
	values.Set("hash", hex.EncodeToString(dataMac.Sum(nil)))
	return values.Encode()
}

func TestValidateTelegramInitData(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	botToken := "123456:test-token"
	raw := signedInitData(t, botToken, url.Values{
		"auth_date": {now.Add(-time.Minute).Format("20060102150405")},
	})
	if _, err := validateTelegramInitData(botToken, raw, now); err == nil {
		t.Fatalf("expected invalid auth_date format")
	}

	raw = signedInitData(t, botToken, url.Values{
		"auth_date": {strconvFormat(now.Add(-time.Minute).Unix())},
		"query_id":  {"query"},
		"user":      {`{"id":42,"first_name":"Yan","username":"yan_op","photo_url":"https://t.me/i/userpic/320/test.jpg"}`},
	})
	user, err := validateTelegramInitData(botToken, raw, now)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if user.ID != 42 || user.Username != "yan_op" {
		t.Fatalf("unexpected user: %+v", user)
	}
	if user.PhotoURL != "https://t.me/i/userpic/320/test.jpg" {
		t.Fatalf("unexpected photo url: %q", user.PhotoURL)
	}

	raw = signedInitData(t, botToken, url.Values{
		"auth_date": {strconvFormat(now.Add(-time.Minute).Unix())},
		"query_id":  {"query"},
		"signature": {"telegram-ed25519-signature"},
		"user":      {`{"id":43,"first_name":"Yan","username":"yan_with_signature"}`},
	})
	user, err = validateTelegramInitData(botToken, raw, now)
	if err != nil {
		t.Fatalf("validate with signature: %v", err)
	}
	if user.ID != 43 || user.Username != "yan_with_signature" {
		t.Fatalf("unexpected signature user: %+v", user)
	}
}

func TestValidateTelegramInitDataRejectsBadHashAndExpiredData(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	botToken := "123456:test-token"
	values := url.Values{
		"auth_date": {strconvFormat(now.Add(-time.Minute).Unix())},
		"user":      {`{"id":42,"username":"yan"}`},
	}
	raw := signedInitData(t, botToken, values)
	if _, err := validateTelegramInitData(botToken, raw+"&extra=1", now); err == nil {
		t.Fatalf("expected bad hash rejection")
	}

	expired := signedInitData(t, botToken, url.Values{
		"auth_date": {strconvFormat(now.Add(-25 * time.Hour).Unix())},
		"user":      {`{"id":42,"username":"yan"}`},
	})
	if _, err := validateTelegramInitData(botToken, expired, now); err != errExpiredInitData {
		t.Fatalf("expected expired error, got %v", err)
	}
}

func TestAuthenticateDevAPI(t *testing.T) {
	t.Parallel()

	s := &Server{devAuth: DevAuthConfig{
		Token:             "dev-secret",
		TelegramUsername:  "Yan_Op",
		WorkspaceUsername: "yan",
		IsAdmin:           true,
	}}
	r := httptest.NewRequest("GET", "/api/v1/me", nil)
	r.Header.Set(monitorDevAuthHeader, "dev-secret")

	user, attempted, ok := s.authenticateDevAPI(r)
	if !attempted || !ok {
		t.Fatalf("expected valid dev auth, attempted=%t ok=%t", attempted, ok)
	}
	if !user.IsDev || user.Username != "yan_op" {
		t.Fatalf("unexpected dev user: %+v", user)
	}

	r = httptest.NewRequest("GET", "/api/v1/me", nil)
	r.Header.Set(monitorDevAuthHeader, "wrong")
	if _, attempted, ok = s.authenticateDevAPI(r); !attempted || ok {
		t.Fatalf("expected attempted invalid dev auth, attempted=%t ok=%t", attempted, ok)
	}

	s = &Server{}
	r = httptest.NewRequest("GET", "/api/v1/me", nil)
	r.Header.Set(monitorDevAuthHeader, "stale")
	if _, attempted, ok = s.authenticateDevAPI(r); attempted || ok {
		t.Fatalf("expected disabled dev auth to ignore header, attempted=%t ok=%t", attempted, ok)
	}

	s = &Server{devAuth: DevAuthConfig{
		Token:             "dev-secret",
		TelegramUsername:  "yan",
		WorkspaceUsername: "yan",
	}}
	r = httptest.NewRequest("GET", "/api/v1/me", nil)
	r.AddCookie(&http.Cookie{Name: "monitor_dev_auth", Value: "dev-secret"})
	if _, attempted, ok = s.authenticateDevAPI(r); attempted || ok {
		t.Fatalf("expected dev auth cookie to be ignored, attempted=%t ok=%t", attempted, ok)
	}
}

func strconvFormat(v int64) string {
	return strconv.FormatInt(v, 10)
}
