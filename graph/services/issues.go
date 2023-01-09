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

type issueService struct {
	exec boil.ContextExecutor
}

func convertIssue(issue *db.Issue) *model.Issue {
	issueURL, err := model.UnmarshalURI(issue.URL)
	if err != nil {
		log.Println("invalid URI", issue.URL)
	}

	return &model.Issue{
		ID:         issue.ID,
		URL:        issueURL,
		Title:      issue.Title,
		Closed:     (issue.Closed == 1),
		Number:     int(issue.Number),
		Author:     &model.User{ID: issue.Author},
		Repository: &model.Repository{ID: issue.Repository},
	}
}

func convertIssueConnection(issues db.IssueSlice, hasPrevPage, hasNextPage bool) *model.IssueConnection {
	var result model.IssueConnection

	for _, dbi := range issues {
		issue := convertIssue(dbi)

		result.Edges = append(result.Edges, &model.IssueEdge{Cursor: issue.ID, Node: issue})
		result.Nodes = append(result.Nodes, issue)
	}
	result.TotalCount = len(issues)

	result.PageInfo = &model.PageInfo{}
	if result.TotalCount != 0 {
		result.PageInfo.StartCursor = &result.Nodes[0].ID
		result.PageInfo.EndCursor = &result.Nodes[result.TotalCount-1].ID
	}
	result.PageInfo.HasPreviousPage = hasPrevPage
	result.PageInfo.HasNextPage = hasNextPage

	return &result
}

func (i *issueService) GetIssueByID(ctx context.Context, id string) (*model.Issue, error) {
	issue, err := db.FindIssue(ctx, i.exec, id,
		db.IssueColumns.ID,
		db.IssueColumns.URL,
		db.IssueColumns.Title,
		db.IssueColumns.Closed,
		db.IssueColumns.Number,
		db.IssueColumns.Author,
		db.IssueColumns.Repository,
	)
	if err != nil {
		return nil, err
	}
	return convertIssue(issue), nil
}

func (i *issueService) GetIssueByRepoAndNumber(ctx context.Context, repoID string, number int) (*model.Issue, error) {
	issue, err := db.Issues(
		qm.Select(
			db.IssueColumns.ID,
			db.IssueColumns.URL,
			db.IssueColumns.Title,
			db.IssueColumns.Closed,
			db.IssueColumns.Number,
			db.IssueColumns.Author,
			db.IssueColumns.Repository,
		),
		db.IssueWhere.Repository.EQ(repoID),
		db.IssueWhere.Number.EQ(int64(number)),
	).One(ctx, i.exec)
	if err != nil {
		return nil, err
	}
	return convertIssue(issue), nil
}

func (i *issueService) ListIssueInRepository(ctx context.Context, repoID string, after *string, before *string, first *int, last *int) (*model.IssueConnection, error) {
	cond := []qm.QueryMod{
		qm.Select(
			db.IssueColumns.ID,
			db.IssueColumns.URL,
			db.IssueColumns.Title,
			db.IssueColumns.Closed,
			db.IssueColumns.Number,
			db.IssueColumns.Author,
			db.IssueColumns.Repository,
		),
		db.IssueWhere.Repository.EQ(repoID),
	}
	var scanDesc bool

	switch {
	case (after != nil) && (before != nil):
		cond = append(cond, db.IssueWhere.ID.GT(*after), db.IssueWhere.ID.LT(*before))
	case after != nil:
		cond = append(cond,
			db.IssueWhere.ID.GT(*after),
			qm.OrderBy(fmt.Sprintf("%s asc", db.IssueColumns.ID)),
		)
		if first != nil {
			cond = append(cond, qm.Limit(*first))
		}
	case before != nil:
		scanDesc = true
		cond = append(cond,
			db.IssueWhere.ID.LT(*before),
			qm.OrderBy(fmt.Sprintf("%s desc", db.IssueColumns.ID)),
		)
		if last != nil {
			cond = append(cond, qm.Limit(*last))
		}
	default:
		switch {
		case last != nil:
			scanDesc = true
			cond = append(cond,
				qm.OrderBy(fmt.Sprintf("%s desc", db.IssueColumns.ID)),
				qm.Limit(*last),
			)
		case first != nil:
			cond = append(cond,
				qm.OrderBy(fmt.Sprintf("%s asc", db.IssueColumns.ID)),
				qm.Limit(*first),
			)
		default:
			cond = append(cond,
				qm.OrderBy(fmt.Sprintf("%s asc", db.IssueColumns.ID)),
			)
		}
	}

	issues, err := db.Issues(cond...).All(ctx, i.exec)
	if err != nil {
		return nil, err
	}

	var hasNextPage, hasPrevPage bool
	if len(issues) != 0 {
		if scanDesc {
			for i, j := 0, len(issues)-1; i < j; i, j = i+1, j-1 {
				issues[i], issues[j] = issues[j], issues[i]
			}
		}
		startCursor, endCursor := issues[0].ID, issues[len(issues)-1].ID

		var err error
		hasPrevPage, err = db.Issues(
			db.IssueWhere.Repository.EQ(repoID),
			db.IssueWhere.ID.LT(startCursor),
		).Exists(ctx, i.exec)
		if err != nil {
			return nil, err
		}
		hasNextPage, err = db.Issues(
			db.IssueWhere.Repository.EQ(repoID),
			db.IssueWhere.ID.GT(endCursor),
		).Exists(ctx, i.exec)
		if err != nil {
			return nil, err
		}
	}

	return convertIssueConnection(issues, hasPrevPage, hasNextPage), nil
}
