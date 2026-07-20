package link

import (
	"context"
	"time"
)

const LinkLimit int64 = 50

type Repo interface {
	CreateLink(ctx context.Context, title, url string, createdBy int32) (int32, error)
	DeleteLastLink(ctx context.Context, createdBy int32) error
	GetLinkByID(ctx context.Context, id int32) (string, error)
	TotalLinksFromUser(ctx context.Context, createdBy int32, start, end time.Time) (int64, error)
}

type CreateParams struct {
	Title string
	URL   string
}
