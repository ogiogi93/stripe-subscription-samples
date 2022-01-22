package main

import (
	"github.com/stripe/stripe-go"
	"time"
)

const (
	CollectionNameSubscription     = "Subscription"
	CollectionNameUserSubscription = "UserSubscription"
)

// Plan サブスクプラン
type Plan struct {
	ID              string     `firestore:"id"`
	Title           string     `firestore:"title"`
	StripeProductID string     `firestore:"stripe_product_id"`
	StripePriceID   string     `firestore:"stripe_product_id"`
	Price           int32      `firestore:"price"`
	Benefits        []*Benefit `firestore:"benefits"`
}

// Benefit サブスク適用のためのデータを定義(割引額等)。今回は触れない
type Benefit struct {
	ID    string `firestore:"id"`
	Title string `firestore:"title"`
	// DiscountValue int32 `firestore:"discount_value"`
}

// Subscription サブスクは複数のプランを保持できる
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

// UserSubscription ユーザー毎の加入サブスクプランの状態を定義
type UserSubscription struct {
	ID                    string                    `firestore:"-" json:"-"`
	CustomerID            string                    `firestore:"customer_id"`
	SubscriptionID        string                    `firestore:"subscription_id"`
	PlanID                string                    `firestore:"plan_id"`
	NextPlanID            string                    `firestore:"next_plan_id"`
	Status                stripe.SubscriptionStatus `firestore:"status"`
	LatestPaymentIntentID string                    `firestore:"latest_payment_intent_id"`
	StartedAt             time.Time                 `firestore:"started_at"`

	CurrentPeriodStart time.Time `firestore:"current_period_start"`
	CurrentPeriodEnd   time.Time `firestore:"current_period_end"`
}

func NewUserSubscription(customerID, subscriptionID, planID string, sub *stripe.Subscription) *UserSubscription {
	return &UserSubscription{
		CustomerID:            customerID,
		SubscriptionID:        subscriptionID,
		PlanID:                planID,
		Status:                sub.Status,
		LatestPaymentIntentID: sub.LatestInvoice.PaymentIntent.ID,
		StartedAt:             time.Now(),
		CurrentPeriodStart:    time.Unix(sub.CurrentPeriodStart, 0),
		CurrentPeriodEnd:      time.Unix(sub.CurrentPeriodEnd, 0),
	}
}
