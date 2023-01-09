package services

import (
	"context"
	"fmt"
	"log"

	"github.com/saki-engineering/graphql-sample/graph/db"
	"github.com/saki-engineering/graphql-sample/graph/model"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type projectService struct {
	exec boil.ContextExecutor
}

func convertProjectV2(project *db.Project) *model.ProjectV2 {
	projectURL, err := model.UnmarshalURI(project.URL)
	if err != nil {
		log.Println("invalid URI", project.URL)
	}

	return &model.ProjectV2{
		ID:     project.ID,
		Title:  project.Title,
		Number: int(project.Number),
		URL:    projectURL,
		Owner:  &model.User{ID: project.Owner},
	}
}

func convertProjectV2Connection(projects db.ProjectSlice, hasPrevPage, hasNextPage bool) *model.ProjectV2Connection {
	var result model.ProjectV2Connection

	for _, dbp := range projects {
		project := convertProjectV2(dbp)

		result.Edges = append(result.Edges, &model.ProjectV2Edge{Cursor: project.ID, Node: project})
		result.Nodes = append(result.Nodes, project)
	}
	result.TotalCount = len(projects)

	result.PageInfo = &model.PageInfo{}
	if result.TotalCount != 0 {
		result.PageInfo.StartCursor = &result.Nodes[0].ID
		result.PageInfo.EndCursor = &result.Nodes[result.TotalCount-1].ID
	}
	result.PageInfo.HasPreviousPage = hasPrevPage
	result.PageInfo.HasNextPage = hasNextPage

	return &result
}

func (p *projectService) GetProjectByID(ctx context.Context, id string) (*model.ProjectV2, error) {
	project, err := db.FindProject(ctx, p.exec, id,
		db.ProjectColumns.ID,
		db.ProjectColumns.Title,
		db.ProjectColumns.Number,
		db.ProjectColumns.URL,
		db.ProjectColumns.Owner,
	)
	if err != nil {
		return nil, err
	}
	return convertProjectV2(project), nil
}

func (p *projectService) GetProjectByOwnerAndNumber(ctx context.Context, ownerID string, number int) (*model.ProjectV2, error) {
	project, err := db.Projects(
		qm.Select(
			db.ProjectColumns.ID,
			db.ProjectColumns.Title,
			db.ProjectColumns.Number,
			db.ProjectColumns.URL,
			db.ProjectColumns.Owner,
		),
		db.ProjectWhere.Owner.EQ(ownerID),
		db.ProjectWhere.Number.EQ(int64(number)),
	).One(ctx, p.exec)
	if err != nil {
		return nil, err
	}
	return convertProjectV2(project), nil
}

func (p *projectService) ListProjectByOwner(ctx context.Context, ownerID string, after *string, before *string, first *int, last *int) (*model.ProjectV2Connection, error) {
	cond := []qm.QueryMod{
		qm.Select(
			db.ProjectColumns.ID,
			db.ProjectColumns.Title,
			db.ProjectColumns.Number,
			db.ProjectColumns.URL,
			db.ProjectColumns.Owner,
		),
		db.ProjectWhere.Owner.EQ(ownerID),
	}
	var scanDesc bool

	switch {
	case (after != nil) && (before != nil):
		cond = append(cond, db.ProjectWhere.ID.GT(*after), db.ProjectWhere.ID.LT(*before))
	case after != nil:
		cond = append(cond,
			db.ProjectWhere.ID.GT(*after),
			qm.OrderBy(fmt.Sprintf("%s asc", db.ProjectColumns.ID)),
		)
		if first != nil {
			cond = append(cond, qm.Limit(*first))
		}
	case before != nil:
		scanDesc = true
		cond = append(cond,
			db.ProjectWhere.ID.LT(*before),
			qm.OrderBy(fmt.Sprintf("%s desc", db.ProjectColumns.ID)),
		)
		if last != nil {
			cond = append(cond, qm.Limit(*last))
		}
	default:
		switch {
		case last != nil:
			scanDesc = true
			cond = append(cond,
				qm.OrderBy(fmt.Sprintf("%s desc", db.ProjectColumns.ID)),
				qm.Limit(*last),
			)
		case first != nil:
			cond = append(cond,
				qm.OrderBy(fmt.Sprintf("%s asc", db.ProjectColumns.ID)),
				qm.Limit(*first),
			)
		default:
			cond = append(cond,
				qm.OrderBy(fmt.Sprintf("%s asc", db.ProjectColumns.ID)),
			)
		}
	}

	projects, err := db.Projects(cond...).All(ctx, p.exec)
	if err != nil {
		return nil, err
	}

	var hasNextPage, hasPrevPage bool
	if len(projects) != 0 {
		if scanDesc {
			for i, j := 0, len(projects)-1; i < j; i, j = i+1, j-1 {
				projects[i], projects[j] = projects[j], projects[i]
			}
		}
		startCursor, endCursor := projects[0].ID, projects[len(projects)-1].ID

		var err error
		hasPrevPage, err = db.Projects(
			db.ProjectWhere.Owner.EQ(ownerID),
			db.ProjectWhere.ID.LT(startCursor),
		).Exists(ctx, p.exec)
		if err != nil {
			return nil, err
		}
		hasNextPage, err = db.Projects(
			db.ProjectWhere.Owner.EQ(ownerID),
			db.ProjectWhere.ID.GT(endCursor),
		).Exists(ctx, p.exec)
		if err != nil {
			return nil, err
		}
	}

	return convertProjectV2Connection(projects, hasPrevPage, hasNextPage), nil
}
