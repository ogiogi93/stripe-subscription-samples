package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
)

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	p, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	ev, err := webhook.ConstructEvent(
		p,
		r.Header.Get("Stripe-Signature"),
		os.Getenv("STRIPE_WEBHOOK_SIGNATURE"),
	)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch ev.Type {
	case "invoice.payment_succeeded", "invoice.payment_failed":
		var invoice stripe.Invoice
		err := json.Unmarshal(ev.Data.Raw, &invoice)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = renewalUserSubscription(context.Background(), invoice)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func renewalUserSubscription(ctx context.Context, inv stripe.Invoice) error {
	line := inv.Lines.Data[0]

	subscriptionID := line.Metadata["subscription_id"]
	planID := line.Metadata["plan_id"]

	err := fsClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		sub, _ := GetSubscriptionTx(tx, subscriptionID)
		ub, _ := GetUserSubscriptionTx(tx, sub.UserSubscriptionID(inv.Customer.ID))

		// Stripe上のSubscriptionを取得する(自動更新後の状態)
		stripeSub, _ := client.Subscriptions.Get(ub.StripeSubscriptionID, nil)

		// 次回更新時にプラン変更するパターン
		if ub.NextPlanID != "" {
			ub.PlanID = planID
			ub.NextPlanID = ""
		}
		ub.RenewalAll(planID, stripeSub)
		return UpdateUserSubscriptionTx(tx, ub)
	})
	if err != nil {
		return err
	}
	return nil
}
