package config

import (
	"os"
	"strconv"

	"github.com/robfig/cron/v3"
)

var ENVIRONMENT string
var DATABASE_URL string
var FIREBASE_CREDENTIALS string
var HOST string
var SECRET string
var REDIS_HOST string
var REDIS_PORT string
var REDIS_PASSWORD string
var REDIS_DB_NUMBER int
var SMTP_FROM string
var SMTP_PASSWORD string
var SMTP_HOST string
var SMTP_PORT int
var SCHEDULER = cron.New()
var API_URL string
var EMAIL_LOGO_URL string
var LOG_PATH string

// var MONGO_URL string
// var MONGO_DB string
// var MONGO_COLLECTION string

func LoadEnv(path string) {
	// err := godotenv.Load(path + ".env")
	// if err != nil {
	// 	logger.Log.Error(err)
	// }

	DATABASE_URL = os.Getenv("DATABASE_URL")
	FIREBASE_CREDENTIALS = os.Getenv("FIREBASE_CREDENTIALS")
	HOST = os.Getenv("HOST")

	// Auth Configuration
	SECRET = os.Getenv("SECRET")

	// Redis Configuration
	REDIS_HOST = os.Getenv("REDIS_HOST")
	REDIS_PORT = os.Getenv("REDIS_PORT")
	REDIS_PASSWORD = os.Getenv("REDIS_PASSWORD")
	REDIS_DB_NUMBER, _ = strconv.Atoi(os.Getenv("REDIS_DB_NUMBER"))

	// SMTP Configuration
	SMTP_FROM = os.Getenv("SMTP_FROM")
	SMTP_PASSWORD = os.Getenv("SMTP_PASSWORD")
	SMTP_HOST = os.Getenv("SMTP_HOST")
	SMTP_PORT, _ = strconv.Atoi(os.Getenv("SMTP_PORT"))

	// MongoDB Configuration
	// MONGO_URL = os.Getenv("MONGO_URL")
	// MONGO_DB = os.Getenv("MONGO_DB")
	// MONGO_COLLECTION = os.Getenv("MONGO_COLLECTION")

	EMAIL_LOGO_URL = os.Getenv("EMAIL_LOGO_URL")
	API_URL = os.Getenv("API_URL")
	LOG_PATH = os.Getenv("LOG_PATH")

}
