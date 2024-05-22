package controllers

import (
	"fmt"
	"hyperpage/initializers"
	"hyperpage/models"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nikita-vanyasin/tinkoff"
	"gorm.io/gorm"
)

func Pending(c *fiber.Ctx) error {

	var requestBody map[string]interface{}

	if err := c.BodyParser(&requestBody); err != nil {
		return err
	}

	fmt.Println(requestBody)

	var response = requestBody

	// Access the value of the PaymentId field
	paymentIDFloat := response["PaymentId"].(float64)

	// Convert the PaymentId to an integer
	paymentID := int64(paymentIDFloat)

	if response["Status"] == "CONFIRMED" { // Use == for comparison
		var payment models.Payments
		if err := initializers.DB.Where("payment_id = ?  AND status = ?", fmt.Sprintf("%d", paymentID), "NEW").First(&payment).Error; err != nil {
			// Handle error if the record is not found or other issues
			return err
		}
		// Update the status to "applied"
		payment.Status = "applied"
		if err := initializers.DB.Save(&payment).Error; err != nil {
			// Handle error if the update fails
			return err
		}

		decimalAmount := float64(payment.Amount) / 100.0

		// Update the amount field in the balance table
		updateBalanceErr := initializers.DB.Model(&models.Billing{}).
			Where("user_id = ?", payment.UserID).
			Updates(map[string]interface{}{
				"amount": gorm.Expr("amount + ?", decimalAmount),
			}).Error
		if updateBalanceErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to update balance",
			})
		}

		transaction := models.Transaction{
			UserID:      payment.UserID,
			Total:       `0`,
			Amount:      decimalAmount,
			Description: `Пополнение баланса c карты банка`,
			Module:      `Payment`,
			Type:        `profit`,
			Status:      `CLOSED_1`,
		}

		createTransactionErr := initializers.DB.Create(&transaction).Error
		if createTransactionErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Failed to create transaction",
			})
		}

	}

	// return the city names as a JSON response
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   "initRes",
	})
}

func CreateInvoice(c *fiber.Ctx) error {
	// get all city names from the database
	var terminalKey = "1692629881262DEMO"
	var terminalPassword = "38cfo5j546t0ln2h"

	client := tinkoff.NewClient(terminalKey, terminalPassword)

	orderID := strconv.FormatInt(time.Now().UnixNano(), 10)

	user := c.Locals("user")
	sum := c.Get("amount")

	type Receipt struct {
		Name     string `json:"name"`
		Price    int    `json:"price"`
		Quantity int    `json:"quantity"`
	}

	receipt := new(Receipt)
	if err := c.BodyParser(&receipt); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Failed to parse request body: %v", err),
		})
	}

	amount, err := strconv.ParseUint(sum, 10, 64)
	if err != nil {
		// Handle the error if the conversion fails
		return err
	}

	userResp := user.(models.UserResponse)

	initReq := &tinkoff.InitRequest{
		Amount:          uint64(receipt.Price) * uint64(receipt.Quantity),
		OrderID:         orderID,
		CustomerKey:     userResp.Name,
		Description:     "Пополнение баланса в личном кабинете",
		PayType:         tinkoff.PayTypeOneStep,
		RedirectDueDate: tinkoff.Time(time.Now().Add(4 * time.Hour * 24)), // ссылка истечет через 4 дня
		Receipt: &tinkoff.Receipt{
			Email: userResp.Email,
			Items: []*tinkoff.ReceiptItem{
				{
					Price:         uint64(receipt.Price),
					Quantity:      string(rune(receipt.Quantity)),
					Amount:        uint64(receipt.Price) * uint64(receipt.Quantity),
					Name:          "Баланс на сумму " + strconv.FormatUint(amount, 10),
					Tax:           tinkoff.VATNone,
					PaymentMethod: tinkoff.PaymentMethodFullPayment,
					PaymentObject: tinkoff.PaymentObjectIntellectualActivity,
				},
			},
			Taxation: tinkoff.TaxationUSNIncome,
			Payments: &tinkoff.ReceiptPayments{
				Electronic: amount,
			},
		},

		//custom fields for tinkoff
		Data: map[string]string{
			"": "",
		},
	}
	initRes, err := client.Init(initReq)
	if err != nil {
		// Handle the error here, if needed
		fmt.Println("Error:", err)
	} else {
		fmt.Println(initRes)

		payments := models.Payments{
			UserID:    userResp.ID,
			Amount:    float64(amount),
			Status:    "NEW",
			PaymentId: initRes.PaymentID, // Store as a string directly
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Create the database record
		if err := initializers.DB.Create(&payments).Error; err != nil {
			log.Println("Could not create payment:", err)
		} else {
			fmt.Println("Payment record created successfully")
		}
	}

	// return the city names as a JSON response
	return c.JSON(fiber.Map{
		"status": "success",
		"data":   initRes,
	})
}
