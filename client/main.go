package main

import (
	"bufio"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"

	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/xans-me/grpc-chat-docker/protobuff"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var client protobuff.BroadcastClient
var wait *sync.WaitGroup

func init() {
	wait = &sync.WaitGroup{}
}

func connect(user *protobuff.User) error {
	var streamError error

	stream, err := client.CreateStream(context.Background(), &protobuff.Connect{
		User:   user,
		Active: true,
	})

	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}

	wait.Add(1)
	go func(str protobuff.Broadcast_CreateStreamClient) {
		defer wait.Done()

		for {
			msg, err := str.Recv()
			if err != nil {
				streamError = fmt.Errorf("error reading message: %v", err)
				break
			}

			fmt.Printf("%v \t: %s\n", msg.User.Name, msg.Content)

		}
	}(stream)

	return streamError
}

func main() {
	timestamp := time.Now()
	done := make(chan int)

	name := flag.String("N", "Anon", "The name of the user")
	flag.Parse()

	id := sha256.Sum256([]byte(timestamp.String() + *name))

	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Couldnt connect to service: %v", err)
	}

	client = protobuff.NewBroadcastClient(conn)
	user := &protobuff.User{
		Id:   hex.EncodeToString(id[:]),
		Name: *name,
	}

	_ = connect(user)

	wait.Add(1)
	go func() {
		defer wait.Done()

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			msg := &protobuff.Message{
				Id:        user.Id,
				Content:   scanner.Text(),
				Timestamp: timestamp.String(),
				User:      user,
			}

			_, err := client.BroadcastMessage(context.Background(), msg)
			if err != nil {
				fmt.Printf("Error Sending Message: %v", err)
				break
			}
		}

	}()

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
}
