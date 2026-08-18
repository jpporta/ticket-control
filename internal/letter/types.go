package letter

import (
	"context"
	"time"
)

const LetterLimit int64 = 20

type Repo interface {
	CreateLetter(ctx context.Context, title, content, recipient, sender string, createdBy int32) (int32, error)
	DeleteLastLetter(ctx context.Context, createdBy int32) error
	TotalLettersFromUser(ctx context.Context, createdBy int32, start, end time.Time) (int64, error)
}

// CreateParams mirrors what handlers pass to Service.CreateLetter.
type CreateParams struct {
	Title    string
	Content  string
	To       string
	ToLabel  string
	From     string
	SignOff  string
	Date     string
	Font     string
	FontSize string
	Justify  bool
}
