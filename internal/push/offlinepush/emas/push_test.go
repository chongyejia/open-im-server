package emas

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openimsdk/open-im-server/v3/internal/push/offlinepush/options"
)

func validOpts() *options.Opts {
	return &options.Opts{Signal: &options.Signal{ClientMsgID: "message-1"}, SendID: "sender", SessionType: 1, Ex: "test-extra"}
}

func TestPushContract(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "POST" || r.Header.Get("X-Internal-Token") != "test-token" {
			t.Error("missing authenticated POST")
		}
		var p pushRequest
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Fatal(err)
		}
		if len(p.UserIDList) != 1 || p.UserIDList[0] != "receiver" || p.SendID != "sender" || p.ClientMsgID != "message-1" || p.SessionType != 1 || p.Ex != "test-extra" || p.Title != "title" || p.Description != "body" {
			t.Errorf("unexpected payload: %+v", p)
		}
		w.Write([]byte(`{"success":true,"data":true}`))
	}))
	defer server.Close()
	c, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Push(context.Background(), []string{"receiver", "sender", "receiver"}, "title", "body", validOpts()); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("got %d calls", calls)
	}
}

func TestPushRejectsInvalidAndNeverBroadcasts(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.Write([]byte(`{"success":true,"data":true}`)) }))
	defer server.Close()
	c, _ := NewClient(server.URL, "test-token")
	group := validOpts()
	group.SessionType = 3
	cases := []struct {
		ids     []string
		opts    *options.Opts
		wantErr bool
	}{
		{nil, validOpts(), false}, {[]string{"sender"}, validOpts(), false},
		{[]string{"receiver"}, nil, true}, {[]string{"receiver"}, group, true},
		{[]string{""}, validOpts(), true}, {make([]string, 1001), validOpts(), true},
	}
	for _, tc := range cases {
		if err := c.Push(context.Background(), tc.ids, "t", "b", tc.opts); (err != nil) != tc.wantErr {
			t.Errorf("error=%v expected=%v", err, tc.wantErr)
		}
	}
	if calls != 0 {
		t.Fatalf("invalid recipients sent %d requests", calls)
	}
}

func TestPushFailuresAndRedirectDoNotRetryOrLeak(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"server", 500, `test-token`}, {"unauthorized", 401, ``}, {"redirect", 307, ``},
		{"business", 200, `{"success":false,"data":true}`}, {"false-data", 200, `{"success":true,"data":false}`}, {"invalid-json", 200, `test-token`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.Header().Set("Location", "/redirected")
				w.WriteHeader(tc.status)
				w.Write([]byte(tc.body))
			}))
			defer server.Close()
			c, _ := NewClient(server.URL, "test-token")
			err := c.Push(context.Background(), []string{"receiver"}, "title", "body", validOpts())
			if err == nil || strings.Contains(err.Error(), "test-token") {
				t.Fatalf("unsafe or missing error: %v", err)
			}
			if calls != 1 {
				t.Errorf("unexpected retry/redirect: %d", calls)
			}
		})
	}
}

func TestConfigurationFailsClosed(t *testing.T) {
	for _, endpoint := range []string{"", "ftp://host", "https://user:secret@host", "https://host?token=secret"} {
		if _, err := NewClient(endpoint, "token"); err == nil {
			t.Errorf("accepted invalid endpoint %q", endpoint)
		}
	}
	if _, err := NewClient("http://mp-api/internal/im/offline-push", " "); err == nil {
		t.Fatal("accepted blank token")
	}
}

func TestCanceledContextDoesNotSend(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++ }))
	defer server.Close()
	c, _ := NewClient(server.URL, "test-token")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := c.Push(ctx, []string{"receiver"}, "t", "b", validOpts()); err == nil {
		t.Fatal("canceled send succeeded")
	}
	if calls != 0 {
		t.Fatal("sent after cancellation")
	}
}
