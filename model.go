package main

import (
	"fmt"
	"time"

	"github.com/stripe/stripe-go"
)

const (
	CollectionNameSubscription     = "Subscription"
	CollectionNameUserSubscription = "UserSubscription"
)

// Plan サブスクリプションのプラン
type Plan struct {
	ID              string     `firestore:"id"`
	Title           string     `firestore:"title"`
	StripeProductID string     `firestore:"stripe_product_id"`
	StripePriceID   string     `firestore:"stripe_product_id"`
	Price           int32      `firestore:"price"`
	Benefits        []*Benefit `firestore:"benefits"`
}

// Benefit サブスクリプション適用のためのデータを定義(割引額等)。今回は触れない
type Benefit struct {
	ID    string `firestore:"id"`
	Title string `firestore:"title"`
	// DiscountValue int32 `firestore:"discount_value"`
}

// Subscription サブスクリプションは複数のプランを保持できる
type Subscription struct {
	ID    string  `firestore:"-"`
	Title string  `firestore:"title"`
	Plans []*Plan `firestore:"plans"`
}

func (s *Subscription) Plan(planID string) *Plan {
	for _, plan := range s.Plans {
		if planID == plan.ID {
			return plan
		}
	}
	return nil
}

func (s *Subscription) UserSubscriptionID(customerID string) string {
	return fmt.Sprintf("%s-%s", customerID, s.ID)
}

// UserSubscription ユーザー毎のサブスクリプションプランの状態を定義
type UserSubscription struct {
	ID                    string                    `firestore:"-"`
	CustomerID            string                    `firestore:"customer_id"`
	SubscriptionID        string                    `firestore:"subscription_id"`
	PlanID                string                    `firestore:"plan_id"`
	NextPlanID            string                    `firestore:"next_plan_id"`
	Status                stripe.SubscriptionStatus `firestore:"status"`
	LatestPaymentIntentID string                    `firestore:"latest_payment_intent_id"`
	StartedAt             time.Time                 `firestore:"started_at"`

	StripeSubscriptionID     string `firestore:"stripe_subscription_id"`
	StripeSubscriptionItemID string `firestore:"stripe_subscription_item_id"`

	CurrentPeriodStart time.Time `firestore:"current_period_start"`
	CurrentPeriodEnd   time.Time `firestore:"current_period_end"`
}

func (us *UserSubscription) Renewal(planID string) {
	us.PlanID = planID
	us.NextPlanID = ""
}

func (us *UserSubscription) RenewalAll(planID string, sub *stripe.Subscription) {
	us.Renewal(planID)

	us.StripeSubscriptionID = sub.ID
	us.StripeSubscriptionItemID = sub.Items.Data[0].ID
	us.Status = sub.Status
	us.CurrentPeriodStart = time.Unix(sub.CurrentPeriodStart, 0)
	us.CurrentPeriodEnd = time.Unix(sub.CurrentPeriodEnd, 0)
	us.LatestPaymentIntentID = sub.LatestInvoice.PaymentIntent.ID
}

func NewUserSubscription(id, customerID, subscriptionID, planID string, sub *stripe.Subscription) *UserSubscription {
	return &UserSubscription{
		ID:                       id,
		CustomerID:               customerID,
		SubscriptionID:           subscriptionID,
		PlanID:                   planID,
		StripeSubscriptionID:     sub.ID,
		StripeSubscriptionItemID: sub.Items.Data[0].ID,
		Status:                   sub.Status,
		LatestPaymentIntentID:    sub.LatestInvoice.PaymentIntent.ID,
		StartedAt:                time.Now(),
		CurrentPeriodStart:       time.Unix(sub.CurrentPeriodStart, 0),
		CurrentPeriodEnd:         time.Unix(sub.CurrentPeriodEnd, 0),
	}
}
