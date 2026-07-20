package list

import (
	"context"
	"time"
)

const ListLimit int64 = 10

type Repo interface {
	CreateList(ctx context.Context, title, content string, createdBy int32) (int32, error)
	DeleteLastList(ctx context.Context, createdBy int32) error
	TotalListsFromUser(ctx context.Context, createdBy int32, start, end time.Time) (int64, error)
}

// CreateParams mirrors what handlers pass to Service.CreateList.
type CreateParams struct {
	Title string
	Items []string
}
