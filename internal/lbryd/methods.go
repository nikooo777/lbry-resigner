package lbryd

import "context"

func (c *Client) Status(ctx context.Context) (*StatusResponse, error) {
	out := new(StatusResponse)
	err := c.do(ctx, "status", map[string]any{}, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) AccountList(ctx context.Context, page, pageSize uint64) (*AccountListResponse, error) {
	out := new(AccountListResponse)
	err := c.do(ctx, "account_list", map[string]any{
		"page":      page,
		"page_size": pageSize,
	}, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) AccountBalance(ctx context.Context, accountID *string) (*AccountBalanceResponse, error) {
	out := new(AccountBalanceResponse)
	err := c.do(ctx, "account_balance", map[string]any{
		"account_id": accountID,
	}, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type channelListRequest struct {
	Page     uint64 `json:"page"`
	PageSize uint64 `json:"page_size"`
	ChannelListOptions
}

func (c *Client) ChannelList(ctx context.Context, page, pageSize uint64, opts ChannelListOptions) (*ChannelListResponse, error) {
	out := new(ChannelListResponse)
	err := c.do(ctx, "channel_list", channelListRequest{
		Page:               page,
		PageSize:           pageSize,
		ChannelListOptions: opts,
	}, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) StreamList(ctx context.Context, page, pageSize uint64) (*StreamListResponse, error) {
	out := new(StreamListResponse)
	err := c.do(ctx, "stream_list", map[string]any{
		"page":      page,
		"page_size": pageSize,
	}, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) UTXOList(ctx context.Context, accountID *string, page, pageSize uint64) (*UTXOListResponse, error) {
	out := new(UTXOListResponse)
	err := c.do(ctx, "utxo_list", map[string]any{
		"account_id": accountID,
		"page":       page,
		"page_size":  pageSize,
	}, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) AccountFund(ctx context.Context, fromAccount, toAccount, amount string, outputs uint64, everything bool) (*AccountFundResponse, error) {
	out := new(AccountFundResponse)
	err := c.do(ctx, "account_fund", map[string]any{
		"from_account": fromAccount,
		"to_account":   toAccount,
		"amount":       amount,
		"outputs":      outputs,
		"everything":   everything,
		"broadcast":    true,
	}, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type streamUpdateRequest struct {
	ClaimID  string `json:"claim_id"`
	Blocking bool   `json:"blocking"`
	StreamUpdateOptions
}

func (c *Client) StreamUpdate(ctx context.Context, claimID string, opts StreamUpdateOptions) (*StreamUpdateResponse, error) {
	out := new(StreamUpdateResponse)
	err := c.do(ctx, "stream_update", streamUpdateRequest{
		ClaimID:             claimID,
		Blocking:            true,
		StreamUpdateOptions: opts,
	}, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
