package services

import (
	"fmt"
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// ReceiptService represents receipt service
type ReceiptService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a receipt service singleton instance
var (
	Receipts = &ReceiptService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetReceiptByReceiptId returns a receipt model according to receipt id
func (s *ReceiptService) GetReceiptByReceiptId(c core.Context, uid int64, receiptId int64) (*models.Receipt, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if receiptId <= 0 {
		return nil, errs.ErrReceiptIdInvalid
	}

	receipt := &models.Receipt{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(receiptId).Where("uid=? AND deleted=?", uid, false).Get(receipt)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrReceiptNotFound
	}

	return receipt, nil
}

// GetAllReceipts returns all receipts of user
func (s *ReceiptService) GetAllReceipts(c core.Context, uid int64) ([]*models.Receipt, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var receipts []*models.Receipt
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).OrderBy("transaction_time desc").Find(&receipts)

	return receipts, err
}

// GetReceiptsByReceiptIds returns receipts by receipt ids
func (s *ReceiptService) GetReceiptsByReceiptIds(c core.Context, uid int64, receiptIds []int64) ([]*models.Receipt, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if len(receiptIds) < 1 {
		return nil, nil
	}

	var receipts []*models.Receipt
	err := s.UserDataDB(uid).NewSession(c).
		Where("uid=? AND deleted=?", uid, false).
		In("receipt_id", receiptIds).
		Find(&receipts)

	return receipts, err
}

// CreateReceipt saves a new receipt to database
func (s *ReceiptService) CreateReceipt(c core.Context, receipt *models.Receipt) error {
	if receipt.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	uuids := s.GenerateUuids(uuid.UUID_TYPE_RECEIPT, 1)

	if len(uuids) < 1 {
		return errs.ErrSystemIsBusy
	}

	receipt.ReceiptId = uuids[0]
	receipt.TransactionTime = utils.GetMinTransactionTimeFromUnixTime(utils.GetUnixTimeFromTransactionTime(receipt.TransactionTime))
	receipt.CreatedUnixTime = now
	receipt.UpdatedUnixTime = now

	_, err := s.UserDataDB(receipt.Uid).NewSession(c).Insert(receipt)

	return err
}

// ModifyReceipt saves an existed receipt to database
func (s *ReceiptService) ModifyReceipt(c core.Context, receipt *models.Receipt, prevAccountId int64, prevStatus models.TransactionStatus) error {
	if receipt.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	receipt.TransactionTime = utils.GetMinTransactionTimeFromUnixTime(utils.GetUnixTimeFromTransactionTime(receipt.TransactionTime))
	receipt.UpdatedUnixTime = now

	return s.UserDataDB(receipt.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		cols := []string{"transaction_time", "timezone_utc_offset", "account_id", "place", "comment", "updated_unix_time", "status"}
		updatedRows, err := sess.ID(receipt.ReceiptId).
			Cols(cols...).
			Where("uid=? AND deleted=?", receipt.Uid, false).Update(receipt)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrReceiptNotFound
		}

		if prevAccountId != receipt.AccountId && receipt.AccountId > 0 {
			txnUpdate := &models.Transaction{
				AccountId:      receipt.AccountId,
				UpdatedUnixTime: now,
			}

			_, err = sess.Cols("account_id", "updated_unix_time").
				Where("uid=? AND deleted=? AND receipt_id=?", receipt.Uid, false, receipt.ReceiptId).
				Update(txnUpdate)

			if err != nil {
				return err
			}
		}

		// Sync status to child transactions if status changed
		if prevStatus != receipt.Status {
			err = s.syncReceiptStatusToTransactions(c, sess, receipt.Uid, receipt.ReceiptId, receipt.Status, prevStatus, now)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// syncReceiptStatusToTransactions syncs the receipt status to all child transactions and adjusts balances
func (s *ReceiptService) syncReceiptStatusToTransactions(c core.Context, sess *xorm.Session, uid int64, receiptId int64, newStatus models.TransactionStatus, oldStatus models.TransactionStatus, now int64) error {
	var transactions []*models.Transaction
	err := sess.Where("uid=? AND deleted=? AND receipt_id=?", uid, false, receiptId).Find(&transactions)
	if err != nil {
		return err
	}

	for _, t := range transactions {
		if t.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			return errs.ErrTransactionTypeInvalid
		}

		if t.Status == newStatus {
			continue
		}

		sourceAccount, destinationAccount, err := Transactions.getAccountModels(sess, t)
		if err != nil {
			return err
		}

		if sourceAccount.Hidden || (destinationAccount != nil && destinationAccount.Hidden) {
			return errs.ErrCannotModifyTransactionInHiddenAccount
		}

		// Update transaction status
		updateModel := &models.Transaction{
			Status:          newStatus,
			UpdatedUnixTime: now,
		}
		updateCols := []string{"status", "updated_unix_time"}

		_, err = sess.ID(t.TransactionId).Cols(updateCols...).Where("uid=? AND deleted=?", uid, false).Update(updateModel)
		if err != nil {
			return err
		}

		// Update related transfer transaction status
		if t.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			relatedUpdateModel := &models.Transaction{
				Status:          newStatus,
				UpdatedUnixTime: now,
			}
			_, err = sess.ID(t.RelatedId).Cols(updateCols...).Where("uid=? AND deleted=?", uid, false).Update(relatedUpdateModel)
			if err != nil {
				return err
			}
		}

		// Adjust balance based on status change
		needsApply := oldStatus == models.TRANSACTION_STATUS_UNCONFIRMED && newStatus != models.TRANSACTION_STATUS_UNCONFIRMED
		needsReverse := oldStatus != models.TRANSACTION_STATUS_UNCONFIRMED && newStatus == models.TRANSACTION_STATUS_UNCONFIRMED

		if needsApply || needsReverse {
			sourceAccount.UpdatedUnixTime = now
			var balanceMod int64

			switch t.Type {
			case models.TRANSACTION_DB_TYPE_MODIFY_BALANCE:
				balanceMod = t.RelatedAccountAmount
			case models.TRANSACTION_DB_TYPE_INCOME:
				balanceMod = t.Amount
			case models.TRANSACTION_DB_TYPE_EXPENSE:
				balanceMod = -t.Amount
			case models.TRANSACTION_DB_TYPE_TRANSFER_OUT:
				balanceMod = -t.Amount
			default:
				return errs.ErrTransactionTypeInvalid
			}

			if needsReverse {
				balanceMod = -balanceMod
			}

			var balanceExpr string
			if balanceMod >= 0 {
				balanceExpr = fmt.Sprintf("balance+(%d)", balanceMod)
			} else {
				balanceExpr = fmt.Sprintf("balance-(%d)", -balanceMod)
			}

			_, err = sess.ID(sourceAccount.AccountId).SetExpr("balance", balanceExpr).Cols("updated_unix_time").Where("uid=? AND deleted=?", sourceAccount.Uid, false).Update(sourceAccount)
			if err != nil {
				return err
			}

			// Handle destination account for transfers
			if t.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT && destinationAccount != nil {
				destinationAccount.UpdatedUnixTime = now
				var destBalanceMod int64 = t.RelatedAccountAmount

				if needsReverse {
					destBalanceMod = -destBalanceMod
				}

				var destBalanceExpr string
				if destBalanceMod >= 0 {
					destBalanceExpr = fmt.Sprintf("balance+(%d)", destBalanceMod)
				} else {
					destBalanceExpr = fmt.Sprintf("balance-(%d)", -destBalanceMod)
				}

				_, err = sess.ID(destinationAccount.AccountId).SetExpr("balance", destBalanceExpr).Cols("updated_unix_time").Where("uid=? AND deleted=?", destinationAccount.Uid, false).Update(destinationAccount)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// GetReceiptMapByReceiptIds returns a map of receipt id to receipt model
func (s *ReceiptService) GetReceiptMapByReceiptIds(c core.Context, uid int64, receiptIds []int64) (map[int64]*models.Receipt, error) {
	receipts, err := s.GetReceiptsByReceiptIds(c, uid, receiptIds)

	if err != nil {
		return nil, err
	}

	receiptMap := make(map[int64]*models.Receipt, len(receipts))

	for i := 0; i < len(receipts); i++ {
		receiptMap[receipts[i].ReceiptId] = receipts[i]
	}

	return receiptMap, nil
}

// DeleteReceipt deletes an existed receipt from database and unlinks its transactions
func (s *ReceiptService) DeleteReceipt(c core.Context, uid int64, receiptId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		receipt := &models.Receipt{}
		has, err := sess.ID(receiptId).Where("uid=? AND deleted=?", uid, false).Get(receipt)

		if err != nil {
			return err
		} else if !has {
			return errs.ErrReceiptNotFound
		}

		unlinkModel := &models.Transaction{
			ReceiptId:      int64(0),
			UpdatedUnixTime: now,
		}

		_, err = sess.Cols("receipt_id", "updated_unix_time").
			Where("uid=? AND deleted=? AND receipt_id=?", uid, false, receiptId).
			Update(unlinkModel)

		if err != nil {
			log.Errorf(c, "[receipts.DeleteReceipt] failed to unlink transactions from receipt \"id:%d\", because %s", receiptId, err.Error())
			return err
		}

		updateModel := &models.Receipt{
			Deleted:         true,
			DeletedUnixTime: now,
		}

		deletedRows, err := sess.ID(receiptId).Cols("deleted", "deleted_unix_time").
			Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrReceiptNotFound
		}

		return nil
	})
}

// AddTransactionsToReceipt adds transactions to a receipt
func (s *ReceiptService) AddTransactionsToReceipt(c core.Context, uid int64, receiptId int64, transactionIds []int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		receipt := &models.Receipt{}
		has, err := sess.ID(receiptId).Where("uid=? AND deleted=?", uid, false).Get(receipt)

		if err != nil {
			return err
		} else if !has {
			return errs.ErrReceiptNotFound
		}

		updateModel := &models.Transaction{
			ReceiptId:      receiptId,
			UpdatedUnixTime: now,
		}

		_, err = sess.Cols("receipt_id", "updated_unix_time").
			Where("uid=? AND deleted=?", uid, false).
			In("transaction_id", transactionIds).
			Update(updateModel)

		if err != nil {
			log.Errorf(c, "[receipts.AddTransactionsToReceipt] failed to add transactions to receipt \"id:%d\", because %s", receiptId, err.Error())
			return err
		}

		return s.recalculateReceiptTotal(c, sess, uid, receiptId, now)
	})
}

