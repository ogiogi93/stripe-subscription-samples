package main

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v72"
	stripeClient "github.com/stripe/stripe-go/v72/client"
)

var (
	client   *stripeClient.API
	fsClient *firestore.Client
)

func init() {
	client = stripeClient.New(os.Getenv("STRIPE_API_KEY"), nil)
	cli, err := firestore.NewClient(context.Background(), os.Getenv("GCP_PROJECT"))
	if err != nil {
		log.Fatalf("Failed to create firestore client: %v", err)
	}
	fsClient = cli
}

func main() {
	ctx := context.Background()

	sub := &Subscription{
		ID:    uuid.New().String(),
		Title: "味噌ラーメンわくわく定額プラン",
		Plans: []*Plan{
			{
				ID:    uuid.New().String(),
				Title: "毎日ラーメン1杯無料プラン",
				Price: 3000,
				Benefits: []*Benefit{
					{
						ID: uuid.New().String(),
					},
				},
			},
			{
				ID:    uuid.New().String(),
				Title: "トッピング毎回1品無料",
				Price: 350,
				Benefits: []*Benefit{
					{
						ID: uuid.New().String(),
					},
				},
			},
		},
	}

	for _, plan := range sub.Plans {
		// Subscriptionの商品及び価格の詳細はこちら: https://stripe.com/docs/billing/prices-guide

		// Productの作成 https://stripe.com/docs/api/products/create
		productParams := &stripe.ProductParams{
			Name:                stripe.String(plan.Title),
			StatementDescriptor: stripe.String("Chompy"), // 明細書に記載する文字列. 5 ~ 22文字でアルファベットと数字のみなので注意. https://stripe.com/docs/statement-descriptors
		}
		productParams.AddMetadata("subscription_id", sub.ID)
		productParams.AddMetadata("plan_id", plan.ID)
		product, _ := client.Products.New(productParams)

		// Priceの作成 https://stripe.com/docs/api/prices/create
		priceParams := &stripe.PriceParams{
			Currency: stripe.String(string(stripe.CurrencyJPY)), // 通貨の設定, JPYを設定する
			Product:  stripe.String(product.ID),                 // 上記で作成したProductのIDを設定する
			Recurring: &stripe.PriceRecurringParams{ // サブスク期間の設定
				Interval:      stripe.String("day"), // 日毎
				IntervalCount: stripe.Int64(30),     // 30日
			},
			UnitAmount: stripe.Int64(3000), // 料金, 3000円
		}
		priceParams.AddMetadata("subscription_id", sub.ID)
		priceParams.AddMetadata("plan_id", plan.ID)
		price, _ := client.Prices.New(priceParams)

		plan.StripeProductID = product.ID
		plan.StripePriceID = price.ID
	}

	// DBに保存
	dr := fsClient.Collection("Subscription").Doc(sub.ID)
	if _, err := dr.Set(ctx, sub); err != nil {
		log.Fatalf("Failed to create subscription. err=%v", err)
	}
}
