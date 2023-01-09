package services

import (
	"context"
	"fmt"

	"github.com/saki-engineering/graphql-sample/graph/db"
	"github.com/saki-engineering/graphql-sample/graph/model"

	"github.com/google/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type projectItemService struct {
	exec boil.ContextExecutor
}

func convertProjectV2Item(item *db.Projectcard) *model.ProjectV2Item {
	result := &model.ProjectV2Item{
		ID:      item.ID,
		Project: &model.ProjectV2{ID: item.Project},
	}
	if item.Issue.Valid {
		result.Content = &model.Issue{ID: item.Issue.String}
	}
	if item.Pullrequest.Valid {
		result.Content = &model.PullRequest{ID: item.Pullrequest.String}
	}
	return result
}

func convertProjectV2ItemConnection(items db.ProjectcardSlice, hasPrevPage, hasNextPage bool) *model.ProjectV2ItemConnection {
	var result model.ProjectV2ItemConnection

	for _, dbi := range items {
		item := convertProjectV2Item(dbi)

		result.Edges = append(result.Edges, &model.ProjectV2ItemEdge{Cursor: item.ID, Node: item})
		result.Nodes = append(result.Nodes, item)
	}
	result.TotalCount = len(items)

	result.PageInfo = &model.PageInfo{}
	if result.TotalCount != 0 {
		result.PageInfo.StartCursor = &result.Nodes[0].ID
		result.PageInfo.EndCursor = &result.Nodes[result.TotalCount-1].ID
	}
	result.PageInfo.HasPreviousPage = hasPrevPage
	result.PageInfo.HasNextPage = hasNextPage

	return &result
}

func (p *projectItemService) GetProjectItemByID(ctx context.Context, id string) (*model.ProjectV2Item, error) {
	item, err := db.FindProjectcard(ctx, p.exec, id,
		db.ProjectcardColumns.ID,
		db.ProjectcardColumns.Project,
		db.ProjectcardColumns.Issue,
		db.ProjectcardColumns.Pullrequest,
	)
	if err != nil {
		return nil, err
	}
	return convertProjectV2Item(item), nil
}

func (p *projectItemService) ListProjectItemOwnedByProject(ctx context.Context, projectID string, after *string, before *string, first *int, last *int) (*model.ProjectV2ItemConnection, error) {
	cond := []qm.QueryMod{
		qm.Select(
			db.ProjectcardColumns.ID,
			db.ProjectcardColumns.Project,
			db.ProjectcardColumns.Issue,
			db.ProjectcardColumns.Pullrequest,
		),
		db.ProjectcardWhere.Project.EQ(projectID),
	}
	var scanDesc bool

	switch {
	case (after != nil) && (before != nil):
		cond = append(cond, db.ProjectcardWhere.ID.GT(*after), db.ProjectcardWhere.ID.LT(*before))
	case after != nil:
		cond = append(cond,
			db.ProjectcardWhere.ID.GT(*after),
			qm.OrderBy(fmt.Sprintf("%s asc", db.ProjectcardColumns.ID)),
		)
		if first != nil {
			cond = append(cond, qm.Limit(*first))
		}
	case before != nil:
		scanDesc = true
		cond = append(cond,
			db.ProjectcardWhere.ID.LT(*before),
			qm.OrderBy(fmt.Sprintf("%s desc", db.ProjectcardColumns.ID)),
		)
		if last != nil {
			cond = append(cond, qm.Limit(*last))
		}
	default:
		switch {
		case last != nil:
			scanDesc = true
			cond = append(cond,
				qm.OrderBy(fmt.Sprintf("%s desc", db.ProjectcardColumns.ID)),
				qm.Limit(*last),
			)
		case first != nil:
			cond = append(cond,
				qm.OrderBy(fmt.Sprintf("%s asc", db.ProjectcardColumns.ID)),
				qm.Limit(*first),
			)
		default:
			cond = append(cond,
				qm.OrderBy(fmt.Sprintf("%s asc", db.ProjectcardColumns.ID)),
			)
		}
	}

	items, err := db.Projectcards(cond...).All(ctx, p.exec)
	if err != nil {
		return nil, err
	}

	var hasNextPage, hasPrevPage bool
	if len(items) != 0 {
		if scanDesc {
			for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
				items[i], items[j] = items[j], items[i]
			}
		}
		startCursor, endCursor := items[0].ID, items[len(items)-1].ID

		var err error
		hasPrevPage, err = db.Projectcards(
			db.ProjectcardWhere.Project.EQ(projectID),
			db.ProjectcardWhere.ID.LT(startCursor),
		).Exists(ctx, p.exec)
		if err != nil {
			return nil, err
		}
		hasNextPage, err = db.Projectcards(
			db.ProjectcardWhere.Project.EQ(projectID),
			db.ProjectcardWhere.ID.GT(endCursor),
		).Exists(ctx, p.exec)
		if err != nil {
			return nil, err
		}
	}

	return convertProjectV2ItemConnection(items, hasPrevPage, hasNextPage), nil
}

