package transactionapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListTransactionsUsesBackendResponseEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/transactions/" {
			t.Fatalf("expected /transactions/, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("limit") != "2" {
			t.Fatalf("expected limit=2, got %s", r.URL.RawQuery)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": [
				{
					"id": 1,
					"title": "Coffee",
					"amount": 4.5,
					"category": "food",
					"notes": "",
					"transaction_date": "2026-05-31T00:00:00Z",
					"type": "E",
					"created_at": "2026-05-31T10:00:00Z",
					"updated_at": "2026-05-31T10:00:00Z"
				}
			]
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	transactions, err := client.ListTransactions(context.Background(), 2, 0)
	if err != nil {
		t.Fatal(err)
	}

	if len(transactions) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(transactions))
	}
	if transactions[0].Amount != 4.5 {
		t.Fatalf("expected amount 4.5, got %f", transactions[0].Amount)
	}
	if transactions[0].TransactionDate.Format("2006-01-02") != "2026-05-31" {
		t.Fatalf("expected date 2026-05-31, got %s", transactions[0].TransactionDate.Format("2006-01-02"))
	}
	if transactions[0].Type != "E" {
		t.Fatalf("expected type E, got %s", transactions[0].Type)
	}
}

func TestCreateTransactionSendsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/transactions/" {
			t.Fatalf("expected /transactions/, got %s", r.URL.Path)
		}

		var payload SaveTransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}

		if payload.Title != "Lunch" {
			t.Fatalf("expected title Lunch, got %s", payload.Title)
		}
		if payload.Amount != 12.50 {
			t.Fatalf("expected amount 12.50, got %f", payload.Amount)
		}
		if payload.Type != "E" {
			t.Fatalf("expected type E, got %s", payload.Type)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"data": {
				"id": 10,
				"title": "Lunch",
				"amount": 12.5,
				"category": "food",
				"notes": "",
				"transaction_date": "2026-05-31T00:00:00Z",
				"type": "E",
				"created_at": "2026-05-31T10:00:00Z",
				"updated_at": "2026-05-31T10:00:00Z"
			}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	transaction, err := client.CreateTransaction(context.Background(), SaveTransactionRequest{
		Title:  "Lunch",
		Amount: 12.50,
		Type:   "E",
	})
	if err != nil {
		t.Fatal(err)
	}

	if transaction.ID != 10 {
		t.Fatalf("expected id 10, got %d", transaction.ID)
	}
	if transaction.Amount != 12.50 {
		t.Fatalf("expected amount 12.50, got %f", transaction.Amount)
	}
	if transaction.Type != "E" {
		t.Fatalf("expected type E, got %s", transaction.Type)
	}
}
