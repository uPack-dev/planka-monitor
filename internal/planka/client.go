package planka

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Client struct {
	baseURL  string
	email    string
	password string
	http     *http.Client

	mu    sync.Mutex
	token string
}

func New(baseURL, email, password string) *Client {
	return &Client{
		baseURL:  strings.TrimRight(baseURL, "/"),
		email:    email,
		password: password,
		http:     &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) Configured() bool {
	return c.baseURL != "" && c.email != "" && c.password != ""
}

// ---------- auth ----------

type tokenResp struct {
	Item string `json:"item"`
}

func (c *Client) login(ctx context.Context) (string, error) {
	body, _ := json.Marshal(map[string]string{
		"emailOrUsername": c.email,
		"password":        c.password,
	})
	req, _ := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/api/access-tokens", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("planka login %d: %s", resp.StatusCode, string(b))
	}
	var tr tokenResp
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", err
	}
	if tr.Item == "" {
		return "", errors.New("planka: empty token")
	}
	return tr.Item, nil
}

func (c *Client) getToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	t := c.token
	c.mu.Unlock()
	if t != "" {
		return t, nil
	}
	t, err := c.login(ctx)
	if err != nil {
		return "", err
	}
	c.mu.Lock()
	c.token = t
	c.mu.Unlock()
	return t, nil
}

func (c *Client) resetToken() {
	c.mu.Lock()
	c.token = ""
	c.mu.Unlock()
}

// ---------- avatar ----------

// UploadAvatar — POST multipart на /api/users/:id/avatar з полем "file".
// Planka сама згенерує перевʼю через sharp і оновить user_account.avatar.
func (c *Client) UploadAvatar(ctx context.Context, userID, filename string, data []byte) error {
	doRequest := func(token string) (*http.Response, error) {
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		fw, err := mw.CreateFormFile("file", filename)
		if err != nil {
			return nil, err
		}
		if _, err := fw.Write(data); err != nil {
			return nil, err
		}
		mw.Close()

		req, _ := http.NewRequestWithContext(ctx, "POST",
			fmt.Sprintf("%s/api/users/%s/avatar", c.baseURL, userID), &buf)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+token)
		return c.http.Do(req)
	}

	token, err := c.getToken(ctx)
	if err != nil {
		return err
	}
	resp, err := doRequest(token)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		c.resetToken()
		token, err = c.getToken(ctx)
		if err != nil {
			return err
		}
		resp, err = doRequest(token)
		if err != nil {
			return err
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload avatar %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// DownloadAvatar — тягне файл аватара з protected-маршруту Planka.
// Шлях: GET {baseURL}/user-avatars/{uploadedFileID}/{variant}.{ext}
// Авторизація — через cookie accessToken=<JWT> (як це робить браузер).
//
// variant: "cover-180" (мініатюра) або "original" (повний розмір).
func (c *Client) DownloadAvatar(ctx context.Context, uploadedFileID, ext, variant string) ([]byte, error) {
	if uploadedFileID == "" || ext == "" {
		return nil, errors.New("empty avatar identifiers")
	}
	if variant == "" {
		variant = "cover-180"
	}

	doRequest := func(token string) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/user-avatars/%s/%s.%s", c.baseURL, uploadedFileID, variant, ext), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Cookie", "accessToken="+token)
		return c.http.Do(req)
	}

	token, err := c.getToken(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := doRequest(token)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		c.resetToken()
		token, err = c.getToken(ctx)
		if err != nil {
			return nil, err
		}
		resp, err = doRequest(token)
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download avatar %d: %s", resp.StatusCode, string(b))
	}
	return io.ReadAll(resp.Body)
}
