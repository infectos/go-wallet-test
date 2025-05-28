package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"r.drannikov/wallet-test/internal/model"
	"r.drannikov/wallet-test/internal/repository"
	"r.drannikov/wallet-test/internal/service"
)

const version = "1.0.0"

type application struct {
	logger        *log.Logger
	config        Config
	walletService service.IWalletService
}

type Config struct {
	Port int `env:"SERVER_PORT,required"`
}

func main() {
	if err := godotenv.Load("config.env"); err != nil {
		log.Printf("No config.env file found, using environment variables")
	}

	cfg := Config{}

	if err := env.Parse(&cfg); err != nil {
		log.Println(err)
	}
	fmt.Printf("%+v", cfg)

	db, err := gorm.Open(postgres.Open(os.Getenv("DB_DSN")), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&model.Wallet{}, &model.Transaction{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	walletRepo := repository.NewWalletRepository(db)
	walletService := service.NewWalletService(walletRepo)

	app := &application{
		logger:        logger,
		config:        cfg,
		walletService: walletService,
	}

	router := app.getRoutes()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("Starting server on %s", srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}
