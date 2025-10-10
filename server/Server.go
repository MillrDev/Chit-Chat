package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	pb "MA3/grpc"

	"google.golang.org/grpc"
)

type chitChatServer struct {
	pb.UnimplementedChitChatServiceServer
	mu          sync.Mutex
	subscribers map[int]chan *pb.MessageRequest
	nextID      int
}

func newServer() *chitChatServer {
	return &chitChatServer{
		subscribers: make(map[int]chan *pb.MessageRequest),
	}
}

// Client opens a stream to receive messages
func (s *chitChatServer) Subscribe(_ *pb.Empty, stream pb.ChitChatService_SubscribeServer) error {
	id := s.registerSubscriber()
	defer s.unregisterSubscriber(id)

	fmt.Printf("Client %d subscribed\n", id)
	ch := s.subscribers[id]

	for msg := range ch {
		if err := stream.Send(msg); err != nil {
			fmt.Printf("Client %d disconnected: %v\n", id, err)
			return nil
		}
	}
	return nil
}

// Called whenever a client publishes a message
func (s *chitChatServer) Publish(ctx context.Context, msg *pb.MessageRequest) (*pb.Empty, error) {
	fmt.Printf("Broadcasting: %s\n", msg.Text)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Send message to all connected subscribers
	for _, ch := range s.subscribers {
		select {
		case ch <- msg:
		default:
			// Drop message if subscriber channel is full
		}
	}

	return &pb.Empty{}, nil
}

// --- Helpers for managing subscribers ---

func (s *chitChatServer) registerSubscriber() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	id := s.nextID
	s.subscribers[id] = make(chan *pb.MessageRequest, 10)
	return id
}

func (s *chitChatServer) unregisterSubscriber(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.subscribers[id]; ok {
		close(ch)
		delete(s.subscribers, id)
	}
}

// --- Main ---

func main() {
	lis, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterChitChatServiceServer(grpcServer, newServer())

	fmt.Println("💬 ChitChat Server running on port 5050...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
