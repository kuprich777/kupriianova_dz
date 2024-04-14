package repo

import (
	//configs "DZ_ITOG/config"
	"DZ_ITOG/models"
	"fmt"

	//"github.com/stretchr/testify/assert"

	//"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO items").
		WithArgs("1", "value").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := Create(models.Item{ID: "1", Value: "value"}, db); err != nil {
		t.Errorf("error was not expected while creating item: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetTransactionByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	testDate, _ := time.Parse(time.RFC3339, "2024-04-14T00:00:00Z")

	rows := sqlmock.NewRows([]string{"transaction_id", "user_id", "amount", "currency", "transaction_type", "category", "date", "description"}).
		AddRow(1, 4, 500.00, "USD", "перевод", "перевод", testDate, "Оплата услуг")

	mock.ExpectQuery("SELECT .* FROM transactions WHERE transaction_id =").
		WithArgs(1).
		WillReturnRows(rows)

	transaction, err := GetTransactionByID(1, db)
	if err != nil {
		t.Errorf("error was not expected while fetching transaction: %s", err)
	}
	if transaction == nil {
		t.Fatal("transaction was expected, got nil")
	}
	if transaction.ID != 1 {
		t.Errorf("expected transaction ID 1, got %d", transaction.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDeleteTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	transactionID1 := int64(1)
	transactionID2 := int64(2)

	mock.ExpectExec("DELETE FROM transactions WHERE transaction_id =").
		WithArgs(transactionID1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := DeleteTransaction(transactionID1, db); err != nil {
		t.Errorf("error was not expected while deleting transaction: %s", err)
	}

	mock.ExpectExec("DELETE FROM transactions WHERE transaction_id =").
		WithArgs(transactionID2).
		WillReturnError(fmt.Errorf("delete error"))

	if err := DeleteTransaction(transactionID2, db); err == nil {
		t.Errorf("expected error, got none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateCommission(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mockCommission := models.Commission{
		TransactionID:   1,
		Amount:          100.0,
		Currency:        "USD",
		TransactionType: "commission",
		Commission:      10.0,
		Date:            "2024-04-14",
		Description:     "Commission for transaction",
	}

	mock.ExpectExec("INSERT INTO commissions").
		WithArgs(mockCommission.TransactionID, mockCommission.Amount, mockCommission.Currency, mockCommission.TransactionType, mockCommission.Commission, mockCommission.Date, mockCommission.Description).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := CreateCommission(db, mockCommission); err != nil {
		t.Errorf("error was not expected while creating commission: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestUpdateTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mockTransaction := models.Transaction{
		ID:              1,
		UserID:          1,
		Amount:          500.0,
		Currency:        "USD",
		TransactionType: "transfer",
		Category:        "transfer",
		Date:            time.Now(),
		Description:     "Updated transaction",
	}

	mock.ExpectExec("UPDATE transactions").
		WithArgs(mockTransaction.UserID, mockTransaction.Amount, mockTransaction.Currency, mockTransaction.TransactionType, mockTransaction.Category, mockTransaction.Description, int64(mockTransaction.ID)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := UpdateTransaction(int64(mockTransaction.ID), mockTransaction, db); err != nil {
		t.Errorf("error was not expected while updating transaction: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateCommissionSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	newCommission := models.Commission{
		TransactionID:   1,
		Amount:          100.00,
		Currency:        "USD",
		TransactionType: "transfer",
		Commission:      10.00,
		Date:            "2024-04-14",
		Description:     "Commission for transfer",
	}

	mock.ExpectExec("INSERT INTO commissions").
		WithArgs(newCommission.TransactionID, newCommission.Amount, newCommission.Currency, newCommission.TransactionType, newCommission.Commission, newCommission.Date, newCommission.Description).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = CreateCommission(db, newCommission)

	if err != nil {
		t.Errorf("unexpected error while creating commission: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateCommissionWithInvalidData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	invalidCommission := models.Commission{
		TransactionID:   1,
		Amount:          100.00,
		Currency:        "USD",
		TransactionType: "transfer",
		Commission:      1000.00,
		Date:            "2024-04-14",
		Description:     "Commission for transfer",
	}

	mock.ExpectExec("INSERT INTO commissions").
		WithArgs(invalidCommission.TransactionID, invalidCommission.Amount, invalidCommission.Currency, invalidCommission.TransactionType, invalidCommission.Commission, invalidCommission.Date, invalidCommission.Description).
		WillReturnError(fmt.Errorf("commission exceeds maximum allowed amount"))

	err = CreateCommission(db, invalidCommission)

	if err == nil {
		t.Error("expected error, got none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateCommissionWithSQLConstraintError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	duplicateCommission := models.Commission{
		TransactionID:   1,
		Amount:          100.00,
		Currency:        "USD",
		TransactionType: "transfer",
		Commission:      10.00,
		Date:            "2024-04-14",
		Description:     "Commission for transfer",
	}

	mock.ExpectExec("INSERT INTO commissions").
		WithArgs(duplicateCommission.TransactionID, duplicateCommission.Amount, duplicateCommission.Currency, duplicateCommission.TransactionType, duplicateCommission.Commission, duplicateCommission.Date, duplicateCommission.Description).
		WillReturnError(fmt.Errorf("duplicate key violation"))

	err = CreateCommission(db, duplicateCommission)

	if err == nil {
		t.Error("expected error, got none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestGetAllTransactions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"transaction_id", "user_id", "amount", "currency", "transaction_type", "category", "date", "description"}).
		AddRow(1, 1, 100.00, "USD", "expense", "food", time.Now(), "Dinner out").
		AddRow(2, 1, 200.00, "USD", "income", "salary", time.Now(), "Monthly salary")

	mock.ExpectQuery(`SELECT transaction_id, user_id, amount, currency, transaction_type, category, date, description FROM transactions`).
		WillReturnRows(rows)

	transactions, err := GetAllTransactions(db)
	if err != nil {
		t.Errorf("error was not expected while fetching transactions: %s", err)
	}

	if len(transactions) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(transactions))
	}

	if transactions[0].Amount != 100.00 {
		t.Errorf("expected amount of first transaction to be 100.00, got %f", transactions[0].Amount)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
func TestCreateTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	expectedID := 1

	mock.ExpectQuery("INSERT INTO transactions").
		WithArgs(1, 500.0, "USD", "transfer", "transfer", "Test transaction").
		WillReturnRows(sqlmock.NewRows([]string{"transaction_id"}).AddRow(expectedID))

	testTransaction := models.Transaction{
		UserID:          1,
		Amount:          500.0,
		Currency:        "USD",
		TransactionType: "transfer",
		Category:        "transfer",
		Description:     "Test transaction",
	}

	id, err := CreateTransaction(testTransaction, db)
	if err != nil {
		t.Errorf("error was not expected while inserting transaction: %s", err)
		return
	}

	if id != expectedID {
		t.Errorf("returned transaction ID does not match expected ID. Got %d, expected %d", id, expectedID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
