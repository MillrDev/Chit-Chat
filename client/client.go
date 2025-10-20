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
	"strings"
	"time"

	"google.golang.org/grpc"
)

func main() {
	timestamp := 0

	// Connect to the server
	conn, err := grpc.Dial("localhost:5050", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	timestamp++

	client := pb.NewChitChatServiceClient(conn)
	timestamp++

	// Start listening for broadcasts in a background goroutine
	go subscribeForMessages(client)
	timestamp++

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
			timestamp++
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err := client.Publish(ctx, &pb.MessageRequest{Text: text})
		cancel()
		timestamp++
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
		parts := strings.Split(msg.GetText(), ",")
		message := parts[0]
		serverTimestamp := strings.TrimSpace(parts[1])

		fmt.Printf("\n [" + serverTimestamp + "] " + message + "\n")
		fmt.Print("> ") // Reprint prompt after message
	}
}
