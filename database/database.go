package database

import (
	"context"
	"database/sql"
	"errors"

	"go.shu.run/bootstrap/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var ErrNotOpen = errors.New("database not open")

var defaultConfig = &gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}, PrepareStmt: true}

func New(cfg Config, log logger.Logger) (*DB, error) {
	log = log.Prefix("DB")
	log.Debugf("dsn: %s", cfg)
	gdb, err := gorm.Open(mysql.Open(cfg.String()), defaultConfig)
	if err != nil {
		return nil, err
	}
	gdb.Logger = &l{Logger: log}
	return &DB{
		log:   log.Prefix("DB"),
		gdb:   gdb,
		debug: cfg.Debug,
	}, err
}

type DB struct {
	log   logger.Logger
	gdb   *gorm.DB
	debug bool
}

func (d *DB) GetDB(ctx context.Context) *gorm.DB {
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
