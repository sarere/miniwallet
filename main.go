package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	mc "github.com/miniwallet/model/customer"
	mw "github.com/miniwallet/model/wallet"
)

// Rest
func auth(c *gin.Context) (mc.Customer, bool) {
	var customer mc.Customer
	var status = false
	auth := c.GetHeader("Authorization")

	if auth != "" {
		splitAuth := strings.Split(auth, "Token ")
		if len(splitAuth) >= 2 {
			token := splitAuth[1]
			customer, _, status = mc.GetCustomerByToken(token)
		}
	}

	return customer, status
}

func createAccount(c *gin.Context) {
	var token string
	id := c.PostForm("customer_xid")

	if id == "" {
		c.JSON(400, gin.H{
			"data": gin.H{
				"error": gin.H{
					"customer_xid": []string{"Missing data for required field."},
				},
			},
			"status": "fail",
		})

		return
	}

	data, i, ok := mc.GetCustomerById(id)

	if ok {
		data.Token = mc.TokenGenerator()
		updatedCustomer, _ := mc.Update(i, data)
		token = updatedCustomer.Token
	} else {
		var customer mc.Customer
		customer.ID = id
		customer.Token = mc.TokenGenerator()
		newCustomer, _ := mc.Create(customer)
		token = newCustomer.Token
	}

	c.JSON(201, gin.H{
		"data": gin.H{
			"token": token,
		},
		"status": "success",
	})
}

func disableWallet(c *gin.Context) {
	var err gin.H
	var status int
	isDisabled := c.PostForm("is_disabled")
	customer, ok := auth(c)

	if !ok {
		c.JSON(401, gin.H{
			"data": gin.H{
				"message": "Token invalid",
			},
			"status": "fail",
		})

		return
	}

	if isDisabled == "" {
		c.JSON(400, gin.H{
			"data": gin.H{
				"is_disabled": []string{"Missing data for required field."},
			},
			"status": "fail",
		})

		return
	}

	if isDisabled == "true" {
		wallet, index, ok := mw.GetWalletByCustomer(customer)

		if ok && wallet.Status == "enabled" {
			disabled := time.Now()
			wallet.Status = "disabled"
			wallet.DisableAt = disabled.String()

			if wallet, ok := mw.Update(index, wallet); ok {
				c.JSON(200, gin.H{
					"data": gin.H{
						"id":          wallet.ID,
						"owned_by":    wallet.CustomerId,
						"status":      wallet.Status,
						"disabled_at": wallet.DisableAt,
						"balance":     wallet.Balance,
					},

					"status": "success",
				})

				return
			} else {
				c.JSON(500, gin.H{
					"message": "Something went wrong",
					"status":  "error",
				})

				return
			}
		} else if ok && wallet.Status == "disabled" {
			status = 404
			err = gin.H{
				"error": "Wallet Already Disabled",
			}
		} else {
			status = 404
			err = gin.H{
				"error": "Customer doesn't have a wallet",
			}
		}
	} else if isDisabled == "false" {
		status = 400
		err = gin.H{
			"error": "Can't enabled wallet using this endpoint",
		}
	}

	c.JSON(status, gin.H{
		"data":   err,
		"status": "fail",
	})
}

func enableWallet(c *gin.Context) {
	var data mw.Wallet
	customer, ok := auth(c)

	if !ok {
		c.JSON(401, gin.H{
			"data": gin.H{
				"message": "Token invalid",
			},
			"status": "fail",
		})

		return
	}

	wallet, index, ok := mw.GetWalletByCustomer(customer)

	if ok && wallet.Status == "disabled" {
		wallet.Status = "enabled"

		if wallet, ok := mw.Update(index, wallet); ok {
			data = wallet
		} else {
			c.JSON(500, gin.H{
				"message": "Something went wrong",
				"status":  "error",
			})
		}
	} else if ok && wallet.Status == "enabled" {
		c.JSON(400, gin.H{
			"data": gin.H{
				"error": "Already enabled",
			},
			"status": "fail",
		})

		return
	} else {
		id := uuid.New()
		enable := time.Now()
		data.ID = id.String()
		data.CustomerId = customer.ID
		data.EnableAt = enable.String()
		data.Status = "enabled"
		data.Balance = 0
		data, _ = mw.Create(data)
	}

	c.JSON(201, gin.H{
		"data": gin.H{
			"wallet": gin.H{
				"id":         data.ID,
				"owned_by":   data.CustomerId,
				"status":     data.Status,
				"enabled_at": data.EnableAt,
				"balance":    data.Balance,
			},
		},
		"status": "success",
	})
}

func getWallet(c *gin.Context) {
	customer, ok := auth(c)

	if !ok {
		c.JSON(401, gin.H{
			"data": gin.H{
				"message": "Token invalid",
			},
			"status": "fail",
		})

		return
	}

	wallet, _, ok := mw.GetWalletByCustomer(customer)

	if ok && wallet.Status == "enabled" {
		c.JSON(200, gin.H{
			"status": "success",
			"data": gin.H{
				"wallet": gin.H{
					"id":         wallet.ID,
					"owned_by":   wallet.CustomerId,
					"status":     wallet.Status,
					"enabled_at": wallet.EnableAt,
					"balance":    wallet.Balance,
				},
			},
		})
	} else if ok && wallet.Status == "disabled" {
		c.JSON(404, gin.H{
			"status": "fail",
			"data": gin.H{
				"error": "Disabled",
			},
		})
	} else {
		c.JSON(404, gin.H{
			"status": "fail",
			"data": gin.H{
				"message": "Customer doesn't have a wallet",
			},
		})
	}
}

