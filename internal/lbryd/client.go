package lbryd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	endpoint   string
	httpClient *http.Client
}

func NewClient(endpoint string) *Client {
	return &Client{
		endpoint:   endpoint,
		httpClient: &http.Client{},
	}
}

type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message)
}

func Ptr[T any](v T) *T {
	return &v
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
	ID      int    `json:"id"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *RPCError       `json:"error"`
}

func (c *Client) do(ctx context.Context, method string, params any, out any) error {
	body, err := json.Marshal(rpcRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	})
	if err != nil {
		return fmt.Errorf("%s: marshal request: %w", method, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("%s: build request: %w", method, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", method, err)
	}
	defer resp.Body.Close()

	envelope := new(rpcResponse)
	err = json.NewDecoder(resp.Body).Decode(envelope)
	if err != nil {
		return fmt.Errorf("%s: decode response: %w", method, err)
	}
	if envelope.Error != nil {
		return envelope.Error
	}
	if out == nil {
		return nil
	}
	err = json.Unmarshal(envelope.Result, out)
	if err != nil {
		return fmt.Errorf("%s: decode result: %w", method, err)
	}
	return nil
}
