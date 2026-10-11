package redis_locks

import (
	"context"
	"github.com/redis/go-redis/v9"
	redis_storage "github.com/storage-lock/go-redis-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
	storage_lock_factory "github.com/storage-lock/go-storage-lock-factory"
)

var redisClientStorageLockFactoryBeanFactory *storage_lock_factory.StorageLockFactoryBeanFactory[*redis.Client, *redis.Client] = storage_lock_factory.NewStorageLockFactoryBeanFactory[*redis.Client, *redis.Client]()

func NewRedisLockByClient(ctx context.Context, client *redis.Client, lockId string) (*storage_lock.StorageLock, error) {
	factory, err := GetRedisLockFactoryByClient(ctx, client)
	if err != nil {
		return nil, err
	}
	return factory.CreateLock(lockId)
}

func NewRedisLockByClientWithOptions(ctx context.Context, client *redis.Client, options *storage_lock.StorageLockOptions) (*storage_lock.StorageLock, error) {
	factory, err := GetRedisLockFactoryByClient(ctx, client)
	if err != nil {
		return nil, err
	}
	return factory.CreateLockWithOptions(options)
}

func GetRedisLockFactoryByClient(ctx context.Context, client *redis.Client) (*storage_lock_factory.StorageLockFactory[*redis.Client], error) {
	return redisClientStorageLockFactoryBeanFactory.GetOrInit(ctx, client, func(ctx context.Context) (*storage_lock_factory.StorageLockFactory[*redis.Client], error) {
		connectionManager := redis_storage.NewRedisConnectionManagerFromClient(client)
		options := redis_storage.NewRedisStorageOptions().SetConnectionManager(connectionManager)
		storage, err := redis_storage.NewRedisStorage(ctx, options)
		if err != nil {
			return nil, err
		}
		factory := storage_lock_factory.NewStorageLockFactory(storage, options.ConnectionManager)
		return factory, nil
	})
}
