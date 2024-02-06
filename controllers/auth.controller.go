package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"hyperpage/initializers"
	"hyperpage/models"
	"hyperpage/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SignUpUser(c *fiber.Ctx) error {
	config, _ := initializers.LoadConfig(".")

	var payload *models.SignUpInput
	language := c.Query("language")

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	errors := models.ValidateStruct(payload)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "errors": errors})

	}

	if payload.Password != payload.PasswordConfirm {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Passwords do not match"})

	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)

	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	// Generate a unique directory name
	dirName := utils.GenerateUniqueDirName()

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Join(config.IMGStorePath, dirName), 0755); err != nil {
		// handle error
		_ = err
	}

	src := filepath.Join(config.IMGStorePath, "default.jpg")
	dst := filepath.Join(config.IMGStorePath, dirName, "default.jpg")
	if err := os.Symlink(src, dst); err != nil {
		// Handle error
		_ = err
	}

	newUser := models.User{
		Name:          payload.Name,
		DeviceIOS:     payload.DevicesIOS,
		DeviceIOSVOIP: payload.DevicesIOSVOIP,
		Email:         strings.ToLower(payload.Email),
		Storage:       dirName,
		Password:      string(hashedPassword),
		Photo:         dirName + "/default.jpg",
	}

	result := initializers.DB.Create(&newUser)

	if result.Error != nil && strings.Contains(result.Error.Error(), "duplicate key value violates unique") {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"status": "fail", "message": "User with that email already exists"})
	} else if result.Error != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "message": "Something bad happened"})
	}

	code := make([]byte, 20)

	if _, err := rand.Read(code); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to generate password reset token",
		})
	}

	verification_code := hex.EncodeToString(code)
	TokenCode := hex.EncodeToString(code)

	newUser.VerificationCode = verification_code
	newUser.TelegramToken = TokenCode
	initializers.DB.Save(newUser)

	var firstName = newUser.Name

	if strings.Contains(firstName, " ") {
		firstName = strings.Split(firstName, " ")[1]
	}

	// ? Send Email
	emailData := utils.EmailData{
		URL:       "https://" + config.ClientOrigin + "/auth/verify/" + verification_code,
		FirstName: firstName,
	}

	switch language {
	case "en":
		emailData.Subject = "Paxintrade account activation"
	case "ru":
		emailData.Subject = "Paxintrade активация аккаунта"
	case "es":
		emailData.Subject = "Paxintrade activación de cuenta"
	case "ke":
		emailData.Subject = "Paxintrade ანგარიშის გააქტიურება"
	default:
		emailData.Subject = "Paxintrade account activation"
	}

	billing := models.Billing{
		UserID: newUser.ID,
		Amount: 100,
	}

	transaction := models.Transaction{
		UserID:      newUser.ID,
		Total:       `0`,
		Amount:      100,
		Description: `Бонус за регистрацию`,
		Module:      `Registration`,
		Type:        `profit`,
		Status:      `CLOSED_1`,
	}

	// Create and save the OnlineStorage object to the database
	onlineStorage := models.OnlineStorage{
		UserID: newUser.ID,
		Year:   time.Now().Year(), // Set the current year
		Data:   []byte("[]"),      // Set the desired data
	}

	// Create and save the OnlineStorage object to the database
	// profileId := models.Profile{
	// 	UserID: newUser.ID,
	// }

	// Create and save the OnlineStorage object to the database
	// Profile := models.Profile{
	// 	UserID: newUser.ID,
	// }

	// initializers.DB.Create(&Profile)
	initializers.DB.Create(&onlineStorage)
	initializers.DB.Create(&transaction)
	initializers.DB.Create(&billing)
	// initializers.DB.Create(&profileId)

	utils.SendEmail(&newUser, &emailData, "verificationCode", language)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success", "data": fiber.Map{"user": models.FilterUserRecord(&newUser, language)}})
}

