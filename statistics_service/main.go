package main

import (
	"context"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/ClickHouse/clickhouse-go/v2"

	pb "statistics_service/proto"

	"common"
)

type server struct {
	pb.UnimplementedStatsServiceServer
	db clickhouse.Conn
}

func (s *server) GetTaskStats(ctx context.Context, req *pb.TaskRequest) (*pb.TaskStatsResponse, error) {
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
			countIf(event_type = 'LIKE') AS likes_count,
			countIf(event_type = 'VIEW') AS views_count
        FROM
			task_events FINAL
        WHERE
            task_id = ?
		GROUP BY task_id
    `
	if err := s.db.QueryRow(ctx, query, req.TaskId).Scan(&likesCount, &viewsCount); err != nil {
		return nil, err
	}
	return &pb.TaskStatsResponse{
		TaskId:     req.TaskId,
		ViewsCount: viewsCount,
		LikesCount: likesCount,
	}, nil
}

func (s *server) GetTopTasks(ctx context.Context, req *pb.TopTasksRequest) (*pb.TopTasksResponse, error) {
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
			countIf(event_type = 'LIKE') AS likes_count,
			countIf(event_type = 'VIEW') AS views_count
		FROM
			task_events FINAL
		GROUP BY
			(task_id, author_id)
        ORDER BY
            ` + metric + ` DESC
        LIMIT 5
    `
	rows, err := s.db.Query(ctx, query)
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
		if err := rows.Scan(&taskId, &authorId, &likesCount, &viewsCount); err != nil {
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

func (s *server) GetTopUsers(ctx context.Context, req *pb.TopUsersRequest) (*pb.TopUsersResponse, error) {
	_, err := common.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	query := `
        SELECT
            author_id,
            countIf(event_type = 'LIKE') AS likes_count
        FROM
            task_events FINAL
        GROUP BY
            author_id
        ORDER BY
            likes_count DESC
        LIMIT 3
    `
	rows, err := s.db.Query(ctx, query)
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

func main() {
	ctx := context.Background()
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"clickhouse:9000"},
		Auth: clickhouse.Auth{
			Database: os.Getenv("CLICKHOUSE_DB"),
			Username: os.Getenv("CLICKHOUSE_USER"),
			Password: os.Getenv("CLICKHOUSE_PASSWORD"),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := conn.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterStatsServiceServer(grpcServer, &server{db: conn})
	log.Printf("server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
