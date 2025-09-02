package printer

import (
	"fmt"
	"image"
	_ "image/png"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

type taskInput struct {
	Title           string
	Description     string
	PriorityDisplay string
	CreatedBy       string
	CreatedAt       time.Time
}

func (p *Printer) PrintTask(
	title, description string,
	priority int32,
	createdBy string,
	createdAt time.Time,
) error {
	if !p.Enabled {
		p.queue = append(p.queue, func() error {
			return p.PrintTask(title, description, priority, createdBy, createdAt)
		})
		return fmt.Errorf("Printer is disabled, queuing task: %s\n", title)
	}
	close, err := p.start()
	if err != nil {
		fmt.Println("Error starting printer:", err)
		return err
	}
	defer close()
	template, ok := p.templates["task"]
	if !ok {
		return fmt.Errorf("task template not found")
	}

	file, err := os.CreateTemp("", "task-*.typ")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}
	var priorityDisplay string
	if priority < -1 || priority > 5 {
		priority = 0
	}
	switch priority {
	case -1:
		priorityDisplay = ""
	case 0:
		priorityDisplay = ""
	default:
		priorityDisplay = strings.TrimSpace(strings.Repeat(" ", int(priority)))
	}

	template.Execute(file, taskInput{
		Title:           title,
		Description:     description,
		PriorityDisplay: priorityDisplay,
		CreatedBy:       createdBy,
		CreatedAt:       createdAt,
	})
	cmd := exec.Command("typst", "c", file.Name(), "-f", "png")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error executing typst command: %w", err)
	}

	img_raw, err := os.Open(strings.Replace(file.Name(), ".typ", ".png", 1))
	if err != nil {
		return fmt.Errorf("error opening image file: %w", err)
	}
	defer img_raw.Close()
	img, _, err := image.Decode(img_raw)
	if err != nil {
		return fmt.Errorf("error decoding image: %w", err)
	}
	if img.Bounds().Max.Y%8 != 0 {
		cropRect := image.Rect(0, 0, img.Bounds().Max.X, img.Bounds().Max.Y-(img.Bounds().Max.Y%8))
		img = img.(interface {
			SubImage(r image.Rectangle) image.Image
		}).SubImage(cropRect)
	}
	p.Reset()
	err = p.printImage(img)
	if err != nil {
		return fmt.Errorf("error printing image: %w", err)
	}
	return nil
}

type TaskInput struct {
	Title, Description string
	Priority           int32
	CreatedBy          string
	CreatedAt          time.Time
}

func (p Printer) printSingleTask(task TaskInput, template *template.Template) error {
	file, err := os.CreateTemp("", "task-*.typ")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}
	defer os.Remove(file.Name())
	var priorityDisplay string
	priority := task.Priority
	if priority < -1 || priority > 5 {
		priority = 0
	}
	switch priority {
	case -1:
		priorityDisplay = ""
	case 0:
		priorityDisplay = ""
	default:
		priorityDisplay = strings.TrimSpace(strings.Repeat(" ", int(priority)))
	}

	err = template.Execute(file, taskInput{
		Title:           task.Title,
		Description:     task.Description,
		PriorityDisplay: priorityDisplay,
		CreatedBy:       task.CreatedBy,
		CreatedAt:       task.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}
	cmd := exec.Command("typst", "c", file.Name(), "-f", "png")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error executing typst command: %w", err)
	}

	img_raw, err := os.Open(strings.Replace(file.Name(), ".typ", ".png", 1))
	if err != nil {
		return fmt.Errorf("error opening image file: %w", err)
	}
	defer img_raw.Close()
	img, _, err := image.Decode(img_raw)
	if err != nil {
		return fmt.Errorf("error decoding image: %w", err)
	}
	if img.Bounds().Max.Y%8 != 0 {
		cropRect := image.Rect(0, 0, img.Bounds().Max.X, img.Bounds().Max.Y-(img.Bounds().Max.Y%8))
		img = img.(interface {
			SubImage(r image.Rectangle) image.Image
		}).SubImage(cropRect)
	}
	err = p.printImage(img)
	if err != nil {
		return fmt.Errorf("error printing image: %w", err)
	}
	return nil
}
func (p *Printer) PrintTasks(tasks []TaskInput) error {
	if !p.Enabled {
		p.queue = append(p.queue, func() error {
			return p.PrintTasks(tasks)
		})
		return fmt.Errorf("Printer is disabled, queuing tasks\n")
	}
	close, err := p.start()
	if err != nil {
		fmt.Println("Error starting printer:", err)
		return err
	}
	defer close()
	template, ok := p.templates["task"]
	if !ok {
		return fmt.Errorf("task template not found")
	}

	p.Reset()
	for _, task := range tasks {
		err := p.printSingleTask(task, template)
		if err != nil {
			log.Println("error printing task:", err, "Task:", task)
		}
	}
	return nil
}