func SignUpBot(c *fiber.Ctx) error {
	config, _ := initializers.LoadConfig(".")

	var payload *models.SignUpInput
	language := c.Query("language")

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	errors := models.ValidateStruct(payload)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "errors": errors})

	}

	if payload.Password != payload.PasswordConfirm {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Passwords do not match"})

	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)

	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	// Generate a unique directory name
	dirName := utils.GenerateUniqueDirName()

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Join(config.IMGStorePath, dirName), 0755); err != nil {
		// handle error
		_ = err
	}

	src := filepath.Join(config.IMGStorePath, "default.jpg")
	dst := filepath.Join(config.IMGStorePath, dirName, "default.jpg")
	if err := os.Symlink(src, dst); err != nil {
		// Handle error
		_ = err
	}

	newUser := models.User{
		Name:          payload.Name,
		DeviceIOS:     payload.DevicesIOS,
		DeviceIOSVOIP: payload.DevicesIOSVOIP,
		Email:         strings.ToLower(payload.Email),
		Storage:       dirName,
		Password:      string(hashedPassword),
		Photo:         dirName + "/default.jpg",
		IsBot:         true,
	}

	result := initializers.DB.Create(&newUser)

	if result.Error != nil && strings.Contains(result.Error.Error(), "duplicate key value violates unique") {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"status": "fail", "message": "User with that email already exists"})
	} else if result.Error != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "message": "Something bad happened"})
	}

	code := make([]byte, 20)

	if _, err := rand.Read(code); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to generate password reset token",
		})
	}

	verification_code := hex.EncodeToString(code)
	TokenCode := hex.EncodeToString(code)

	newUser.VerificationCode = verification_code
	newUser.TelegramToken = TokenCode
	initializers.DB.Save(newUser)

	var firstName = newUser.Name

	if strings.Contains(firstName, " ") {
		firstName = strings.Split(firstName, " ")[1]
	}

	// ? Send Email
	emailData := utils.EmailData{
		URL:       "https://" + config.ClientOrigin + "/auth/verify/" + verification_code,
		FirstName: firstName,
	}

	switch language {
	case "en":
		emailData.Subject = "Paxintrade account activation"
	case "ru":
		emailData.Subject = "Paxintrade активация аккаунта"
	case "es":
		emailData.Subject = "Paxintrade activación de cuenta"
	case "ke":
		emailData.Subject = "Paxintrade ანგარიშის გააქტიურება"
	default:
		emailData.Subject = "Paxintrade account activation"
	}

	billing := models.Billing{
		UserID: newUser.ID,
		Amount: 100,
	}

	transaction := models.Transaction{
		UserID:      newUser.ID,
		Total:       `0`,
		Amount:      100,
		Description: `Бонус за регистрацию`,
		Module:      `Registration`,
		Type:        `profit`,
		Status:      `CLOSED_1`,
	}

	// Create and save the OnlineStorage object to the database
	onlineStorage := models.OnlineStorage{
		UserID: newUser.ID,
		Year:   time.Now().Year(), // Set the current year
		Data:   []byte("[]"),      // Set the desired data
	}

	// Create and save the OnlineStorage object to the database
	// profileId := models.Profile{
	// 	UserID: newUser.ID,
	// }

	// Create and save the OnlineStorage object to the database
	// Profile := models.Profile{
	// 	UserID: newUser.ID,
	// }

	// initializers.DB.Create(&Profile)
	initializers.DB.Create(&onlineStorage)
	initializers.DB.Create(&transaction)
	initializers.DB.Create(&billing)
	// initializers.DB.Create(&profileId)

	utils.SendEmail(&newUser, &emailData, "verificationCode", language)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success", "data": fiber.Map{"user": models.FilterUserRecord(&newUser, language)}})
}

func VerifyEmail(c *fiber.Ctx) error {
	code := c.Params("verificationCode")

	var updatedUser models.User
	result := initializers.DB.First(&updatedUser, "verification_code = ?", code)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid verification code or user doesn't exist"})
	}

	if updatedUser.Verified {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"status": "fail", "message": "User already verified"})
	}

	updatedUser.VerificationCode = ""
	updatedUser.Verified = true
	initializers.DB.Save(&updatedUser)

	Profile := models.Profile{
		UserID: updatedUser.ID,
	}

	initializers.DB.Create(&Profile)

	return c.JSON(fiber.Map{"status": "success", "message": "Email verified successfully"})
}

