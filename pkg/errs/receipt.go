package errs

import "net/http"

// Error codes related to receipt
var (
	ErrReceiptIdInvalid           = NewNormalError(NormalSubcategoryTransaction, 44, http.StatusBadRequest, "receipt id is invalid")
	ErrReceiptNotFound            = NewNormalError(NormalSubcategoryTransaction, 45, http.StatusBadRequest, "receipt not found")
	ErrReceiptNoTransactions      = NewNormalError(NormalSubcategoryTransaction, 46, http.StatusBadRequest, "receipt has no transactions")
	ErrReceiptTransactionNotFound = NewNormalError(NormalSubcategoryTransaction, 47, http.StatusBadRequest, "transaction not found in receipt")
)
