package repo

import "go.mongodb.org/mongo-driver/mongo/options"

func replaceUpsert() *options.ReplaceOptions {
	upsert := true
	return &options.ReplaceOptions{Upsert: &upsert}
}
