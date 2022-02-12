package utils

import (
	"os"

	"go.mongodb.org/mongo-driver/bson"
)

// get env vars; return empty string if not found.
func Env(key string) string {
	return os.Getenv(key)
}

// convert map to bson.M for mongoDB docs.
func MapToBson(data map[string]interface{}) bson.M {
	return bson.M(data)
}
