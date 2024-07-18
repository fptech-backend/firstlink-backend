package main

import (
	"certification/config"
	"certification/database"
	_ "certification/docs"
	"certification/logger"
	"certification/router"
	"certification/socket"
	"flag"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var (
	version string
	date    string
	hash    string
)

// @title FirstLink API
// @version 2.0
// @description This is a swagger for FirstLink
// @termsOfService http://swagger.io/terms/
// @contact.name FPG Tech
// @contact.email support@firstpaviliontech.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host 127.0.0.1:8080
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @BasePath /
func main() {
	app := fiber.New(fiber.Config{
		BodyLimit: 500 * 1024 * 1024, // 500 MB
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	Initialize(app)
	GetBuildVersion()

	err := app.Listen(config.HOST)
	if err != nil {
		logger.Log.Fatal(err)
	}
}

func Initialize(app *fiber.App) {
	// var initializer database.Initializer

	config.LoadEnv("")
	logger.Initialize("", config.LOG_PATH)
	defer logger.Log.Sync()

	var initializer = database.Initializer{}

	initializer.ConnectDB(config.DATABASE_URL)
	initializer.MigrateDB()
	initializer.ConnectRedis(config.REDIS_HOST, config.REDIS_PORT, config.REDIS_PASSWORD, config.REDIS_DB_NUMBER)

	// initializer.ConnectFirebase(config.FIREBASE_CREDENTIALS)
	// initializer.ConnectS3(config.AWS_ACCESS_KEY_ID, config.AWS_SECRET_ACCESS_KEY, config.AWS_REGION)
	// // database.ConnectMongoDB(&initializer)

	router.SetupRoutes(app, &initializer)
	socket.InitializeWebSocket(app, &initializer)
}

func GetBuildVersion() {
	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "Show version information")
	flag.Parse()

	if showVersion {
		build := fmt.Sprintf("Version: %s | Build Time: %s | Git Hash: %s",
			version, date, hash)
		logger.Log.Info(build)
	}
}