// RemoveTransactionFromReceipt removes a transaction from a receipt
func (s *ReceiptService) RemoveTransactionFromReceipt(c core.Context, uid int64, receiptId int64, transactionId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		updateModel := &models.Transaction{
			ReceiptId:      int64(0),
			UpdatedUnixTime: now,
		}

		updatedRows, err := sess.Cols("receipt_id", "updated_unix_time").
			Where("uid=? AND deleted=? AND transaction_id=? AND receipt_id=?", uid, false, transactionId, receiptId).
			Update(updateModel)

		if err != nil {
			log.Errorf(c, "[receipts.RemoveTransactionFromReceipt] failed to remove transaction \"id:%d\" from receipt \"id:%d\", because %s", transactionId, receiptId, err.Error())
			return err
		} else if updatedRows < 1 {
			return errs.ErrReceiptTransactionNotFound
		}

		return s.recalculateReceiptTotal(c, sess, uid, receiptId, now)
	})
}

// recalculateReceiptTotal recalculates the total amount of a receipt
func (s *ReceiptService) recalculateReceiptTotal(c core.Context, sess *xorm.Session, uid int64, receiptId int64, now int64) error {
	var transactions []*models.Transaction
	err := sess.Where("uid=? AND deleted=? AND receipt_id=?", uid, false, receiptId).Find(&transactions)

	if err != nil {
		log.Errorf(c, "[receipts.recalculateReceiptTotal] failed to get transactions for receipt \"id:%d\", because %s", receiptId, err.Error())
		return err
	}

	totalAmount := int64(0)

	for _, t := range transactions {
		switch t.Type {
		case models.TRANSACTION_DB_TYPE_INCOME:
			totalAmount += t.Amount
		case models.TRANSACTION_DB_TYPE_EXPENSE, models.TRANSACTION_DB_TYPE_TRANSFER_OUT:
			totalAmount -= t.Amount
		}
	}

	receiptUpdate := &models.Receipt{
		TotalAmount:    totalAmount,
		UpdatedUnixTime: now,
	}

	_, err = sess.ID(receiptId).Cols("total_amount", "updated_unix_time").
		Where("uid=? AND deleted=?", uid, false).Update(receiptUpdate)

	return err
}

