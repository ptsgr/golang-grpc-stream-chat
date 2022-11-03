package main

import (
	"log"
	"net"
	"os"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	gServer "github.com/ptsgr/golang-grpc-stream-chat/internal/grpc"
	"github.com/ptsgr/golang-grpc-stream-chat/internal/service"
)

const (
	grpcPort = "grpc.port"
)

func main() {
	if err := initConfig(); err != nil {
		log.Fatal(err)
	}
	logger := grpclog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
	svc := service.NewService(logger)

	listener, err := net.Listen("tcp", ":"+viper.GetString(grpcPort))
	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(opts...)

	gServer.RegisterChatServer(grpcServer, svc)
	logger.Infoln("Starting grpc server on port: " + viper.GetString(grpcPort))
	logger.Fatal(grpcServer.Serve(listener))
}

func initConfig() error {
	viper.AddConfigPath("config")
	viper.SetConfigName("local")
	return viper.ReadInConfig()
}
