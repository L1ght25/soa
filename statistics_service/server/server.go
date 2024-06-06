package server

import (
	"common"
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"

	pb "statistics_service/proto"
)

type Server struct {
	pb.UnimplementedStatsServiceServer
	Db clickhouse.Conn
}

func (s *Server) GetTaskStats(ctx context.Context, req *pb.TaskRequest) (*pb.TaskStatsResponse, error) {
	_, err := common.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	var (
		viewsCount uint64
		likesCount uint64
	)
	query :=
		`
        SELECT
			sum(likes_count) AS likes_count,
			sum(views_count) AS views_count
        FROM
			mv_likes_views FINAL
        WHERE
            task_id = ?
		GROUP BY task_id
    `
	if err := s.Db.QueryRow(ctx, query, req.TaskId).Scan(&likesCount, &viewsCount); err != nil {
		return nil, err
	}
	return &pb.TaskStatsResponse{
		TaskId:     req.TaskId,
		ViewsCount: viewsCount,
		LikesCount: likesCount,
	}, nil
}

func (s *Server) GetTopTasks(ctx context.Context, req *pb.TopTasksRequest) (*pb.TopTasksResponse, error) {
	_, err := common.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	metric := "views_count"
	if req.Metric == pb.EventType_LIKE {
		metric = "likes_count"
	}
	query :=
		`
        SELECT
            task_id,
			author_id,
            sum(views_count) AS views_count,
            sum(likes_count) AS likes_count
        FROM
			mv_likes_views FINAL
        GROUP BY
            (task_id, author_id)
        ORDER BY
            ` + metric + ` DESC
        LIMIT 5
    `
	rows, err := s.Db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*pb.TaskStatsResponse
	for rows.Next() {
		var (
			taskId     int32
			authorId   int32
			viewsCount uint64
			likesCount uint64
		)
		if err := rows.Scan(&taskId, &authorId, &viewsCount, &likesCount); err != nil {
			return nil, err
		}
		tasks = append(tasks, &pb.TaskStatsResponse{
			TaskId:     taskId,
			AuthorId:   authorId,
			ViewsCount: viewsCount,
			LikesCount: likesCount,
		})
	}
	return &pb.TopTasksResponse{Tasks: tasks}, nil
}

func (s *Server) GetTopUsers(ctx context.Context, req *pb.TopUsersRequest) (*pb.TopUsersResponse, error) {
	_, err := common.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	query := `
        SELECT
            author_id,
            sum(likes_count) AS likes_count
        FROM
            mv_user_likes FINAL
        GROUP BY
            author_id
        ORDER BY
            likes_count DESC
        LIMIT 3
    `
	rows, err := s.Db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*pb.UserStats
	for rows.Next() {
		var (
			userId     int32
			likesCount uint64
		)
		if err := rows.Scan(&userId, &likesCount); err != nil {
			return nil, err
		}
		users = append(users, &pb.UserStats{
			UserId:     userId,
			LikesCount: likesCount,
		})
	}
	return &pb.TopUsersResponse{Users: users}, nil
}