// UpdateReceiptStatus updates the status of a receipt and all its transactions
func (s *ReceiptService) UpdateReceiptStatus(c core.Context, uid int64, receiptId int64, newStatus models.TransactionStatus) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if receiptId <= 0 {
		return errs.ErrReceiptIdInvalid
	}

	now := time.Now().Unix()

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		receipt := &models.Receipt{}
		has, err := sess.ID(receiptId).Where("uid=? AND deleted=?", uid, false).Get(receipt)

		if err != nil {
			return err
		} else if !has {
			return errs.ErrReceiptNotFound
		}

		if receipt.Status == newStatus {
			return errs.ErrNothingWillBeUpdated
		}

		oldStatus := receipt.Status

		// Update receipt status
		receiptUpdateModel := &models.Receipt{
			Status:          newStatus,
			UpdatedUnixTime: now,
		}
		updatedRows, err := sess.ID(receiptId).Cols("status", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(receiptUpdateModel)
		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrReceiptNotFound
		}

		return s.syncReceiptStatusToTransactions(c, sess, uid, receiptId, newStatus, oldStatus, now)
	})
}

// DeleteAllReceipts deletes all existed receipts from database
func (s *ReceiptService) DeleteAllReceipts(c core.Context, uid int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.Receipt{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		}

		return nil
	})
}

// RecalculateReceiptTotalById recalculates the total amount of a receipt by its id
func (s *ReceiptService) RecalculateReceiptTotalById(c core.Context, uid int64, receiptId int64) error {
	if receiptId <= 0 {
		return nil
	}

	now := time.Now().Unix()

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		return s.recalculateReceiptTotal(c, sess, uid, receiptId, now)
	})
}
