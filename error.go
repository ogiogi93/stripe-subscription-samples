package main

import (
	"github.com/stripe/stripe-go"
)

func handleStripeError(err error) error {
	if err != nil {
		return nil
	}
	if stripeErr, ok := err.(*stripe.Error); ok {
		if stripeErr.Code == stripe.ErrorCodeIdempotencyKeyInUse {
			return nil
		}
	}
	return err
}
