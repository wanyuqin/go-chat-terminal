package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/sashabaranov/go-openai"

	"go-chat-terminal/config"
	pb "go-chat-terminal/gen/proto/v1"
)

type TerminalServer struct {
	pb.UnimplementedTerminalServer
}

func NewTerminalServer() *TerminalServer {
	return &TerminalServer{}
}

type ConversationBody struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func NewConversationBody(content string) ConversationBody {
	return ConversationBody{
		Content: content,
		Role:    openai.ChatMessageRoleUser,
	}
}

type Conversion struct {
	mux  sync.Mutex
	Body []ConversationBody `json:"body"`
}

var (
	conversion *Conversion
	once       sync.Once
)

func GetInstance() *Conversion {
	once.Do(func() {
		conversion = &Conversion{
			Body: make([]ConversationBody, 0),
		}
	})

	return conversion
}

func (c *Conversion) SetMessage(content string) *Conversion {

	c.Body = append(c.Body, ConversationBody{
		Content: content,
		Role:    openai.ChatMessageRoleUser,
	})

	return c
}

func (c *Conversion) GetBody() []ConversationBody {
	return c.Body
}

func (t *TerminalServer) Chat(in *pb.ChatRequest, stream pb.Terminal_ChatServer) error {
	cb := NewConversationBody(in.GetQuestion())
	ci := GetInstance().SetMessage(cb.Content)
	chat(ci, stream)
	return nil
}

func chat(ci *Conversion, ps pb.Terminal_ChatServer) {
	oc := openai.NewClient(config.GetConfig().OpenAIKey)
	ocm := make([]openai.ChatCompletionMessage, 0, len(ci.GetBody()))
	for _, v := range ci.GetBody() {
		ocm = append(ocm, openai.ChatCompletionMessage{
			Role:    v.Role,
			Content: v.Content,
		})
	}
	req := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 1000,
		Messages:  ocm,
		Stream:    true,
	}
	stream, err := oc.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)

	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return
		}

		if err != nil {
			log.Println(err)
			return
		}
		// 发送消息
		ps.Send(
			&pb.ChatReply{
				Answer: response.Choices[0].Delta.Content,
			})
	}

}

func (t *TerminalServer) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshReply, error) {
	ci := GetInstance()
	ci.mux.Lock()
	defer ci.mux.Unlock()

	ci.Body = make([]ConversationBody, 0)
	fmt.Printf("\033[1;31;40m%s\033[0m", "refresh successfully")

	return &pb.RefreshReply{}, nil
}
