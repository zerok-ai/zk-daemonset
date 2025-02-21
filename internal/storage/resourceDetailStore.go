package storage

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	zklogger "github.com/zerok-ai/zk-utils-go/logs"
	storage "github.com/zerok-ai/zk-utils-go/storage/redis/config"
	"time"
	"zk-daemonset/internal/config"
	"zk-daemonset/internal/models"
)

const (
	resourceDetailExpiry time.Duration = time.Minute * 30
	podDetailsLogTag     string        = "PodDetailsStore"
)

var ctx = context.Background()

type ResourceDetailStore struct {
	redisClient *redis.Client
}

func GetNewPodDetailsStore(configs config.AppConfigs) *ResourceDetailStore {
	dbName := "imageStore"
	redisConfig := configs.Redis
	zklogger.Debug(podDetailsLogTag, "Host: %s, Port: %s, db = %d", redisConfig.Host, redisConfig.Port, redisConfig.DBs[dbName])

	_redisClient := storage.GetRedisConnection(dbName, redisConfig)
	imgRedis := &ResourceDetailStore{
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

func (resourceDetailStore ResourceDetailStore) SetPodDetails(podIP string, podDetails models.PodDetails) error {
	podItems := map[string]interface{}{}
	podItems["spec"] = getSerialisedValue(podDetails.Spec)
	podItems["metadata"] = getSerialisedValue(podDetails.Metadata)
	podItems["status"] = getSerialisedValue(podDetails.Status)
	if _, err := resourceDetailStore.redisClient.HMSet(ctx, podIP, podItems).Result(); err != nil {
		zklogger.Error(podDetailsLogTag, "error in SetPodDetails %v\n", err)
		return err
	}
	_, err := resourceDetailStore.redisClient.Expire(ctx, podIP, resourceDetailExpiry).Result()
	if err != nil {
		return err
	}
	return nil
}

func (resourceDetailStore ResourceDetailStore) SetServiceDetails(serviceIP string, serviceDetails models.ServiceDetails) error {
	items := map[string]interface{}{}
	items["metadata"] = getSerialisedValue(serviceDetails.Metadata)
	if _, err := resourceDetailStore.redisClient.HMSet(ctx, serviceIP, items).Result(); err != nil {
		zklogger.Error(podDetailsLogTag, "error in SetServiceDetails %v\n", err)
		return err
	}
	_, err := resourceDetailStore.redisClient.Expire(ctx, serviceIP, resourceDetailExpiry).Result()
	if err != nil {
		return err
	}
	return nil
}
