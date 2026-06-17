package planka

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) CompleteCardDue(ctx context.Context, cardID string) error {
	return c.doJSON(ctx, http.MethodPatch, "/api/cards/"+cardID, map[string]any{
		"isDueCompleted": true,
	}, nil)
}

func (c *Client) CompleteCardDueAs(ctx context.Context, token, cardID string) error {
	return c.doJSONWithToken(ctx, token, http.MethodPatch, "/api/cards/"+cardID, map[string]any{
		"isDueCompleted": true,
	}, nil)
}

func (c *Client) AddCardLabelAs(ctx context.Context, token, cardID, labelID string) error {
	return c.doJSONWithToken(ctx, token, http.MethodPost, "/api/cards/"+cardID+"/card-labels", map[string]any{
		"labelId": labelID,
	}, nil)
}

func (c *Client) CompleteTask(ctx context.Context, taskID string) error {
	return c.doJSON(ctx, http.MethodPatch, "/api/tasks/"+taskID, map[string]any{
		"isCompleted": true,
	}, nil)
}

func (c *Client) CompleteTaskAs(ctx context.Context, token, taskID string) error {
	return c.doJSONWithToken(ctx, token, http.MethodPatch, "/api/tasks/"+taskID, map[string]any{
		"isCompleted": true,
	}, nil)
}

func (c *Client) CreateComment(ctx context.Context, cardID, text string) error {
	return c.doJSON(ctx, http.MethodPost, "/api/cards/"+cardID+"/comments", map[string]any{
		"text": text,
	}, nil)
}

func (c *Client) CreateCommentAs(ctx context.Context, token, cardID, text string) error {
	return c.doJSONWithToken(ctx, token, http.MethodPost, "/api/cards/"+cardID+"/comments", map[string]any{
		"text": text,
	}, nil)
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, out any) error {
	if !c.Configured() {
		return errors.New("planka api is not configured")
	}

	doRequest := func(token string) (*http.Response, error) {
		var payload io.Reader
		if body != nil {
			data, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			payload = bytes.NewReader(data)
		}
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, payload)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
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
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("planka %s %s %d: %s", method, path, resp.StatusCode, string(data))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *Client) doJSONWithToken(ctx context.Context, token, method, path string, body any, out any) error {
	if c.baseURL == "" {
		return errors.New("planka api is not configured")
	}
	if token == "" {
		return errors.New("planka user token is required")
	}

	var payload io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, payload)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("planka %s %s %d: %s", method, path, resp.StatusCode, string(data))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
