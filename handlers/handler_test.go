package handlers

import (
	"DZ_ITOG/models"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetAllTransactions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{
		"transaction_id", "user_id", "amount", "currency", "transaction_type", "category", "date", "description",
	}).AddRow(
		1, 10, 100.0, "USD", "перевод", "category1", time.Now(), "description1",
	).AddRow(
		2, 11, 200.0, "EUR", "покупка", "category2", time.Now(), "description2",
	)

	mock.ExpectQuery("^SELECT (.+) FROM transactions$").WillReturnRows(rows)

	r := gin.New()
	r.GET("/transactions", func(c *gin.Context) {
		c.Set("db", db)
		GetAllTransactions(c)
	})

	req, _ := http.NewRequest(http.MethodGet, "/transactions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "USD")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestCreateTransaction(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	r := gin.Default()
	r.POST("/transaction", func(c *gin.Context) {
		c.Set("db", db)
		CreateTransaction(c)
	})

	transaction := models.Transaction{
		UserID:          1,
		Amount:          100.00,
		Currency:        "USD",
		TransactionType: "перевод",
		Category:        "test",
		Description:     "test transaction",
	}

	jsonValue, _ := json.Marshal(transaction)
	req, _ := http.NewRequest(http.MethodPost, "/transaction", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("db", db)

	mock.ExpectQuery(`^INSERT INTO transactions`).WithArgs(transaction.UserID, transaction.Amount, transaction.Currency, transaction.TransactionType, transaction.Category, transaction.Description).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	commissionDescription := fmt.Sprintf("Комиссия 2.00%% от суммы %.2f", transaction.Amount*0.02)

	commission := models.Commission{
		TransactionID:   1,
		Amount:          transaction.Amount,
		TransactionType: transaction.TransactionType,
		Currency:        transaction.Currency,
		Commission:      math.Round(transaction.Amount*0.02*100) / 100,
		Date:            time.Now().Format("2006-01-02"),
		Description:     commissionDescription,
	}

	mock.ExpectQuery(`^INSERT INTO commissions`).WithArgs(
		commission.TransactionID,
		commission.Amount,
		commission.Currency,
		commission.TransactionType,
		commission.Commission,
		commission.Date,
		commission.Description,
	).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	responseBody := w.Body.String()
	assert.Contains(t, responseBody, `"commission":`)
	assert.Contains(t, responseBody, `"transaction":`)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestGetTransactionByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.Default()
	r.GET("/transaction/:id", GetTransactionByID)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	tests := []struct {
		description        string
		ID                 string
		setupMock          func()
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			description: "Correct ID",
			ID:          "1",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id", "amount", "currency", "transaction_type", "category", "date", "description"}).
					AddRow(1, 10, 100.50, "USD", "transfer", "business", "2024-01-01", "Test transaction")
				mock.ExpectQuery(`SELECT \* FROM transactions WHERE transaction_id = \?`).WithArgs(1).WillReturnRows(rows)
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   "{\"id\":1,\"user_id\":10,\"amount\":100.50,\"currency\":\"USD\",\"transaction_type\":\"transfer\",\"category\":\"business\",\"date\":\"2024-01-01\",\"description\":\"Test transaction\"}",
		},
		{
			description:        "Invalid ID format",
			ID:                 "abc",
			setupMock:          func() {},
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "{\"error\":\"Invalid transaction ID format\"}",
		},
		{
			description: "Transaction not found",
			ID:          "999",
			setupMock: func() {
				mock.ExpectQuery(`SELECT \* FROM transactions WHERE transaction_id = \?`).WithArgs(999).WillReturnError(sql.ErrNoRows)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   "{\"error\":\"Transaction not found\"}",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			test.setupMock()

			req, _ := http.NewRequest(http.MethodGet, "/transaction/"+test.ID, nil)
			w := httptest.NewRecorder()

			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			ctx.Set("db", db)

			r.ServeHTTP(w, req)

			assert.Equal(t, test.expectedStatusCode, w.Code)
			assert.JSONEq(t, test.expectedResponse, w.Body.String())

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