func (p *projectItemService) ListProjectItemOwnedByIssue(ctx context.Context, issueID string, after *string, before *string, first *int, last *int) (*model.ProjectV2ItemConnection, error) {
	cond := []qm.QueryMod{
		qm.Select(
			db.ProjectcardColumns.ID,
			db.ProjectcardColumns.Project,
			db.ProjectcardColumns.Issue,
			db.ProjectcardColumns.Pullrequest,
		),
		db.ProjectcardWhere.Pullrequest.EQ(null.StringFrom(issueID)),
	}
	var scanDesc bool

	switch {
	case (after != nil) && (before != nil):
		cond = append(cond, db.ProjectcardWhere.ID.GT(*after), db.ProjectcardWhere.ID.LT(*before))
	case after != nil:
		cond = append(cond,
			db.ProjectcardWhere.ID.GT(*after),
			qm.OrderBy(fmt.Sprintf("%s asc", db.ProjectcardColumns.ID)),
		)
		if first != nil {
			cond = append(cond, qm.Limit(*first))
		}
	case before != nil:
		scanDesc = true
		cond = append(cond,
			db.ProjectcardWhere.ID.LT(*before),
			qm.OrderBy(fmt.Sprintf("%s desc", db.ProjectcardColumns.ID)),
		)
		if last != nil {
			cond = append(cond, qm.Limit(*last))
		}
	default:
		switch {
		case last != nil:
			scanDesc = true
			cond = append(cond,
				qm.OrderBy(fmt.Sprintf("%s desc", db.ProjectcardColumns.ID)),
				qm.Limit(*last),
			)
		case first != nil:
			cond = append(cond,
				qm.OrderBy(fmt.Sprintf("%s asc", db.ProjectcardColumns.ID)),
				qm.Limit(*first),
			)
		default:
			cond = append(cond,
				qm.OrderBy(fmt.Sprintf("%s asc", db.ProjectcardColumns.ID)),
			)
		}
	}

	items, err := db.Projectcards(cond...).All(ctx, p.exec)
	if err != nil {
		return nil, err
	}

	var hasNextPage, hasPrevPage bool
	if len(items) != 0 {
		if scanDesc {
			for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
				items[i], items[j] = items[j], items[i]
			}
		}
		startCursor, endCursor := items[0].ID, items[len(items)-1].ID

		var err error
		hasPrevPage, err = db.Projectcards(
			db.ProjectcardWhere.Pullrequest.EQ(null.StringFrom(issueID)),
			db.ProjectcardWhere.ID.LT(startCursor),
		).Exists(ctx, p.exec)
		if err != nil {
			return nil, err
		}
		hasNextPage, err = db.Projectcards(
			db.ProjectcardWhere.Pullrequest.EQ(null.StringFrom(issueID)),
			db.ProjectcardWhere.ID.GT(endCursor),
		).Exists(ctx, p.exec)
		if err != nil {
			return nil, err
		}
	}

	return convertProjectV2ItemConnection(items, hasPrevPage, hasNextPage), nil
}

