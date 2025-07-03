package redix

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	"sync"
)

var (
	redisMu sync.RWMutex
	clients = make(map[string]*redis.Client)
)

// InitConn 初始化Redis连接
func InitConn(name string, redisClient *redis.Client) {
	if name == ""{
		return
	}
	clients[name] = redisClient
	log.Printf("[app.redix] redis success, name: %s", name)
}

// Wrap 获取Redis客户端
func Wrap(ctx context.Context, name string) *redis.Client {
	redisMu.RLock()
	defer redisMu.RUnlock()
	return clients[name]
}