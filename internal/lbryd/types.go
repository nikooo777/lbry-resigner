package lbryd

type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AccountListResponse struct {
	Items []Account `json:"items"`
}

type AccountBalanceResponse struct {
	Available string `json:"available"`
}

type Thumbnail struct {
	URL string `json:"url"`
}

type ClaimValue struct {
	Thumbnail Thumbnail `json:"thumbnail"`
}

type Channel struct {
	ClaimID string      `json:"claim_id"`
	Name    string      `json:"name"`
	Txid    string      `json:"txid"`
	Nout    uint64      `json:"nout"`
	IsSpent bool        `json:"is_spent"`
	Value   *ClaimValue `json:"value,omitempty"`
}

type ChannelListResponse struct {
	Items []Channel `json:"items"`
}

// SigningChannel is the daemon's polymorphic signing-channel object: when the
// stream signature is valid, ClaimID is populated and ChannelID is null; when
// invalid, only ChannelID is populated (the orphan channel id).
type SigningChannel struct {
	ClaimID   *string `json:"claim_id,omitempty"`
	ChannelID *string `json:"channel_id,omitempty"`
}

type Stream struct {
	ClaimID                 string          `json:"claim_id"`
	Name                    string          `json:"name"`
	IsSpent                 bool            `json:"is_spent"`
	IsChannelSignatureValid bool            `json:"is_channel_signature_valid"`
	SigningChannel          *SigningChannel `json:"signing_channel,omitempty"`
}

type StreamListResponse struct {
	Items []Stream `json:"items"`
}

type UTXO struct {
	Amount        string `json:"amount"`
	IsMyOutput    bool   `json:"is_my_output"`
	Type          string `json:"type"`
	Confirmations int    `json:"confirmations"`
}

type UTXOListResponse struct {
	Items []UTXO `json:"items"`
}

type StatusWallet struct {
	Blocks       int `json:"blocks"`
	BlocksBehind int `json:"blocks_behind"`
}

type StatusResponse struct {
	Wallet StatusWallet `json:"wallet"`
}

type AccountFundResponse struct {
	Txid string `json:"txid"`
}

type StreamUpdateResponse struct {
	Txid string `json:"txid"`
}

type ChannelListOptions struct {
	AccountID *string `json:"account_id,omitempty"`
	IsSpent   *bool   `json:"is_spent,omitempty"`
}

type StreamUpdateOptions struct {
	ChannelID         *string  `json:"channel_id,omitempty"`
	FundingAccountIDs []string `json:"funding_account_ids,omitempty"`
	ClearTags         *bool    `json:"clear_tags,omitempty"`
	ClearLanguages    *bool    `json:"clear_languages,omitempty"`
	ClearLocations    *bool    `json:"clear_locations,omitempty"`
}
