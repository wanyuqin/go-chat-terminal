package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"go-chat-terminal/chat"
)

var startCommand = &cobra.Command{
	Use: "start",
	Run: daemonStart,
}

func init() {
	rootCmd.AddCommand(startCommand)
}

func daemonStart(cmd *cobra.Command, args []string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("终端GPT！")

	for {
		fmt.Print("请输入一段文本：")
		text, _ := reader.ReadString('\n')
		chat.GetInstance().SetMessage(text).Chat()
		input, _ := reader.ReadString('\n')
		if input == "n\n" {
			break
		}
	}

	fmt.Println("程序结束！")
}
