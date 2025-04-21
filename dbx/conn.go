package dbx

import (
	"facebyte/pkg/tl/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"log"
	"runtime"
	"time"
)

var defaultDatabase = "facebyte"

var defaultPoolMaxOpen = runtime.NumCPU()*2 + 5 // 连接池最大连接数量4c*2+4只读副本+1主实例

const (
	defaultPoolMaxIdle     = 2                                 // 连接池空闲连接数量
	defaultConnMaxLifeTime = time.Second * time.Duration(7200) // MySQL默认长连接时间为8个小时,可根据高并发业务持续时间合理设置该值
	defaultConnMaxIdleTime = time.Second * time.Duration(600)  // 设置连接10分钟没有用到就断开连接(内存要求较高可降低该值)
	levelInfo              = "info"
	LevelWarn              = "warn"
	LevelError             = "error"
)

type dbPoolCfg struct {
	maxIdleConn int64 //空闲连接数
	maxOpenConn int64 //最大连接数
	maxLifeTime int64 //连接可重用的最大时间
	maxIdleTime int64 //在关闭连接之前, 连接可能处于空闲状态的最大时间
}

type dbConfig struct {
	name       string
	dsn        string
	replicaDsn string
	logLevel   string
	poolCfg    *dbPoolCfg
	gormCfg    *gorm.Config
}

// initDB init db
func initDB(cfg dbConfig) *gorm.DB {
	var err error

	var level logger.LogLevel
	switch cfg.logLevel {
	case levelInfo:
		level = logger.Info
	case LevelWarn:
		level = logger.Warn
	case LevelError:
		level = logger.Error
	default:
		level = logger.Info
	}

	if cfg.gormCfg == nil {
		cfg.gormCfg = &gorm.Config{
			Logger: logx.Default(level),
		}
	} else {
		if cfg.gormCfg.Logger == nil {
			cfg.gormCfg.Logger = logx.Default(level)
		}
	}

	Db, err := gorm.Open(mysql.Open(cfg.dsn), cfg.gormCfg)
	if err != nil {
		log.Printf("[app.dbx] mysql open fail, err:%s", err)
		panic(err)
	}
	if cfg.replicaDsn != "" {
		err = Db.Use(dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{mysql.Open(cfg.replicaDsn)},
		}))
		if err != nil {
			log.Printf("[app.dbx] mysql replica open fail, err:%s", err)
			panic(err)
		}
		registerReplicaCallbacks(Db)
	}

	cfg.setDefaultPoolConfig(Db)
	//registerCallbacks(Db)
	registerBeforeCallbacks(Db)

	err = DbSurvive(Db)
	if err != nil {
		log.Printf("[app.dbx] mysql survive fail, err:%s", err)
		panic(err)
	}

	log.Printf("[app.dbx] mysql success, name: %s", cfg.name)
	return Db
}

// DbSurvive mysql survive
func DbSurvive(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	err = sqlDB.Ping()
	if err != nil {
		return err
	}
	return nil
}

func (c *dbConfig) setDefaultPoolConfig(db *gorm.DB) {
	d, err := db.DB()
	if err != nil {
		log.Printf("[app.dbx] mysql db fail, err: %s", err.Error())
		panic(err)
	}
	var cfg = c.poolCfg
	if cfg == nil {
		d.SetMaxOpenConns(defaultPoolMaxOpen)
		d.SetMaxIdleConns(defaultPoolMaxIdle)
		d.SetConnMaxLifetime(defaultConnMaxLifeTime)
		d.SetConnMaxIdleTime(defaultConnMaxIdleTime)
		return
	}

	if cfg.maxOpenConn == 0 {
		d.SetMaxOpenConns(defaultPoolMaxOpen)
	} else {
		d.SetMaxOpenConns(int(cfg.maxOpenConn))
	}

	if cfg.maxIdleConn == 0 {
		d.SetMaxIdleConns(defaultPoolMaxIdle)
	} else {
		d.SetMaxIdleConns(int(cfg.maxIdleConn))
	}

	if cfg.maxLifeTime == 0 {
		d.SetConnMaxLifetime(defaultConnMaxLifeTime)
	} else {
		d.SetConnMaxLifetime(time.Second * time.Duration(cfg.maxLifeTime))
	}

	if cfg.maxIdleTime == 0 {
		d.SetConnMaxIdleTime(defaultConnMaxIdleTime)
	} else {
		d.SetConnMaxIdleTime(time.Second * time.Duration(cfg.maxIdleTime))
	}
}

func InitConn(name string, dsn string, opts ...interface{}) {
	if name == "" || dsn == "" {
		return
	}
	cfg := dbConfig{
		name:     name,
		dsn:      dsn,
		poolCfg:  &dbPoolCfg{},
		logLevel: levelInfo,
	}
	if len(opts) >= 1 {
		replicaDsn, ok := opts[0].(string)
		if ok && replicaDsn != "" {
			cfg.replicaDsn = replicaDsn
		}
		gormCfg, ok2 := opts[len(opts)-1].(*gorm.Config)
		if ok2 {
			cfg.gormCfg = gormCfg
		}
	}
	setGormDB(name, initDB(cfg))
}
