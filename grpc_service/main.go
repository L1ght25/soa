package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "grpc_service/proto"
)

var (
	secretKey = os.Getenv("SECRET_KEY")
	db        *sql.DB
)

func verifyToken(tokenString string) (int32, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return -1, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := int32(claims["userID"].(float64))
		return userID, nil
	} else {
		return -1, err
	}
}

func (s *server) authenticate(ctx context.Context) (int32, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return -1, status.Error(codes.Unauthenticated, "missing metadata")
	}

	tokenHeader, ok := md["x-access-token"]

	if !ok || len(tokenHeader) == 0 {
		return -1, status.Error(codes.Unauthenticated, "missing token")
	}

	token := strings.TrimPrefix(tokenHeader[0], "Bearer ")

	userID, err := verifyToken(token)
	if err != nil {
		return -1, status.Errorf(codes.Unauthenticated, "invalid authorization token: %v", err)
	}

	return userID, nil
}

func connectDB() (*sql.DB, error) {
	connStr := os.Getenv("DATA_SOURCE")

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

type server struct {
	pb.UnimplementedTaskServiceServer
}

func (s *server) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.Task, error) {
	userID, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	var taskID int32
	err = db.QueryRowContext(ctx, "INSERT INTO tasks (title, content, createdBy) VALUES ($1, $2, $3) RETURNING id", req.Title, req.Content, userID).Scan(&taskID)
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

func (s *server) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.Task, error) {
	userID, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	result, err := db.ExecContext(ctx, "UPDATE tasks SET title = $1, content = $2 WHERE id = $3 AND createdBy = $4", req.Title, req.Content, req.TaskId, userID)
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

func (s *server) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	userID, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	result, err := db.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1 AND createdBy = $2", req.TaskId, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update task: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete task: %v", err)
	}

	if rowsAffected == 0 {
		result, _ := db.ExecContext(ctx, "SELECT FROM tasks WHERE id = $1", req.TaskId)
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

func (s *server) GetTaskById(ctx context.Context, req *pb.GetTaskByIdRequest) (*pb.Task, error) {
	userID, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	var task pb.Task
	err = db.QueryRowContext(ctx, "SELECT id, title, content, createdBy FROM tasks WHERE id = $1", req.TaskId).Scan(&task.Id, &task.Title, &task.Content, &task.CreatedByUserID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "task not found")
	}

	if userID != task.CreatedByUserID {
		return nil, status.Errorf(codes.PermissionDenied, "Access denied")
	}

	return &task, nil
}

func (s *server) GetTaskListWithPagination(ctx context.Context, req *pb.GetTaskListRequest) (*pb.GetTaskListResponse, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, title, content FROM tasks LIMIT $1 OFFSET $2", req.PageSize, req.PageNumber*req.PageSize)
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

func main() {
	var err error
	db, err = connectDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTaskServiceServer(s, &server{})
	fmt.Println("Server running on port :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
