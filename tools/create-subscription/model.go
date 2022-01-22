package main

import "time"

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
	ID string `firestore:"id"`
	// DiscountValue int32 `firestore:"discount_value"`
}

// Subscription サブスクは複数のプランを持っている
type Subscription struct {
	ID    string  `firestore:"-"`
	Title string  `firestore:"title"`
	Plans []*Plan `firestore:"plans"`
}

// UserSubscription ユーザー毎の加入サブスクプランの状態を定義
type UserSubscription struct {
	ID                    string    `firestore:"-" json:"-"`
	SubscriptionID        string    `firestore:"subscription_id" json:"subscription_id"`
	UserID                string    `firestore:"user_id" json:"user_id"`
	PlanID                string    `firestore:"plan_id" json:"plan_id"`
	NextPlanID            string    `firestore:"next_plan_id"`
	Status                string    `firestore:"status" json:"status"`
	PaymentIssueType      string    `firestore:"payment_issue_type" json:"payment_issue_type"`
	LatestPaymentIntentID string    `firestore:"latest_payment_intent_id" json:"latest_payment_intent_id"`
	StartedAt             time.Time `firestore:"started_at" json:"started_at"`

	CurrentPeriodStart time.Time `firestore:"current_period_start" json:"current_period_start"`
	CurrentPeriodEnd   time.Time `firestore:"current_period_end" json:"current_period_end"`
}
