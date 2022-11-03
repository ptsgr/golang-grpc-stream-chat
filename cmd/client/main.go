package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	gServer "github.com/ptsgr/golang-grpc-stream-chat/internal/grpc"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

const (
	grpcPort = "grpc.port"
)

type grpcClient struct {
	chatClient gServer.ChatClient
	wg         *sync.WaitGroup
	error      chan error
}

func NewGrpcClient(chatClient gServer.ChatClient) *grpcClient {
	return &grpcClient{
		chatClient: chatClient,
		wg:         &sync.WaitGroup{},
		error:      make(chan error),
	}
}

func main() {
	if err := initConfig(); err != nil {
		log.Fatal(err)
	}
	user := &gServer.User{}
	fmt.Println(" --- Hello form golang-grpc-stream-chat --- ")
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Input your username: ")
	if scanner.Scan() {
		user.Name = scanner.Text()
	}
	if user.Name == "" {
		log.Fatalf("cannot read username")
	}

	conn, err := grpc.Dial("localhost:"+viper.GetString(grpcPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Couldnt connect to service: %v", err)
	}
	client := NewGrpcClient(gServer.NewChatClient(conn))

	client.wg.Add(1)
	go client.connectAndListen(user)

	client.wg.Add(1)
	go func() {
		defer client.wg.Done()

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			msgData := strings.SplitN(scanner.Text(), " >:", 2)
			if len(msgData) < 2 {
				fmt.Printf("error parse message, msg: %v\n", msgData)
				continue
			}
			msg := &gServer.Message{
				From:    user.Name,
				To:      msgData[0],
				Content: msgData[1],
			}

			_, err := client.chatClient.SendMessage(context.Background(), msg)
			if err != nil {
				fmt.Printf("Error Sending Message: %v", err)
				break
			}
		}

	}()

	go fmt.Println(<-client.error)
	client.wg.Wait()
	close(client.error)

}

func (c *grpcClient) connectAndListen(user *gServer.User) {
	defer c.wg.Done()
	var streamerror error

	stream, err := c.chatClient.CreateDirectStream(context.Background(), user)
	if err != nil {
		c.error <- fmt.Errorf("connection failed: %v", err)
		return
	}

	for {
		msg, err := stream.Recv()
		if err != nil {
			streamerror = fmt.Errorf("Error reading message: %v", err)
			break
		}

		fmt.Printf("%v <:%s\n", msg.From, msg.Content)
	}

	c.error <- streamerror
}

func initConfig() error {
	viper.AddConfigPath("config")
	viper.SetConfigName("local")
	return viper.ReadInConfig()
}
