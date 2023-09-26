package storage

import (
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	storage "github.com/zerok-ai/zk-utils-go/storage/redis/config"
	"time"
	"zk-daemonset/internal/config"
	"zk-daemonset/internal/models"
)

const (
	podDetailExpiry time.Duration = time.Hour * 24 * 30
)

type PodDetailStore struct {
	redisClient *redis.Client
}

func GetNewPodDetailsStore(configs config.AppConfigs) *PodDetailStore {
	dbName := "imageStore"
	redisConfig := configs.Redis
	fmt.Printf("Host: %s, Port: %s, db = %d\n", redisConfig.Host, redisConfig.Port, redisConfig.DBs[dbName])

	_redisClient := storage.GetRedisConnection(dbName, redisConfig)
	imgRedis := &PodDetailStore{
		redisClient: _redisClient,
	}
	return imgRedis
}

func getSerialisedValue(value interface{}) string {
	serialized, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(serialized)
}

func (podDetailStore PodDetailStore) SetPodDetails(podIP string, podDetails models.PodDetails) error {
	podItems := map[string]interface{}{}
	podItems["spec"] = getSerialisedValue(podDetails.Spec)
	podItems["metadata"] = getSerialisedValue(podDetails.Metadata)
	podItems["status"] = getSerialisedValue(podDetails.Status)
	if err := podDetailStore.redisClient.HMSet(ctx, podIP, podItems); err != nil {
		fmt.Printf("error in SetPodDetails %v\n", err)
	}
	return nil
}