func SignInUser(c *fiber.Ctx) error {
	Authorization := c.Get("session")

	var payload *models.SignInInput

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	errors := models.ValidateStruct(payload)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "errors": errors})
	}

	message := "Invalid email or password"

	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "errors": errors})
	}

	err := initializers.DB.First(&user, "email = ?", strings.ToLower(payload.Email)).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": message})
		} else {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": err.Error()})

		}
	}

	if !user.Verified {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "errors": "Account was not verified"})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": message})
	}

	config, _ := initializers.LoadConfig(".")

	accessTokenDetails, err := utils.CreateToken(user.ID.String(), config.AccessTokenExpiresIn, config.AccessTokenPrivateKey)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	refreshTokenDetails, err := utils.CreateToken(user.ID.String(), config.RefreshTokenExpiresIn, config.RefreshTokenPrivateKey)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	user.Session = Authorization
	user.Online = true

	userID := user.ID.String()

	utils.UserActivity("userOnline", userID)

	// Save the updated user data to the database
	err = initializers.DB.Save(&user).Error
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	// Set user data in the context
	c.Locals("user", &user)

	err = utils.SendPersonalMessageToClient(Authorization, "Hello Client")
	if err != nil {
		return err
	}

	// jsonBytes, err := json.Marshal(user)
	// if err != nil {
	// 	// Handle the error if necessary
	// 	return err
	// }

	// jsonString := string(jsonBytes)

	// c.Cookie(&fiber.Cookie{
	// 	Name:     "datas",
	// 	Value:    jsonString,
	// 	Path:     "/",
	// 	MaxAge:   config.AccessTokenMaxAge * 60,
	// 	Secure:   true,
	// 	HTTPOnly: true,
	// 	Domain:   config.ClientOrigin,
	// })

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    *accessTokenDetails.Token,
		Path:     "/",
		SameSite: "Lax",
		MaxAge:   config.AccessTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: false,
		Domain:   config.ClientOrigin,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "user_id",
		Value:    userID,
		Path:     "/",
		SameSite: "Lax",
		MaxAge:   config.AccessTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: false,
		Domain:   config.ClientOrigin,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "authenticated",
		Value:    "true",
		Path:     "/",
		SameSite: "Lax",
		MaxAge:   config.AccessTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: false,
		Domain:   config.ClientOrigin,
	})

	// c.Cookie(&fiber.Cookie{
	// 	Name:     "refresh_token",
	// 	Value:    *refreshTokenDetails.Token,
	// 	Path:     "/",
	// 	MaxAge:   config.RefreshTokenMaxAge * 60,
	// 	Secure:   true,
	// 	HTTPOnly: true,
	// 	Domain:   config.ClientOrigin,
	// })

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "access_token": accessTokenDetails.Token, "refresh_token": refreshTokenDetails})
}

func CheckTokenExp(c *fiber.Ctx) error {
	message := "could not find access token"

	token := c.Query("access_token")
	if token == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": message})
	}

	config, _ := initializers.LoadConfig(".")

	tokenClaims, err := utils.ValidateToken(token, config.AccessTokenPublicKey)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	var user models.User
	err = initializers.DB.First(&user, "id = ?", tokenClaims.UserID).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": "the user belonging to this token no longer exists"})
		} else {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": err.Error()})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "access token is valid"})
}

func RefreshAccessToken(c *fiber.Ctx) error {
	message := "could not refresh access token"

	refresh_token := c.Params("refreshToken")

	if refresh_token == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": message})
	}

	config, _ := initializers.LoadConfig(".")

	tokenClaims, err := utils.ValidateToken(refresh_token, config.RefreshTokenPublicKey)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	var user models.User
	err = initializers.DB.First(&user, "id = ?", tokenClaims.UserID).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": "the user belonging to this token no longer exists"})
		} else {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": err.Error()})
		}
	}

	accessTokenDetails, err := utils.CreateToken(user.ID.String(), config.AccessTokenExpiresIn, config.AccessTokenPrivateKey)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    *accessTokenDetails.Token,
		Path:     "/",
		SameSite: "Lax",
		MaxAge:   config.AccessTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: false,
		Domain:   config.ClientOrigin,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "access_token": accessTokenDetails.Token})
}

