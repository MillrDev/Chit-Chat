package server

import (
	pb "MA3/grpc"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

/**
- broadcasts alle beskeder til alle aktive deltagere og vedhæfter en Lamport logical timestamp til hver broadcast (inkl. join/leave-notifikationer)
*/

type ChitChatServiceServer struct {
	pb.UnimplementedChitChatServiceServer
	subscribers map[int]chan *pb.MessageRequest
}

func (s *ChitChatServiceServer) Publish(ctx context.Context, in *string) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func (s *ChitChatServiceServer) Subscribe(ctx context.Context, in *pb.Empty) (*pb.MessageRequest, error) {
	return &pb.MessageRequest{}, nil
}

func main() {
	server := &ChitChatServiceServer{}
	server.start_server()

}

func (s *ChitChatServiceServer) start_server() {
	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	pb.RegisterChitChatServiceServer(grpcServer, s)
	grpcServer.Serve(listener)
	fmt.Println("Clock gRPC server is running on port 5050...")
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
