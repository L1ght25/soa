package main

import (
	"context"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	pb "grpc_service/proto"
)

type mockTask struct {
	title     string
	content   string
	createdBy int32
}

type mockServer struct {
	mockTasks map[int32]mockTask
	pb.UnimplementedTaskServiceServer
}

func (s *mockServer) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.Task, error) {
	taskId := int32(len(s.mockTasks))
	s.mockTasks[taskId] = mockTask{
		title:     req.Title,
		content:   req.Content,
		createdBy: 43,
	}

	return &pb.Task{
		Id:              taskId,
		Title:           req.Title,
		Content:         req.Content,
		CreatedByUserID: 43,
	}, nil
}

func (s *mockServer) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.Task, error) {
	s.mockTasks[req.TaskId] = mockTask{
		title:     req.Title,
		content:   req.Content,
		createdBy: 43,
	}

	return &pb.Task{
		Id:              req.TaskId,
		Title:           req.Title,
		Content:         req.Content,
		CreatedByUserID: 43,
	}, nil
}

func (s *mockServer) GetTaskById(ctx context.Context, req *pb.GetTaskByIdRequest) (*pb.Task, error) {

	taskInfo := s.mockTasks[req.TaskId]

	return &pb.Task{
		Id:              req.TaskId,
		Title:           taskInfo.title,
		Content:         taskInfo.content,
		CreatedByUserID: 43,
	}, nil

}

func TestGet(t *testing.T) {
	s := mockServer{
		mockTasks: make(map[int32]mockTask),
	}

	req := &pb.CreateTaskRequest{
		Title:   "title",
		Content: "content",
	}
	res, err := s.CreateTask(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int32(0), res.Id)

	resGet, err := s.GetTaskById(context.Background(), &pb.GetTaskByIdRequest{
		TaskId: 0,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resGet)
	assert.Equal(t, int32(0), resGet.Id)
	assert.Equal(t, "content", resGet.Content)
	assert.Equal(t, "title", resGet.Title)
	assert.Equal(t, int32(43), resGet.CreatedByUserID)
}

func TestUpdate(t *testing.T) {
	s := mockServer{
		mockTasks: make(map[int32]mockTask),
	}

	req := &pb.CreateTaskRequest{
		Title:   "title",
		Content: "content",
	}
	res, err := s.CreateTask(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int32(0), res.Id)

	reqUpd := &pb.UpdateTaskRequest{
		TaskId:  0,
		Title:   "title_second",
		Content: "content_second",
	}
	resUpd, err := s.UpdateTask(context.Background(), reqUpd)

	assert.NoError(t, err)
	assert.NotNil(t, resUpd)
	assert.Equal(t, int32(0), res.Id)

	resGet, err := s.GetTaskById(context.Background(), &pb.GetTaskByIdRequest{
		TaskId: 0,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resGet)
	assert.Equal(t, int32(0), resGet.Id)
	assert.Equal(t, "content_second", resGet.Content)
	assert.Equal(t, "title_second", resGet.Title)
	assert.Equal(t, int32(43), resGet.CreatedByUserID)
}
