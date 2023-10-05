package utils

import (
	"errors"
	"fmt"
	"hyperpage/initializers"
	"hyperpage/models"
	"strconv"

	uuid "github.com/satori/go.uuid"

	"gorm.io/gorm"
)

func DeductAmountFromUserBalance(userID uuid.UUID, amount float64, total float64, module string, elementId uint64) error {

	// Calculate 5% of the amount
	fee := amount

	// Retrieve user's balance from database
	balance := new(models.Billing)
	if err := initializers.DB.Where("user_id = ?", userID).First(balance).Error; err != nil {
		fmt.Println(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No balance record found for the user, create a new one
			balance = &models.Billing{
				UserID: userID,
				Amount: 0,
			}
			if err := initializers.DB.Create(balance).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Check if user has sufficient balance
	if balance.Amount < amount {
		return errors.New("insufficient balance")
	}


	// Deduct amount from user's balance and save to database
	balance.Amount -= fee
	if err := initializers.DB.Save(balance).Error; err != nil {
		return err
	}

	description := `Списание за публикацию объявления`
	// Create transaction log for the deduction
	transaction := &models.Transaction{
		UserID:      userID,
		Amount:      fee,
		Status: 	 `OPENED`,
		Module: 	module,
		ElementId: elementId,
    	Total: 		 strconv.FormatFloat(total, 'f', 2, 64),
		Description: description,
		Type:        "deduction",
	}
	if err := initializers.DB.Create(transaction).Error; err != nil {
		return err
	}


	return nil
}
