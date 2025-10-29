package main

import (
	pb "MA3/grpc"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
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
		log.Printf("[Lamport=%d][Server] | Event=Error | Message=%v", s.timestamp, err)
	}

	go func() {
		<-stream.Context().Done()
		fmt.Printf("Client %d disconnected\n", id)
		s.unregisterSubscriber(id)

		leaveMsg := fmt.Sprintf("Client %d has left the chat", id)
		_, err := s.Publish(context.Background(), &pb.MessageRequest{Text: leaveMsg})
		if err != nil {
			log.Printf("[Lamport=%d][Server] | Event=Error | Message= %v", s.timestamp, err)
		}
	}()

	for msg := range ch {
		if err := stream.Send(msg); err != nil {
			log.Fatalf("[Lamport=%d][Server] | Event=Error | Message= %v", s.timestamp, err)
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
		log.Printf("[Lamport=%d][Server] | Event=Receive | Message=\"%s\"", s.timestamp, text)
	}

	s.timestamp++ //increase timestamp for the sending of a message
	for id, ch := range s.subscribers {
		m := *msg
		m.Text = fmt.Sprintf("%s,%d", text, s.timestamp)
		select {
		case ch <- &m:
		default:
		}
		log.Printf("[Lamport=%d][Server] | Event=Broadcast | ReceivingClientID=%d | Message=\"%s\"", s.timestamp, id, text)
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
	log.Printf("[Lamport=%d][Server] | Event=ClientJoin | ClientID=%d | Message=\"Client %d joined the chat\"", s.timestamp, id, id)
	return id
}

func (s *chitChatServer) unregisterSubscriber(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timestamp++

	if ch, ok := s.subscribers[id]; ok {
		close(ch)
		delete(s.subscribers, id)
		log.Printf("[Lamport=%d][Server] | Event=Leave | ClientID=%d", s.timestamp, id)
	}
}

func main() {
	lis, err := net.Listen("tcp", ":5060")
	file, err := os.OpenFile("logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("[Lamport=0][Server] | Event=Error | Message= %v", err)
	}
	log.SetOutput(file)

	grpcServer := grpc.NewServer()
	pb.RegisterChitChatServiceServer(grpcServer, newServer())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Printf("[Lamport=?][Server] | Event=Shutdown | Message=Server shutting down")
		os.Exit(0)
	}()

	fmt.Println("ChitChat Server running on port 5060...")
	log.Printf("[Lamport=1][Server] | Event=Listening | Message=Server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("[Lamport=1][Server] | Event=Error | Message= %v", err)
	}
}
