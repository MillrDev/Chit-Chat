package main

import (
	pb "MA3/grpc"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
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
		timestamp:   1,
	}
}

func (s *chitChatServer) Subscribe(_ *pb.Empty, stream pb.ChitChatService_SubscribeServer) error {
	id := s.registerSubscriber()
	defer s.unregisterSubscriber(id)

	fmt.Printf("Client %d subscribed\n", id)
	ch := s.subscribers[id]

	text := fmt.Sprintf("Client %d has joined the chat", id)

	_, err := s.Publish(context.Background(), &pb.MessageRequest{Text: text})

	if err != nil {
		log.Printf("[Server][Error]Error publishing client joined the chat: %v\n", err)
	}

	go func() {
		<-stream.Context().Done()
		fmt.Printf("Client %d disconnected\n", id)
		s.unregisterSubscriber(id)

		leaveMsg := fmt.Sprintf("Client %d has left the chat", id)
		_, err := s.Publish(context.Background(), &pb.MessageRequest{Text: leaveMsg})
		if err != nil {
			log.Printf("[Server][Error] Error publishing leave message: %v\n", err)
		}
	}()

	for msg := range ch {
		if err := stream.Send(msg); err != nil {
			fmt.Printf("Error sending to client %d: %v\n", id, err)
			return nil
		}
	}
	return nil
}

func (s *chitChatServer) Publish(ctx context.Context, msg *pb.MessageRequest) (*pb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	parts := strings.SplitN(msg.Text, ",", 2)
	text := strings.TrimSpace(parts[0])
	clientTime := 0
	if len(parts) > 1 {
		clientTime, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
		s.timestamp = max(s.timestamp, clientTime) + 1 //increases timestamp for the receival of a message
		log.Printf("[Server][Receive] Event=Publish | from= A Client | Message=\"%s\" | Lamport=%d", text, s.timestamp)
	}

	fmt.Printf("Broadcasting: %s\n", text)

	s.timestamp++ //increase timestamp for the sending of a message
	for id, ch := range s.subscribers {
		m := *msg
		m.Text = fmt.Sprintf("%s,%d", text, s.timestamp)
		select {
		case ch <- &m:
		default:
		}
		log.Printf("[Server][Send] Event=Broadcast | To=ClientID=%d | Message=\"%s\" | Lamport=%d", id, text, s.timestamp)
	}

	return &pb.Empty{}, nil
}

func (s *chitChatServer) registerSubscriber() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.timestamp++
	s.nextID++
	id := s.nextID
	s.subscribers[id] = make(chan *pb.MessageRequest, 10)
	log.Printf("[Server][Join] ClientID=%d | Event=ClientJoined | Message=\"Client %d joined the chat\" | Lamport=%d ", id, id, s.timestamp)
	return id
}

func (s *chitChatServer) unregisterSubscriber(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timestamp++

	if ch, ok := s.subscribers[id]; ok {
		close(ch)
		delete(s.subscribers, id)
		log.Printf("[Server][Leave] Client id %d left the chat | Lamport=%d", id, s.timestamp)
	}
}

func main() {
	lis, err := net.Listen("tcp", ":5060")
	file, err := os.OpenFile("logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(file)

	grpcServer := grpc.NewServer()
	pb.RegisterChitChatServiceServer(grpcServer, newServer())

	fmt.Println("ChitChat Server running on port 5060...")
	log.Printf("[Server][Listening] Server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("[Server][Fail]Failed to serve: %v", err)
	}
	log.Printf("[Server][Shutdown] Server shutting down")
}
