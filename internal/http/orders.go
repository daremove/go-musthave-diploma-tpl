package router

import (
	"errors"
	"fmt"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/middlewares"
	"github.com/daremove/go-musthave-diploma-tpl/tree/master/internal/services"
	"net/http"
)

func PostOrders(w http.ResponseWriter, r *http.Request) {
	orderId := middlewares.GetParsedTextData(w, r)

	if len(orderId) == 0 {
		http.Error(w, "Order id is empty", http.StatusUnprocessableEntity)
		return
	}

	orderService := middlewares.GetServiceFromContext[services.OrderService](w, r, middlewares.OrderServiceKey)

	if !orderService.VerifyOrderId(orderId) {
		http.Error(w, "Order id is invalid", http.StatusUnprocessableEntity)
		return
	}

	user := middlewares.GetUserFromContext(w, r)

	if err := orderService.CreateOrder(r.Context(), orderId, user.ID); err != nil {
		if errors.Is(err, services.ErrDuplicateOrderByOriginalUser) {
			w.WriteHeader(http.StatusOK)
			return
		}

		if errors.Is(err, services.ErrDuplicateOrder) {
			http.Error(w, "Order was created by another user", http.StatusConflict)
			return
		}

		http.Error(w, fmt.Sprintf("Error occurred during creating order: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
