package adapter

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jpporta/ticket-control/internal/repository"
	"github.com/jpporta/ticket-control/internal/schedule"
)

func (a Schedule) CreateScheduleTask(ctx context.Context, name, title, description, cronExpression, checkFunction string, createdBy int32) (int32, error) {
	return a.Q.CreateScheduleTask(ctx, repository.CreateScheduleTaskParams{
		Name:           name,
		Title:          title,
		Description:    pgtype.Text{String: description, Valid: description != ""},
		CronExpression: cronExpression,
		CreatedBy:      createdBy,
		CheckFunction:  pgtype.Text{String: checkFunction, Valid: checkFunction != ""},
	})
}

func (a Schedule) ToggleScheduleTask(ctx context.Context, id, createdBy int32) (schedule.ToggleScheduleRow, error) {
	row, err := a.Q.ToggleScheduleTask(ctx, repository.ToggleScheduleTaskParams{ID: id, CreatedBy: createdBy})
	if err != nil {
		return schedule.ToggleScheduleRow{}, err
	}
	return schedule.ToggleScheduleRow{
		ID:             row.ID,
		Name:           row.Name,
		Title:          row.Title,
		Description:    row.Description.String,
		CronExpression: row.CronExpression,
		Enabled:        row.Enabled,
		CreatedBy:      createdBy,
		CheckFunction:  row.CheckFunction.String,
	}, nil
}

func (a Schedule) GetUserScheduleTasks(ctx context.Context, createdBy int32) ([]schedule.ScheduleRow, error) {
	rows, err := a.Q.GetUserScheduleTasks(ctx, createdBy)
	if err != nil {
		return nil, err
	}
	out := make([]schedule.ScheduleRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, schedule.ScheduleRow{
			ID:             r.ID,
			Name:           r.Name,
			Title:          r.Title,
			Description:    r.Description.String,
			CronExpression: r.CronExpression,
			Enabled:        r.Enabled,
			CreatedAt:      r.CreatedAt.Time,
			CheckFunction:  r.CheckFunction.String,
		})
	}
	return out, nil
}

func (a Schedule) GetAllEnabledScheduleTasks(ctx context.Context) ([]schedule.EnabledScheduleRow, error) {
	rows, err := a.Q.GetAllEnabledScheduleTasks(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]schedule.EnabledScheduleRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, schedule.EnabledScheduleRow{
			ID:             r.ID,
			Name:           r.Name,
			Title:          r.Title,
			Description:    r.Description.String,
			CronExpression: r.CronExpression,
			CreatedBy:      r.CreatedBy,
			CheckFunction:  r.CheckFunction.String,
		})
	}
	return out, nil
}
