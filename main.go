package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if isRunningLocally() {
		loadEnvironmentVariablesFromDotEnvFile()
	}

	runMigrations()
	runApplication()
}

func isRunningLocally() bool {
	return os.Getenv("GIN_MODE") == ""
}

func loadEnvironmentVariablesFromDotEnvFile() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func runApplication() {
	log.Println("Starting application...")

	e := gin.Default()

	allowAllOriginsForCORS(e)

	db := getPostgresDb()
	defer db.Close()

	jwtClient := newJwtClient()

	sendgrid := newSendGridClient()
	uniUri := newUniUriStringGenerator()

	c := Controller{
		Database:              db,
		TokenClient:           jwtClient,
		Emailer:               sendgrid,
		RandomStringGenerator: uniUri,
		RedirectUrlAfterSuccessfulEmailVerification: getEnv("REDIRECT_URL_AFTER_SUCCESSFUL_EMAIL_VERIFICATION"),
	}

	v1 := e.Group("/v1")
	{
		auth := v1.Group("/auth")
		{
			verify := auth.Group("/verifyemail")
			{
				verify.GET("/:code", c.VerifyEmail)
			}

			c.ConfirmationUri = getEnv("BASE_URL") + verify.BasePath()
			auth.POST("/register", c.RegisterUser)

			auth.POST("/login", c.LoginUser)
		}

		user := v1.Group("/user")
		{
			user.Use(c.SetAuthenticatedUser)
			user.POST("/tracklink", c.TrackLink)
		}

		v1.GET("/redirect/:path", c.Redirect)
	}

	e.Run()
}

func allowAllOriginsForCORS(engine *gin.Engine) {
	engine.Use(cors.Default())
}

type Controller struct {
	TokenClient                                 TokenClient
	Database                                    Database
	Emailer                                     Emailer
	ErrorRedirectUrl                            string
	ConfirmationUri                             string
	RedirectUrlAfterSuccessfulEmailVerification string
	RandomStringGenerator                       RandomStringGenerator
}

type RandomStringGenerator interface {
	Generate() string
}

type Database interface {
	UserIdExists(ctx context.Context, userId string) (bool, error)
	EmailExists(ctx context.Context, userId string) (bool, error)
	CreateUser(ctx context.Context, user CreateUserRequest) error
	GetUser(ctx context.Context, email string) (ResultForGetUserRequest, error)
	AddRedirect(ctx context.Context, request AddRedirectRequest) error
	GetRedirectRecord(ctx context.Context, path string) (RedirectRecord, error)
	IncrementNumberOfTimesLinkClicked(ctx context.Context, path string) error
	ConfirmEmailVerified(ctx context.Context, code string) error
}

type TokenClient interface {
	GenerateToken(userId string) (string, error)
	GetUserId(token string) (string, error)
}

type CreateUserRequest struct {
	Email            string
	Password         string
	VerificationCode string
}

type Emailer interface {
	SendEmail(request SendEmailRequest) error
}

type ResultForGetUserRequest struct {
	Found bool
	User  UserRecordInDatabase
}

type UserRecordInDatabase struct {
	HashedPassword string
	IsVerified     bool
	Id             string
}

type AddRedirectRequest struct {
	UserId       string
	Url          string
	Path         string
	EmailSubject string
}

type LinkClickNotificationRequest struct {
	Email string
	Url   string
}

type RedirectRecord struct {
	RedirectUrl          string
	UserEmail            string
	EmailSubject         string
	NumberOfTimesClicked int
	UserId               string
}
