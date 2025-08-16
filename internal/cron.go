package internal

import (
	"context"
	"fmt"
	"log"

	"github.com/go-co-op/gocron/v2"
	"github.com/jpporta/ticket-control/internal/utils"
)

type CronJob struct {
	s    gocron.Scheduler
	jobs []gocron.Job
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
	c.s = utils.Must(gocron.NewScheduler())
	c.jobs = make([]gocron.Job, 0)

	jobs := utils.Must(a.Q.GetAllEnabledScheduleTasks(ctx))

	for _, job := range jobs {
		j, err := c.s.NewJob(
			gocron.CronJob(job.CronExpression, false),
			gocron.NewTask(
				a.CreateTask,
				ctx,
				job.Title,
				job.Description.String,
				int32(0),
				job.CreatedBy,
			),
			gocron.WithTags(string(job.ID)),
		)
		if err != nil {
			log.Println("Failed to create job:", job.ID, job.Name, err)
			continue
		}
		log.Println("Created job:", job.ID, job.Name, "with cron expression:", job.CronExpression)
		c.jobs = append(c.jobs, j)
	}
	c.s.Start()
	return nil
}
func (c *CronJob) AddJob(ctx context.Context, a *Application, job Job) error {
	if c.s == nil {
		panic("CronJob scheduler not started")
	}
	j, err := c.s.NewJob(
		gocron.CronJob(job.CronExpression, false),
		gocron.NewTask(
			a.CreateTask,
			ctx,
			job.Title,
			job.Description,
			int32(0),
			job.CreatedBy,
		),
		gocron.WithTags(string(job.ID)),
	)
	if err != nil {
		return fmt.Errorf("Failed to create job:", job.ID, job.Name, err)
	}
	log.Println("Created job:", job.ID, job.Name, "with cron expression:", job.CronExpression)
	c.jobs = append(c.jobs, j)
	return nil
}

func (c *CronJob) RemoveJob(id int32) error {
	if c.s == nil {
		return fmt.Errorf("CronJob scheduler not started")
	}
	for i, job := range c.jobs {
		if job.Tags()[0] == string(id) {
			c.s.RemoveJob(job.ID())
			c.jobs = append(c.jobs[:i], c.jobs[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Job with ID %d not found", id)
}

func (c *CronJob) Stop() error {
	return c.s.Shutdown()
}
