package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthCreds struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (r *Controller) RegisterUser(c *gin.Context) {
	registerRequest := AuthCreds{}

	if err := c.ShouldBindJSON(&registerRequest); err != nil {
		log.Printf("invalid register request: %s", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if valid := emailAndPasswordAreValid(registerRequest.Email, registerRequest.Password); !valid {
		log.Printf("invalid email %s or password %s", registerRequest.Email, registerRequest.Password)
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid email or password")
		return
	}

	exists, err := r.Database.EmailExists(c, registerRequest.Email)
	if err != nil {
		log.Printf("error checking email %s exists: %s", registerRequest.Email, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if exists {
		log.Println("email exists:", registerRequest.Email)
		c.AbortWithStatusJSON(http.StatusBadRequest, "Email exists")
		return
	}

	hashedPassword, err := generateHashedPassword(registerRequest.Password)
	if err != nil {
		log.Println("error hashing password:", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	code := uuid.New().String()
	cur := CreateUserRequest{Email: registerRequest.Email, Password: hashedPassword, VerificationCode: code}

	err = r.Database.CreateUser(c, cur)
	if err != nil {
		log.Printf("error creating user with email %s: %s", registerRequest.Email, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	subject := "Confirm your registration with LinkUp"
	content_template := "Click this link to confirm your registration with LinkUp: %s/%s"
	content := fmt.Sprintf(content_template, r.ConfirmationUri, code)

	ser := SendEmailRequest{Email: registerRequest.Email, Subject: subject, Content: content}
	err = r.Emailer.SendEmail(ser)
	if err != nil {
		log.Printf("error sending email confirmation to %s: %s", registerRequest.Email, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusCreated)
}

func emailAndPasswordAreValid(email string, password string) bool {
	if email == "" || password == "" {
		return false
	}

	return true
}

func generateHashedPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func (r *Controller) LoginUser(c *gin.Context) {
	loginRequest := AuthCreds{}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		log.Println("invalid request: ", loginRequest)
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid request")
		return
	}

	result, err := r.Database.GetUser(c, loginRequest.Email)
	if err != nil {
		log.Printf("error fetching user for email %s: %s", loginRequest.Email, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if !result.Found {
		log.Printf("email %s doesn't exist", loginRequest.Email)
		c.AbortWithStatusJSON(http.StatusBadRequest, "Email doesn't exist")
		return
	}

	if !result.User.IsVerified {
		log.Printf("email %s has not been verified", loginRequest.Email)
		c.AbortWithStatusJSON(http.StatusBadRequest, "Email has not been verified")
		return
	}

	err = checkPasswordIsCorrect(loginRequest.Password, result.User.HashedPassword)
	if err != nil {
		log.Printf("incorrect password %s for email %s: ", loginRequest.Password, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, "Incorrect password")
		return
	}

	token, err := r.TokenClient.GenerateToken(result.User.Id)
	if err != nil {
		log.Printf("error generating jwt token for email %s: %s", loginRequest.Email, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.String(http.StatusOK, token)
}

func checkPasswordIsCorrect(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (r *Controller) VerifyEmail(c *gin.Context) {
	code := c.Param("code")

	if lengthOfString(code) == 0 {
		log.Printf("no verification code given")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err := r.Database.ConfirmEmailVerified(c, code)
	if err != nil {
		log.Printf("error confirming email verified for code %s: %s", code, err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, r.EmailVerifiedUrl)
}
