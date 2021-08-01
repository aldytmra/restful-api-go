package main

import (
	"github.com/aldytmra/restful-api-go/api"
	"github.com/go-redis/redis/v7"
)

var client *redis.Client

func main() {
	api.Run()
}
