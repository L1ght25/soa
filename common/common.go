package common

import (
	"context"
	"os"
	"strings"

	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	secretKey = os.Getenv("SECRET_KEY")
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

func Authenticate(ctx context.Context) (int32, error) {
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
