package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"r.drannikov/wallet-test/internal/model"
)

type IWalletRepository interface {
	CreateWallet(ctx context.Context) (*model.Wallet, error)
	GetWallet(ctx context.Context, id string) (*model.Wallet, error)
	UpdateBalance(ctx context.Context, id string, amount int64, operationType string) (*model.Wallet, error)
}

type WalletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) IWalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) CreateWallet(ctx context.Context) (*model.Wallet, error) {
	wallet := &model.Wallet{}
	result := r.db.WithContext(ctx).Create(wallet)
	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return nil, fmt.Errorf("wallet already exists")
	}
	return wallet, result.Error
}

func (r *WalletRepository) GetWallet(ctx context.Context, id string) (*model.Wallet, error) {
	var wallet model.Wallet
	err := r.db.WithContext(ctx).First(&wallet, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("wallet not found")
	}
	return &wallet, err
}

func (r *WalletRepository) UpdateBalance(ctx context.Context, id string, amount int64, operationType string) (*model.Wallet, error) {
	var wallet model.Wallet

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&wallet, "id = ?", id).Error; err != nil {
			return err
		}

		if operationType == "WITHDRAW" && wallet.Balance < amount {
			return fmt.Errorf("insufficient funds")
		}

		if operationType == "DEPOSIT" {
			wallet.Balance += amount
		} else {
			wallet.Balance -= amount
		}

		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		transaction := model.Transaction{
			WalletID:      wallet.ID,
			OperationType: operationType,
			Amount:        amount,
		}

		return tx.Create(&transaction).Error
	})

	return &wallet, err
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&model.Wallet{}, &model.Transaction{})
}
