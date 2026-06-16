package transactionapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultTimeout = 8 * time.Second

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type Transaction struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	Amount          float64   `json:"amount"`
	Category        string    `json:"category"`
	Notes           string    `json:"notes"`
	TransactionDate time.Time `json:"transaction_date"`
	Type            string    `json:"type"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type SaveTransactionRequest struct {
	Title           string  `json:"title"`
	Amount          float64 `json:"amount"`
	Category        string  `json:"category,omitempty"`
	Notes           string  `json:"notes,omitempty"`
	TransactionDate string  `json:"transaction_date,omitempty"`
	Type            string  `json:"type"`
}

type transactionResponse struct {
	Data Transaction `json:"data"`
}

type transactionsResponse struct {
	Data []Transaction `json:"data"`
}

type apiErrorResponse struct {
	Error string `json:"error"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func (c *Client) ListTransactions(ctx context.Context, limit int, offset int) ([]Transaction, error) {
	query := url.Values{}
	query.Set("limit", strconv.Itoa(limit))
	query.Set("offset", strconv.Itoa(offset))

	var response transactionsResponse
	if err := c.do(ctx, http.MethodGet, "/transactions/?"+query.Encode(), nil, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (c *Client) GetTransaction(ctx context.Context, id int64) (Transaction, error) {
	var response transactionResponse
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/transactions/%d", id), nil, &response); err != nil {
		return Transaction{}, err
	}

	return response.Data, nil
}

func (c *Client) CreateTransaction(ctx context.Context, payload SaveTransactionRequest) (Transaction, error) {
	var response transactionResponse
	if err := c.do(ctx, http.MethodPost, "/transactions/", payload, &response); err != nil {
		return Transaction{}, err
	}

	return response.Data, nil
}

func (c *Client) UpdateTransaction(ctx context.Context, id int64, payload SaveTransactionRequest) (Transaction, error) {
	var response transactionResponse
	if err := c.do(ctx, http.MethodPut, fmt.Sprintf("/transactions/%d", id), payload, &response); err != nil {
		return Transaction{}, err
	}

	return response.Data, nil
}

func (c *Client) DeleteTransaction(ctx context.Context, id int64) error {
	return c.do(ctx, http.MethodDelete, fmt.Sprintf("/transactions/%d", id), nil, nil)
}

func (c *Client) do(ctx context.Context, method string, path string, payload any, target any) error {
	var body io.Reader
	if payload != nil {
		buffer := bytes.NewBuffer(nil)
		if err := json.NewEncoder(buffer).Encode(payload); err != nil {
			return err
		}
		body = buffer
	}

	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return decodeAPIError(response)
	}

	if target == nil || response.StatusCode == http.StatusNoContent {
		return nil
	}

	return json.NewDecoder(response.Body).Decode(target)
}

func decodeAPIError(response *http.Response) error {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("api returned %s", response.Status)
	}

	var apiError apiErrorResponse
	if err := json.Unmarshal(body, &apiError); err == nil && apiError.Error != "" {
		return fmt.Errorf("api returned %s: %s", response.Status, apiError.Error)
	}

	message := strings.TrimSpace(string(body))
	if message == "" {
		message = response.Status
	}

	return fmt.Errorf("api returned %s: %s", response.Status, message)
}
