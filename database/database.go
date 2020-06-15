package database

import (
	"context"
	"database/sql"
	"errors"
	"go.shu.run/bootstrap/logger"
	"gorm.io/driver/mysql"
	"sync"
	"time"

	"go.shu.run/bootstrap/config"
	"go.shu.run/bootstrap/utils/strs"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const keyRoot = "/sys/database"
const keyDebug = keyRoot + "/debug"
const keyDSN = keyRoot + "/dsn"

var ErrNotOpen = errors.New("database not open")

var defaultConfig = &gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}, PrepareStmt: true}

func New(cfg config.Loader, log logger.Logger) *DB {
	db := &DB{cfg: cfg, log: log.Prefix("DB")}
	db.Init(context.Background())
	return db
}

type DB struct {
	cfg config.Loader
	log logger.Logger

	dsnBu string
	dsn   string
	debug bool
	sync  bool

	gdb *gorm.DB

	sync.Mutex
}

func (d *DB) Init(ctx context.Context) error {
	d.log.Debugf("数据库初始化")
	//return d.cfg.Watch(ctx, d)
	d.dsn = "root:root@tcp(127.0.0.1:3306)/example"
	return nil
}

func (d *DB) Key() string {
	return keyRoot
}

func (d *DB) OnUpdate(key string, val string) {
	d.log.Debugf("key: %s, val: %s", key, val)
	switch key {
	case keyDebug:
		d.debug = strs.V(val).Bool(d.debug)
	case keyDSN:
		d.Lock()
		defer d.Unlock()
		d.dsnBu = d.dsn
		d.dsn = strs.V(val).String()
	}
}

func (d *DB) GetDB(ctx context.Context) *gorm.DB {
	d.log.Debugf("try open database: ...%s", d.dsn)

	d.Lock()
	for {
		gdb, err := gorm.Open(mysql.Open(d.dsn), defaultConfig)
		if err != nil {
			d.log.Errorf("try open database: %v", err)
			select {
			case <-ctx.Done():
				d.log.Errorf("try open database: %v", ctx.Err())
				return nil
			case <-time.After(time.Second * 3):
				continue
			}
		}

		gdb.Logger = &l{d.log}
		d.gdb = gdb
		d.log.Debugf("database is opened")

		err = d.Ping(ctx)
		d.log.Debugf("database ping: %v", err)
		break
	}
	d.Unlock()

	if d.debug {
		return d.gdb.Debug().WithContext(ctx)
	}
	return d.gdb.WithContext(ctx)
}

func (d *DB) Ping(ctx context.Context) error {
	if d.gdb != nil {
		return ping(ctx, d.gdb)
	}
	return ErrNotOpen
}

// Exec execute raw sql
func (d *DB) Exec(ctx context.Context, sql string, values ...interface{}) *gorm.DB {
	return d.GetDB(ctx).Exec(sql, values)
}

// Transaction start a transaction as a block, return error will rollback, otherwise to commit.
func (d *DB) Transaction(ctx context.Context, fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) (err error) {
	return d.GetDB(ctx).Transaction(fc, opts...)
}

func (d *DB) Raw(ctx context.Context, sql string, values ...interface{}) *gorm.DB {
	return d.GetDB(ctx).Raw(sql, values)
}

func (d *DB) Model(ctx context.Context, model interface{}) *gorm.DB {
	return d.GetDB(ctx).Model(model)
}

func (d *DB) Table(ctx context.Context, table string) *gorm.DB {
	return d.GetDB(ctx).Table(table)
}

func (d *DB) Create(ctx context.Context, model interface{}) *gorm.DB {
	return d.Model(ctx, model).Create(model)
}

func (d *DB) Update(ctx context.Context, model interface{}) *gorm.DB {
	return d.Model(ctx, model).Updates(model)
}

// Begin begins a transaction
func (d *DB) Begin(ctx context.Context, opts ...*sql.TxOptions) *gorm.DB {
	return d.GetDB(ctx).Begin(opts...)
}
