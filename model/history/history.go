package mh

import (
	"time"

	"github.com/google/uuid"
)

type Histories []History

type History struct {
	ID            string  `json:"id"`
	WalletID      string  `json:"wallet_id"`
	CustomerID    string  `json:"customer_id"`
	TransactionAt string  `json:"transaction_at"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"`
	Balance       float64 `json:"balance"`
	ReferenceID   string  `json:"reference_id"`
}

var histories Histories

func IsReferenceAvailable(referenceId string) string {
	for _, data := range histories {
		if referenceId == data.ReferenceID {
			return "fail"
		}
	}

	return "success"
}

func Create(amount float64, referenceId string, customerId string, balance float64, walletId string, transaction string) (History, string) {
	id := uuid.New()
	time := time.Now()

	var message string
	var history History
	var total float64
	var status = IsReferenceAvailable(referenceId)

	if status == "fail" {
		message = "reference_id"
	} else {
		if transaction == "deposit" {
			total = balance + amount
		} else if transaction == "withdraw" && balance-amount >= 0 {
			total = balance - amount
		} else {
			message = "balance"
			status = "fail"
		}
	}

	history.Amount = amount
	history.CustomerID = customerId
	history.Balance = total
	history.TransactionAt = time.String()
	history.ID = id.String()
	history.Status = status
	history.WalletID = walletId
	history.ReferenceID = referenceId

	histories = append(histories, history)

	return history, message
}
