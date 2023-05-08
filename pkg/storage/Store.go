package storage

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"time"
	"zerok-deamonset/internal/config"
	"zerok-deamonset/internal/models"
)

const (
	defaultExpiry    time.Duration = time.Hour * 24 * 30
	hashTableName    string        = "zk_img_proc_map"
	hashTableVersion string        = "zk_img_proc_version"
)

type ImageStore struct {
	redisClient   *redis.Client
	hashTableName string
}

func GetNewImageStore(appConfig config.AppConfigs) *ImageStore {

	redisConfig := appConfig.Redis
	readTimeout := time.Duration(redisConfig.ReadTimeout) * time.Second
	_redisClient := redis.NewClient(&redis.Options{
		Addr:        fmt.Sprint(redisConfig.Host, ":", redisConfig.Port),
		Password:    "",
		DB:          7,
		ReadTimeout: readTimeout,
	})

	//_redisClient.Expire(hashTableName, defaultExpiry)

	imgRedis := &ImageStore{
		redisClient:   _redisClient,
		hashTableName: hashTableName,
	}

	//if _redisClient.Del(hashTableName).Err() != nil {
	//	fmt.Println("couldn't delete hashtable " + hashTableName)
	//}

	return imgRedis
}

func (zkRedis ImageStore) SetContainerRuntime(key string, value models.ContainerRuntime) error {

	// serialize the ContainerRuntime struct to JSON
	serialized, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// set the new value
	if err := zkRedis.redisClient.HSet(zkRedis.hashTableName, key, string(serialized)).Err(); err != nil {
		return err
	}

	// increment the store version
	return zkRedis.incrementStoreVersion()
}

func (zkRedis ImageStore) SetContainerRuntimes(containerRuntimeObjects []models.ContainerRuntime) error {

	// serialize the ContainerRuntime struct to JSON
	valuesToSet := map[string]interface{}{}
	for _, value := range containerRuntimeObjects {
		// serialize the ContainerRuntime struct to JSON
		serialized, err := json.Marshal(value)
		if err != nil {
			return err
		}
		valuesToSet[value.Image] = serialized
	}

	// set the new value
	if err := zkRedis.redisClient.HMSet(zkRedis.hashTableName, valuesToSet).Err(); err != nil {
		return err
	}

	// increment the store version
	return zkRedis.incrementStoreVersion()
}

func (zkRedis ImageStore) GetContainerRuntime(key string) (*models.ContainerRuntime, error) {

	// get the value against the key in hashset
	output := zkRedis.redisClient.HGet(zkRedis.hashTableName, key)
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

func (zkRedis ImageStore) GetAllContainerRuntimes() map[string]*models.ContainerRuntime {
	mapContainerRuntime := map[string]*models.ContainerRuntime{}

	output := zkRedis.redisClient.HGetAll(zkRedis.hashTableName)
	err := output.Err()
	if err != nil {
		return mapContainerRuntime
	}
	mapOfContainerRuntimeStrings := output.Val()

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

func (zkRedis ImageStore) Delete(key string) error {
	return zkRedis.redisClient.HDel(zkRedis.hashTableName, key).Err()
}

func (zkRedis ImageStore) Length() (int64, error) {
	// get the number of hash key-value pairs
	return zkRedis.redisClient.HLen(zkRedis.hashTableName).Result()
}

func (zkRedis ImageStore) GetStoreVersion() (int64, error) {
	return zkRedis.redisClient.Get(hashTableVersion).Int64()
}

func (zkRedis ImageStore) incrementStoreVersion() error {
	return zkRedis.redisClient.IncrBy(hashTableVersion, 1).Err()
}
