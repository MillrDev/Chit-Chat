/**
- Skal kunne joine, publish og leave som de vil
- Hver klient viser og logger beskeder + timestamps.
*/

package main

import (
	pb "MA3/grpc"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
)

func main() {
	// Connect to the server
	conn, err := grpc.Dial("localhost:5050", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewChitChatServiceClient(conn)

	// Start listening for broadcasts in a background goroutine
	go subscribeForMessages(client)

	// Main loop: read user input and publish messages
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Connected to ChitChat — start typing messages!")

	for {
		var text string
		var Continue = true
		for Continue {
			fmt.Print("> ")
			text, _ = reader.ReadString('\n')
			if len(text) > 128 {
				fmt.Print("The message is too long, max input length is 128 characters! \nPlease input again")
			} else {
				Continue = false
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err := client.Publish(ctx, &pb.MessageRequest{Text: text})
		cancel()

		if err != nil {
			log.Printf("Error publishing message: %v\n", err)
		}
	}
}

// Listens for broadcasted messages and prints them
func subscribeForMessages(client pb.ChitChatServiceClient) {
	stream, err := client.Subscribe(context.Background(), &pb.Empty{})
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("Stream closed: %v", err)
			return
		}
		fmt.Printf("\n📩 %s> %s", time.Now().Format("15:04:05"), msg.Text)
		fmt.Print("> ") // Reprint prompt after message
	}
}
