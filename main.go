package main

import (
	"context"
	"log"
	"net/http"
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

	c := Controller{
		Database:               db,
		TokenClient:            newJwtClient(),
		Emailer:                newSendGridClient(),
		RedirectPathLength:     getInt(getEnv("REDIRECT_PATH_LENGTH")),
		MaxNumberOfEmailAlerts: getInt(getEnv("MAX_NUMBER_OF_EMAIL_ALERTS")),
		EmailVerifiedUrl:       getEnv("EMAIL_VERIFIED_URL"),
		ErrorRedirectUrl:       getEnv("ERROR_REDIRECT_URL"),
		BotChecker:             newBotChecker(),
	}

	baseUrl := getEnv("BASE_URL")

	redirect := e.Group("/r")
	{
		redirect.GET("/:path", c.Redirect)
	}

	v1 := e.Group("/v1")
	{
		auth := v1.Group("/auth")
		{
			verify := auth.Group("/verifyemail")
			{
				verify.GET("/:code", c.VerifyEmail)
			}

			c.ConfirmationUri = baseUrl + verify.BasePath()
			auth.POST("/register", c.RegisterUser)

			auth.POST("/login", c.LoginUser)
		}

		c.RedirectUri = baseUrl + redirect.BasePath()

		user := v1.Group("/user")
		{
			user.Use(c.SetAuthenticatedUser)
			user.POST("/tracklink", c.TrackLink)
		}
	}

	e.Run()
}

func allowAllOriginsForCORS(e *gin.Engine) {
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Content-Length", "Accept", "Authorization"}
	e.Use(cors.New(config))
}

type Controller struct {
	TokenClient            TokenClient
	Database               Database
	Emailer                Emailer
	BotChecker             BotChecker
	ErrorRedirectUrl       string
	ConfirmationUri        string
	EmailVerifiedUrl       string
	RedirectUri            string
	RedirectPathLength     int
	MaxNumberOfEmailAlerts int
}

type Database interface {
	UserIdExists(ctx context.Context, userId string) (bool, error)
	EmailExists(ctx context.Context, userId string) (bool, error)
	CreateUser(ctx context.Context, user CreateUserRequest) error
	GetUser(ctx context.Context, email string) (ResultForGetUserRequest, error)
	AddRedirect(ctx context.Context, request AddRedirectRequest) error
	GetRedirectRecord(ctx context.Context, path string) (RedirectRecord, error)
	GetRedirectUrl(ctx context.Context, path string) (string, error)
	AddLinkClick(ctx context.Context, linkId string) error
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
	UserId string
	Url    string
	Path   string
	Tag    string
}

type LinkClickNotificationRequest struct {
	Email string
	Url   string
}

type RedirectRecord struct {
	RedirectUrl          string
	UserEmail            string
	Tag                  string
	NumberOfTimesClicked int
	LinkId               string
	Path                 string
}

type BotChecker interface {
	IsBotRequest(req *http.Request) (bool, error)
}
