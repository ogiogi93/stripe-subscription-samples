package main

import (
	"cloud.google.com/go/firestore"
)

func GetSubscriptionTx(tx *firestore.Transaction, id string) (*Subscription, error) {
	dr := fsClient.Collection(CollectionNameSubscription).Doc(id)
	ds, err := tx.Get(dr)
	if err != nil {
		return nil, err
	}
	var s Subscription
	if err := ds.DataTo(&s); err != nil {
		return nil, err
	}
	s.ID = ds.Ref.ID
	return &s, nil
}

func GetUserSubscriptionTx(tx *firestore.Transaction, id string) (*UserSubscription, error) {
	dr := fsClient.Collection(CollectionNameUserSubscription).Doc(id)
	ds, err := tx.Get(dr)
	if err != nil {
		return nil, err
	}
	var s UserSubscription
	if err := ds.DataTo(&s); err != nil {
		return nil, err
	}
	s.ID = ds.Ref.ID
	return &s, nil
}

func CreateUserSubscriptionTx(tx *firestore.Transaction, ub *UserSubscription) (*UserSubscription, error) {
	dr := fsClient.Collection(CollectionNameUserSubscription).Doc(ub.ID)
	if err := tx.Set(dr, ub); err != nil {
		return nil, err
	}
	return ub, nil
}

func UpdateUserSubscriptionTx(tx *firestore.Transaction, ub *UserSubscription) error {
	dr := fsClient.Collection(CollectionNameUserSubscription).Doc(ub.ID)
	return tx.Set(dr, ub)
}
