package database

import (
	"certification/cache"
	"certification/logger"
	model_account "certification/model/account"
	model_company "certification/model/company"
	model_token "certification/model/token"
	model_user "certification/model/user"
	"context"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"

	"google.golang.org/api/option"

	s3config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Initializer struct {
	DB  *gorm.DB
	RDB *cache.Cache
	FB  *firestore.Client
	S3  *s3.Client
}

// ----------------- Postgres -----------------

func (initializer *Initializer) ConnectDB(url string) error {
	gdb, err := gorm.Open(postgres.Open(url), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return err
	}
	initializer.DB = gdb
	return nil
}

func (initializer *Initializer) CloseDB() {
	sqlDB, err := initializer.DB.DB()
	if err != nil {
		logger.Log.Error(err)
	}
	sqlDB.Close()
}

func (initializer *Initializer) MigrateDB() {
	err := initializer.DB.AutoMigrate(
		model_account.Account{},
		model_company.Company{},
		model_user.User{},
		model_token.Token{},
	)
	if err != nil {
		logger.Log.Error(err)
	} else {
		logger.Log.Info("Database migrated")
	}
}

// ----------------- Redis -----------------
func (initializer *Initializer) ConnectRedis(host, port, password string, db int) error {
	err := cache.Redis.ConnectRedis(host, port, password, db)
	// err := initializer.RDB.ConnectRedis(host, port, password, db)
	if err != nil {
		return err
	}
	return nil
}

// ----------------- Firebase -----------------

func (initializer *Initializer) ConnectFirebase(credential string) error {
	opt := option.WithCredentialsFile(credential)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return err
	}

	client, err := app.Firestore(context.Background())
	if err != nil {
		return err
	}
	initializer.FB = client
	return nil
}

func (initializer *Initializer) CloseFirebase() {
	err := initializer.FB.Close()
	if err != nil {
		logger.Log.Error(err)
	}
}

// ----------------- S3 -----------------

func (initializer *Initializer) ConnectS3(accessKeyID, secretAccessKey, region string) error {
	customProvider := credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")

	cfg, err := s3config.LoadDefaultConfig(context.TODO(),
		s3config.WithCredentialsProvider(customProvider),
		s3config.WithRegion(region),
	)

	if err != nil {
		return err
	}

	initializer.S3 = s3.NewFromConfig(cfg)
	return nil
}
