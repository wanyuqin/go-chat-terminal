package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"go-chat-terminal/config"
)

var gchatLogDesc = "`gchat 是一个终端连接Chat GPT 的聊天工具，可以在终端完成与Chat GPT的对话，通过命令 gchat start 即可进入对话，更多功能和在进入对话后" +
	"使用help 命令获取更多的信息`"

var rootCmd = &cobra.Command{
	Use:     "gchat",
	Short:   "gchat 一个终端chat GPT 聊天工具",
	Long:    gchatLogDesc,
	PreRunE: loadConfig,
	Run:     start,
}

func loadConfig(cmd *cobra.Command, args []string) error {
	err := config.LoadConfig()
	cfg := config.GetConfig()
	if cfg.OpenAIKey == "" {
		return errors.New("unknown openai key")
	}
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
		return
	}
}
