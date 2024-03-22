package floodcontrol

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type FloodControlService struct {
	lastCallTime  int64
	timeInterval  time.Duration
	maxCallsCount int64
	redis         *redis.Client
	mu            sync.Mutex
}

func NewFloodControlService(timeIntervalInt64, maxCallsCount int64, redis *redis.Client) *FloodControlService {
	return &FloodControlService{
		lastCallTime:  0,
		timeInterval:  time.Duration(timeIntervalInt64),
		maxCallsCount: maxCallsCount,
		redis:         redis,
		mu:            sync.Mutex{},
	}
}

// Check возвращает false если достигнут лимит максимально разрешенного
// кол-ва запросов согласно заданным правилам флуд контроля.
func (fcs *FloodControlService) Check(ctx context.Context, userID int64) (bool, error) {
	fcs.mu.Lock()

	now := time.Now().Unix()
	jsonUser, err := fcs.redis.Get(ctx, fmt.Sprint(userID)).Result()
	if err != nil {
		log.Printf("error while get user from redis: %s\n", err.Error())
		return false, err
	}
	fcs.mu.Unlock()

	user := User{}
	err = json.Unmarshal([]byte(jsonUser), &user)
	if err != nil {
		log.Printf("error while unmarshaling json: %s\n", err.Error())
		return false, err
	}

	user.TimeToCalls[now] = user.TimeToCalls[fcs.lastCallTime] + 1
	fcs.lastCallTime = now

	byteUser, err := json.Marshal(user)
	if err != nil {
		log.Printf("error while marshaling json: %s\n", err.Error())
		return false, err
	}

	fcs.mu.Lock()
	err = fcs.redis.Set(ctx, fmt.Sprint(userID), byteUser, 0).Err()
	if err != nil {
		log.Printf("error while set json to redis: %s\n", err.Error())
		return false, nil
	}
	fcs.mu.Unlock()

	prevCalls := user.TimeToCalls[now-int64(fcs.timeInterval)]
	if prevCalls == 0 {
		for time, calls := range user.TimeToCalls {
			if time > 0 && time < now-int64(fcs.timeInterval) && calls > prevCalls {
				prevCalls = calls
			}
		}
	}
	// log.Println(user, user.TimeToCalls[now], prevCalls)
	return user.TimeToCalls[now]-prevCalls <= fcs.maxCallsCount, nil
}

type User struct {
	Id          int64
	TimeToCalls map[int64]int64 `json:"timeToCalls"`
}
