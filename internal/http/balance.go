package router

import (
	"fmt"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/middlewares"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/models"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/services"
	"net/http"
)

func GetBalance(w http.ResponseWriter, r *http.Request) {
	balanceService := middlewares.GetServiceFromContext[services.BalanceService](w, r, middlewares.BalanceServiceKey)
	user := middlewares.GetUserFromContext(w, r)

	balance, err := balanceService.GetUserBalance(r.Context(), user.ID)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error occurred during getting balance: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	middlewares.EncodeJSONResponse(w, balance)
}

func CreateWithdrawal(w http.ResponseWriter, r *http.Request) {
	data := middlewares.GetParsedJSONData[models.Withdrawal](w, r)

	if data.ID == nil || data.Sum == nil {
		http.Error(w, "Request doesn't contain order or sum", http.StatusBadRequest)
		return
	}

	if len(*data.ID) == 0 {
		http.Error(w, "Order id is empty", http.StatusUnprocessableEntity)
		return
	}

	orderService := middlewares.GetServiceFromContext[services.OrderService](w, r, middlewares.OrderServiceKey)
	balanceService := middlewares.GetServiceFromContext[services.BalanceService](w, r, middlewares.BalanceServiceKey)

	if !orderService.VerifyOrderId(*data.ID) {
		http.Error(w, "Order id is invalid", http.StatusUnprocessableEntity)
		return
	}

	user := middlewares.GetUserFromContext(w, r)
	balance, err := balanceService.GetUserBalance(r.Context(), user.ID)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error occurred during getting balance: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if balance.Current < *data.Sum {
		http.Error(w, "There is not enough money", http.StatusPaymentRequired)
		return
	}

	if err := balanceService.CreateWithdrawal(r.Context(), *data.ID, user.ID, *data.Sum); err != nil {
		http.Error(w, fmt.Sprintf("Error occurred during creating withdrawal: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	balanceService := middlewares.GetServiceFromContext[services.BalanceService](w, r, middlewares.BalanceServiceKey)
	user := middlewares.GetUserFromContext(w, r)

	withdrawalFlow, err := balanceService.GetWithdrawalFlow(r.Context(), user.ID)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error occurred during getting withdrawals: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if len(withdrawalFlow) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	middlewares.EncodeJSONResponse(w, withdrawalFlow)
}
