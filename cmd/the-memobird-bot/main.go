package main

import (
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/tevino/log"

	"github.com/awesome-memobird/the-memobird-bot/bot"
	"github.com/awesome-memobird/the-memobird-bot/memobird"
	"github.com/awesome-memobird/the-memobird-bot/model"
	"github.com/awesome-memobird/the-memobird-bot/service"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // Database Driver
	_ "github.com/jinzhu/gorm/dialects/sqlite"   // Database Driver
)

// names of environment variables.
const (
	// the token of telegram bot.
	EnvToken = "TELEGRAM_TOKEN"
	// the access key of memobird application.
	EnvAccessKey = "MEMOBIRD_AK"
	// the port to listen as required by Heroku.
	EnvPort = "PORT"
)

func newDB() *gorm.DB {
	var driver string
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		driver = "sqlite3"
		dbURL = "test.db"
	} else {
		driver = "postgres"
	}
	log.Infof("Using %s: %s", driver, dbURL)

	db, err := gorm.Open(driver, dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database")
	}

	// Migrate the schema
	db.AutoMigrate(&model.User{})
	db.AutoMigrate(&model.Device{})
	db.AutoMigrate(&model.Content{})

	return db
}

func newBot(config *bot.Config) *bot.Bot {
	b, err := bot.New(config)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func listenHTTPIfRequired() {
	port := os.Getenv(EnvPort)

	if port != "" {
		pingHandler := func(w http.ResponseWriter, req *http.Request) {
			io.WriteString(w, "I'm alive!\n")
		}
		http.HandleFunc("/ping", pingHandler)
		go http.ListenAndServe(":"+port, nil)
	}
}

func main() {
	// Check mandantory environment variables.
	accessKey := os.Getenv(EnvAccessKey)
	if accessKey == "" {
		log.Fatalf("Please specify access key for memobird application via environment variable %s", EnvAccessKey)
	}
	token := os.Getenv(EnvToken)
	if token == "" {
		log.Fatalf("Please specify token for telegram bot via environment variable %s", EnvToken)
	}

	listenHTTPIfRequired()

	// initialization
	rand.Seed(time.Now().UnixNano())
	db := newDB()
	defer db.Close()

	birdApp := memobird.NewApp(&memobird.AppConfig{
		AccessKey: accessKey,
		Timeout:   30 * time.Second,
	})

	// services
	deviceService := &service.Device{DB: db}
	userService := &service.User{DB: db}
	birdService := &service.Bird{BirdApp: birdApp}

	b := newBot(&bot.Config{
		Token:         token,
		PollerTimeout: time.Second * 10,

		UserService:   userService,
		DeviceService: deviceService,
		BirdService:   birdService,
	})

	// Starting the bot
	b.Start()
}
