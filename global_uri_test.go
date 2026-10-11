package redis_locks

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"testing"
	"time"
)

var redisServerCmd2 *exec.Cmd

func TestMain(m *testing.M) {
	envName := "STORAGE_LOCK_REDIS_URI"
	uri := os.Getenv(envName)
	if uri != "" {
		os.Exit(m.Run())
	}

	// 尝试启动临时 redis-server
	redisPort := "6399"
	redisServerCmd2 = exec.Command("redis-server",
		"--port", redisPort,
		"--save", "",
		"--appendonly", "no",
		"--daemonize", "no",
	)
	redisServerCmd2.Stdout = os.Stderr
	redisServerCmd2.Stderr = os.Stderr
	if err := redisServerCmd2.Start(); err != nil {
		os.Exit(0)
	}

	uri = "redis://localhost:" + redisPort
	redisCliCmd := exec.Command("redis-cli", "-p", redisPort, "PING")
	for i := 0; i < 30; i++ {
		if output, err := redisCliCmd.Output(); err == nil && len(output) > 0 {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	os.Setenv(envName, uri)
	exitCode := m.Run()

	if redisServerCmd2 != nil && redisServerCmd2.Process != nil {
		_ = redisServerCmd2.Process.Kill()
		_ = redisServerCmd2.Wait()
	}

	os.Exit(exitCode)
}

func TestNewRedisLockByURI(t *testing.T) {
	envName := "STORAGE_LOCK_REDIS_URI"
	uri := os.Getenv(envName)
	if uri == "" {
		t.Skip("STORAGE_LOCK_REDIS_URI not set and redis-server not available")
	}

	ctx := context.Background()
	lockId := "test-redis-lock-uri"

	factory, err := GetRedisLockFactoryByURI(ctx, uri)
	assert.Nil(t, err)
	assert.NotNil(t, factory)

	lock, err := NewRedisLockByURI(ctx, uri, lockId)
	assert.Nil(t, err)
	assert.NotNil(t, lock)

	// 加锁
	ownerId := "test-owner"
	err = lock.Lock(ctx, ownerId)
	assert.Nil(t, err)

	// 解锁
	err = lock.UnLock(ctx, ownerId)
	assert.Nil(t, err)
}
