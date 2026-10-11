package redis_locks

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	storage_lock "github.com/storage-lock/go-storage-lock"
	"github.com/stretchr/testify/assert"
)

// testRedisClient 用 TestMain 起好的 redis（或环境变量）建直连 client
func testRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	uri := os.Getenv("STORAGE_LOCK_REDIS_URI")
	if uri == "" {
		t.Skip("STORAGE_LOCK_REDIS_URI not set, skipping")
	}
	opts, err := redis.ParseURL(uri)
	assert.Nil(t, err)
	return redis.NewClient(opts)
}

func TestNewRedisLockByClient(t *testing.T) {
	client := testRedisClient(t)
	defer client.Close()

	ctx := context.Background()
	factory, err := GetRedisLockFactoryByClient(ctx, client)
	assert.Nil(t, err)
	assert.NotNil(t, factory)

	// 缓存命中
	factory2, err := GetRedisLockFactoryByClient(ctx, client)
	assert.Nil(t, err)
	assert.Equal(t, factory, factory2)

	lock, err := NewRedisLockByClient(ctx, client, "test-redis-lock-client")
	assert.Nil(t, err)
	assert.NotNil(t, lock)

	assert.Nil(t, lock.Lock(ctx, "owner-1"))
	assert.Nil(t, lock.UnLock(ctx, "owner-1"))
}

func TestNewRedisLockByClientWithOptions(t *testing.T) {
	client := testRedisClient(t)
	defer client.Close()

	options := storage_lock.NewStorageLockOptionsWithLockId("test-redis-lock-client-opt")
	lock, err := NewRedisLockByClientWithOptions(context.Background(), client, options)
	assert.Nil(t, err)
	assert.NotNil(t, lock)
}

func TestRedisLockByClient_DeadServer(t *testing.T) {
	// 指向不存在服务的 client：Take 直接返回 preset，但 Init 的 Ping 失败
	dead := redis.NewClient(&redis.Options{Addr: "127.0.0.1:19999"})
	defer dead.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := GetRedisLockFactoryByClient(ctx, dead)
	assert.NotNil(t, err)
	_, err = NewRedisLockByClient(ctx, dead, "x")
	assert.NotNil(t, err)
	_, err = NewRedisLockByClientWithOptions(ctx, dead, storage_lock.NewStorageLockOptionsWithLockId("x"))
	assert.NotNil(t, err)
}

func TestRedisLockByURI_Errors(t *testing.T) {
	// 非法 URI：ParseURL 失败
	_, err := GetRedisLockFactoryByURI(context.Background(), "://bad-uri")
	assert.NotNil(t, err)
	_, err = NewRedisLockByURI(context.Background(), "://bad-uri", "x")
	assert.NotNil(t, err)
	_, err = NewRedisLockByURIWithOptions(context.Background(), "://bad-uri", storage_lock.NewStorageLockOptionsWithLockId("x"))
	assert.NotNil(t, err)

	// 格式合法但连不上：Ping 失败
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err = GetRedisLockFactoryByURI(ctx, "redis://127.0.0.1:19999")
	assert.NotNil(t, err)
}

func TestNewRedisLockByURIWithOptions(t *testing.T) {
	uri := os.Getenv("STORAGE_LOCK_REDIS_URI")
	if uri == "" {
		t.Skip("STORAGE_LOCK_REDIS_URI not set, skipping")
	}
	options := storage_lock.NewStorageLockOptionsWithLockId("test-redis-lock-uri-opt")
	lock, err := NewRedisLockByURIWithOptions(context.Background(), uri, options)
	assert.Nil(t, err)
	assert.NotNil(t, lock)
}
