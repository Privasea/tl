## 此次升级的日志包，比较少的业务改动（请使用最新版本）

- https://github.com/Privasea/tl


#### 引入tl

```
go get github.com/Privasea/tl
import "github.com/Privasea/tl"
```

#### 增加配置

```
"env_api_log": true,//开启apilog，默认开启
"env_db_log": true,//开启dblog，默认开启
"env_dbw_log": true,//开启dblog,写操作，默认开启
"env_dbr_log": true,//开启dblog，读操作,默认开启
"env_debug_log": true,//开启debuglog，默认开启
```

#### 初始化

```
tl.SetParamLog(env_api_log)
tl.SetDbLog(env_db_log)
tl.SetDbWLog(env_db_log)
tl.SetDbRLog(env_db_log)
tl.SetDebugLog(env_db_log)

tl.SetLogger(appName,"dev","file","/yourpath") 
or tl.SetLogger(appName,"dev","console","")

```
### 异常注入开启
```
	tl.SetLogErr(true)
```
#### 替换gin的全局中间件

```
// 新增tl.GinInterceptor中间件（日志打印和链路追踪）
r.Use(tl.GinInterceptor)
```

#### 修改数据库连接和查询（count不能用preload）（需要context必须从最外层传递下来）

```
// 替换原来的数据库连接池
tl.InitConn(数据库名, 连接dsn)

// 获取一个连接
func Orm(ctx *gin.Context) *gorm.DB {
db := tl.WarpMysql(ctx, 数据库名)
return db.WithContext(ctx)
}

Orm(ctx).model().where().Find()
```

#### 修改所有打印日志的地方（按照原则应用不需要打印日志，可以直接删除）（需要context必须从最外层传递下来）（移除原有的glogs）

```
// 把原有的打印替换为tl提供的3个方法
tl.InfoF()
tl.WarnF()
tl.ErrorF()
```
