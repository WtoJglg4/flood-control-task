package main

import (
	"context"
	"encoding/json"
	"fmt"
	floodcontrol "github/flood-control-task/floodControl"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

// FloodControl интерфейс, который нужно реализовать.
// Рекомендуем создать директорию-пакет, в которой будет находиться реализация.
type FloodControl interface {
	// Check возвращает false если достигнут лимит максимально разрешенного
	// кол-ва запросов согласно заданным правилам флуд контроля.
	Check(ctx context.Context, userID int64) (bool, error)
}

func main() {
	//initialize a configs
	if err := initConfig(); err != nil {
		log.Fatalf("error initializing configs: %s\n", err.Error())
	}

	//initialize storage
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", viper.GetString("redis.host"), viper.GetString("redis.port")),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})

	//create an example User object
	ExampleUser := floodcontrol.User{
		Id:          52,
		TimeToCalls: make(map[int64]int64),
	}
	ExampleUser.TimeToCalls[0] = 0

	//marshal user
	jsonUser, err := json.Marshal(ExampleUser)
	if err != nil {
		log.Printf("error while marshaling json: %s\n", err.Error())
		return
	}

	ctx := context.Background()

	//set new user to redis
	err = client.Set(ctx, fmt.Sprint(ExampleUser.Id), jsonUser, 0).Err()
	if err != nil {
		log.Printf("error while set json to redis: %s\n", err.Error())
		return
	}

	//initialize a new FloodControleService
	fc := floodcontrol.NewFloodControlService(viper.GetInt64("timeInterval"), viper.GetInt64("checkCallsCount"), client)

	//small test
	for i := 0; i < 10; i++ {
		res, err := fc.Check(ctx, ExampleUser.Id)
		if err != nil {
			log.Printf("error while checking: %s\n", err.Error())
		}

		if res {
			log.Println("OK")
		} else {
			log.Println("NOT OK")
		}
	}
}

// initializing config
func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
