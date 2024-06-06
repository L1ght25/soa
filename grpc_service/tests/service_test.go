package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	pb "grpc_service/proto"
	"grpc_service/server"
)

type CustomClaims struct {
	UserID int `json:"userID"`
	jwt.StandardClaims
}

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func initDB(t *testing.T, connStr string) *sql.DB {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
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
		t.Fatalf("failed to create table: %v", err)
	}

	return db
}

func createGRPCServer(db *sql.DB) *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterTaskServiceServer(s, &server.Server{
		Db: db,
	})
	return s
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestCreateTask(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	postgresC, err := postgres.RunContainer(ctx, testcontainers.CustomizeRequest(
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		}),
	)
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}
	defer postgresC.Terminate(ctx)

	time.Sleep(time.Second * 5)

	host, err := postgresC.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}

	port, err := postgresC.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("failed to get container port: %v", err)
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=test password=test dbname=test sslmode=disable",
		host, port.Port(),
	)

	db := initDB(t, connStr)
	defer db.Close()

	lis = bufconn.Listen(bufSize)
	grpcServer := createGRPCServer(db)

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

	client := pb.NewTaskServiceClient(conn)

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

	resp, err := client.CreateTask(ctx, &pb.CreateTaskRequest{
		Title:   "Test Task",
		Content: "This is a test task.",
	})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	assert.NotNil(t, resp)
	assert.Equal(t, "Test Task", resp.Title)
	assert.Equal(t, "This is a test task.", resp.Content)
	assert.Equal(t, int32(12345), resp.CreatedByUserID)

	respGet, err := client.GetTaskById(ctx, &pb.GetTaskByIdRequest{
		TaskId: 1,
	})
	if err != nil {
		t.Fatalf("GetTaskById failed: %v", err)
	}

	assert.NotNil(t, respGet)
	assert.Equal(t, "Test Task", respGet.Title)
	assert.Equal(t, "This is a test task.", respGet.Content)
	assert.Equal(t, int32(12345), respGet.CreatedByUserID)
}
