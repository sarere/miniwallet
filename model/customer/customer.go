package mc

import (
	"crypto/rand"
	"fmt"
)

type Customers []Customer

var customers Customers

type Customer struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

func TokenGenerator() string {
	b := make([]byte, 20)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func GetCustomerByToken(token string) (Customer, int, bool) {
	for i, data := range customers {
		if data.Token == token {
			return data, i, true
		}
	}

	var customer Customer

	return customer, -1, false
}

func GetCustomerById(id string) (Customer, int, bool) {
	for i, data := range customers {
		if data.ID == id {
			return data, i, true
		}
	}

	var customer Customer

	return customer, -1, false
}

func Update(index int, data Customer) (Customer, bool) {

	if customer := &customers[index]; customer != nil {
		(*customer).ID = data.ID
		(*customer).Token = data.Token
		data = *customer
		return data, true
	}

	return data, false
}

func Create(data Customer) (Customer, bool) {
	customers = append(customers, data)

	return data, true
}

func GetAll() Customers {
	return customers
}
