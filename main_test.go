package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gofr.dev/pkg/gofr"
)

// CreditCard struct to match your database structure
type CreditCard struct {
	ID        int       `json:"id"`
	Number    string    `json:"card_number"`
	IsValid   bool      `json:"is_valid"`
	CreatedAt time.Time `json:"created_at"`
	Owner     string    `json:"owner"` // Added Owner field
}

// TestValidateCreditCardEndpoint tests the /validate endpoint
func TestValidateCreditCardEndpoint(t *testing.T) {
	app := gofr.New()
	setupHandlers(app)

	testCases := []struct {
		name       string
		inputJSON  string
		expected   string
		statusCode int
	}{
		{
			name:       "ValidCreditCard",
			inputJSON:  `{"card_number": "4111111111111111"}`,
			expected:   `{"is_valid":true}`,
			statusCode: http.StatusOK,
		},
		{
			name:       "InvalidCreditCard",
			inputJSON:  `{"card_number": "1234567890123456"}`,
			expected:   `{"is_valid":false}`,
			statusCode: http.StatusOK,
		},
		{
			name:       "InvalidJSON",
			inputJSON:  `{"invalid_field": "value"}`,
			expected:   `{"error":"invalid request: json: unknown field \"invalid_field\""}`,
			statusCode: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/validate", strings.NewReader(testCase.inputJSON))
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, testCase.statusCode, rec.Code)
			assert.JSONEq(t, testCase.expected, rec.Body.String())
		})
	}
}

// TestAddCreditCardEndpoint tests the /card endpoint
func TestAddCreditCardEndpoint(t *testing.T) {
	app := gofr.New()
	setupHandlers(app)

	testCases := []struct {
		name       string
		inputJSON  string
		expected   string
		statusCode int
	}{
		{
			name:       "ValidCreditCard",
			inputJSON:  `{"card_number": "4111111111111111", "is_valid": true, "owner": "John Doe"}`,
			expected:   `{"ID":1,"card_number":"4111111111111111","is_valid":true,"created_at":"0001-01-01T00:00:00Z","owner":"John Doe"}`,
			statusCode: http.StatusOK,
		},
		{
			name:       "InvalidCreditCard",
			inputJSON:  `{"card_number": "1234567890123456", "is_valid": false, "owner": "Jane Doe"}`,
			expected:   `{"error":"pq: duplicate key value violates unique constraint \"credit_cards_pkey\""}`,
			statusCode: http.StatusInternalServerError,
		},
		{
			name:       "InvalidJSON",
			inputJSON:  `{"invalid_field": "value"}`,
			expected:   `{"error":"invalid request: json: unknown field \"invalid_field\""}`,
			statusCode: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/card", strings.NewReader(testCase.inputJSON))
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, testCase.statusCode, rec.Code)
			assert.JSONEq(t, testCase.expected, rec.Body.String())
		})
	}
}

// TestGetAllCreditCardsEndpoint tests the /cards endpoint
func TestGetAllCreditCardsEndpoint(t *testing.T) {
	app := gofr.New()
	setupHandlers(app)

	t.Run("GetAllCreditCards", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/cards", nil)
		rec := httptest.NewRecorder()

		app.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, rec.Body.String())
	})
}

// TestDeleteCreditCardEndpoint tests the /card/:id endpoint
func TestDeleteCreditCardEndpoint(t *testing.T) {
	app := gofr.New()
	setupHandlers(app)

	testCases := []struct {
		name       string
		id         string
		expected   string
		statusCode int
	}{
		{
			name:       "ValidDelete",
			id:         "1",
			expected:   `"Card with ID 1 deleted successfully"`,
			statusCode: http.StatusOK,
		},
		{
			name:       "InvalidID",
			id:         "invalid",
			expected:   `"invalid ID: strconv.Atoi: parsing \"invalid\": invalid syntax"`,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "NonExistentID",
			id:         "999",
			expected:   `"No card found with specified ID"`,
			statusCode: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/card/"+testCase.id, nil)
			rec := httptest.NewRecorder()

			app.ServeHTTP(rec, req)

			assert.Equal(t, testCase.statusCode, rec.Code)
			assert.JSONEq(t, testCase.expected, rec.Body.String())
		})
	}
}

// TestMain tests the main function
func TestMain(t *testing.T) {
	// Test the main function (entry point)
	// Note: This is a basic test to ensure there's no panic on startup.
	main()
}

// TestIsValidCreditCard tests the isValidCreditCard function
func TestIsValidCreditCard(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"ValidCreditCard", "4111111111111111", true},
		{"InvalidCreditCard", "1234567890123456", false},
		{"InvalidNonNumeric", "abc", false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := isValidCreditCard(testCase.input)
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func setupHandlers(app *gofr.App) {
	app.POST("/validate", func(c *gofr.Context) (interface{}, error) {
		var request struct {
			CardNumber string `json:"card_number"`
		}

		if err := c.Bind(&request); err != nil {
			return nil, fmt.Errorf("invalid request: %v", err)
		}

		isValid := isValidCreditCard(request.CardNumber)
		return map[string]bool{"is_valid": isValid}, nil
	})

	app.POST("/card", func(c *gofr.Context) (interface{}, error) {
		var card CreditCard
		if err := c.Bind(&card); err != nil {
			return nil, fmt.Errorf("invalid request: %v", err)
		}

		_, err := db.Exec("INSERT INTO credit_cards (card_number, is_valid, owner) VALUES (?, ?, ?)", card.Number, card.IsValid, card.Owner)
		if err != nil {
			return nil, err
		}

		return card, nil
	})

	app.GET("/cards", func(c *gofr.Context) (interface{}, error) {
		rows, err := db.Query("SELECT id, card_number, is_valid, created_at FROM credit_cards")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var cards []CreditCard
		for rows.Next() {
			var card CreditCard
			var createdAt string

			if err := rows.Scan(&card.ID, &card.Number, &card.IsValid, &createdAt); err != nil {
				return nil, err
			}

			card.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAt)
			if err != nil {
				return nil, err
			}

			cards = append(cards, card)
		}

		return cards, nil
	})

	app.DELETE("/card/:id", func(c *gofr.Context) (interface{}, error) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Println("Error converting ID:", err)
			return nil, fmt.Errorf("invalid ID: %v", err)
		}

		result, err := db.Exec("DELETE FROM credit_cards WHERE id = ?", id)
		if err != nil {
			log.Println("Error executing DELETE:", err)
			return nil, err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println("Error getting rows affected:", err)
			return nil, err
		}

		if rowsAffected == 0 {
			return "No card found with specified ID", nil
		}

		return fmt.Sprintf("Card with ID %d deleted successfully", id), nil
	})
}
