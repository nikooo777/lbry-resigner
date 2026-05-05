package lbryd

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type recorder struct {
	server *httptest.Server
	body   map[string]any
}

func newRecorder(t *testing.T, response string) *recorder {
	t.Helper()
	rec := &recorder{}
	handler := func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		err = json.Unmarshal(raw, &rec.body)
		if err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_, _ = io.WriteString(w, response)
	}
	rec.server = httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(rec.server.Close)
	return rec
}

func (r *recorder) params(t *testing.T) map[string]any {
	t.Helper()
	params, ok := r.body["params"].(map[string]any)
	if !ok {
		t.Fatalf("params is not an object: %v", r.body["params"])
	}
	return params
}

func TestEnvelopeShape(t *testing.T) {
	rec := newRecorder(t, `{"jsonrpc":"2.0","result":{"wallet":{"blocks":1,"blocks_behind":0}}}`)
	client := NewClient(rec.server.URL)

	_, err := client.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got := rec.body["jsonrpc"]; got != "2.0" {
		t.Errorf("jsonrpc=%v want 2.0", got)
	}
	if got := rec.body["method"]; got != "status" {
		t.Errorf("method=%v want status", got)
	}
	if _, ok := rec.body["id"].(float64); !ok {
		t.Errorf("id=%v (%T) want number", rec.body["id"], rec.body["id"])
	}
}

func TestRPCErrorDecoding(t *testing.T) {
	rec := newRecorder(t, `{"jsonrpc":"2.0","error":{"code":-1,"message":"boom"}}`)
	client := NewClient(rec.server.URL)

	_, err := client.Status(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) {
		t.Fatalf("got %T (%v), want *RPCError", err, err)
	}
	if rpcErr.Code != -1 {
		t.Errorf("code=%d want -1", rpcErr.Code)
	}
	if rpcErr.Message != "boom" {
		t.Errorf("message=%q want boom", rpcErr.Message)
	}
}

func TestChannelListIsSpent(t *testing.T) {
	cases := []struct {
		name    string
		opts    ChannelListOptions
		present bool
		value   bool
	}{
		{"nil omits", ChannelListOptions{}, false, false},
		{"true sends true", ChannelListOptions{IsSpent: Ptr(true)}, true, true},
		{"false sends false", ChannelListOptions{IsSpent: Ptr(false)}, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := newRecorder(t, `{"jsonrpc":"2.0","result":{"items":[]}}`)
			client := NewClient(rec.server.URL)
			_, err := client.ChannelList(context.Background(), 1, 10, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			params := rec.params(t)
			value, present := params["is_spent"]
			if present != tc.present {
				t.Fatalf("is_spent present=%v want %v (params=%v)", present, tc.present, params)
			}
			if tc.present {
				got, ok := value.(bool)
				if !ok {
					t.Fatalf("is_spent type=%T want bool", value)
				}
				if got != tc.value {
					t.Errorf("is_spent=%v want %v", got, tc.value)
				}
			}
		})
	}
}

func TestAccountFundBroadcast(t *testing.T) {
	rec := newRecorder(t, `{"jsonrpc":"2.0","result":{"txid":"abc"}}`)
	client := NewClient(rec.server.URL)
	_, err := client.AccountFund(context.Background(), "a", "b", "1.0", 1, false)
	if err != nil {
		t.Fatal(err)
	}
	params := rec.params(t)
	value, ok := params["broadcast"].(bool)
	if !ok {
		t.Fatalf("broadcast type=%T want bool", params["broadcast"])
	}
	if !value {
		t.Errorf("broadcast=%v want true", value)
	}
}

func TestStreamUpdateBlocking(t *testing.T) {
	rec := newRecorder(t, `{"jsonrpc":"2.0","result":{"txid":"abc"}}`)
	client := NewClient(rec.server.URL)
	_, err := client.StreamUpdate(context.Background(), "claimid", StreamUpdateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	params := rec.params(t)
	value, ok := params["blocking"].(bool)
	if !ok {
		t.Fatalf("blocking type=%T want bool", params["blocking"])
	}
	if !value {
		t.Errorf("blocking=%v want true", value)
	}
}

func authHeaderServer(t *testing.T, seen *string) *httptest.Server {
	t.Helper()
	handler := func(w http.ResponseWriter, r *http.Request) {
		*seen = r.Header.Get(AuthTokenHeader)
		_, _ = io.WriteString(w, `{"jsonrpc":"2.0","result":{"wallet":{"blocks":1,"blocks_behind":0}}}`)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	t.Cleanup(server.Close)
	return server
}

func TestAuthHeaderSetWhenConfigured(t *testing.T) {
	var seen string
	server := authHeaderServer(t, &seen)
	client := NewClient(server.URL, WithAuthToken("xyz"))
	_, err := client.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if seen != "xyz" {
		t.Errorf("auth header=%q want xyz", seen)
	}
}

func TestAuthHeaderAbsentWithoutOption(t *testing.T) {
	var seen string
	server := authHeaderServer(t, &seen)
	client := NewClient(server.URL)
	_, err := client.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if seen != "" {
		t.Errorf("auth header=%q want empty", seen)
	}
}

func TestAccountBalanceOmitsAccountIDWhenNil(t *testing.T) {
	rec := newRecorder(t, `{"jsonrpc":"2.0","result":{"available":"0"}}`)
	client := NewClient(rec.server.URL)
	_, err := client.AccountBalance(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	params := rec.params(t)
	if _, ok := params["account_id"]; ok {
		t.Errorf("account_id should be absent, params=%v", params)
	}

	rec2 := newRecorder(t, `{"jsonrpc":"2.0","result":{"available":"0"}}`)
	client2 := NewClient(rec2.server.URL)
	id := "acc1"
	_, err = client2.AccountBalance(context.Background(), &id)
	if err != nil {
		t.Fatal(err)
	}
	params2 := rec2.params(t)
	if got := params2["account_id"]; got != "acc1" {
		t.Errorf("account_id=%v want acc1", got)
	}
}

func TestUTXOListOmitsAccountIDWhenNil(t *testing.T) {
	rec := newRecorder(t, `{"jsonrpc":"2.0","result":{"items":[]}}`)
	client := NewClient(rec.server.URL)
	_, err := client.UTXOList(context.Background(), nil, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	params := rec.params(t)
	if _, ok := params["account_id"]; ok {
		t.Errorf("account_id should be absent, params=%v", params)
	}

	rec2 := newRecorder(t, `{"jsonrpc":"2.0","result":{"items":[]}}`)
	client2 := NewClient(rec2.server.URL)
	id := "acc1"
	_, err = client2.UTXOList(context.Background(), &id, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	params2 := rec2.params(t)
	if got := params2["account_id"]; got != "acc1" {
		t.Errorf("account_id=%v want acc1", got)
	}
}
