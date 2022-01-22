package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go"
	"log"
	"net/http"
)

type CreateUserSubscriptionRequest struct {
	CustomerID     string `json:"customer_id"`
	SubscriptionID string `json:"subscription_id"`
	PlanID         string `json:"plan_id"`
}

type CreateUserSubscriptionResponse struct {
	Status       stripe.PaymentIntentStatus `json:"status"`
	ClientSecret string                     `json:"client_secret"`
}

func createUserSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req *CreateUserSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}

	idempotencyKey := uuid.New().String()
	var intent *stripe.PaymentIntent
	err := fsClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// DBからSubscriptionを取得する
		sub, _ := getSubscription(tx, req.SubscriptionID)
		plan := sub.Plan(req.PlanID)

		// Stripe上にてSubscriptionを作成する https://stripe.com/docs/api/subscriptions/create
		params := &stripe.SubscriptionParams{
			Customer: stripe.String(req.CustomerID),
			Items: []*stripe.SubscriptionItemsParams{
				{
					Price:    stripe.String(plan.StripePriceID), // ユーザーが選択したサブスクリプションプランのPriceIDをセットする
					Quantity: stripe.Int64(1),                   // 数量、今回は1プランを契約する
				},
			},
			CancelAtPeriodEnd: stripe.Bool(false),                                              // 自動更新有無、falseにすることで期限が切れたらStripe側で自動更新される
			ProrationBehavior: stripe.String(string(stripe.SubscriptionProrationBehaviorNone)), // 日割り計算に関するパラメータ。今回は日割りなしを想定しているのでNoneを選択する https://stripe.com/docs/billing/subscriptions/prorations
			PaymentBehavior:   stripe.String("allow_incomplete"),                               // 支払い処理に関するパラメータ。決済処理まで一気に処理をすすめる場合は allow_incompleteを選択する
		}
		params.AddExpand("latest_invoice.payment_intent") // レスポンスとして最新のInvoiceに紐づくPaymentIntentを取得したいためAddExpandに指定しておく
		params.SetIdempotencyKey(idempotencyKey)          // 冪等キー

		s, err := client.Subscriptions.New(params)
		err = handleStripeError(err) // 冪等チェックエラーだった場合は処理を続行する
		if err != nil {
			return err
		}
		intent = s.LatestInvoice.PaymentIntent
		ub := NewUserSubscription(req.CustomerID, sub.ID, plan.ID, s)
		ub, _ = createSubscription(tx, ub)
		return nil
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("createUserSubscriptionHandler: %v", err)
		return
	}
	res := CreateUserSubscriptionResponse{
		Status:       intent.Status,
		ClientSecret: intent.ClientSecret,
	}
	if err = json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("createUserSubscriptionHandler: %v", err)
		return
	}
}

func getSubscription(tx *firestore.Transaction, id string) (*Subscription, error) {
	dr := fsClient.Collection(CollectionNameSubscription).Doc(id)
	ds, err := tx.Get(dr)
	if err != nil {
		return nil, err
	}
	var s Subscription
	if err := ds.DataTo(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

func createSubscription(tx *firestore.Transaction, ub *UserSubscription) (*UserSubscription, error) {
	dr := fsClient.Collection(CollectionNameUserSubscription).NewDoc()
	if err := tx.Set(dr, ub); err != nil {
		return nil, err
	}
	ub.ID = dr.ID
	return ub, nil
}
