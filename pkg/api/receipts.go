package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// ReceiptsApi represents receipt api
type ReceiptsApi struct {
	*TransactionsApi
	receipts *services.ReceiptService
}

// Initialize a receipt api singleton instance
var (
	ReceiptsAPI = &ReceiptsApi{
		TransactionsApi: Transactions,
		receipts:        services.Receipts,
	}
)

// ReceiptListHandler returns all receipts of current user
func (a *ReceiptsApi) ReceiptListHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()

	receipts, err := a.receipts.GetAllReceipts(c, uid)

	if err != nil {
		log.Errorf(c, "[receipts.ReceiptListHandler] failed to get all receipts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	result := make(models.ReceiptInfoResponseSlice, len(receipts))

	for i := 0; i < len(receipts); i++ {
		result[i] = receipts[i].ToReceiptInfoResponse()
	}

	return result, nil
}

// ReceiptGetHandler returns one specific receipt with its transactions
func (a *ReceiptsApi) ReceiptGetHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.ReceiptGetRequest
	err := c.ShouldBindQuery(&req)

	if err != nil {
		log.Warnf(c, "[receipts.ReceiptGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[receipts.ReceiptGetHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[receipts.ReceiptGetHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	receipt, err := a.receipts.GetReceiptByReceiptId(c, uid, req.Id)

	if err != nil {
		log.Errorf(c, "[receipts.ReceiptGetHandler] failed to get receipt \"id:%d\" for user \"uid:%d\", because %s", req.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	resp := receipt.ToReceiptInfoResponse()

	transactions, err := a.transactions.GetTransactionsByReceiptId(c, uid, req.Id)

	if err != nil {
		log.Errorf(c, "[receipts.ReceiptGetHandler] failed to get transactions for receipt \"id:%d\" for user \"uid:%d\", because %s", req.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if len(transactions) > 0 {
		accountMap, categoryMap, tagMap, allTransactionTagIds, pictureInfoMap, err := a.getTransactionEssentialDataByTransactionIds(c, user, transactions, false, false, false)

		if err != nil {
			log.Errorf(c, "[receipts.ReceiptGetHandler] failed to get essential data for receipt \"id:%d\" for user \"uid:%d\", because %s", req.Id, uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		transactionResult, err := a.getTransactionResponseListResult(c, user, transactions, accountMap, categoryMap, tagMap, allTransactionTagIds, pictureInfoMap, clientTimezone, false, false, false, false)

		if err != nil {
			log.Errorf(c, "[receipts.ReceiptGetHandler] failed to assemble transaction result for receipt \"id:%d\" for user \"uid:%d\", because %s", req.Id, uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		resp.Transactions = transactionResult
	}

	return resp, nil
}

// ReceiptCreateHandler creates a new receipt
func (a *ReceiptsApi) ReceiptCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.ReceiptCreateRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[receipts.ReceiptCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	receipt := &models.Receipt{
		Uid:               uid,
		TransactionTime:   utils.GetMinTransactionTimeFromUnixTime(req.Time),
		TimezoneUtcOffset: req.UtcOffset,
		AccountId:         req.AccountId,
		Place:             req.Place,
		Comment:           req.Comment,
		CreatedIp:         c.ClientIP(),
		TotalAmount:       0,
	}

	err = a.receipts.CreateReceipt(c, receipt)

	if err != nil {
		log.Errorf(c, "[receipts.ReceiptCreateHandler] failed to create receipt for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[receipts.ReceiptCreateHandler] user \"uid:%d\" has created a new receipt \"id:%d\" successfully", uid, receipt.ReceiptId)

	return receipt.ToReceiptInfoResponse(), nil
}

// ReceiptModifyHandler modifies an existing receipt
func (a *ReceiptsApi) ReceiptModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.ReceiptModifyRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[receipts.ReceiptModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	receipt, err := a.receipts.GetReceiptByReceiptId(c, uid, req.Id)

	if err != nil {
		log.Errorf(c, "[receipts.ReceiptModifyHandler] failed to get receipt \"id:%d\" for user \"uid:%d\", because %s", req.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	prevAccountId := receipt.AccountId
	receipt.TransactionTime = utils.GetMinTransactionTimeFromUnixTime(req.Time)
	receipt.TimezoneUtcOffset = req.UtcOffset
	receipt.AccountId = req.AccountId
	receipt.Place = req.Place
	receipt.Comment = req.Comment

	err = a.receipts.ModifyReceipt(c, receipt, prevAccountId)

	if err != nil {
		log.Errorf(c, "[receipts.ReceiptModifyHandler] failed to modify receipt \"id:%d\" for user \"uid:%d\", because %s", req.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[receipts.ReceiptModifyHandler] user \"uid:%d\" has updated receipt \"id:%d\" successfully", uid, receipt.ReceiptId)

	return receipt.ToReceiptInfoResponse(), nil
}

// ReceiptDeleteHandler deletes a receipt and unlinks its transactions
func (a *ReceiptsApi) ReceiptDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.ReceiptDeleteRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[receipts.ReceiptDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	err = a.receipts.DeleteReceipt(c, uid, req.Id)

	if err != nil {
		log.Errorf(c, "[receipts.ReceiptDeleteHandler] failed to delete receipt \"id:%d\" for user \"uid:%d\", because %s", req.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[receipts.ReceiptDeleteHandler] user \"uid:%d\" has deleted receipt \"id:%d\"", uid, req.Id)

	return true, nil
}

// ReceiptAddTransactionsHandler adds transactions to a receipt
func (a *ReceiptsApi) ReceiptAddTransactionsHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.ReceiptAddTransactionsRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[receipts.ReceiptAddTransactionsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	transactionIds, err := utils.StringArrayToInt64Array(req.TransactionIds)

	if err != nil {
		log.Warnf(c, "[receipts.ReceiptAddTransactionsHandler] parse transaction ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionIdInvalid
	}

	err = a.receipts.AddTransactionsToReceipt(c, uid, req.ReceiptId, transactionIds)

	if err != nil {
		log.Errorf(c, "[receipts.ReceiptAddTransactionsHandler] failed to add transactions to receipt \"id:%d\" for user \"uid:%d\", because %s", req.ReceiptId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[receipts.ReceiptAddTransactionsHandler] user \"uid:%d\" has added %d transactions to receipt \"id:%d\"", uid, len(transactionIds), req.ReceiptId)

	return true, nil
}

// ReceiptRemoveTransactionHandler removes a transaction from a receipt
func (a *ReceiptsApi) ReceiptRemoveTransactionHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.ReceiptRemoveTransactionRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[receipts.ReceiptRemoveTransactionHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	err = a.receipts.RemoveTransactionFromReceipt(c, uid, req.ReceiptId, req.TransactionId)

	if err != nil {
		log.Errorf(c, "[receipts.ReceiptRemoveTransactionHandler] failed to remove transaction \"id:%d\" from receipt \"id:%d\" for user \"uid:%d\", because %s", req.TransactionId, req.ReceiptId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[receipts.ReceiptRemoveTransactionHandler] user \"uid:%d\" has removed transaction \"id:%d\" from receipt \"id:%d\"", uid, req.TransactionId, req.ReceiptId)

	return true, nil
}
