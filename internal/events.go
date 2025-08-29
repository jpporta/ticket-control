package internal

import (
	"context"
	"fmt"
	"time"
)

type Event struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Title string    `json:"title"`
}

func (a *Application) CreateEvents(ctx context.Context, events []Event, userId int32) (int, error) {
	applicationTasks := []CreateTaskParams{}
	for _, event := range events {
		description := event.Start.Format("15:04") + " - " + event.End.Format("15:04")
		applicationTasks = append(applicationTasks, CreateTaskParams{
			Title:       event.Title,
			Description: description,
			Priority:    -1,
		})
	}
	noCreated, err := a.CreateTasks(ctx, applicationTasks, userId)
	if err != nil {
		return 0, fmt.Errorf("Error creating events: %w", err)
	}
	return noCreated, nil
}
