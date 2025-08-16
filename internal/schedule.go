package internal

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jpporta/ticket-control/internal/repository"
)

type Schedule struct {
	Name            string `json:"name"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	Cron_expression string `json:"cron_expression"`
	UserId          int32
}

func (a *Application) CreateSchedule(ctx context.Context, schedule *Schedule) error {
	err := a.Q.CreateScheduleTask(ctx, repository.CreateScheduleTaskParams{
		Name:           schedule.Name,
		Title:          schedule.Title,
		Description:    pgtype.Text{String: schedule.Description, Valid: schedule.Description != ""},
		CronExpression: schedule.Cron_expression,
		CreatedBy:      schedule.UserId,
	})
	if err != nil {
		return err
	}
	err = a.Cron.AddJob(ctx, a, Job{
		Name:           schedule.Name,
		Title:          schedule.Title,
		Description:    schedule.Description,
		CronExpression: schedule.Cron_expression,
		CreatedBy:      schedule.UserId,
	})
	if err != nil {
		return err
	}
	return nil
}

type response struct {
	ID             int32     `json:"id"`
	Name           string    `json:"name"`
	Title          string    `json:"title"`
	Enabled        bool      `json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	CronExpression string    `json:"cron_expression"`
}

func (a *Application) GetSchedules(ctx context.Context, userId int32) ([]response, error) {
	schedules, err := a.Q.GetUserScheduleTasks(ctx, userId)
	if err != nil {
		return nil, err
	}

	var result []response
	for _, s := range schedules {
		result = append(result, response{
			ID:             s.ID,
			Name:           s.Name,
			Title:          s.Title,
			Enabled:        s.Enabled,
			CreatedAt:      s.CreatedAt.Time,
			CronExpression: s.CronExpression,
		})
	}
	return result, nil
}

func (a *Application) ToggleSchedule(ctx context.Context, id, userId int32) error {
	job, err := a.Q.ToggleScheduleTask(ctx, repository.ToggleScheduleTaskParams{
		ID:        id,
		CreatedBy: userId,
	})
	if err != nil {
		return err
	}
	if job.Enabled {
		err = a.Cron.AddJob(ctx, a, Job{
			ID:             job.ID,
			Name:           job.Name,
			Title:          job.Title,
			Description:    job.Description.String,
			CronExpression: job.CronExpression,
			CreatedBy:      userId,
		})
		if err != nil {
			return err
		}
	} else {
		err = a.Cron.RemoveJob(job.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
