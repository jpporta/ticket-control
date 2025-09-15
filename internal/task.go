package internal

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jpporta/ticket-control/internal/printer"
	"github.com/jpporta/ticket-control/internal/repository"
)

var TASK_LIMIT int64 = 50

func (a *Application) UserHasReachedTaskLimit(ctx context.Context, userId int32) (bool, error) {
	startYear, startMonth, startDay := time.Now().Date()
	startTime := time.Date(startYear, startMonth, startDay, 0, 0, 0, 0, time.UTC)

	total, err := a.Q.GetNoUsersTask(ctx, repository.GetNoUsersTaskParams{
		CreatedBy:   userId,
		CreatedAt:   pgtype.Timestamp{Time: startTime, Valid: true},
		CreatedAt_2: pgtype.Timestamp{Time: startTime.Add(time.Hour * 24), Valid: true},
	})
	if err != nil {
		return false, err
	}

	return (total >= TASK_LIMIT), nil
}

func (a *Application) CreateTask(ctx context.Context, title, description string, priority int32, userId int32) (int32, error) {
	// Create in DB
	id, err := a.Q.CreateTask(ctx, repository.CreateTaskParams{
		Title:       title,
		Description: pgtype.Text{String: description, Valid: description != ""},
		Priority:    pgtype.Int4{Int32: priority, Valid: priority > 0 && priority <= 5},
		CreatedBy:   userId,
	})
	if err != nil {
		return 0, fmt.Errorf("Error creating task: %w", err)
	}

	user, err := a.Q.GetUserById(ctx, userId)
	err = a.Printer.PrintTask(id, title, description, priority, user.Name, time.Now())
	if err != nil {
		err_2 := a.Q.DeleteLastTask(ctx, userId)
		if err_2 != nil {
			return 0, fmt.Errorf("Error deleting task after printer start failure: %w", err)
		}
		return 0, fmt.Errorf("Error starting printer: %w", err)
	}
	return id, nil
}

type CreateTaskParams struct {
	Title, Description string
	Priority           int32
}

func (a *Application) CreateTasks(ctx context.Context, tasks []CreateTaskParams, userId int32) (int, error) {
	// Create in DB
	printerTasks := []printer.TaskInput{}
	user, err := a.Q.GetUserById(ctx, userId)
	for _, task := range tasks {
		ID, err := a.Q.CreateTask(ctx, repository.CreateTaskParams{
			Title:       task.Title,
			Description: pgtype.Text{String: task.Description, Valid: task.Description != ""},
			Priority:    pgtype.Int4{Int32: task.Priority, Valid: task.Priority > 0 && task.Priority <= 5},
			CreatedBy:   userId,
		})
		if err != nil {
			log.Println("Error creating task:", err, "Task:", task)
			continue
		}
		printerTasks = append(printerTasks, printer.TaskInput{
			ID:          ID,
			Title:       task.Title,
			Description: task.Description,
			Priority:    task.Priority,
			CreatedBy:   user.Name,
			CreatedAt:   time.Now()})
	}

	err = a.Printer.PrintTasks(printerTasks)
	if err != nil {
		return 0, fmt.Errorf("Error creating tasks: %w", err)
	}
	return len(printerTasks), nil
}

type openTasks struct {
	ID        int32     `json:"id"`
	Title     string    `json:"title"`
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}

func (a *Application) GetOpenTasks(ctx context.Context, userId int32) ([]openTasks, error) {
	tasks := []openTasks{}
	tasks_db, err := a.Q.GetOpenTasks(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("Error getting open tasks: %w", err)
	}
	for _, task := range tasks_db {
		tasks = append(tasks, openTasks{
			ID:        task.ID,
			Title:     task.Title,
			Priority:  int(task.Priority.Int32),
			CreatedAt: task.CreatedAt.Time,
		})
	}
	return tasks, nil
}

func (a *Application) MarkTaskAsDone(ctx context.Context, taskId, userId int32) error {
	// Mark as done in DB
	err := a.Q.MarkTaskAsDone(ctx, repository.MarkTaskAsDoneParams{
		ID:          taskId,
		CompletedBy: pgtype.Int4{Int32: userId, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("Error marking task as done: %w", err)
	}
	err = a.Printer.PrintBip()
	if err != nil {
		return fmt.Errorf("Error printing bip: %w", err)
	}
	return nil
}
