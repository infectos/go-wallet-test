package repository

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=postgres_user password=postgres_password dbname=postgres_db port=5429 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)

	// Clean up before test
	db.Exec("DROP TABLE IF EXISTS transactions, wallets")
	Migrate(db)
	return db
}

func TestWalletRepository_CreateWallet(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)

	t.Run("Create new wallet", func(t *testing.T) {
		wallet, err := repo.CreateWallet(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, int64(0), wallet.Balance)
	})
}

func TestWalletRepository_GetWallet(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)

	t.Run("Get existing wallet", func(t *testing.T) {
		baseWallet, err := repo.CreateWallet(context.Background())
		assert.NoError(t, err)

		wallet, err := repo.GetWallet(context.Background(), baseWallet.ID)
		assert.NoError(t, err)
		assert.Equal(t, baseWallet.ID, wallet.ID)
	})

	t.Run("Get non-existent wallet", func(t *testing.T) {
		_, err := repo.GetWallet(context.Background(), "b716d883-8566-4e0d-be29-d07211f027a2")
		assert.ErrorContains(t, err, "wallet not found")
	})
}

func TestWalletRepository_UpdateBalance(t *testing.T) {
	db := setupTestDB(t)
	repo := NewWalletRepository(db)
	ctx := context.Background()

	t.Run("Successful deposit", func(t *testing.T) {
		baseWallet, err := repo.CreateWallet(context.Background())
		assert.NoError(t, err)

		wallet, err := repo.UpdateBalance(ctx, baseWallet.ID, 100, "DEPOSIT")
		assert.NoError(t, err)
		assert.Equal(t, int64(100), wallet.Balance)
	})

	t.Run("Successful withdrawal", func(t *testing.T) {
		baseWallet, err := repo.CreateWallet(context.Background())
		assert.NoError(t, err)
		_, err = repo.UpdateBalance(ctx, baseWallet.ID, 200, "DEPOSIT")
		assert.NoError(t, err)

		wallet, err := repo.UpdateBalance(ctx, baseWallet.ID, 100, "WITHDRAW")
		assert.NoError(t, err)
		assert.Equal(t, int64(100), wallet.Balance)
	})

	t.Run("Insufficient funds", func(t *testing.T) {
		baseWallet, err := repo.CreateWallet(context.Background())
		assert.NoError(t, err)

		_, err = repo.UpdateBalance(ctx, baseWallet.ID, 100, "WITHDRAW")
		assert.ErrorContains(t, err, "insufficient funds")
	})

	t.Run("Concurrent updates", func(t *testing.T) {
		baseWallet, err := repo.CreateWallet(context.Background())
		assert.NoError(t, err)

		var wg sync.WaitGroup
		count := 100
		wg.Add(count)

		for i := 0; i < count; i++ {
			go func() {
				defer wg.Done()
				_, err := repo.UpdateBalance(ctx, baseWallet.ID, 1, "DEPOSIT")
				assert.NoError(t, err)
			}()
		}

		wg.Wait()
		wallet, err := repo.GetWallet(ctx, baseWallet.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(count), wallet.Balance)
	})
}