func (p *projectItemService) ListProjectItemOwnedByPullRequest(ctx context.Context, pullRequestID string, after *string, before *string, first *int, last *int) (*model.ProjectV2ItemConnection, error) {
	cond := []qm.QueryMod{
		qm.Select(
			db.ProjectcardColumns.ID,
			db.ProjectcardColumns.Project,
			db.ProjectcardColumns.Issue,
			db.ProjectcardColumns.Pullrequest,
		),
		db.ProjectcardWhere.Pullrequest.EQ(null.StringFrom(pullRequestID)),
	}
	var scanDesc bool

	switch {
	case (after != nil) && (before != nil):
		cond = append(cond, db.ProjectcardWhere.ID.GT(*after), db.ProjectcardWhere.ID.LT(*before))
	case after != nil:
		cond = append(cond,
			db.ProjectcardWhere.ID.GT(*after),
			qm.OrderBy(fmt.Sprintf("%s asc", db.ProjectcardColumns.ID)),
		)
		if first != nil {
			cond = append(cond, qm.Limit(*first))
		}
	case before != nil:
		scanDesc = true
		cond = append(cond,
			db.ProjectcardWhere.ID.LT(*before),
			qm.OrderBy(fmt.Sprintf("%s desc", db.ProjectcardColumns.ID)),
		)
		if last != nil {
			cond = append(cond, qm.Limit(*last))
		}
	default:
		switch {
		case last != nil:
			scanDesc = true
			cond = append(cond,
				qm.OrderBy(fmt.Sprintf("%s desc", db.ProjectcardColumns.ID)),
				qm.Limit(*last),
			)
		case first != nil:
			cond = append(cond,
				qm.OrderBy(fmt.Sprintf("%s asc", db.ProjectcardColumns.ID)),
				qm.Limit(*first),
			)
		default:
			cond = append(cond,
				qm.OrderBy(fmt.Sprintf("%s asc", db.ProjectcardColumns.ID)),
			)
		}
	}

	items, err := db.Projectcards(cond...).All(ctx, p.exec)
	if err != nil {
		return nil, err
	}

	var hasNextPage, hasPrevPage bool
	if len(items) != 0 {
		if scanDesc {
			for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
				items[i], items[j] = items[j], items[i]
			}
		}
		startCursor, endCursor := items[0].ID, items[len(items)-1].ID

		var err error
		hasPrevPage, err = db.Projectcards(
			db.ProjectcardWhere.Pullrequest.EQ(null.StringFrom(pullRequestID)),
			db.ProjectcardWhere.ID.LT(startCursor),
		).Exists(ctx, p.exec)
		if err != nil {
			return nil, err
		}
		hasNextPage, err = db.Projectcards(
			db.ProjectcardWhere.Pullrequest.EQ(null.StringFrom(pullRequestID)),
			db.ProjectcardWhere.ID.GT(endCursor),
		).Exists(ctx, p.exec)
		if err != nil {
			return nil, err
		}
	}

	return convertProjectV2ItemConnection(items, hasPrevPage, hasNextPage), nil
}

func (p *projectItemService) AddIssueInProjectV2(ctx context.Context, projectID, issueID string) (*model.ProjectV2Item, error) {
	itemID := uuid.New()
	item := &db.Projectcard{
		ID:      itemID.String(),
		Project: projectID,
		Issue:   null.StringFrom(issueID),
	}
	if err := item.Insert(ctx, p.exec, boil.Infer()); err != nil {
		return nil, err
	}
	return convertProjectV2Item(item), nil
}

func (p *projectItemService) AddPullRequestInProjectV2(ctx context.Context, projectID, pullRequestID string) (*model.ProjectV2Item, error) {
	itemID := uuid.New()
	item := &db.Projectcard{
		ID:          itemID.String(),
		Project:     projectID,
		Pullrequest: null.StringFrom(pullRequestID),
	}
	if err := item.Insert(ctx, p.exec, boil.Infer()); err != nil {
		return nil, err
	}
	return convertProjectV2Item(item), nil
}
