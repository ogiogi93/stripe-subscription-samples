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

type UpdateUserSubscriptionImmediatelyRequest struct {
	CustomerID     string `json:"customer_id"`
	SubscriptionID string `json:"subscription_id"`
	PlanID         string `json:"plan_id"`
}

type UpdateUserSubscriptionImmediatelyResponse struct {
	Status       stripe.PaymentIntentStatus `json:"status"`
	ClientSecret string                     `json:"client_secret"`
}

func UpdateUserSubscriptionImmediatelyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req *UpdateUserSubscriptionImmediatelyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}

	idempotencyKey := uuid.New().String()
	var intent *stripe.PaymentIntent
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

		// Subscriptionの更新処理を実行する
		subParams := &stripe.SubscriptionParams{
			BillingCycleAnchorNow: stripe.Bool(true),
			ProrationBehavior:     stripe.String(string(stripe.SubscriptionProrationBehaviorNone)),
			CancelAtPeriodEnd:     stripe.Bool(false),
		}
		subParams.AddMetadata("plan_id", plan.ID)
		subParams.AddExpand("latest_invoice.payment_intent") // レスポンスとして最新のInvoiceに紐づくPaymentIntentを取得したいためAddExpandに指定しておく
		subParams.SetIdempotencyKey(idempotencyKey)          // 冪等キー
		s, err := client.Subscriptions.Update(ub.StripeSubscriptionID, subParams)
		err = handleStripeError(err) // 冪等チェックエラーだった場合は処理を続行する
		if err != nil {
			return err
		}
		intent = s.LatestInvoice.PaymentIntent
		// サブスクリプションプランのデータを更新する
		ub.RenewalAll(plan.ID, s)
		return UpdateUserSubscriptionTx(tx, ub)
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("updateUserSubscriptionImmediatelyHandler: %v", err)
		return
	}
	res := UpdateUserSubscriptionImmediatelyResponse{
		Status:       intent.Status,
		ClientSecret: intent.ClientSecret,
	}
	if err = json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("updateUserSubscriptionImmediatelyHandler: %v", err)
		return
	}
}
