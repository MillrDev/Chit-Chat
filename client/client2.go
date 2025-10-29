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
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
)

func main() {
	timestamp := 0

	fmt.Println("Enter address you want to connect to: ")
	var address string
	fmt.Scanln(&address)

	conn, err := grpc.Dial(address+":5060", grpc.WithInsecure())
	log.Printf("[Client][Connect] Connected to server")
	if err != nil {
		log.Fatalf("[Client][Error] did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewChitChatServiceClient(conn)

	go subscribeForMessages2(client, &timestamp)
	timestamp++

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
		timestamp++
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		textTimestamp := fmt.Sprintf("%s,%d", text, timestamp)
		log.Printf("[Client][Publish] Published message: %s at timestamp: %d", strings.TrimSpace(text), timestamp)
		_, err := client.Publish(ctx, &pb.MessageRequest{Text: textTimestamp})
		cancel()
		if err != nil {
			log.Printf("[Client][Error]Error publishing message: %v\n", err)
		}
	}
}

// Listens for broadcasted messages and prints them
func subscribeForMessages2(client pb.ChitChatServiceClient, timestamp *int) {
	stream, err := client.Subscribe(context.Background(), &pb.Empty{})
	if err != nil {
		log.Fatalf("[Client][Fail] Failed to subscribe: %v", err)
	}

	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("[Client][Close] Stream closed: %v", err)
			return
		}
		parts := strings.Split(msg.GetText(), ",")
		message := parts[0]

		serverTime, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
		*timestamp = max(*timestamp, serverTime) + 1

		fmt.Printf("\n [" + strconv.Itoa(*timestamp) + "] " + message + "\n")
		fmt.Print("> ")
		log.Println("[Client][Receive] Received message: " + message + " at timestamp: " + strconv.Itoa(*timestamp))
	}
}
