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
	groups      map[string][]string
	logger      grpclog.LoggerV2
}

func NewService(logger grpclog.LoggerV2) *Service {
	var connections []*Connection
	groups := make(map[string][]string)
	return &Service{
		connections: connections,
		groups:      groups,
		logger:      logger,
	}
}

func (s *Service) addConnection(conn *Connection) {
	s.connections = append(s.connections, conn)
}

func (s *Service) getGroupMember(grp, usr string) int {
	for k, v := range s.groups[grp] {
		if v == usr {
			return k
		}
	}
	return -1
}

func (s *Service) removeGroupMember(grp string, index int) {
	s.groups[grp] = append(s.groups[grp][:index], s.groups[grp][index+1:]...)
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
func (s *Service) JoinGroupChat(ctx context.Context, group *grpc.Group) (*grpc.Response, error) {
	s.logger.Infof(" --- User \"%s\" try to join group: \"%s\" ---", group.Username, group.Name)
	if _, ok := s.groups[group.Name]; !ok {
		return &grpc.Response{
			Error: grpc.ErrGroupNotFound,
		}, nil
	}
	if s.getGroupMember(group.Name, group.Username) < 0 {
		s.groups[group.Name] = append(s.groups[group.Name], group.Username)
	}
	return &grpc.Response{}, nil
}
func (s *Service) LeftGroupChat(ctx context.Context, group *grpc.Group) (*grpc.Response, error) {
	s.logger.Infof(" --- User \"%s\" try to left group: \"%s\" ---", group.Username, group.Name)
	if _, ok := s.groups[group.Name]; !ok {
		return &grpc.Response{
			Error: grpc.ErrGroupNotFound,
		}, nil
	}
	if len(s.groups[group.Name]) <= 1 && s.groups[group.Name][0] == group.Username {
		s.logger.Infof(" --- Group \"%s\" deleted ---", group.Name)
		delete(s.groups, group.Name)
	} else {
		memberIndex := s.getGroupMember(group.Name, group.Username)
		if memberIndex < 0 {
			return &grpc.Response{
				Error: grpc.ErrNotGroupMember,
			}, nil
		}
		s.removeGroupMember(group.Name, memberIndex)
		s.logger.Infof(" --- User \"%s\" successfully left group \"%s\" ---", group.Username, group.Name)
	}
	return &grpc.Response{}, nil
}
func (s *Service) CreateGroupChat(ctx context.Context, group *grpc.Group) (*grpc.Response, error) {
	s.logger.Infof(" --- Create group: \"%s\" ---", group.Name)
	if _, ok := s.groups[group.Name]; ok {
		return &grpc.Response{
			Error: grpc.ErrGroupExists,
		}, nil
	}
	s.groups[group.Name] = []string{group.Username}
	return &grpc.Response{}, nil
}
