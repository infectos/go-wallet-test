package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"r.drannikov/wallet-test/internal/model"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateWallet(ctx context.Context) (*model.Wallet, error) {
	args := m.Called(ctx)
	return args.Get(0).(*model.Wallet), args.Error(1)
}

func (m *MockRepository) GetWallet(ctx context.Context, id string) (*model.Wallet, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Wallet), args.Error(1)
}

func (m *MockRepository) UpdateBalance(ctx context.Context, id string, amount int64, operationType string) (*model.Wallet, error) {
	args := m.Called(ctx, id, amount, operationType)
	return args.Get(0).(*model.Wallet), args.Error(1)
}

func TestWalletService_CreateWallet(t *testing.T) {
	repo := new(MockRepository)
	svc := NewWalletService(repo)

	t.Run("Create new wallet", func(t *testing.T) {
		repo.On("CreateWallet", context.Background()).Return(&model.Wallet{}, nil)

		_, err := svc.CreateWallet(context.Background())
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

func TestWalletService_ProcessTransaction(t *testing.T) {
	repo := new(MockRepository)
	svc := NewWalletService(repo)
	ctx := context.Background()

	t.Run("Valid deposit", func(t *testing.T) {
		id := "ac0d0bfb-c11b-40d5-97fe-7581980c7f5c"
		repo.On("UpdateBalance", ctx, id, int64(100), "DEPOSIT").Return(
			&model.Wallet{ID: id, Balance: 100}, nil,
		)

		balance, err := svc.ProcessTransaction(ctx, id, "DEPOSIT", 100)
		assert.NoError(t, err)
		assert.Equal(t, int64(100), balance)
		repo.AssertExpectations(t)
	})

	t.Run("Insufficient funds", func(t *testing.T) {
		id := "27ce1329-de8d-4e04-8f81-d3aa6220671a"
		repo.On("UpdateBalance", ctx, id, int64(100), "WITHDRAW").Return(
			(*model.Wallet)(nil), errors.New("insufficient funds"),
		)

		_, err := svc.ProcessTransaction(ctx, id, "WITHDRAW", 100)
		assert.ErrorIs(t, err, ErrInsufficientFunds)
		repo.AssertExpectations(t)
	})
}
