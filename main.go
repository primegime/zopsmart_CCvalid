package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"gofr.dev/pkg/gofr"
	_ "github.com/go-sql-driver/mysql"
)

// CreditCard struct to match your database structure
type CreditCard struct {
	ID        int       `json:"id"`
	Number    string    `json:"card_number"`
	IsValid   bool      `json:"is_valid"`
	CreatedAt time.Time `json:"created_at"`
	Owner     string    `json:"owner"` // Added Owner field
}

func main() {
	// Database connection (assuming a MySQL database)
	dataSourceName := "blog_user:password@tcp(localhost:3306)/credit_card_db"
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := gofr.New()

	// POST: Validate a credit card number
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

	// POST: Add a new credit card
	app.POST("/card", func(c *gofr.Context) (interface{}, error) {
		var card CreditCard
		if err := c.Bind(&card); err != nil {
			return nil, fmt.Errorf("invalid request: %v", err)
		}

		// Insert into database
		_, err := db.Exec("INSERT INTO credit_cards (card_number, is_valid, owner) VALUES (?, ?, ?)", card.Number, card.IsValid, card.Owner)
		if err != nil {
			return nil, err
		}

		return card, nil
	})

	// GET: Retrieve all credit cards
	app.GET("/cards", func(c *gofr.Context) (interface{}, error) {
		rows, err := db.Query("SELECT id, card_number, is_valid, created_at FROM credit_cards")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var cards []CreditCard
		for rows.Next() {
			var card CreditCard
			var createdAt string // or use []byte for createdAt

			// Scan the created_at into a string
			if err := rows.Scan(&card.ID, &card.Number, &card.IsValid, &createdAt); err != nil {
				return nil, err
			}

			// Parse the string into time.Time
			card.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAt) // Adjust format as per your database format
			if err != nil {
				return nil, err
			}

			cards = append(cards, card)
		}

		return cards, nil
	})

	// DELETE: Remove a credit card by ID
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

	app.Start()
}

// isValidCreditCard applies the Luhn algorithm to validate a credit card number
func isValidCreditCard(number string) bool {
	var sum int
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false // Non-numeric digit found, invalid card number
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = (digit % 10) + 1
			}
		}

		sum += digit
		alternate = !alternate
	}

	// A valid credit card number will result in a sum that is a multiple of 10
	return sum%10 == 0
}
