// user_service_test.go
package main

import (
	"context"
	"testing"

	pb "statistics_service/proto"

	"github.com/stretchr/testify/assert"
)

// Mock implementation of Statistics service
type mockStatsServiceServer struct {
	pb.UnimplementedStatsServiceServer
}

func (s *mockStatsServiceServer) GetTaskStats(ctx context.Context, req *pb.TaskRequest) (*pb.TaskStatsResponse, error) {
	return &pb.TaskStatsResponse{
		TaskId:     req.TaskId,
		AuthorId:   43,
		ViewsCount: 42,
		LikesCount: 44,
	}, nil
}

func (s *mockStatsServiceServer) GetTopTasks(ctx context.Context, req *pb.TopTasksRequest) (*pb.TopTasksResponse, error) {
	var tasks []*pb.TaskStatsResponse
	for i := range 42 {
		tasks = append(tasks, &pb.TaskStatsResponse{
			TaskId:     int32(i),
			AuthorId:   int32(i + 100),
			ViewsCount: 42,
			LikesCount: 43,
		})
	}
	return &pb.TopTasksResponse{Tasks: tasks}, nil
}

func TestTaskInfo(t *testing.T) {
	s := mockStatsServiceServer{}

	req := &pb.TaskRequest{
		TaskId: 1,
	}
	res, err := s.GetTaskStats(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int32(1), res.TaskId)
	assert.Equal(t, int32(43), res.AuthorId)
	assert.Equal(t, uint64(44), res.LikesCount)
	assert.Equal(t, uint64(42), res.ViewsCount)
}

func TestTopTasks(t *testing.T) {
	s := mockStatsServiceServer{}

	req := &pb.TopTasksRequest{
		Metric: pb.EventType_LIKE,
	}
	res, err := s.GetTopTasks(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	for i, task := range res.Tasks {
		assert.Equal(t, int32(i), task.TaskId)
		assert.Equal(t, int32(i+100), task.AuthorId)
		assert.Equal(t, uint64(42), task.ViewsCount)
		assert.Equal(t, uint64(43), task.LikesCount)
	}
}
