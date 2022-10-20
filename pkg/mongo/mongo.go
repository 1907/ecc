package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var Db *mongo.Database

func Conn(dns string, db string) error {
	client, err := mongo.NewClient(options.Client().ApplyURI(dns))
	if err != nil {
		return err
	}

	err = client.Connect(nil)
	if err != nil {
		return err
	}

	err = client.Ping(nil, readpref.Primary())
	if err != nil {
		return err
	}

	Db = client.Database(db)
	return nil
}

func GetDb() *mongo.Database {
	return Db
}

func CheckCollection(collection string) bool {
	collections, err := GetDb().ListCollectionNames(context.TODO(), bson.M{})
	if err != nil {
		return false
	}

	for _, c := range collections {
		if c == collection {
			return true
		}
	}

	return false
}

func DelCollection(collection string) error {
	err := GetCollection(collection).Drop(context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func CreateCollection(collection string) error {
	return GetDb().CreateCollection(context.TODO(), collection)
}

func GetCollection(name string) *mongo.Collection {
	return GetDb().Collection(name)
}

func GetObjectId(v string) primitive.ObjectID {
	id, _ := primitive.ObjectIDFromHex(v)
	return id
}

func GetCollectionAllData(collection string) ([]bson.M, error) {
	var res []bson.M
	cursor, err := GetCollection(collection).Find(context.TODO(), bson.M{})
	if err != nil {
		return res, err
	}

	for cursor.Next(context.TODO()) {
		var result bson.M
		err := cursor.Decode(&result)
		if err != nil {
			return res, err
		}

		res = append(res, result)
	}

	return res, nil
}

func DeleteMany(collection string, filter bson.M) error {
	_, err := GetCollection(collection).DeleteMany(context.TODO(), filter)
	return err
}

func CreateDocs(c *mongo.Collection, docs interface{}) (id primitive.ObjectID, e error) {
	re, err := c.InsertOne(context.TODO(), docs)
	return re.InsertedID.(primitive.ObjectID), err
}
