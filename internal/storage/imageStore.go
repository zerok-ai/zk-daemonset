package storage

import (
	"context"
	"encoding/json"
	"fmt"
	redis "github.com/redis/go-redis/v9"
	storage "github.com/zerok-ai/zk-utils-go/storage/redis/config"
	"log"
	"time"
	"zk-daemonset/internal/config"
	"zk-daemonset/internal/models"
)

const (
	defaultExpiry  time.Duration = time.Hour * 24 * 30
	_hashSetName   string        = "zk_img_proc_map"
	hashSetVersion string        = "zk_img_proc_version"
)

var (
	ctx = context.Background()
)

type ImageStore struct {
	redisClient *redis.Client
	hashSetName string
}

func GetNewImageStore(appConfig config.AppConfigs) *ImageStore {

	dbName := "imageStore"
	redisConfig := appConfig.Redis
	fmt.Printf("Host: %s, Port: %s, db = %d\n", redisConfig.Host, redisConfig.Port, redisConfig.DBs[dbName])

	_redisClient := storage.GetRedisConnection(dbName, redisConfig)
	imgRedis := &ImageStore{
		redisClient: _redisClient,
		hashSetName: _hashSetName,
	}
	return imgRedis
}

func (imageStore ImageStore) SetPodDetails(podIP string, podDetails models.PodDetails) {
	if err := imageStore.redisClient.Set(ctx, podIP, podDetails, defaultExpiry).Err(); err != nil {
		log.Default().Printf("error in SetPodDetails %v\n", err)
	}
}

func (imageStore ImageStore) SetContainerRuntimes(containerRuntimeObjects []models.ContainerRuntime) error {

	containerRuntimeObjects = imageStore.getOnlyWriteEligibleRuntimeObjects(containerRuntimeObjects)
	if len(containerRuntimeObjects) < 1 {
		return nil
	}
	log.Default().Printf("found %d new containerRuntimeObjects %v", len(containerRuntimeObjects), containerRuntimeObjects)

	// serialize the ContainerRuntime struct to JSON
	valuesToSet := map[string]interface{}{}
	for _, value := range containerRuntimeObjects {
		// serialize the ContainerRuntime struct to JSON
		serialized, err := json.Marshal(value)
		if err != nil {
			return err
		}
		valuesToSet[value.Image] = string(serialized)
	}

	// set the new value
	if err := imageStore.redisClient.HMSet(ctx, imageStore.hashSetName, valuesToSet).Err(); err != nil {
		return err
	}

	// increment the store version
	return imageStore.incrementStoreVersion()
}

func (imageStore ImageStore) GetContainerRuntime(key string) (*models.ContainerRuntime, error) {

	// get the value against the key in hashset
	output := imageStore.redisClient.HGet(ctx, imageStore.hashSetName, key)
	if err := output.Err(); err != nil {
		return nil, err
	}
	value := output.Val()

	// deserialize the string into ContainerRuntime
	var containerRuntime models.ContainerRuntime
	if err := json.Unmarshal([]byte(value), &containerRuntime); err != nil {
		return nil, err
	}

	return &containerRuntime, nil
}

func (imageStore ImageStore) GetContainerRuntimes(keys []string) map[string]*models.ContainerRuntime {
	valuesFromRedis, err := imageStore.redisClient.HMGet(ctx, imageStore.hashSetName, keys...).Result()
	if err != nil {
		return nil
	}
	mapOfContainerRuntimeStrings := map[string]string{}
	for index, key := range keys {
		if valuesFromRedis[index] != nil {
			mapOfContainerRuntimeStrings[key] = valuesFromRedis[index].(string)
		}
	}
	return deserializeContainerRuntimeStrings(mapOfContainerRuntimeStrings)
}

func (imageStore ImageStore) GetAllContainerRuntimes() map[string]*models.ContainerRuntime {
	output := imageStore.redisClient.HGetAll(ctx, imageStore.hashSetName)
	if output.Err() != nil {
		return nil
	}
	return deserializeContainerRuntimeStrings(output.Val())
}

func (imageStore ImageStore) Delete(key string) error {
	return imageStore.redisClient.HDel(ctx, imageStore.hashSetName, key).Err()
}

func (imageStore ImageStore) Length() (int64, error) {
	// get the number of hash key-value pairs
	return imageStore.redisClient.HLen(ctx, imageStore.hashSetName).Result()
}

func (imageStore ImageStore) GetStoreVersion() (int64, error) {
	return imageStore.redisClient.Get(ctx, hashSetVersion).Int64()
}

func (imageStore ImageStore) incrementStoreVersion() error {
	return imageStore.redisClient.IncrBy(ctx, hashSetVersion, 1).Err()
}

func (imageStore ImageStore) getOnlyWriteEligibleRuntimeObjects(containerResultsFromPods []models.ContainerRuntime) []models.ContainerRuntime {

	// 1. get existing value for container runtime from persistent store
	runtimeKeys := make([]string, len(containerResultsFromPods))
	for index, containerResult := range containerResultsFromPods {
		runtimeKeys[index] = containerResult.Image
	}
	containerRuntimeMapFromRedis := imageStore.GetContainerRuntimes(runtimeKeys)

	// Find the diff between the data in imageStore and the data from pods
	diffMapContainerRuntime := []models.ContainerRuntime{}
	for _, containerRuntime := range containerResultsFromPods {

		pushNewValue := false

		// get object from image store
		imgStoreContainerRuntime, ok := containerRuntimeMapFromRedis[containerRuntime.Image]
		if ok {
			// if present, compare if the values are different
			pushNewValue = !imgStoreContainerRuntime.Equals(containerRuntime)
		} else {
			// not found, push the containerRuntime
			pushNewValue = true
		}

		// if the containerRuntime is different push in the `diffMapContainerRuntime`
		if pushNewValue {
			diffMapContainerRuntime = append(diffMapContainerRuntime, containerRuntime)
		}
	}

	return diffMapContainerRuntime
}

func deserializeContainerRuntimeStrings(mapOfContainerRuntimeStrings map[string]string) map[string]*models.ContainerRuntime {
	mapContainerRuntime := map[string]*models.ContainerRuntime{}

	// deserialize the value against each key into `models.ContainerRuntime`
	for key, value := range mapOfContainerRuntimeStrings {
		var containerRuntime models.ContainerRuntime

		if err := json.Unmarshal([]byte(value), &containerRuntime); err != nil {
			fmt.Printf("Unable to unmarshal value for key `%s`. Error: %v", key, err)
			continue
		}
		mapContainerRuntime[key] = &containerRuntime
	}
	return mapContainerRuntime
}
