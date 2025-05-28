package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"r.drannikov/wallet-test/internal/service"
)

type WalletOperationRequet struct {
	WalletID      string `json:"walletId"`
	OperationType string `json:"operationType"`
	Amount        string `json:"amount"`
}

func (app *application) HandleWalletOperation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WalletID      string `json:"walletId"`
		OperationType string `json:"operationType"`
		Amount        int64  `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "Invalid request")
		return
	}

	if _, err := uuid.Parse(req.WalletID); err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "Invalid wallet ID")
		return
	}

	balance, err := app.walletService.ProcessTransaction(r.Context(), req.WalletID, req.OperationType, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrWalletNotFound):
			app.notFoundResponse(w, r)
		case errors.Is(err, service.ErrInsufficientFunds),
			errors.Is(err, service.ErrInvalidAmount),
			errors.Is(err, service.ErrInvalidOperation):
			app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	response := envelope{
		"walletId": req.WalletID,
		"balance":  balance,
	}

	err = app.writeJSON(w, 200, response, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) HandleWalletBalance(w http.ResponseWriter, r *http.Request) {
	walletUUID := chi.URLParam(r, "wallet_uuid")
	if _, err := uuid.Parse(walletUUID); err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "Invalid wallet ID")
		return
	}

	balance, err := app.walletService.GetBalance(r.Context(), walletUUID)
	if err != nil {
		if errors.Is(err, service.ErrWalletNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	response := envelope{
		"walletId": walletUUID,
		"balance":  balance,
	}

	err = app.writeJSON(w, 200, response, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) HandleCreateWallet(w http.ResponseWriter, r *http.Request) {
	result, err := app.walletService.CreateWallet(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, service.ErrWalletExists):
			app.errorResponse(w, r, http.StatusConflict, err.Error())
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	response := envelope{
		"id":      result.ID,
		"balance": result.Balance,
	}

	err = app.writeJSON(w, 202, response, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
