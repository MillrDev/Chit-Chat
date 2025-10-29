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
	file, err := os.OpenFile("logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("[Lamport=%d][Client] | Event=Error | Message= %v", timestamp, err)
	}
	log.SetOutput(file)

	fmt.Println("Enter address you want to connect to: ")
	var address string
	fmt.Scanln(&address)

	conn, err := grpc.Dial(address+":5060", grpc.WithInsecure())
	if err != nil {
		timestamp++
		log.Fatalf("[Lamport=%d][Client] | Event=Error | Message= %v", timestamp, err)
	}
	defer conn.Close()

	client := pb.NewChitChatServiceClient(conn)
	timestamp++
	log.Printf("[Lamport=%d][Client] | Event=Connect | Message=Connected to server", timestamp)
	go subscribeForMessages(client, &timestamp)

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
		log.Printf("[Lamport=%d][Client] | Event=Publish | Msg=\"%s\"", timestamp, strings.TrimSpace(text))
		_, err := client.Publish(ctx, &pb.MessageRequest{Text: textTimestamp})
		cancel()
		if err != nil {
			log.Printf("[Lamport=%d][Client] | Event=Error | Message=%v", timestamp, err)
		}
	}
}

// Listens for broadcasted messages and prints them
func subscribeForMessages(client pb.ChitChatServiceClient, timestamp *int) {
	stream, err := client.Subscribe(context.Background(), &pb.Empty{})
	if err != nil {
		log.Fatalf("[Lamport=%d][Client] | Event=Error | Message= %v", timestamp, err)
	}

	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("[Lamport=%d][Client] | Event=StreamClosed | Message= %v", timestamp, err)
			return
		}
		parts := strings.Split(msg.GetText(), ",")
		message := parts[0]

		serverTime, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
		*timestamp = max(*timestamp, serverTime) + 1

		fmt.Printf("\n [" + strconv.Itoa(*timestamp) + "] " + message + "\n")
		fmt.Print("> ")
		log.Printf("[Lamport=%d][Client] | Event=Receive | Msg=\"%s\"", *timestamp, message)
	}
}
