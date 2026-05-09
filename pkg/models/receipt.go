package models

import (
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// Receipt represents a receipt/check that groups multiple transactions
type Receipt struct {
	ReceiptId         int64  `xorm:"PK"`
	Uid               int64  `xorm:"UNIQUE(UQE_receipt_uid_time) INDEX(IDX_receipt_uid_deleted_time) NOT NULL"`
	Deleted           bool   `xorm:"INDEX(IDX_receipt_uid_deleted_time) NOT NULL"`
	TransactionTime   int64  `xorm:"UNIQUE(UQE_receipt_uid_time) INDEX(IDX_receipt_uid_deleted_time) NOT NULL"`
	TimezoneUtcOffset int16  `xorm:"NOT NULL"`
	AccountId         int64  `xorm:"NOT NULL DEFAULT 0"`
	Place             string `xorm:"VARCHAR(255) NOT NULL DEFAULT ''"`
	Comment           string `xorm:"VARCHAR(255) NOT NULL DEFAULT ''"`
	TotalAmount       int64  `xorm:"NOT NULL DEFAULT 0"`
	CreatedIp         string `xorm:"VARCHAR(39)"`
	CreatedUnixTime   int64
	UpdatedUnixTime   int64
	DeletedUnixTime   int64
}

// ReceiptCreateRequest represents receipt creation request
type ReceiptCreateRequest struct {
	Time      int64  `json:"time" binding:"required,min=1"`
	UtcOffset int16  `json:"utcOffset" binding:"min=-720,max=840"`
	AccountId int64  `json:"accountId,string,omitempty"`
	Place     string `json:"place" binding:"max=255"`
	Comment   string `json:"comment" binding:"max=255"`
}

// ReceiptModifyRequest represents receipt modification request
type ReceiptModifyRequest struct {
	Id        int64  `json:"id,string" binding:"required,min=1"`
	Time      int64  `json:"time" binding:"required,min=1"`
	UtcOffset int16  `json:"utcOffset" binding:"min=-720,max=840"`
	AccountId int64  `json:"accountId,string"`
	Place     string `json:"place" binding:"max=255"`
	Comment   string `json:"comment" binding:"max=255"`
}

// ReceiptDeleteRequest represents receipt deletion request
type ReceiptDeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

// ReceiptGetRequest represents receipt get request
type ReceiptGetRequest struct {
	Id int64 `form:"id,string" binding:"required,min=1"`
}

// ReceiptAddTransactionsRequest represents request to add transactions to a receipt
type ReceiptAddTransactionsRequest struct {
	ReceiptId      int64    `json:"receiptId,string" binding:"required,min=1"`
	TransactionIds []string `json:"transactionIds" binding:"required"`
}

// ReceiptRemoveTransactionRequest represents request to remove a transaction from a receipt
type ReceiptRemoveTransactionRequest struct {
	ReceiptId     int64 `json:"receiptId,string" binding:"required,min=1"`
	TransactionId int64 `json:"transactionId,string" binding:"required,min=1"`
}

// ReceiptSummaryResponse represents a lightweight view-object of receipt for embedding in transaction response
type ReceiptSummaryResponse struct {
	Id          int64  `json:"id,string"`
	Time        int64  `json:"time"`
	UtcOffset   int16  `json:"utcOffset"`
	Place       string `json:"place"`
	Comment     string `json:"comment"`
	TotalAmount int64  `json:"totalAmount"`
	AccountId   int64  `json:"accountId,string,omitempty"`
}

// ReceiptInfoResponse represents a view-object of receipt
type ReceiptInfoResponse struct {
	Id           int64                        `json:"id,string"`
	Time         int64                        `json:"time"`
	UtcOffset    int16                        `json:"utcOffset"`
	AccountId    int64                        `json:"accountId,string,omitempty"`
	Place        string                       `json:"place"`
	Comment      string                       `json:"comment"`
	TotalAmount  int64                        `json:"totalAmount"`
	Transactions TransactionInfoResponseSlice `json:"transactions,omitempty"`
}

// ReceiptInfoResponseSlice represents the slice data structure of ReceiptInfoResponse
type ReceiptInfoResponseSlice []*ReceiptInfoResponse

func (s ReceiptInfoResponseSlice) Len() int      { return len(s) }
func (s ReceiptInfoResponseSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ReceiptInfoResponseSlice) Less(i, j int) bool {
	if s[i].Time != s[j].Time {
		return s[i].Time > s[j].Time
	}
	return s[i].Id > s[j].Id
}

// ToReceiptInfoResponse returns a view-object according to database model
func (r *Receipt) ToReceiptInfoResponse() *ReceiptInfoResponse {
	return &ReceiptInfoResponse{
		Id:          r.ReceiptId,
		Time:        utils.GetUnixTimeFromTransactionTime(r.TransactionTime),
		UtcOffset:   r.TimezoneUtcOffset,
		AccountId:   r.AccountId,
		Place:       r.Place,
		Comment:     r.Comment,
		TotalAmount: r.TotalAmount,
	}
}

// ToReceiptSummaryResponse returns a lightweight view-object for embedding in transaction response
func (r *Receipt) ToReceiptSummaryResponse() *ReceiptSummaryResponse {
	return &ReceiptSummaryResponse{
		Id:          r.ReceiptId,
		Time:        utils.GetUnixTimeFromTransactionTime(r.TransactionTime),
		UtcOffset:   r.TimezoneUtcOffset,
		Place:       r.Place,
		Comment:     r.Comment,
		TotalAmount: r.TotalAmount,
		AccountId:   r.AccountId,
	}
}
