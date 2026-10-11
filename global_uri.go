package redis_locks

import (
	"context"
	"github.com/redis/go-redis/v9"
	redis_storage "github.com/storage-lock/go-redis-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
	storage_lock_factory "github.com/storage-lock/go-storage-lock-factory"
)

var redisURIStorageLockFactoryBeanFactory *storage_lock_factory.StorageLockFactoryBeanFactory[string, *redis.Client] = storage_lock_factory.NewStorageLockFactoryBeanFactory[string, *redis.Client]()

func NewRedisLockByURI(ctx context.Context, uri string, lockId string) (*storage_lock.StorageLock, error) {
	factory, err := GetRedisLockFactoryByURI(ctx, uri)
	if err != nil {
		return nil, err
	}
	return factory.CreateLock(lockId)
}

func NewRedisLockByURIWithOptions(ctx context.Context, uri string, options *storage_lock.StorageLockOptions) (*storage_lock.StorageLock, error) {
	factory, err := GetRedisLockFactoryByURI(ctx, uri)
	if err != nil {
		return nil, err
	}
	return factory.CreateLockWithOptions(options)
}

func GetRedisLockFactoryByURI(ctx context.Context, uri string) (*storage_lock_factory.StorageLockFactory[*redis.Client], error) {
	return redisURIStorageLockFactoryBeanFactory.GetOrInit(ctx, uri, func(ctx context.Context) (*storage_lock_factory.StorageLockFactory[*redis.Client], error) {
		connectionManager := redis_storage.NewRedisConnectionManagerFromURI(uri)
		options := redis_storage.NewRedisStorageOptions().SetConnectionManager(connectionManager)
		storage, err := redis_storage.NewRedisStorage(ctx, options)
		if err != nil {
			return nil, err
		}
		factory := storage_lock_factory.NewStorageLockFactory(storage, options.ConnectionManager)
		return factory, nil
	})
}
