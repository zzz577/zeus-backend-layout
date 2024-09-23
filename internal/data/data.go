package data

import (
	"context"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	"fmt"
	"zeus-backend-layout/internal/biz"
	"zeus-backend-layout/internal/conf"
	"zeus-backend-layout/internal/data/ent"
	"zeus-backend-layout/internal/data/ent/migrate"
	"zeus-backend-layout/internal/data/ent/transaction"

	atlas "ariga.io/atlas/sql/migrate"
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
	return transaction.WithTx(ctx, d.db, func(tx *ent.Tx) error {
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
	client, err := ent.Open(c.Database.Driver, fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=True&loc=Local",
		c.Database.Username,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Database,
	))
	if err != nil {
		log.Errorf("failed opening connection to sqlite: %v", err)
		return nil, nil, err
	}

	ctx := context.Background()
	// Create a local migration directory able to understand Atlas migration file format for replay.
	dir, err := atlas.NewLocalDir("internal/data/ent/migrate/migrations")
	if err != nil {
		log.Errorf("failed creating atlas migration directory: %v", err)
		return nil, nil, err
	}
	// Migrate diff options.
	opts := []schema.MigrateOption{
		schema.WithDir(dir),                         // provide migration directory
		schema.WithMigrationMode(schema.ModeReplay), // provide migration mode
		schema.WithDialect(dialect.MySQL),           // Ent dialect to use
		schema.WithFormatter(atlas.DefaultFormatter),
	}
	// Generate migrations using Atlas support for MySQL (note the Ent dialect option passed above).
	err = migrate.Diff(ctx, fmt.Sprintf("%s://%s:%s@%s:%d/%s",
		c.Database.Driver,
		c.Database.Username,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Database), opts...)
	if err != nil {
		log.Errorf("failed generating migration file: %v", err)
		return nil, nil, err
	}

	return &Data{
		db: client,
	}, cleanup, nil
}
