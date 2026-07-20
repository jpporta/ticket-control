package printer

import (
	"fmt"
	"image"
	"log/slog"
	"strings"
	"text/template"
	"time"

	"github.com/jpporta/ticket-control/internal/printer/render"
)

type taskInput struct {
	ID              int32
	Title           string
	Description     string
	PriorityDisplay string
	CreatedBy       string
	CreatedAt       time.Time
}

func (p *Printer) PrintTask(
	id int32,
	title, description string,
	priority int32,
	createdBy string,
	createdAt time.Time,
) error {
	if !p.Enabled {
		p.queue = append(p.queue, func() error {
			return p.PrintTask(id, title, description, priority, createdBy, createdAt)
		})
		return fmt.Errorf("%w: queuing task: %s", errPrinterOffline, title)
	}
	tmpl, ok := p.templates["task"]
	if !ok {
		return fmt.Errorf("task template not found")
	}

	var priorityDisplay string
	switch priority {
	case -2:
		priorityDisplay = "❀"
	case -1:
		priorityDisplay = "⏳"
	case 1, 2, 3, 4, 5:
		priorityDisplay = strings.TrimSpace(strings.Repeat("! ", int(priority)))
	default:
		priorityDisplay = "·"
	}

	pngFile, cleanup, err := render.Render(tmpl, taskInput{
		ID:              id,
		Title:           title,
		Description:     description,
		PriorityDisplay: priorityDisplay,
		CreatedBy:       createdBy,
		CreatedAt:       createdAt,
	}, "task")
	if err != nil {
		return err
	}
	defer cleanup()

	close, err := p.start()
	if err != nil {
		slog.Error("printer start", "err", err)
		return err
	}
	defer close()

	img, _, err := image.Decode(pngFile)
	pngFile.Close()
	if err != nil {
		return fmt.Errorf("decode task png: %w", err)
	}
	img = render.CropHeight8(img)

	p.Reset()
	return p.printImage(img)
}

type TaskInput struct {
	ID                 int32
	Title, Description string
	Priority           int32
	CreatedBy          string
	CreatedAt          time.Time
}

func (p *Printer) printSingleTask(task TaskInput, tmpl *template.Template) error {
	pngFile, cleanup, err := render.Render(tmpl, taskInput{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		PriorityDisplay: priorityDisplayFor(task.Priority),
		CreatedBy:   task.CreatedBy,
		CreatedAt:   task.CreatedAt,
	}, "task")
	if err != nil {
		return err
	}
	defer cleanup()

	img, _, err := image.Decode(pngFile)
	pngFile.Close()
	if err != nil {
		return fmt.Errorf("decode task png: %w", err)
	}
	img = render.CropHeight8(img)
	return p.printImage(img)
}

func priorityDisplayFor(priority int32) string {
	switch {
	case priority == -1:
		return "⏳"
	case priority >= 1 && priority <= 5:
		return strings.TrimSpace(strings.Repeat("! ", int(priority)))
	default:
		return "·"
	}
}

func (p *Printer) PrintTasks(tasks []TaskInput) error {
	if !p.Enabled {
		p.queue = append(p.queue, func() error {
			return p.PrintTasks(tasks)
		})
		return fmt.Errorf("%w: queuing tasks", errPrinterOffline)
	}
	tmpl, ok := p.templates["task"]
	if !ok {
		return fmt.Errorf("task template not found")
	}

	close, err := p.start()
	if err != nil {
		slog.Error("printer start", "err", err)
		return err
	}
	defer close()
	p.Reset()
	for _, task := range tasks {
		if err := p.printSingleTask(task, tmpl); err != nil {
			slog.Error("print task", "task", task, "err", err)
		}
	}
	return nil
}
