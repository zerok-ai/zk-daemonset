package storage

import (
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

func (podDetailStore PodDetailStore) SetPodDetails(podIP string, podDetails models.PodDetails) {
	podDetailStore.redisClient.HMSet(ctx, podIP, map[string]interface{}{"spec": podDetails.Spec, "metadata": podDetails.Metadata})
}
