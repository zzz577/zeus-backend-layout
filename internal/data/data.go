package data

import (
	"context"
	"zeus-backend-layout/internal/biz"
	"zeus-backend-layout/internal/conf"
	"zeus-backend-layout/internal/data/ent/ent"

	"github.com/go-kratos/kratos/v2/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewTransaction, NewGreeterRepo)

// Data .
type Data struct {
	db *ent.Client
}

type contextTxKey struct{}

func (d *Data) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return ent.WithTx(ctx, d.db, func(tx *ent.Tx) error {
		ctx = context.WithValue(ctx, contextTxKey{}, tx)
		return fn(ctx)
	})
}

func (d *Data) DB(ctx context.Context) *ent.Client {
	tx, ok := ctx.Value(contextTxKey{}).(*ent.Client)
	if ok {
		return tx
	}
	return d.db
}

// NewTransaction .
func NewTransaction(d *Data) biz.Transaction {
	return d
}

// NewData .
func NewData(c *conf.Data) (*Data, func(), error) {
	cleanup := func() {
		log.Info("closing the data resources")
	}
	client, err := ent.Open("mysql", "root:xxxxx@tcp(127.0.0.1:3306)/test?parseTime=True")
	if err != nil {
		log.Errorf("failed opening connection to sqlite: %v", err)
		return nil, nil, err
	}
	return &Data{
		db: client,
	}, cleanup, nil
}
