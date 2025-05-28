package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"r.drannikov/wallet-test/internal/model"
	"r.drannikov/wallet-test/internal/repository"
)

var (
	ErrWalletExists      = errors.New("wallet already exists")
	ErrWalletNotFound    = errors.New("wallet not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidAmount     = errors.New("amount must be positive")
	ErrInvalidOperation  = errors.New("invalid operation type")
)

type IWalletService interface {
	CreateWallet(ctx context.Context) (*model.Wallet, error)
	GetBalance(ctx context.Context, walletID string) (int64, error)
	ProcessTransaction(ctx context.Context, walletID string, operationType string, amount int64) (int64, error)
}

type WalletService struct {
	repo repository.IWalletRepository
}

func NewWalletService(repo repository.IWalletRepository) IWalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) CreateWallet(ctx context.Context) (*model.Wallet, error) {
	wallet, err := s.repo.CreateWallet(ctx)
	if err != nil {
		if err.Error() == "wallet already exists" {
			return nil, ErrWalletExists
		}
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}
	return wallet, nil
}

func (s *WalletService) GetBalance(ctx context.Context, walletID string) (int64, error) {
	wallet, err := s.repo.GetWallet(ctx, walletID)
	if err != nil {
		return 0, ErrWalletNotFound
	}
	return wallet.Balance, nil
}

func (s *WalletService) ProcessTransaction(ctx context.Context, walletID string, operationType string, amount int64) (int64, error) {
	if amount <= 0 {
		return 0, ErrInvalidAmount
	}

	if operationType != "DEPOSIT" && operationType != "WITHDRAW" {
		return 0, ErrInvalidOperation
	}

	wallet, err := s.repo.UpdateBalance(ctx, walletID, amount, operationType)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return 0, ErrWalletNotFound
		case err.Error() == "insufficient funds":
			return 0, ErrInsufficientFunds
		default:
			return 0, fmt.Errorf("transaction failed: %w", err)
		}
	}

	return wallet.Balance, nil
}
