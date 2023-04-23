package common

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "go-chat-terminal/gen/proto/v1"
	"go-chat-terminal/server"
	"go-chat-terminal/service"
)

type ServerImpl struct {
	config *service.Config

	listener net.Listener
}

func NewServerImpl(config *service.Config) *ServerImpl {
	return &ServerImpl{
		config: config,
	}
}

func (s *ServerImpl) Run() error {

	srv := grpc.NewServer()
	pb.RegisterTerminalServer(srv, server.NewTerminalServer())
	err := srv.Serve(s.config.Listener)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil

}

func (s *ServerImpl) Stop() error {
	err := s.listener.Close()
	if err != nil {
		fmt.Println(err)
		return err

	}
	return nil
}
