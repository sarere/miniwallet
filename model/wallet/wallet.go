package mw

import (
	mc "github.com/miniwallet/model/customer"
	mh "github.com/miniwallet/model/history"
)

type Wallets []Wallet

type Wallet struct {
	ID         string  `json:"id"`
	CustomerId string  `json:"owned_by"`
	EnableAt   string  `json:"enable_at"`
	DisableAt  string  `json:"disable_at"`
	Status     string  `json:"status"`
	Balance    float64 `json:"balance"`
}

var wallets Wallets

func GetWalletByCustomer(cust mc.Customer) (Wallet, int, bool) {
	for i, data := range wallets {
		if data.CustomerId == cust.ID {
			return wallets[i], i, true
		}
	}

	var wallet Wallet

	return wallet, -1, false
}

func Update(index int, data Wallet) (Wallet, bool) {

	if wallet := &wallets[index]; wallet != nil {
		(*wallet).EnableAt = data.EnableAt
		(*wallet).DisableAt = data.DisableAt
		(*wallet).Status = data.Status
		(*wallet).Balance = data.Balance

		return *wallet, true
	}

	return data, false
}

func Create(data Wallet) (Wallet, bool) {
	wallets = append(wallets, data)

	return data, true
}

func Deposit(amount float64, referenceId string, wallet Wallet, index int) (mh.History, string) {
	var history, message = mh.Create(amount, referenceId, wallet.CustomerId, wallet.Balance, wallet.ID, "deposit")

	if history.Status == "success" {
		wallet.Balance = history.Balance
		Update(index, wallet)
	}

	return history, message
}

func Withdraw(amount float64, referenceId string, wallet Wallet, index int) (mh.History, string) {
	var history, message = mh.Create(amount, referenceId, wallet.CustomerId, wallet.Balance, wallet.ID, "withdraw")

	if history.Status == "success" {
		wallet.Balance = history.Balance
		Update(index, wallet)
	}

	return history, message
}
