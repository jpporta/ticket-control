package task

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	_ "embed"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jpporta/ticket-control/internal/repository"
)

//goo:embed models/task.typ
var task_template string

var TASK_LIMIT int64 = 50

type Server struct {
	queries  *repository.Queries
	template *template.Template
}

type taskInput struct {
	Title           string
	Description     string
	PriorityDisplay string
	CreatedBy       string
	CreatedAt       time.Time
}

func NewServer(conn *pgx.Conn) *Server {
	queries := repository.New(conn)
	layout, err := template.New("task").Parse(task_template)
	if err != nil {
		log.Fatalf("Error parsing task template: %v", err)
	}
	return &Server{
		queries:  queries,
		template: layout,
	}
}


func (s *Server) UserHasReachedTaskLimit(ctx context.Context, userId int32) (bool, error) {
	start := time.Now()
	defer func() {
		log.Printf("UserHasReachedTaskLimit took %s", time.Since(start))
	}()

	startYear, startMonth, startDay := time.Now().Date()
	startTime := time.Date(startYear, startMonth, startDay, 0, 0, 0, 0, time.UTC)

	total, err := s.queries.GetNoUsersTask(ctx, repository.GetNoUsersTaskParams{
		CreatedBy:   userId,
		CreatedAt:   pgtype.Timestamp{Time: startTime, Valid: true},
		CreatedAt_2: pgtype.Timestamp{Time: startTime.Add(time.Hour * 24), Valid: true},
	})
	if err != nil {
		return false, err
	}

	return (total >= TASK_LIMIT), nil
}

func (s *Server) Create(ctx context.Context, r *CreateTaskRequest) (*CreateTaskResponse, error) {
	start := time.Now()
	defer func() {
		log.Printf("Create took %s", time.Since(start))
	}()
	limitReached, err := s.UserHasReachedTaskLimit(ctx, r.UserId)
	if err != nil {
		return nil, fmt.Errorf("Error checking task limit: %w", err)
	}

	if limitReached {
		return nil, fmt.Errorf("You have reached your task limit for today")
	}
	id, err := s.queries.CreateTask(ctx, repository.CreateTaskParams{
		Title:       r.Title,
		Description: pgtype.Text{String: r.Description, Valid: r.Description != ""},
		Priority:    pgtype.Int4{Int32: r.Priority, Valid: r.Priority > 0 && r.Priority <= 5},
		CreatedBy:   r.UserId,
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating task")
	}

	file, err := os.CreateTemp("", "task-*.typ")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %w", err)
	}
	s.template.Execute(file, taskInput{
		Title:           r.Title,
		Description:     r.Description,
		PriorityDisplay: strings.TrimSpace(strings.Repeat(" ", int(r.Priority))),
		CreatedBy:       r.UserName,
		CreatedAt:       time.Now(),
	})
	cmd := exec.Command("typst", "c", file.Name(), "-f", "png")
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error executing typst command: %w", err)
	}

	img, err := os.ReadFile(strings.Replace(file.Name(), ".typ", ".png", 1))
	if err != nil {
		return nil, fmt.Errorf("error opening image file: %w", err)
	}

	err = PrintImage(ctx, img)
	if err != nil {
		err_2 := s.queries.DeleteLastTask(ctx, r.UserId)
		if err_2 != nil {
			return nil, fmt.Errorf("Error deleting task after printer start failure: %w", err)
		}
		return nil, fmt.Errorf("Error starting printer: %w", err)
	}

	return &CreateTaskResponse{
		TaskId: id,
	}, nil

}
