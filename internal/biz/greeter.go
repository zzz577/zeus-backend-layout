package biz

import (
	"context"

	v1 "zeus-backend-layout/api/helloworld/v1"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

var (
	// ErrUserNotFound is user not found.
	ErrUserNotFound = errors.NotFound(v1.ErrorReason_USER_NOT_FOUND.String(), "user not found")
)

// Greeter is a Greeter model.
type Greeter struct {
	Hello string
}

// GreeterRepo is a Greater repo.
type GreeterRepo interface {
	Save(context.Context, *Greeter) (*Greeter, error)
	Update(context.Context, *Greeter) (*Greeter, error)
	FindByID(context.Context, int64) (*Greeter, error)
	ListByHello(context.Context, string) ([]*Greeter, error)
	ListAll(context.Context) ([]*Greeter, error)
}

// GreeterUsecase is a Greeter usecase.
type GreeterUsecase struct {
	repo GreeterRepo
	tm   Transaction
}

// NewGreeterUsecase new a Greeter usecase.
func NewGreeterUsecase(repo GreeterRepo, tm Transaction) *GreeterUsecase {
	return &GreeterUsecase{repo: repo, tm: tm}
}

// CreateGreeter creates a Greeter, and returns the new Greeter.
func (uc *GreeterUsecase) CreateGreeter(ctx context.Context, g *Greeter) (*Greeter, error) {
	log.Infof("CreateGreeter: %v", g.Hello)
	var save *Greeter
	var err error
	err = uc.tm.InTx(ctx, func(ctx context.Context) error {
		save, err = uc.repo.Save(ctx, g)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Error(err)
	}
	return save, err
}
