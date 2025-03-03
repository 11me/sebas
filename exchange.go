package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const binanceBaseURL = "https://api.binance.com"

type ExchangeClient struct {
	client *http.Client
	conf   *Config
}

func NewExchangeClient(conf *Config) *ExchangeClient {
	return &ExchangeClient{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		conf: conf,
	}
}

func (c *ExchangeClient) GetDelistings() ([]DelistSchedule, error) {
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	params := url.Values{}
	params.Set("timestamp", ts)

	mac := hmac.New(sha256.New, []byte(c.conf.Binance.Secret))
	mac.Write([]byte(params.Encode()))
	signature := hex.EncodeToString(mac.Sum(nil))
	params.Set("signature", signature)

	endpoint := binanceBaseURL + "/sapi/v1/spot/delist-schedule?" + params.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("X-MBX-APIKEY", c.conf.Binance.Key)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received bad code: %w", err)
	}

	var res []DelistSchedule

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return res, nil
}

type delistTime int64

func (t delistTime) ToTime() time.Time {
	return time.UnixMilli(int64(t))
}

type DelistSchedule struct {
	DelistTime delistTime `json:"delistTime"`
	Symbols    []string   `json:"symbols"`
}
