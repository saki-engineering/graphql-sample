package services

import (
	"context"

	"github.com/saki-engineering/graphql-sample/graph/model"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

//go:generate mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../../mock/$GOPACKAGE/service_mock.go
type Services interface {
	UserService
	RepoService
	IssueService
	PullRequestService
	ProjectService
	ProjectItemService
}

type UserService interface {
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByName(ctx context.Context, name string) (*model.User, error)
	ListUsersByID(ctx context.Context, IDs []string) ([]*model.User, error)
}

type RepoService interface {
	GetRepoByID(ctx context.Context, id string) (*model.Repository, error)
	GetRepoByFullName(ctx context.Context, owner, name string) (*model.Repository, error)
}

type IssueService interface {
	GetIssueByID(ctx context.Context, id string) (*model.Issue, error)
	GetIssueByRepoAndNumber(ctx context.Context, repoID string, number int) (*model.Issue, error)
	ListIssueInRepository(ctx context.Context, repoID string, after *string, before *string, first *int, last *int) (*model.IssueConnection, error)
}

type PullRequestService interface {
	GetPullRequestByID(ctx context.Context, id string) (*model.PullRequest, error)
	GetPullRequestByRepoAndNumber(ctx context.Context, repoID string, number int) (*model.PullRequest, error)
	ListPullRequestInRepository(ctx context.Context, repoID string, after *string, before *string, first *int, last *int) (*model.PullRequestConnection, error)
}

type ProjectService interface {
	GetProjectByID(ctx context.Context, id string) (*model.ProjectV2, error)
	GetProjectByOwnerAndNumber(ctx context.Context, ownerID string, number int) (*model.ProjectV2, error)
	ListProjectByOwner(ctx context.Context, ownerID string, after *string, before *string, first *int, last *int) (*model.ProjectV2Connection, error)
}

type ProjectItemService interface {
	GetProjectItemByID(ctx context.Context, id string) (*model.ProjectV2Item, error)
	ListProjectItemOwnedByProject(ctx context.Context, projectID string, after *string, before *string, first *int, last *int) (*model.ProjectV2ItemConnection, error)
	ListProjectItemOwnedByIssue(ctx context.Context, issueID string, after *string, before *string, first *int, last *int) (*model.ProjectV2ItemConnection, error)
	ListProjectItemOwnedByPullRequest(ctx context.Context, pullRequestID string, after *string, before *string, first *int, last *int) (*model.ProjectV2ItemConnection, error)
	AddIssueInProjectV2(ctx context.Context, projectID, issueID string) (*model.ProjectV2Item, error)
	AddPullRequestInProjectV2(ctx context.Context, projectID, pullRequestID string) (*model.ProjectV2Item, error)
}

type services struct {
	*userService
	*repoService
	*issueService
	*pullRequestService
	*projectService
	*projectItemService
}

func New(exec boil.ContextExecutor) Services {
	return &services{
		userService:        &userService{exec: exec},
		repoService:        &repoService{exec: exec},
		issueService:       &issueService{exec: exec},
		pullRequestService: &pullRequestService{exec: exec},
		projectService:     &projectService{exec: exec},
		projectItemService: &projectItemService{exec: exec},
	}
}
