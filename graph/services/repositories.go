package services

import (
	"context"

	"github.com/saki-engineering/graphql-sample/graph/db"
	"github.com/saki-engineering/graphql-sample/graph/model"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type repoService struct {
	exec boil.ContextExecutor
}

func convertRepository(repo *db.Repository) *model.Repository {
	return &model.Repository{
		ID:        repo.ID,
		Owner:     &model.User{ID: repo.Owner},
		Name:      repo.Name,
		CreatedAt: repo.CreatedAt,
	}
}

func (r *repoService) GetRepoByID(ctx context.Context, id string) (*model.Repository, error) {
	repo, err := db.FindRepository(ctx, r.exec, id,
		db.RepositoryColumns.ID, db.RepositoryColumns.Name, db.RepositoryColumns.Owner, db.RepositoryColumns.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return convertRepository(repo), nil
}

func (r *repoService) GetRepoByFullName(ctx context.Context, owner, name string) (*model.Repository, error) {
	repo, err := db.Repositories(
		qm.Select(
			db.RepositoryColumns.ID,
			db.RepositoryColumns.Name,
			db.RepositoryColumns.Owner,
			db.RepositoryColumns.CreatedAt,
		),
		db.RepositoryWhere.Owner.EQ(owner),
		db.RepositoryWhere.Name.EQ(name),
	).One(ctx, r.exec)
	if err != nil {
		return nil, err
	}
	return convertRepository(repo), nil
}
