package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/stripe/stripe-go"
)

type CancelUserSubscriptionRequest struct {
	CustomerID     string `json:"customer_id"`
	SubscriptionID string `json:"subscription_id"`
}

func CancelUserSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req *CancelUserSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}

	err := fsClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		sub, _ := GetSubscriptionTx(tx, req.SubscriptionID)
		ub, _ := GetUserSubscriptionTx(tx, sub.UserSubscriptionID(req.CustomerID))

		// 自動更新を無効にする https://stripe.com/docs/billing/subscriptions/cancel
		params := &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true),
		}
		_, _ = client.Subscriptions.Update(ub.StripeSubscriptionID, params)
		return nil
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("cancelSubscriptionHandler: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
