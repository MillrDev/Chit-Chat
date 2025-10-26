package main

import (
	pb "MA3/grpc"
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"google.golang.org/grpc"
)

type chitChatServer struct {
	pb.UnimplementedChitChatServiceServer
	mu          sync.Mutex
	subscribers map[int]chan *pb.MessageRequest
	nextID      int
	timestamp   int
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

	text := fmt.Sprintf("Client %d has joined the chat", id)
	_, err := s.Publish(context.Background(), &pb.MessageRequest{Text: text})

	if err != nil {
		log.Printf("[Server][Error]Error publishing message: %v\n", err)
	}

	for msg := range ch {
		if err := stream.Send(msg); err != nil {
			fmt.Printf("Client %d disconnected: %v\n", id, err)
			text := fmt.Sprintf("Client %d has left the chat", id)
			_, err := s.Publish(context.Background(), &pb.MessageRequest{Text: text})

			if err != nil {
				log.Printf("[Server][Error]Error publishing message: %v\n", err)
			}
			return nil
		}
	}
	return nil
}

// Called whenever a client publishes a message
func (s *chitChatServer) Publish(ctx context.Context, msg *pb.MessageRequest) (*pb.Empty, error) {
	fmt.Printf("Broadcasting: %s\n", msg.Text)
	text := strings.TrimSpace(msg.Text)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Send message to all connected subscribers
	for id, ch := range s.subscribers {
		s.timestamp++
		m := *msg // copy the value
		m.Text = fmt.Sprintf("%s,%d", text, s.timestamp)
		select {
		case ch <- &m:
		default:
			// Drop message if subscriber channel is full
		}
		log.Printf("[Server][Send] Sent message %s to client %d", text, id)
	}

	return &pb.Empty{}, nil
}

// --- Helpers for managing subscribers ---

func (s *chitChatServer) registerSubscriber() int {
	s.timestamp++
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	id := s.nextID
	s.subscribers[id] = make(chan *pb.MessageRequest, 10)
	log.Printf("[Server][Join]New client id %d joined the chat", id)
	return id
}

func (s *chitChatServer) unregisterSubscriber(id int) {
	s.timestamp++
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.subscribers[id]; ok {
		close(ch)
		delete(s.subscribers, id)
		log.Printf("[Server][Leave] Client id %d left the chat", id)
	}
}

// --- Main ---

func main() {
	lis, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("[Server][Fail] Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterChitChatServiceServer(grpcServer, newServer())

	fmt.Println("💬 ChitChat Server running on port 5050...")
	log.Printf("[Server][Listening] Server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("[Server][Fail]Failed to serve: %v", err)
	}
	log.Printf("[Server][Shutdown] Server shutting down")
}
