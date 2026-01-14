package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/ZaneH/keep-talking/pkg/proto"
)

type GameClient interface {
	CreateGame(ctx context.Context) (sessionID string, err error)
	GetBombs(ctx context.Context, sessionID string) ([]*pb.Bomb, error)
	SendInput(ctx context.Context, input *pb.PlayerInput) (*pb.PlayerInputResult, error)
	Close() error
}

type grpcClient struct {
	conn   *grpc.ClientConn
	client pb.GameServiceClient
}

func New(addr string) (GameClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	return &grpcClient{
		conn:   conn,
		client: pb.NewGameServiceClient(conn),
	}, nil
}

func (c *grpcClient) CreateGame(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := c.client.CreateGame(ctx, &pb.CreateGameRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to create game: %w", err)
	}

	return resp.SessionId, nil
}

func (c *grpcClient) GetBombs(ctx context.Context, sessionID string) ([]*pb.Bomb, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := c.client.GetBombs(ctx, &pb.GetBombsRequest{SessionId: sessionID})
	if err != nil {
		return nil, fmt.Errorf("failed to get bombs: %w", err)
	}

	return resp.Bombs, nil
}

func (c *grpcClient) SendInput(ctx context.Context, input *pb.PlayerInput) (*pb.PlayerInputResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return c.client.SendInput(ctx, input)
}

func (c *grpcClient) Close() error {
	return c.conn.Close()
}
