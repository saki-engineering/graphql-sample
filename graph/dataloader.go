package graph

import (
	"context"

	"github.com/saki-engineering/graphql-sample/graph/model"
	"github.com/saki-engineering/graphql-sample/graph/services"

	"github.com/graph-gophers/dataloader/v7"
)

type Loaders struct {
	// id(string)をwhere句に使ってUserを探すため
	UserLoader dataloader.Interface[string, *model.User]
}

func NewLoaders(Srv services.Services) *Loaders {
	userLoader := &UserLoader{Srv: Srv}

	return &Loaders{
		// dataloader.Loader[string, *model.User]型
		UserLoader: dataloader.NewBatchedLoader[string, *model.User](userLoader.BatchGetUsers),
	}
}

type UserLoader struct {
	Srv services.Services
}

// github.com/graph-gophers/dataloader/v7 の type BatchFunc[K, V]を満たすため
// dataloader.NewBatchedLoader関数の引数にできる
func (u *UserLoader) BatchGetUsers(ctx context.Context, IDs []string) []*dataloader.Result[*model.User] {
	// 引数と戻り値のスライスlenは等しくする
	result := make([]*dataloader.Result[*model.User], 0, len(IDs))

	users, err := u.Srv.ListUsersByID(ctx, IDs)
	for _, user := range users {
		if err != nil {
			result = append(result, &dataloader.Result[*model.User]{
				Error: err,
			})
		} else {
			result = append(result, &dataloader.Result[*model.User]{
				Data: user,
			})
		}
	}
	return result
}
