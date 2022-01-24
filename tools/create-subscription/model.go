package main

// Plan サブスクプラン
type Plan struct {
	ID              string     `firestore:"id"`
	Title           string     `firestore:"title"`
	StripeProductID string     `firestore:"stripe_product_id"`
	StripePriceID   string     `firestore:"stripe_price_id"`
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
