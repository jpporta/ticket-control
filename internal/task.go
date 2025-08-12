package internal

import (
	"context"
	"fmt"
	"os"

	"github.com/jpporta/ticket-control/task"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func CreateTask(ctx context.Context, title, description, userName string, userId, priority int32) (int32, error) {
	task_port := os.Getenv("TASK_PORT")
	if task_port == "" {
		return 0, fmt.Errorf("TASK_PORT environment variable not set")
	}
	conn, err := grpc.NewClient(":"+task_port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return 0, fmt.Errorf("error creating gRPC client: %w", err)
	}
	defer conn.Close()

	tk := task.NewTaskServiceClient(conn)

	job := task.CreateTaskRequest{
		Title:       title,
		Description: description,
		UserName:    userName,
		UserId:      userId,
		Priority:    priority,
	}
	res, err := tk.Create(ctx, &job)
	if err != nil {
		return 0, fmt.Errorf("error creating task: %w", err)
	}
	return res.TaskId, nil
}
