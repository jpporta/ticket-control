package internal

import (
	"context"
	"fmt"
	"log"

	"github.com/jpporta/ticket-control/internal/utils"
	"github.com/robfig/cron/v3"
)

type CronJob struct {
	s    *cron.Cron
	jobs map[int32]cron.EntryID
}

func NewCronJob() *CronJob {
	return &CronJob{}
}

type Job struct {
	ID             int32
	Name           string
	Title          string
	Description    string
	CronExpression string
	CreatedBy      int32
}

func (c *CronJob) Start(ctx context.Context, a *Application) error {
	if c.s != nil {
		return nil // Already started
	}
	c.s = cron.New()
	c.jobs = make(map[int32]cron.EntryID, 0)

	jobs := utils.Must(a.Q.GetAllEnabledScheduleTasks(ctx))

	for _, job := range jobs {
		j, err := c.s.AddFunc(job.CronExpression, func() {
			a.CreateTask(
				ctx,
				job.Title,
				job.Description.String,
				int32(0),
				job.CreatedBy,
			)
		})
		if err != nil {
			log.Println("Failed to create job:", job.ID, job.Name, err)
			continue
		}
		log.Println("Created job:", job.ID, job.Name, "with cron expression:", job.CronExpression)
		c.jobs[job.ID] = j
	}
	c.s.Start()
	return nil
}
func (c *CronJob) AddJob(ctx context.Context, a *Application, job Job) error {
	if c.s == nil {
		panic("CronJob scheduler not started")
	}
	j, err := c.s.AddFunc(job.CronExpression, func() {
		a.CreateTask(
			ctx,
			job.Title,
			job.Description,
			int32(0),
			job.CreatedBy,
		)
	})
	if err != nil {
		return fmt.Errorf("Failed to create job:", job.ID, job.Name, err)
	}
	log.Println("Created job:", job.ID, job.Name, "with cron expression:", job.CronExpression)
	c.jobs[job.ID] = j
	return nil
}

func (c *CronJob) RemoveJob(id int32) error {
	if c.s == nil {
		return fmt.Errorf("CronJob scheduler not started")
	}
	jobId, exits := c.jobs[id]
	if !exits {
		return fmt.Errorf("Job with ID %d not found", id)
	}
	c.s.Remove(jobId)
	delete(c.jobs, id)
	return nil
}

func (c *CronJob) Stop() {
	c.s.Stop()
}
