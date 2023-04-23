package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"go-chat-terminal/chat"
	"go-chat-terminal/config"
	"go-chat-terminal/internal/terminal/colorize"
)

var translateCommand = &cobra.Command{
	Use:     "trans",
	Short:   "翻译源码",
	Long:    "源码翻译",
	PreRunE: preTrans,
	Run:     trans,
}

var (
	input  string
	output string
)

func init() {
	translateCommand.Flags().StringVarP(&input, "input", "i", "", "指定一个代码文件")
	translateCommand.Flags().StringVarP(&output, "output", "o", "", "指定输出到某个文件")
	translateCommand.MarkFlagRequired("port")
	translateCommand.MarkFlagFilename("input", "go", "java", "python", "js")
	startCommand.AddCommand(translateCommand)
}

// 运行前检查
func preTrans(cmd *cobra.Command, args []string) error {
	err := checkKey()
	if err != nil {
		return err
	}
	if input == "" {
		return errors.New("请指定一个代码文件")
	}
	_, err = os.Stat(input)
	if err != nil {
		return err
	}
	config.GetConfig().OpenAIKey = key
	return nil
}

// 执行开始
func trans(cmd *cobra.Command, args []string) {
	fi, err := os.Stat(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, colorize.FgHiRed(err.Error()))
		return
	}
	fb, err := os.ReadFile(input)
	if err != nil {

		return
	}

	cb := chat.NewConversationBody(string(fb))

	ci := chat.GetInstance()
	fmt.Fprintln(os.Stdout, colorize.FgHiGreen("ChatGPT 回答中....."))
	ci.SetMessage(cb.Content)
	ci.SetMessage(strings.Join(args, " "))
	ci.Chat()

	if len(ci.Answer) == 0 {
		fmt.Fprintln(os.Stderr, colorize.FgHiRed("no answer!"))
		return
	}

	if output == "" {
		output = fmt.Sprintf("%s_trans.md", getFileName(fi.Name()))
	}

	_, err = os.Stat(output)

	var of *os.File

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintln(os.Stderr, colorize.FgHiRed(err.Error()))
		return
	}
	defer of.Close()

	if errors.Is(err, os.ErrNotExist) {
		of, err = os.OpenFile(output, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Fprintln(os.Stderr, colorize.FgHiRed(err.Error()))
			return
		}
	} else {
		of, err = os.OpenFile(output, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, colorize.FgHiRed(err.Error()))
			return
		}

		_, err = of.Seek(0, 2)
		if err != nil {
			fmt.Fprintln(os.Stderr, colorize.FgHiRed(err.Error()))
			return
		}
	}

	_, err = of.WriteString(strings.Join(ci.Answer, ""))
	if err != nil {
		fmt.Fprintln(os.Stderr, colorize.FgHiRed(err.Error()))
		return
	}

	err = of.Sync()
	if err != nil {
		fmt.Fprintln(os.Stderr, colorize.FgHiRed(err.Error()))
		return
	}

	fmt.Fprintln(os.Stdout, colorize.FgHiGreen(fmt.Sprintf("回答完毕，内容存放在%s", output)))

	return
}

func getFileName(f string) string {
	index := strings.LastIndex(f, ".")
	return f[:index]
}
