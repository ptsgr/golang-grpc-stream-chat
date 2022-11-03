package service

import (
	"context"
	"sync"

	"google.golang.org/grpc/grpclog"

	"github.com/ptsgr/golang-grpc-stream-chat/internal/grpc"
)

type Connection struct {
	stream   grpc.Chat_CreateDirectStreamServer
	username string
	error    chan error
}

type Service struct {
	connections []*Connection
	logger      grpclog.LoggerV2
}

func NewService(logger grpclog.LoggerV2) *Service {
	var connections []*Connection
	return &Service{
		connections: connections,
		logger:      logger,
	}
}

func (s *Service) addConnection(conn *Connection) {
	s.connections = append(s.connections, conn)
}

func (s *Service) CreateDirectStream(pconn *grpc.User, stream grpc.Chat_CreateDirectStreamServer) error {
	s.logger.Infof(" --- Welcome user: \"%s\" ---", pconn.Name)
	conn := &Connection{
		stream:   stream,
		username: pconn.Name,
		error:    make(chan error),
	}
	s.addConnection(conn)
	return <-conn.error
}

func (s *Service) SendMessage(ctx context.Context, msg *grpc.Message) (*grpc.Response, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)
	for _, conn := range s.connections {
		wait.Add(1)

		go func(msg *grpc.Message, conn *Connection) {
			defer wait.Done()
			if conn.username == msg.To {
				err := conn.stream.Send(msg)
				s.logger.Info("Sending message to: ", conn.stream, " for user: ", conn.username)

				if err != nil {
					s.logger.Errorf("Error with Stream: %v - Error: %v", conn.stream, err)
					conn.error <- err
				}
			}
		}(msg, conn)

	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
	return &grpc.Response{}, nil
}
func (s *Service) JoinGroupChat(context.Context, *grpc.Group) (*grpc.Response, error) {
	return nil, nil
}
func (s *Service) LeftGroupChat(context.Context, *grpc.Group) (*grpc.Response, error) {
	return nil, nil
}
func (s *Service) CreateGroupChat(ctx context.Context, group *grpc.Group) (*grpc.Response, error) {
	return &grpc.Response{}, nil
}
