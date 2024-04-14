package handlers

import (
	"DZ_ITOG/models"
	"DZ_ITOG/repo"
	"DZ_ITOG/service"
	"database/sql"
	"encoding/json"
	"fmt"

	//"log"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

var log = logrus.New()

func Item(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			var item models.Item
			if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if err := repo.Create(item, db); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(models.ItemResponse{Item: item, Ok: true})

		case "GET":
			id := r.URL.Query().Get("id")
			if id != "" {
				item := repo.Read(id, db)
				if item != nil {
					json.NewEncoder(w).Encode(models.ItemResponse{Item: *item, Ok: true})
				} else {
					http.NotFound(w, r)
				}
			} else {
				json.NewEncoder(w).Encode(models.ItemResponse{Item: models.Item{}, Ok: false})
			}
		}
	}
}

func CreateTransaction(c *gin.Context) {
	var transaction models.Transaction
	if err := c.BindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transactionID, err := repo.CreateTransaction(transaction, c.MustGet("db").(*sql.DB))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}
	transaction.ID = transactionID
	var commissionRate float64

	if transaction.TransactionType == "перевод" {
		switch transaction.Currency {
		case "USD":
			commissionRate = 0.02
		case "RUB":
			commissionRate = 0.05
		}
	} else if transaction.TransactionType == "покупка" || transaction.TransactionType == "пополнение" {
		commissionRate = 0
	}

	if commissionRate > 0 {
		commission := models.Commission{
			TransactionID:   transactionID,
			Amount:          transaction.Amount,
			Currency:        transaction.Currency,
			TransactionType: transaction.TransactionType,
			Commission:      transaction.Amount * commissionRate,
			Date:            time.Now().Format("2006-01-02"),
			Description:     fmt.Sprintf("Комиссия %.2f%% от суммы", commissionRate*100),
		}
		log.WithFields(logrus.Fields{
			"module":     "transactionHandler",
			"operation":  "CreateTransaction",
			"commission": commission,
		}).Debug("Commission calculated")
		if err := repo.CreateCommission(c.MustGet("db").(*sql.DB), commission); err != nil {
			log.Printf("Error saving commission: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save commission"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"transaction": transaction, "commission": commission})
	} else {
		c.JSON(http.StatusCreated, gin.H{"transaction": transaction, "commission": nil})
	}
}

func GetAllTransactions(c *gin.Context) {
	dbInterface, ok := c.Get("db")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not available"})
		return
	}

	db, ok := dbInterface.(*sql.DB)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid database connection type"})
		return
	}

	transactions, err := repo.GetAllTransactions(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

func GetTransactionByID(c *gin.Context) {
	idStr := c.Param("id")
	targetCurrency := c.Query("currency")

	log.WithFields(logrus.Fields{
		"module":    "transactionHandler",
		"operation": "GetTransactionByID",
		"id":        idStr,
		"currency":  targetCurrency,
	}).Debug("Received parameters for transaction retrieval")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.WithFields(logrus.Fields{
			"module":    "transactionHandler",
			"operation": "GetTransactionByID",
			"id":        idStr,
		}).Error("Invalid transaction ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID format"})
		return
	}

	db, ok := c.MustGet("db").(*sql.DB)
	if !ok {
		log.WithFields(logrus.Fields{
			"module":    "transactionHandler",
			"operation": "GetTransactionByID",
		}).Error("Database connection not available")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not available"})
		return
	}

	log.WithFields(logrus.Fields{
		"module":    "transactionHandler",
		"operation": "GetTransactionByID",
		"id":        id,
	}).Debug("Attempting to fetch transaction from database")

	transaction, err := repo.GetTransactionByID(id, db)
	if err != nil {
		log.WithFields(logrus.Fields{
			"module":    "transactionHandler",
			"operation": "GetTransactionByID",
			"id":        id,
		}).Error("Transaction not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	response := gin.H{
		"id":               transaction.ID,
		"user_id":          transaction.UserID,
		"amount":           transaction.Amount,
		"currency":         transaction.Currency,
		"transaction_type": transaction.TransactionType,
		"category":         transaction.Category,
		"date":             transaction.Date,
		"description":      transaction.Description,
	}

	if targetCurrency != "" && targetCurrency != transaction.Currency {
		log.WithFields(logrus.Fields{
			"module":       "transactionHandler",
			"operation":    "ConvertAmount",
			"fromCurrency": transaction.Currency,
			"toCurrency":   targetCurrency,
			"amount":       transaction.Amount,
		}).Debug("Attempting to convert currency")

		convertedAmount, err := service.ConvertAmount(transaction.Amount, transaction.Currency, targetCurrency)
		if err != nil {
			log.WithFields(logrus.Fields{
				"module":    "transactionHandler",
				"operation": "ConvertAmount",
				"error":     err,
			}).Error("Failed to convert amount")
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to convert amount: %s", err.Error())})
			return
		}
		response["converted_amount"] = convertedAmount
		response["converted_currency"] = targetCurrency
	}

	log.WithFields(logrus.Fields{
		"module":    "transactionHandler",
		"operation": "GetTransactionByID",
		"id":        id,
	}).Info("Successfully retrieved and processed transaction")

	c.JSON(http.StatusOK, response)
}