func depositWallet(c *gin.Context) {
	var err gin.H
	var status int
	customer, ok := auth(c)

	if !ok {
		c.JSON(401, gin.H{
			"data": gin.H{
				"message": "Token invalid",
			},
			"status": "fail",
		})

		return
	}

	amount := c.PostForm("amount")
	referenceId := c.PostForm("reference_id")

	if referenceId != "" && amount != "" {
		wallet, i, ok := mw.GetWalletByCustomer(customer)

		if ok && wallet.Status == "enabled" {
			decimal, err := strconv.ParseFloat(amount, 64)

			if err != nil {
				c.JSON(400, gin.H{
					"data": gin.H{
						"error": gin.H{
							"amount": []string{"Invalid data"},
						},
					},
					"status": "fail",
				})
				return
			}

			history, message := mw.Deposit(decimal, referenceId, wallet, i)

			if history.Status == "fail" {
				if message == "balance" {
					c.JSON(400, gin.H{
						"data": gin.H{
							"error": gin.H{
								"balance": []string{"Insufficient balance"},
							},
						},
						"status": "fail",
					})
				} else {
					c.JSON(400, gin.H{
						"data": gin.H{
							"error": gin.H{
								"reference_id": []string{"Already Exist"},
							},
						},
						"status": "fail",
					})
				}
			} else {
				c.JSON(201, gin.H{
					"data": gin.H{
						"deposit": gin.H{
							"id":           history.ID,
							"withdrawn_by": history.CustomerID,
							"status":       history.Status,
							"deposited_at": history.TransactionAt,
							"amount":       history.Amount,
							"reference_id": history.ReferenceID,
						},
					},
					"status": history.Status,
				})
			}

			return
		} else if ok && wallet.Status == "disabled" {
			status = 404
			err = gin.H{
				"error": "Disabled",
			}
		} else {
			status = 404
			err = gin.H{
				"error": "Customer doesn't have a wallet",
			}
		}
	} else {
		status = 400

		if referenceId == "" && amount == "" {
			err = gin.H{
				"error": gin.H{
					"amount":       []string{"Missing data for required field."},
					"reference_id": []string{"Missing data for required field."},
				},
			}
		} else if referenceId == "" {
			err = gin.H{
				"error": gin.H{
					"reference_id": []string{"Missing data for required field."},
				},
			}
		} else if amount == "" {
			err = gin.H{
				"error": gin.H{
					"amount": []string{"Missing data for required field."},
				},
			}
		}
	}

	c.JSON(status, gin.H{
		"data":   err,
		"status": "fail",
	})
}

func withdrawWallet(c *gin.Context) {
	var err gin.H
	var status int
	customer, ok := auth(c)

	if !ok {
		c.JSON(401, gin.H{
			"data": gin.H{
				"message": "Token invalid",
			},
			"status": "fail",
		})

		return
	}

	amount := c.PostForm("amount")
	referenceId := c.PostForm("reference_id")

	if referenceId != "" && amount != "" {
		wallet, i, ok := mw.GetWalletByCustomer(customer)

		if ok && wallet.Status == "enabled" {
			decimal, err := strconv.ParseFloat(amount, 64)

			if err != nil {
				c.JSON(400, gin.H{
					"data": gin.H{
						"error": gin.H{
							"amount": []string{"Invalid data"},
						},
					},
					"status": "fail",
				})
				return
			}

			history, message := mw.Withdraw(decimal, referenceId, wallet, i)

			if history.Status == "fail" {
				if message == "balance" {
					c.JSON(400, gin.H{
						"data": gin.H{
							"error": gin.H{
								"balance": []string{"Insufficient balance"},
							},
						},
						"status": "fail",
					})
				} else {
					c.JSON(400, gin.H{
						"data": gin.H{
							"error": gin.H{
								"reference_id": []string{"Already Exist"},
							},
						},
						"status": "fail",
					})
				}
			} else {
				c.JSON(201, gin.H{
					"data": gin.H{
						"deposit": gin.H{
							"id":           history.ID,
							"withdrawn_by": history.CustomerID,
							"status":       history.Status,
							"deposited_at": history.TransactionAt,
							"amount":       history.Amount,
							"reference_id": history.ReferenceID,
						},
					},
					"status": "success",
				})
			}

			return
		} else if ok && wallet.Status == "disabled" {
			status = 404
			err = gin.H{
				"error": "Disabled",
			}
		} else {
			status = 404
			err = gin.H{
				"error": "Customer doesn't have a wallet",
			}
		}
	} else {
		status = 400
		if referenceId == "" && amount == "" {
			err = gin.H{
				"error": gin.H{
					"amount":       []string{"Missing data for required field."},
					"reference_id": []string{"Missing data for required field."},
				},
			}
		} else if referenceId == "" {
			err = gin.H{
				"error": gin.H{
					"reference_id": []string{"Missing data for required field."},
				},
			}
		} else if amount == "" {
			err = gin.H{
				"error": gin.H{
					"amount": []string{"Missing data for required field."},
				},
			}
		}
	}

	c.JSON(status, gin.H{
		"data":   err,
		"status": "fail",
	})
}

func main() {
	router := gin.Default()
	router.POST("/api/v1/init", createAccount)
	router.POST("/api/v1/wallet", enableWallet)
	router.PATCH("/api/v1/wallet", disableWallet)
	router.GET("/api/v1/wallet", getWallet)
	router.POST("/api/v1/wallet/deposits", depositWallet)
	router.POST("/api/v1/wallet/withdrawals", withdrawWallet)

	router.Run("localhost:8080")
}
