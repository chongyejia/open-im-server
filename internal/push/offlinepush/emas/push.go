// Package emas delivers offline notifications through the authenticated business API.
package emas

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/openimsdk/open-im-server/v3/internal/push/offlinepush/options"
)

type Client struct {
	endpoint, token string
	httpClient      *http.Client
}

func NewClient(endpoint, token string) (*Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("invalid EMAS offline push endpoint")
	}
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("EMAS offline push token is required")
	}
	return &Client{endpoint: endpoint, token: token, httpClient: &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}

type pushRequest struct {
	ClientMsgID string   `json:"clientMsgId"`
	SendID      string   `json:"sendId"`
	GroupID     string   `json:"groupId,omitempty"`
	SessionType int32    `json:"sessionType"`
	UserIDList  []string `json:"userIdList"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Ex          string   `json:"ex,omitempty"`
}

func (c *Client) Push(ctx context.Context, userIDs []string, title, content string, opts *options.Opts) error {
	if len(userIDs) == 0 {
		return nil
	}
	if len(userIDs) > 1000 {
		return errors.New("EMAS offline push recipient limit exceeded")
	}
	if opts == nil || opts.Signal == nil || opts.Signal.ClientMsgID == "" || opts.SendID == "" || (opts.SessionType != 1 && opts.SessionType != 3) || (opts.SessionType == 3 && opts.GroupID == "") {
		return errors.New("invalid EMAS offline push message metadata")
	}
	ids := make([]string, 0, len(userIDs))
	seen := make(map[string]bool, len(userIDs))
	for _, id := range userIDs {
		if strings.TrimSpace(id) == "" {
			return errors.New("empty EMAS offline push recipient")
		}
		if id != opts.SendID && !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	if len(ids) == 0 {
		return nil
	}
	payload, err := json.Marshal(pushRequest{opts.Signal.ClientMsgID, opts.SendID, opts.GroupID, opts.SessionType, ids, title, content, opts.Ex})
	if err != nil {
		return errors.New("cannot encode EMAS offline push request")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return errors.New("cannot create EMAS offline push request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", c.token)
	resp, err := c.httpClient.Do(req)
	// Do not log the URL, token, body or recipients, and do not retry an ambiguous send.
	if err != nil {
		return errors.New("EMAS offline push transport failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("EMAS offline push HTTP status %d", resp.StatusCode)
	}
	var result struct {
		Success bool `json:"success"`
		Data    bool `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(&result); err != nil || !result.Success || !result.Data {
		return errors.New("EMAS offline push was not accepted")
	}
	return nil
}