func ForgotPassword(c *fiber.Ctx) error {
	language := c.Query("language")

	// Get the email from the request body
	type RequestBody struct {
		Email string `json:"email"`
	}
	reqBody := new(RequestBody)
	if err := c.BodyParser(reqBody); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	if reqBody.Email == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Email field cannot be empty",
		})
	}

	// TODO: Check if the email exists in the database
	user := new(models.User)
	result := initializers.DB.Where("email = ?", reqBody.Email).First(user)
	if result.Error != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid email",
		})
	}

	// TODO: Generate a password reset token and save it to the database
	resetToken := make([]byte, 20)
	if _, err := rand.Read(resetToken); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to generate password reset token",
		})
	}
	user.PasswordResetToken = hex.EncodeToString(resetToken)
	user.PasswordResetAt = time.Now().Add(time.Minute * 15)
	if err := initializers.DB.Save(&user).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to save password reset token to the database",
		})
	}

	var firstName = user.Name

	// TODO: Send an email to the user with a link to reset their password
	if strings.Contains(firstName, " ") {
		firstName = strings.Split(firstName, " ")[1]
	}

	config, err := initializers.LoadConfig(".")
	if err != nil {
		log.Fatal("Could not load config", err)
	}

	emailData := utils.EmailData{
		URL:       "https://" + config.ClientOrigin + "/auth/reset-password/" + user.PasswordResetToken,
		FirstName: firstName,
	}

	switch language {
	case "en":
		emailData.Subject = "Password reset request (available for 10 minutes)"
	case "ru":
		emailData.Subject = "Запрос на сброс пароля (доступно 10 мин)"
	case "es":
		emailData.Subject = "Solicitud de restablecimiento de contraseña (10 min disponibles)"
	case "ke":
		emailData.Subject = "პაროლის გადატვირთვის მოთხოვნა (ხელმისაწვდომია 10 წთ)"
	default:
		emailData.Subject = "Password reset request (available for 10 minutes)"
	}

	utils.SendEmail(user, &emailData, "resetPassword", language)

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Password reset email sent",
	})

}

func ResetPassword(c *fiber.Ctx) error {
	// Get the reset token from the request params
	resetToken := c.Params("resetToken")

	// Get the password from the request body
	type RequestBody struct {
		Password        string `json:"password"`
		PasswordConfirm string `json:"password_confirm"`
	}

	reqBody := new(RequestBody)

	if err := c.BodyParser(reqBody); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "Invalid request body",
		})
	}
	reqBody.PasswordConfirm = reqBody.Password

	// Check if passwords match
	if strings.TrimSpace(reqBody.Password) != strings.TrimSpace(reqBody.PasswordConfirm) {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "Passwords do not match",
		})
	}

	// Hash the new password
	hashedPassword, err := utils.HashPassword(reqBody.Password)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to hash password",
		})
	}

	// Decode the reset token
	//passwordResetToken, err := utils.Decode(resetToken)
	if reqBody.Password == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "Password cannot be empty",
		})
	}

	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "Invalid reset token",
		})
	}

	// Get the user with the specified reset token
	var user models.User
	result := initializers.DB.Where("password_reset_token = ? AND password_reset_at > ?", resetToken, time.Now()).First(&user)
	if result.Error != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "The reset token is invalid or has expired",
		})
	}

	// Update the user's password and clear the password reset token
	user.Password = hashedPassword
	user.PasswordResetToken = ""
	if err := initializers.DB.Save(&user).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update user password",
		})
	}

	var firstName = user.Name

	// TODO: Send an email to the user with a link to reset their password
	if strings.Contains(firstName, " ") {
		_ = strings.Split(firstName, " ")[1]
	}

	// Clear the user's authentication token
	c.ClearCookie("token")

	// Return success response
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Password updated successfully",
	})
}

func LogoutUser(c *fiber.Ctx) error {
	// message := "Token is invalid or session has expired"

	// refresh_token := c.Get("refresh_token")

	// if refresh_token == "" {
	// 	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": message})
	// }

	// config, _ := initializers.LoadConfig(".")
	// ctx := context.TODO()

	// tokenClaims, err := utils.ValidateToken(refresh_token, config.RefreshTokenPublicKey)
	// if err != nil {
	// 	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	// }

	// access_token_uuid := c.Locals("access_token_uuid").(string)
	// _, err = initializers.RedisClient.Del(ctx, tokenClaims.TokenUuid, access_token_uuid).Result()
	// if err != nil {
	// 	return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	// }
	// config, _ := initializers.LoadConfig(".")

	// c.Cookie(&fiber.Cookie{
	// 	Name:     "access_token",
	// 	Value:    "",
	// 	Path:     "/",
	// 	Secure:   true,
	// 	HTTPOnly: true,
	// 	Domain:   config.ClientOrigin,
	// })

	// c.ClearCookie("access_token")

	// c.Cookie(&fiber.Cookie{
	// 	Name:     "refresh_token",
	// 	Value:    "",
	// 	Path:     "/",
	// 	Secure:   true,
	// 	HTTPOnly: true,
	// 	Domain:   config.ClientOrigin,
	// })

	// c.Cookie(&fiber.Cookie{
	// 	Name:    "logged_in",
	// 	Value:   "",
	// 	Expires: expired,
	// })
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}
