package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/stripe/stripe-go"
)

type UpdateUserSubscriptionPaymentRequest struct {
	CustomerID     string `json:"customer_id"`
	SubscriptionID string `json:"subscription_id"`
	SourceID       string `json:"source_id"`
}

func UpdateUserSubscriptionPaymentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req *UpdateUserSubscriptionPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}

	err := fsClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		sub, _ := GetSubscriptionTx(tx, req.SubscriptionID)
		ub, _ := GetUserSubscriptionTx(tx, sub.UserSubscriptionID(req.CustomerID))

		// 支払い方法を変更する https://stripe.com/docs/api/subscriptions/update
		params := &stripe.SubscriptionParams{
			DefaultSource: stripe.String(req.SourceID),
		}
		_, err := client.Subscriptions.Update(ub.StripeSubscriptionID, params)
		return err
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("updateUserSubscriptionHandler: %v", err)
		return
	}
}
