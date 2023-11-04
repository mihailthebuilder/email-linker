package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (r *Controller) RegisterUser(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	valid := emailAndPasswordAreValid(email, password)

	if !valid {
		log.Printf("invalid email %s or password %s", email, password)
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid email or password")
		return
	}

	exists, err := r.Database.EmailExists(c, email)
	if err != nil {
		log.Printf("error checking email %s exists: %s", email, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if exists {
		log.Println("email exists: ", email)
		c.AbortWithStatusJSON(http.StatusBadRequest, "Email exists")
		return
	}

	hashedPassword, err := generateHashedPassword(password)
	if err != nil {
		log.Println("error hashing password: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	code := r.RandomStringGenerator.Generate()
	cur := CreateUserRequest{Email: email, Password: hashedPassword, VerificationCode: code}

	err = r.Database.CreateUser(c, cur)
	if err != nil {
		log.Printf("error creating email %s: %s", email, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	subject := "Confirm your registration with LinkUp"
	content_template := "Click this link to confirm your registration with LinkUp: %s/%s"
	content := fmt.Sprintf(content_template, r.ConfirmationUri, code)

	ser := SendEmailRequest{Email: email, Subject: subject, Content: content}
	err = r.Emailer.SendEmail(ser)
	if err != nil {
		log.Printf("error sending email confirmation to %s: %s", email, err)
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
	var loginRequest struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		log.Println("invalid request: ", loginRequest)
		c.AbortWithStatusJSON(http.StatusBadRequest, "Invalid request")
		return
	}

	result, err := r.Database.GetUser(c, loginRequest.Email)
	if err != nil {
		log.Printf("error fetching password for email %s: %s", loginRequest.Email, err)
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

	c.Redirect(http.StatusTemporaryRedirect, r.RedirectUrlAfterSuccessfulEmailVerification)
}
