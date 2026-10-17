package utils

import (
	"context"
	"github.com/RandySteven/go-kopi/enums"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func Store[T any](ctx context.Context, client *mongo.Client, collection enums.MongoCollection, request *T) (result *T, err error) {
	coll := client.Database("").Collection(collection.ToString())
	_, err = coll.InsertOne(ctx, request)
	if err != nil {
		return nil, err
	}
	result = request
	return result, nil
}

func Finds[T any](ctx context.Context, client *mongo.Client, collection enums.MongoCollection) (results []*T, err error) {
	coll := client.Database("").Collection(collection.ToString())
	cur, err := coll.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result = new(T)
		err := cur.Decode(&result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}
