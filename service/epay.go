package service

import (
	"encoding/json"
	"epay-bot/model"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type EpayService struct {
	client *http.Client
}

func NewEpayService() *EpayService {
	return &EpayService{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (s *EpayService) GetOrders(domain, pid, key string) ([]model.Order, error) {
	u := fmt.Sprintf("https://%s/api.php", domain)
	params := url.Values{}
	params.Add("act", "orders")
	params.Add("pid", pid)
	params.Add("key", key)
	params.Add("limit", "50")

	reqURL := fmt.Sprintf("%s?%s", u, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("User-Agent", "EpayBot-Client/1.0 (Monitoring Orders & Settlements)")

	resp, err := s.client.Do(req)
	if err != nil {
		errMsg := err.Error()
		if key != "" && strings.Contains(errMsg, key) {
			errMsg = strings.ReplaceAll(errMsg, key, "***")
		}
		return nil, fmt.Errorf("request failed: %s", errMsg)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var result model.EpayResponse[model.Order]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	if result.Code == 1 {
		return result.Data, nil
	}

	// Sometimes code != 1 means error or just no data depending on implementation
	// But usually code=1 is success.
	return nil, fmt.Errorf("api error: %s", result.Msg)
}

func (s *EpayService) GetSettlements(domain, pid, key string) ([]model.Settlement, error) {
	u := fmt.Sprintf("https://%s/api.php", domain)
	params := url.Values{}
	params.Add("act", "settle")
	params.Add("pid", pid)
	params.Add("key", key)
	params.Add("limit", "50")

	reqURL := fmt.Sprintf("%s?%s", u, params.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("User-Agent", "EpayBot-Client/1.0 (Monitoring Orders & Settlements)")

	resp, err := s.client.Do(req)
	if err != nil {
		errMsg := err.Error()
		if key != "" && strings.Contains(errMsg, key) {
			errMsg = strings.ReplaceAll(errMsg, key, "***")
		}
		return nil, fmt.Errorf("request failed: %s", errMsg)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var result model.EpayResponse[model.Settlement]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	if result.Code == 1 {
		return result.Data, nil
	}

	return nil, fmt.Errorf("api error: %s", result.Msg)
}
