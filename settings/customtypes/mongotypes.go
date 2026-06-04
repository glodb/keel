package customtypes

import "go.mongodb.org/mongo-driver/mongo"

type M map[string]interface{}
type A []interface{}

type WriteModel mongo.WriteModel

type GeoPoint struct {
	Type        string    `json:"type" bson:"type"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}
