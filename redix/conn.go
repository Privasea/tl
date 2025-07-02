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
func InitConn(name string, addr string,password string,db int,poolsize int,minidleconns int) {
	if name == "" || addr == "" {
		return
	}

	redisMu.Lock()
	defer redisMu.Unlock()

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,     // no password set
		DB:           db,           // use default DB
		PoolSize:     poolsize,     // 设置连接池大小
		MinIdleConns: minidleconns, // 设置最小空闲连接数
	})

	// 测试连接
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Printf("[app.redix] redis connect fail, err:%s", err)
		panic(err)
	}

	clients[name] = client
	log.Printf("[app.redix] redis success, name: %s", name)
}

// Wrap 获取Redis客户端
func Wrap(ctx context.Context, name string) *redis.Client {
	redisMu.RLock()
	defer redisMu.RUnlock()
	return clients[name]
}