// billsplit/internal/handler/expenses.go
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/fergalhk-lab/apps/billsplit/internal/domain"
	"github.com/fergalhk-lab/apps/billsplit/internal/fxrates"
	"github.com/fergalhk-lab/apps/billsplit/internal/middleware"
	"github.com/fergalhk-lab/apps/billsplit/internal/service"
	"github.com/fergalhk-lab/apps/billsplit/internal/store"
)

func addExpenseHandler(expenses *service.ExpenseService, fxCache *fxrates.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupID := r.PathValue("id")
		var req struct {
			Description string             `json:"description"`
			Amount      float64            `json:"amount"`
			Currency    string             `json:"currency"`
			PaidBy      string             `json:"paidBy"`
			Splits      map[string]float64 `json:"splits"`
		}
		if err := decodeJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		groupCurrency, err := expenses.GetGroupCurrency(r.Context(), groupID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(w, http.StatusNotFound, "group not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to read group")
			return
		}

		inputCurrency := req.Currency
		if inputCurrency == "" {
			inputCurrency = groupCurrency
		}

		amount := req.Amount
		var originalExpense *domain.OriginalExpense
		if inputCurrency != groupCurrency {
			rates, err := fxCache.Get(r.Context())
			if err != nil {
				writeError(w, http.StatusServiceUnavailable, "exchange rates unavailable")
				return
			}
			converted, err := rates.Convert(req.Amount, inputCurrency, groupCurrency)
			if err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			originalExpense = &domain.OriginalExpense{Currency: inputCurrency, Amount: req.Amount}
			amount = converted
		}

		username := middleware.UsernameFromCtx(r)
		eventID, err := expenses.AddExpense(r.Context(), groupID, username, req.Description, req.PaidBy, amount, req.Splits, originalExpense)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"id": eventID})
	}
}

func listEventsHandler(expenses *service.ExpenseService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupID := r.PathValue("id")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		if limit <= 0 {
			limit = 20
		}
		events, total, err := expenses.ListEvents(r.Context(), groupID, limit, offset)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list events")
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"events": events,
			"total":  total,
		})
	}
}

func cancelExpenseHandler(expenses *service.ExpenseService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupID := r.PathValue("id")
		eventID := r.PathValue("eventId")
		username := middleware.UsernameFromCtx(r)
		if err := expenses.CancelExpense(r.Context(), groupID, username, eventID); err != nil {
			if errors.Is(err, service.ErrEventNotFound) {
				writeError(w, http.StatusNotFound, "event not found")
				return
			}
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
