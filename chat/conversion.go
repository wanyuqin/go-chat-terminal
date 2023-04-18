package chat

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/sashabaranov/go-openai"
)

var defaultKey = "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"

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

func (c *Conversion) Chat() {
	key := os.Getenv("CHAT_KEY")
	if key == "" {
		key = defaultKey
	}
	client := openai.NewClient(key)
	ctx := context.Background()
	ocm := make([]openai.ChatCompletionMessage, 0, len(c.GetBody()))
	for _, v := range c.GetBody() {
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
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)

	}
	defer stream.Close()

	fmt.Printf("\n")
	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("\n按回车继续....")
			return
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return
		}
		// 发送消息
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Printf(response.Choices[0].Delta.Content)
	}
}
