package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go"
)

type ReCreateUserSubscriptionRequest struct {
	CustomerID     string `json:"customer_id"`
	SubscriptionID string `json:"subscription_id"`
	PlanID         string `json:"plan_id"`
}

type ReCreateUserSubscriptionResponse struct {
	Status       stripe.PaymentIntentStatus `json:"status"`
	ClientSecret string                     `json:"client_secret"`
}

func ReCreateUserSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req *ReCreateUserSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}
	idempotencyKey := uuid.New().String()
	var intent *stripe.PaymentIntent
	err := fsClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		sub, _ := GetSubscriptionTx(tx, req.SubscriptionID)
		ub, _ := GetUserSubscriptionTx(tx, sub.UserSubscriptionID(req.CustomerID))

		plan := sub.Plan(req.PlanID)

		// 既存のStripe Subscriptionをキャンセルする
		_, err := client.Subscriptions.Cancel(ub.StripeSubscriptionID, nil)
		if err != nil {
			if stripeErr, ok := err.(*stripe.Error); ok {
				if stripeErr.HTTPStatusCode == 404 {
					// already canceled
				} else {
					return err
				}
			} else {
				return err
			}
		}

		// Subscriptionを新規登録する
		// create_user_subscription.go と同様の処理
		params := &stripe.SubscriptionParams{
			Customer: stripe.String(req.CustomerID),
			Items: []*stripe.SubscriptionItemsParams{
				{
					Price:    stripe.String(plan.StripePriceID),
					Quantity: stripe.Int64(1),
				},
			},
			CancelAtPeriodEnd: stripe.Bool(false),
			ProrationBehavior: stripe.String(string(stripe.SubscriptionProrationBehaviorNone)),
			PaymentBehavior:   stripe.String("allow_incomplete"),
		}
		params.AddMetadata("subscription_id", sub.ID)
		params.AddMetadata("plan_id", plan.ID)
		params.AddExpand("latest_invoice.payment_intent")
		params.SetIdempotencyKey(idempotencyKey)

		s, err := client.Subscriptions.New(params)
		err = handleStripeError(err)
		if err != nil {
			return err
		}

		intent = s.LatestInvoice.PaymentIntent
		ub.RenewalAll(plan.ID, s)
		ub, _ = CreateUserSubscriptionTx(tx, ub)
		return nil
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ReCreateUserSubscriptionHandler: %v", err)
		return
	}
	res := ReCreateUserSubscriptionResponse{
		Status:       intent.Status,
		ClientSecret: intent.ClientSecret,
	}
	if err = json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("ReCreateUserSubscriptionHandler: %v", err)
		return
	}
}
