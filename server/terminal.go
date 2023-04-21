package server

import (
	"fmt"

	"go-chat-terminal/chat"
	pb "go-chat-terminal/gen/proto/v1"
)

type TerminalServer struct {
	pb.UnimplementedTerminalServer
}

func NewTerminalServer() *TerminalServer {
	return &TerminalServer{}
}

func (t *TerminalServer) Chat(in *pb.ChatRequest, stream pb.Terminal_ChatServer) error {
	cb := chat.NewConversationBody(in.GetQuestion())
	fmt.Println(cb)
	return nil
}
