package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "go-chat-terminal/gen/proto/v1"
	"go-chat-terminal/service"
	"go-chat-terminal/service/common"
	"go-chat-terminal/terminal"
)

var startCommand = &cobra.Command{
	Use: "start",
	Run: daemonStart,
}

func init() {
	rootCmd.AddCommand(startCommand)
}

func daemonStart(cmd *cobra.Command, args []string) {
	// reader := bufio.NewReader(os.Stdin)
	//
	// fmt.Println("终端GPT！")
	//
	// for {
	// 	fmt.Print("请输入一段文本：")
	// 	text, _ := reader.ReadString('\n')
	// 	chat.GetInstance().SetMessage(text).Chat()
	// 	input, _ := reader.ReadString('\n')
	// 	if input == "n\n" {
	// 		break
	// 	}
	// }
	//
	// fmt.Println("程序结束！")

	status := func() int {
		return execute()
	}()

	os.Exit(status)

}

var defaultPort = 8888
var defaultAddr = "localhost:8888"

func execute() int {
	var server service.Server

	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", defaultPort))
	if err != nil {
		log.Fatal(err)
		return 1
	}
	fmt.Printf("tcp :%s start \n", lis.Addr())

	defer lis.Close()
	server = common.NewServerImpl(
		&service.Config{
			Listener: lis,
		})

	go server.Run()

	var status int
	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	tc := pb.NewTerminalClient(conn)

	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t := terminal.New()
	t.SetTerminalClient(tc).SetContext(ctx).SetConn(conn)
	status, err = t.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if !strings.Contains(err.Error(), "exited") {
			return 1
		}
		return 0
	}
	return status
}
