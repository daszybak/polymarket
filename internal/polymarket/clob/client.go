// Package clob is used to call clob polymarket endpoints.
package clob

import (
	"net/http"
	"time"

	"github.com/daszybak/polymarket/pkg/httpclient"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func New(baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
	}
}

type MarketToken struct {
	Outcome string `json:"outcome"`
	Price
	TokenID string `json:"token_id"`
	Winner  bool   `json:"winner"`
}

type Market struct{}

func (c *Client) GetMarketByConditionID(conditionID string) (*Market, error) {
	return httpclient.GetResource[*Market](c.httpClient, c.baseURL, "/markets"+conditionID, []int{200})
}
