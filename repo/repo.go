package repo

import (
	configs "DZ_ITOG/config"
	"DZ_ITOG/models"
	"database/sql"
	"fmt"

	//"time"

	//"log"
	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

var log = logrus.New()

func init() {
	log.Formatter = &logrus.JSONFormatter{}
	log.Level = logrus.TraceLevel
}

func InitDB(config *configs.Config) (*sql.DB, error) {
	log.WithFields(logrus.Fields{
		"host":     config.Database.Host,
		"port":     config.Database.Port,
		"user":     config.Database.User,
		"database": config.Database.DBName,
	}).Trace("Initializing database connection")

	dbConnStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host, config.Database.Port, config.Database.User, config.Database.Password, config.Database.DBName, config.Database.SSLMode)
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Trace("Failed to open database connection")
		return nil, err
	}

	log.Trace("Database connection opened successfully")

	createDB := `DROP TABLE IF EXISTS items;
	CREATE TABLE IF NOT EXISTS items (
		item_id VARCHAR(255),
		value VARCHAR(255)
	);
	`
	_, err = db.Exec(createDB)
	if err != nil {
		fmt.Println("Exec err on creating items table:", err)
		return nil, err
	}
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		user_id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL
	);
	`
	_, err = db.Exec(createUsersTable)
	if err != nil {
		fmt.Println("Exec err on creating users table:", err)
		return nil, err
	}
	createCommissionsTable := `--DROP TABLE IF EXISTS  transactions, commissions; 
	CREATE TABLE IF NOT EXISTS commissions (
		commission_id SERIAL PRIMARY KEY, 
		transaction_id INT NOT NULL,
		amount FLOAT,
		currency VARCHAR(50),
		transaction_type VARCHAR(50),
		commission FLOAT,
		date VARCHAR(50),
		description TEXT
	);
	`
	_, err = db.Exec(createCommissionsTable)

	if err != nil {
		fmt.Println("Exec err on creating commissions table:", err)
		return nil, err
	}
	createTransactionsTable := `
	CREATE TABLE IF NOT EXISTS transactions (
		transaction_id SERIAL PRIMARY KEY,
		user_id INT NOT NULL,
		amount DECIMAL(10, 2) NOT NULL,
		currency VARCHAR(10) NOT NULL,
		transaction_type VARCHAR(50) NOT NULL,
		category VARCHAR(50),
		date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		description TEXT,
		commission_id INT, 
		FOREIGN KEY (user_id) REFERENCES users(user_id),
		FOREIGN KEY (commission_id) REFERENCES commissions(commission_id) 
	);
	`
	_, err = db.Exec(createTransactionsTable)
	if err != nil {
		fmt.Println("Exec err on creating transactions table:", err)
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		fmt.Println("Ping err")
		return nil, err
	}

	return db, nil
}

func Create(item models.Item, db *sql.DB) error {
	_, err := db.Exec("INSERT INTO items (item_id, value) VALUES ($1, $2)", item.ID, item.Value)
	if err != nil {
		fmt.Println("Exec INSERT")
		return err
	}

	return nil
}

func Read(id string, db *sql.DB) *models.Item {
	var result models.Item
	db.QueryRow("SELECT item_id, value FROM items WHERE item_id = $1", id).Scan(&result)

	return &result
}
func CreateCommission(db *sql.DB, commission models.Commission) error {
	query := `INSERT INTO commissions (transaction_id, amount, currency, transaction_type, commission, date, description) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := db.Exec(query, commission.TransactionID, commission.Amount, commission.Currency, commission.TransactionType, commission.Commission, commission.Date, commission.Description)
	if err != nil {
		log.Println("Error inserting commission:", err)
		return err
	}
	return nil
}
func CreateTransaction(transaction models.Transaction, db *sql.DB) (int, error) {
	log.Info("Inserting transaction into database.")

	var transactionID int
	query := `
        INSERT INTO transactions (user_id, amount, currency, transaction_type, category, description)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING transaction_id;`
	err := db.QueryRow(query, transaction.UserID, transaction.Amount, transaction.Currency, transaction.TransactionType, transaction.Category, transaction.Description).Scan(&transactionID)
	if err != nil {
		log.Printf("Error inserting transaction: %v", err)
		return 0, err
	}
	log.Printf("Transaction inserted with ID: %d", transactionID)
	return transactionID, nil
}

func GetAllTransactions(db *sql.DB) ([]models.Transaction, error) {
	var transactions []models.Transaction
	query := `SELECT transaction_id, user_id, amount, currency, transaction_type, category, date, description FROM transactions`
	rows, err := db.Query(query)
	if err != nil {
		fmt.Println("Error reading transactions:", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var transaction models.Transaction
		if err := rows.Scan(&transaction.ID, &transaction.UserID, &transaction.Amount, &transaction.Currency, &transaction.TransactionType, &transaction.Category, &transaction.Date, &transaction.Description); err != nil {
			fmt.Println("Error scanning transaction:", err)
			continue
		}
		transactions = append(transactions, transaction)
	}
	if err = rows.Err(); err != nil {
		fmt.Println("Error during rows iteration:", err)
		return nil, err
	}
	return transactions, nil
}

func GetTransactionByID(id int64, db *sql.DB) (*models.Transaction, error) {
	var transaction models.Transaction
	query := `SELECT transaction_id, user_id, amount, currency, transaction_type, category, date, description FROM transactions WHERE transaction_id = $1`
	err := db.QueryRow(query, id).Scan(&transaction.ID, &transaction.UserID, &transaction.Amount, &transaction.Currency, &transaction.TransactionType, &transaction.Category, &transaction.Date, &transaction.Description)
	if err != nil {
		fmt.Println("Error getting transaction:", err)
		return nil, err
	}
	return &transaction, nil
}

func UpdateTransaction(id int64, transaction models.Transaction, db *sql.DB) error {
	query := `UPDATE transactions SET user_id = $1, amount = $2, currency = $3, transaction_type = $4, category = $5, description = $6 WHERE transaction_id = $7`
	_, err := db.Exec(query, transaction.UserID, transaction.Amount, transaction.Currency, transaction.TransactionType, transaction.Category, transaction.Description, id)
	if err != nil {
		fmt.Println("Error updating transaction:", err)
		return err
	}
	return nil
}

func DeleteTransaction(id int64, db *sql.DB) error {
	log.Info("Deleting transaction from database.")

	query := `DELETE FROM transactions WHERE transaction_id = $1`
	_, err := db.Exec(query, id)
	if err != nil {
		fmt.Println("Error deleting transaction:", err)
		return err
	}
	return nil
}
