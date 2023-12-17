# zopsmart_CCvalid
Initial commit: Set up a simple Go web application for credit card validation and management.

This commit includes the basic structure of the Go project, integration with the gofr framework, database connection to MySQL, and initial implementation of endpoints for credit card validation, addition, retrieval, and deletion.

- Implemented credit card validation using the Luhn algorithm.
- Added HTTP handlers for validating, adding, retrieving, and deleting credit card information.
- Established a connection to a MySQL database for storing credit card data.
- Included a sample SQLite database file for testing.

# Usage
- Run `go run main.go` to start the web server.
- Access the endpoints:
  - POST /validate: Validate a credit card number.
  - POST /card: Add a new credit card.
  - GET /cards: Retrieve all credit cards.
  - DELETE /card/:id: Delete a credit card by ID.

# Dependencies
- gofr.dev/pkg/gofr: Lightweight Go framework for web applications.
- github.com/go-sql-driver/mysql: MySQL driver for Go's database/sql.

Feel free to explore and expand upon this project!
