package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const telegramInitDataMaxAge = 24 * time.Hour

var (
	errMissingInitData = errors.New("missing telegram init data")
	errInvalidInitData = errors.New("invalid telegram init data")
	errExpiredInitData = errors.New("expired telegram init data")
)

type telegramInitUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	IsDev     bool   `json:"-"`
}

func validateTelegramInitData(botToken, raw string, now time.Time) (telegramInitUser, error) {
	if strings.TrimSpace(raw) == "" {
		return telegramInitUser{}, errMissingInitData
	}
	values, err := url.ParseQuery(raw)
	if err != nil {
		return telegramInitUser{}, errInvalidInitData
	}
	hash := values.Get("hash")
	if hash == "" || botToken == "" {
		return telegramInitUser{}, errInvalidInitData
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		if key == "hash" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	checkParts := make([]string, 0, len(keys))
	for _, key := range keys {
		checkParts = append(checkParts, key+"="+values.Get(key))
	}
	checkString := strings.Join(checkParts, "\n")

	secretMac := hmac.New(sha256.New, []byte("WebAppData"))
	_, _ = secretMac.Write([]byte(botToken))
	secret := secretMac.Sum(nil)

	dataMac := hmac.New(sha256.New, secret)
	_, _ = dataMac.Write([]byte(checkString))
	expected := dataMac.Sum(nil)

	got, err := hex.DecodeString(hash)
	if err != nil || !hmac.Equal(got, expected) {
		return telegramInitUser{}, errInvalidInitData
	}

	authDate, err := strconv.ParseInt(values.Get("auth_date"), 10, 64)
	if err != nil || authDate <= 0 {
		return telegramInitUser{}, errInvalidInitData
	}
	authTime := time.Unix(authDate, 0)
	if now.Sub(authTime) > telegramInitDataMaxAge || authTime.After(now.Add(5*time.Minute)) {
		return telegramInitUser{}, errExpiredInitData
	}

	var user telegramInitUser
	if err := json.Unmarshal([]byte(values.Get("user")), &user); err != nil || user.ID == 0 {
		return telegramInitUser{}, errInvalidInitData
	}
	user.Username = strings.TrimPrefix(strings.TrimSpace(user.Username), "@")
	user.PhotoURL = strings.TrimSpace(user.PhotoURL)
	return user, nil
}