func DeleteTransaction(c *gin.Context) {
	idStr := c.Param("id")

	log.WithFields(logrus.Fields{
		"module":    "transactionHandler",
		"operation": "DeleteTransaction",
		"id":        idStr,
	}).Info("Request to delete transaction")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.WithFields(logrus.Fields{
			"module":    "transactionHandler",
			"operation": "DeleteTransaction",
			"id":        idStr,
			"error":     err.Error(),
		}).Error("Invalid transaction ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	dbInterface, ok := c.Get("db")
	if !ok {
		log.WithFields(logrus.Fields{
			"module":    "transactionHandler",
			"operation": "DeleteTransaction",
		}).Error("Database connection not available")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not available"})
		return
	}

	db, ok := dbInterface.(*sql.DB)
	if !ok {
		log.WithFields(logrus.Fields{
			"module":    "transactionHandler",
			"operation": "DeleteTransaction",
		}).Error("Invalid database connection type")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid database connection type"})
		return
	}

	err = repo.DeleteTransaction(id, db)
	if err != nil {
		log.WithFields(logrus.Fields{
			"module":    "transactionHandler",
			"operation": "DeleteTransaction",
			"id":        id,
			"error":     err.Error(),
		}).Error("Error deleting transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting transaction"})
		return
	}

	log.WithFields(logrus.Fields{
		"module":    "transactionHandler",
		"operation": "DeleteTransaction",
		"id":        id,
	}).Info("Transaction successfully deleted")

	c.JSON(http.StatusOK, gin.H{"message": "Transaction deleted"})
}
func UpdateTransaction(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.WithFields(logrus.Fields{
			"module":    "transactionHandler",
			"operation": "UpdateTransaction",
			"id":        idStr,
		}).Error("Invalid transaction ID format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID format"})
		return
	}

	db, ok := c.MustGet("db").(*sql.DB)
	if !ok {
		log.WithFields(logrus.Fields{
			"module":    "transactionHandler",
			"operation": "UpdateTransaction",
		}).Error("Database connection not available")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not available"})
		return
	}

	var transaction models.Transaction
	if err := c.BindJSON(&transaction); err != nil {
		log.WithFields(logrus.Fields{
			"module":    "transactionHandler",
			"operation": "UpdateTransaction",
		}).Error("Error binding JSON to transaction model")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `UPDATE transactions SET user_id = $1, amount = $2, currency = $3, transaction_type = $4, category = $5, description = $6 WHERE id = $7`
	_, err = db.Exec(query, transaction.UserID, transaction.Amount, transaction.Currency, transaction.TransactionType, transaction.Category, transaction.Description, id)
	if err != nil {
		log.WithFields(logrus.Fields{
			"module":    "transactionHandler",
			"operation": "UpdateTransaction",
			"id":        id,
			"error":     err,
		}).Error("Error updating transaction in the database")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating transaction"})
		return
	}

	log.WithFields(logrus.Fields{
		"module":    "transactionHandler",
		"operation": "UpdateTransaction",
		"id":        id,
	}).Info("Transaction updated successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Transaction updated successfully"})
}
