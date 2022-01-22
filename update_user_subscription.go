package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/stripe/stripe-go"
)

type UpdateUserSubscriptionRequest struct {
	CustomerID     string `json:"customer_id"`
	SubscriptionID string `json:"subscription_id"`
	PlanID         string `json:"plan_id"`
}

func UpdateUserSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req *UpdateUserSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}

	err := fsClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		sub, _ := GetSubscriptionTx(tx, req.SubscriptionID)
		// 新しいサブスクリプションのプランのデータを取得する
		plan := sub.Plan(req.PlanID)

		ub, _ := GetUserSubscriptionTx(tx, sub.UserSubscriptionID(req.CustomerID))

		// Subscriptionに設定されているSubscriptionItemを変更する https://stripe.com/docs/billing/subscriptions/upgrade-downgrade
		itemParams := &stripe.SubscriptionItemParams{
			Price:             stripe.String(plan.StripePriceID),
			ProrationBehavior: stripe.String(string(stripe.SubscriptionProrationBehaviorNone)),
		}
		_, _ = client.SubscriptionItems.Update(ub.StripeSubscriptionItemID, itemParams)

		// SubscriptionのMetadataを新しいプランのIDに更新する
		subParams := &stripe.SubscriptionParams{}
		subParams.AddMetadata("plan_id", plan.ID)
		_, _ = client.Subscriptions.Update(ub.StripeSubscriptionID, subParams)


		// アプリ上で変更後のプランに関する情報を表示するため、DB上に変更後のPlanIDを保持しておく
		ub.NextPlanID = plan.ID
		return UpdateUserSubscriptionTx(tx, ub)
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("updateUserSubscriptionHandler: %v", err)
		return
	}
}
