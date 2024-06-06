package server

import (
	"common"
	"context"
	"database/sql"
	pb "grpc_service/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ConnectDB(connStr string) (*sql.DB, error) {

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS tasks (
		id SERIAL PRIMARY KEY,
		title TEXT,
		content TEXT,
		createdBy INT
	);
	`)

	if err != nil {
		return nil, err
	}

	return db, nil
}

type Server struct {
	pb.UnimplementedTaskServiceServer
	Db *sql.DB
}

func (s *Server) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.Task, error) {
	userID, err := common.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	var taskID int32
	err = s.Db.QueryRowContext(ctx, "INSERT INTO tasks (title, content, createdBy) VALUES ($1, $2, $3) RETURNING id", req.Title, req.Content, userID).Scan(&taskID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create task: %v", err)
	}

	return &pb.Task{
		Id:              taskID,
		Title:           req.Title,
		Content:         req.Content,
		CreatedByUserID: userID,
	}, nil
}

func (s *Server) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.Task, error) {
	userID, err := common.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.Db.ExecContext(ctx, "UPDATE tasks SET title = $1, content = $2 WHERE id = $3 AND createdBy = $4", req.Title, req.Content, req.TaskId, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update task: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update task: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Errorf(codes.PermissionDenied, "Access denied")
	}

	return &pb.Task{
		Id:              req.TaskId,
		Title:           req.Title,
		Content:         req.Content,
		CreatedByUserID: userID,
	}, nil
}

func (s *Server) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	userID, err := common.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.Db.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1 AND createdBy = $2", req.TaskId, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update task: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete task: %v", err)
	}

	if rowsAffected == 0 {
		result, _ := s.Db.ExecContext(ctx, "SELECT id FROM tasks WHERE id = $1", req.TaskId)
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected != 0 {
			return nil, status.Errorf(codes.PermissionDenied, "Access denied")
		}
		return nil, status.Errorf(codes.NotFound, "task not found")
	}

	return &pb.DeleteTaskResponse{
		Success: true,
	}, nil
}

func (s *Server) GetTaskById(ctx context.Context, req *pb.GetTaskByIdRequest) (*pb.Task, error) {
	_, err := common.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	var task pb.Task
	err = s.Db.QueryRowContext(ctx, "SELECT id, title, content, createdBy FROM tasks WHERE id = $1", req.TaskId).Scan(&task.Id, &task.Title, &task.Content, &task.CreatedByUserID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "task not found")
	}

	return &task, nil
}

func (s *Server) GetTaskListWithPagination(ctx context.Context, req *pb.GetTaskListRequest) (*pb.GetTaskListResponse, error) {
	rows, err := s.Db.QueryContext(ctx, "SELECT id, title, content FROM tasks LIMIT $1 OFFSET $2", req.PageSize, req.PageNumber*req.PageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch tasks: %v", err)
	}
	defer rows.Close()

	var tasks []*pb.Task
	for rows.Next() {
		var task pb.Task
		if err := rows.Scan(&task.Id, &task.Title, &task.Content); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan task: %v", err)
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "error iterating over rows: %v", err)
	}

	return &pb.GetTaskListResponse{
		Tasks: tasks,
	}, nil
}
