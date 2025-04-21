package dbx

import (
	"context"
	"facebyte/config"
	"facebyte/pkg/tl/injection"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
	"strings"
	"time"
)

const (
	contextKey    = "otgorm:context"
	tSpanName     = "mysql"
	componentName = "gorm:"
	dbWriteTime   = 3 * time.Second
)

func Wrap(ctx *gin.Context, dbName ...string) *gorm.DB {
	var db *gorm.DB
	if len(dbName) > 0 {
		db = getGormDB(dbName[0])
		ctx.Set("gorm:database", dbName[0])
	} else {
		db = getGormDB(defaultDbName)
		ctx.Set("gorm:database", defaultDatabase)
	}

	return db.Set(contextKey, ctx).WithContext(ctx)
}

func registerCallbacks(db *gorm.DB) {
	prefix := db.Dialector.Name() + ":"

	db.Callback().Create().Before("gorm:begin_transaction").Register("aotgorm_before_create", newBefore(prefix+"create"))
	db.Callback().Create().After("gorm:commit_or_rollback_transaction").Register("otgorm_after_create", newAfter())

	db.Callback().Update().Before("gorm:begin_transaction").Register("otgorm_before_update", newBefore(prefix+"update"))
	db.Callback().Update().After("gorm:commit_or_rollback_transaction").Register("otgorm_after_update", newAfter())

	db.Callback().Query().Before("gorm:query").Register("otgorm_before_query", newBefore(prefix+"query"))
	db.Callback().Query().After("gorm:after_query").Register("otgorm_after_query", newAfter())

	db.Callback().Delete().Before("gorm:begin_transaction").Register("otgorm_before_delete", newBefore(prefix+"delete"))
	db.Callback().Delete().After("gorm:commit_or_rollback_transaction").Register("otgorm_after_delete", newAfter())

	db.Callback().Row().Before("gorm:row").Register("otgorm_before_row", newBefore(prefix+"row"))
	db.Callback().Row().After("gorm:row").Register("otgorm_after_row", newAfter())

	db.Callback().Raw().Before("gorm:raw").Register("otgorm_before_raw", newBefore(prefix+"raw"))
	db.Callback().Raw().After("gorm:raw").Register("otgorm_after_raw", newAfter())
}
func registerBeforeCallbacks(db *gorm.DB) {
	db.Callback().Create().Before("gorm:create").Register("before_create", beforeSql)
	db.Callback().Update().Before("gorm:update").Register("before_update", beforeSql)
	db.Callback().Delete().Before("gorm:delete").Register("before_delete", beforeSql)
	db.Callback().Query().Before("gorm:query").Register("before_query", beforeSql)

}
func beforeSql(tx *gorm.DB) {

	if ctx, ok := tx.Statement.Context.(*gin.Context); ok {
		//start
		if ctx.Request == nil {
			return
		}
		imei := ctx.GetHeader("Imei")
		if imei != "" {
			injection.TrackingPoints(ctx)
			var point = injection.GetTrackingPoint(ctx)
			inErr, _ := injection.DealPoints(config.GetBase().AppName, ctx.Request.URL.Path, ctx.Request.Method, imei, point)
			if inErr == "ErrRecordNotFound" {
				tx.AddError(gorm.ErrRecordNotFound)
			}
			if inErr == "ErrInvalidTransaction" {
				tx.AddError(gorm.ErrInvalidTransaction)
			}
		}

		//end

	}

}

// 记录主库写入时间，在查询的时候动态选择主库或从库
func registerReplicaCallbacks(db *gorm.DB) {
	db.Callback().Create().After("gorm:create").Register("record_write_time", recordWriteTime)
	db.Callback().Update().After("gorm:update").Register("record_write_time", recordWriteTime)
	db.Callback().Delete().After("gorm:delete").Register("record_write_time", recordWriteTime)
	db.Callback().Raw().After("gorm:raw").Register("record_write_time", recordRawWriteTime)
	db.Callback().Query().Before("gorm:query").Register("dynamic_read_write_clauses", dynamicReadWriteClauses)
	db.Callback().Row().Before("gorm:row").Register("dynamic_read_write_clauses", dynamicReadWriteClauses)
}

// 记录执行写入的时间
func recordWriteTime(db *gorm.DB) {
	ctx, ok := db.Statement.Context.(*gin.Context)
	if ok {
		ctx.Set("DBWriteTime", time.Now())
	}
}

// 记录执行写入的时间
func recordRawWriteTime(db *gorm.DB) {
	ctx, ok := db.Statement.Context.(*gin.Context)
	if ok {
		sql := db.Statement.SQL.String()
		if len(sql) >= 6 {
			prefix := strings.ToLower(sql[0:6])
			if prefix == "insert" || prefix == "update" || prefix == "delete" {
				ctx.Set("DBWriteTime", time.Now())
			}
		}
	}
}

// 动态选择使用主库还是从库
// 如果当前请求在 x 秒内刚执行过写入操作，则强制使用主库进行查询操作
func dynamicReadWriteClauses(db *gorm.DB) {
	ctx, ok := db.Statement.Context.(*gin.Context)
	if !ok {
		return
	}
	value, ok2 := ctx.Get("DBWriteTime")
	if !ok2 {
		return
	}
	writeTime, ok3 := value.(time.Time)
	if ok3 && time.Now().Sub(writeTime) <= dbWriteTime {
		db = db.Clauses(dbresolver.Write)
	}
}

func newBefore(name string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		//if v, ok := db.Get(contextKey); ok {
		//	ctx := v.(*gin.Context)
		//	if ctx != nil {
		//		span := ctx.SpanStart(tSpanName)
		//		if nil != span {
		//			newCtx := context.Background()
		//			keepScene(db, newCtx)
		//			ext.Component.Set(span, componentName+name)
		//			setSpan(db, newCtx, span)
		//		}
		//	}
		//}
	}
}

func newAfter() func(*gorm.DB) {
	return func(db *gorm.DB) {
		//span, _ := getSpan(db)
		//if nil != span {
		//	defer func() {
		//		span.Finish()
		//		restoreScene(db)
		//	}()
		//	ext.DBStatement.Set(span, db.Statement.SQL.String())
		//	if db.Error != nil {
		//		if !errors.Is(db.Error, gorm.ErrRecordNotFound) && !errors.Is(db.Error, sql.ErrNoRows) {
		//			ext.LogError(span, db.Error)
		//		}
		//	}
		//}
	}
}

func setSpan(db *gorm.DB, ctx context.Context, span opentracing.Span) {
	db.Set(contextKey, opentracing.ContextWithSpan(ctx, span))
}

func getSpan(db *gorm.DB) (opentracing.Span, context.Context) {
	if v, ok := db.Get(contextKey); ok {
		if ctx, okIf := v.(context.Context); okIf {
			return opentracing.SpanFromContext(ctx), ctx
		}
	}
	return nil, nil
}

const contextSceneKey = "otgorm:context:scene:" + "v1.0.0"

func keepScene(db *gorm.DB, ctx context.Context) {
	db.Set(contextSceneKey, ctx)
}

func restoreScene(db *gorm.DB) {
	if v, ok := db.Get(contextSceneKey); ok {
		db.Set(contextKey, v)
	}
}
