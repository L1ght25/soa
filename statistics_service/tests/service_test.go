package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	testcontainersClickhouse "github.com/testcontainers/testcontainers-go/modules/clickhouse"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	pb "statistics_service/proto"
	"statistics_service/server"
)

type CustomClaims struct {
	UserID int `json:"userID"`
	jwt.StandardClaims
}

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func createGRPCServer(db clickhouse.Conn) *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterStatsServiceServer(s, &server.Server{
		Db: db,
	})
	return s
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func initDB(ctx context.Context, db clickhouse.Conn) error {
	err := db.Exec(ctx, "INSERT INTO task_events (task_id, event_type, author_id, user_id) VALUES (1, 1, 5, 4);")
	if err != nil {
		return fmt.Errorf("failed to execute insert: %v", err)
	}

	return nil
}

func TestCreateTask(t *testing.T) {
	ctx := context.Background()

	clickhouseC, err := testcontainersClickhouse.RunContainer(ctx, testcontainers.WithImage("clickhouse/clickhouse-server:latest"),
		testcontainersClickhouse.WithUsername("test"),
		testcontainersClickhouse.WithPassword("test"),
		testcontainersClickhouse.WithDatabase("statistics"),
		testcontainersClickhouse.WithInitScripts("../init.sql"),
	)
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}
	defer clickhouseC.Terminate(ctx)

	time.Sleep(time.Second * 5)

	host, err := clickhouseC.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}

	port, err := clickhouseC.MappedPort(ctx, "9000")
	if err != nil {
		t.Fatalf("failed to get container port: %v", err)
	}

	chConn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{host + ":" + port.Port()},
		Auth: clickhouse.Auth{
			Database: "statistics",
			Username: "test",
			Password: "test",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := chConn.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	err = initDB(ctx, chConn)
	if err != nil {
		t.Fatalf("Error while initialization of clickhouse: %v", err)
	}

	lis = bufconn.Listen(bufSize)
	grpcServer := createGRPCServer(chConn)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
	defer grpcServer.Stop()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewStatsServiceClient(conn)

	claims := CustomClaims{
		UserID: 12345,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		fmt.Println("Error signing token:", err)
		return
	}

	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(
		`x-access-token`, `Bearer `+tokenString,
	))

	respGet, err := client.GetTaskStats(ctx, &pb.TaskRequest{
		TaskId: 1,
	})
	if err != nil {
		t.Fatalf("GetTaskById failed: %v", err)
	}

	assert.NotNil(t, respGet)
	assert.Equal(t, int32(1), respGet.TaskId)
	assert.Equal(t, uint64(1), respGet.LikesCount)
	assert.Equal(t, uint64(0), respGet.ViewsCount)

	resp, err := client.GetTopUsers(ctx, &pb.TopUsersRequest{})
	if err != nil {
		t.Fatalf("GetTopUsers failed: %v", err)
	}

	assert.NotNil(t, resp)
	assert.Equal(t, 1, len(resp.Users))
	assert.Equal(t, int32(5), resp.Users[0].UserId)
	assert.Equal(t, uint64(1), resp.Users[0].LikesCount)
}
