package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go-chat-terminal/config"
	pb "go-chat-terminal/gen/proto/v1"
	"go-chat-terminal/service"
	"go-chat-terminal/service/common"
	"go-chat-terminal/terminal"
	"go-chat-terminal/terminal/colorize"
)

var (
	key  string // openAI key
	port int64  // grpc port
)

var startCommand = &cobra.Command{
	Use:   "start",
	Short: "启动一个与ChatGPT的对话",
	Long: "启动一个与ChatGPT的对话,默认会监听8888端口，可以通过 port来指定启动端口，\n" +
		"同时关于OpenAI Key, 你可以通过key 来指定，或者添加环境变量GCHAT_KEY来进行指定",
	PreRunE: preRunE,
	Run:     daemonStart,
}

func init() {
	startCommand.Flags().StringVarP(&key, "key", "k", "", "关于OpenAI Key, 你可以通过key 来指定，或者添加环境变量GCHAT_KEY来进行指定")
	startCommand.Flags().Int64VarP(&port, "port", "p", 8888, "grpc listening port")
	rootCmd.AddCommand(startCommand)
}

var logo = "\n ██████╗        ██████╗██╗  ██╗ █████╗ ████████╗\n" +
	"██╔════╝       ██╔════╝██║  ██║██╔══██╗╚══██╔══╝\n" +
	"██║  ███╗█████╗██║     ███████║███████║   ██║   \n" +
	"██║   ██║╚════╝██║     ██╔══██║██╔══██║   ██║   \n" +
	"╚██████╔╝      ╚██████╗██║  ██║██║  ██║   ██║   \n " +
	"╚═════╝        ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   \n   " +
	"                                             \n"

func preRunE(cmd *cobra.Command, args []string) error {
	if key == "" && os.Getenv("GCHAT_KEY") == "" {
		return errors.New("unknown openai key")
	}

	if key == "" {
		key = os.Getenv("GCHAT_KEY")
	}

	fmt.Println(logo)

	cfg := config.GetConfig()
	cfg.OpenAIKey = key
	cfg.Port = port

	fmt.Printf(colorize.FgHiGreen(fmt.Sprintf("OpenAI Key: %s\n", key)))
	fmt.Printf(colorize.FgHiGreen(fmt.Sprintf("Grpc port: %d\n", port)))
	fmt.Println()

	return nil
}

func daemonStart(cmd *cobra.Command, args []string) {
	status := func() int {
		return execute()
	}()

	os.Exit(status)

}

func execute() int {
	var server service.Server

	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", config.GetConfig().Port))
	if err != nil {
		log.Fatal(err)
		return 1
	}

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
	ctx := context.Background()

	t := terminal.New()
	t.SetTerminalClient(tc).SetContext(ctx).SetConn(conn)
	status, err = t.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if !strings.Contains(err.Error(), "exited") {
			return 1
		}
		server.Stop()
		return 0
	}

	return status
}
