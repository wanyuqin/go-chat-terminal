package terminal

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	pb "go-chat-terminal/gen/proto/v1"
	"go-chat-terminal/server"
	"go-chat-terminal/terminal/colorize"
)

type Commands struct {
	cmds []command
	ts   *server.TerminalServer
}

type cmdFunc func(t *Term, args string) error

type command struct {
	aliases []string
	helpMsg string
	cmdFn   cmdFunc
}

type byFirstAlias []command

func (b byFirstAlias) Len() int {
	return len(b)
}

func (b byFirstAlias) Less(i, j int) bool {
	return b[i].aliases[0] < b[j].aliases[0]
}

func (b byFirstAlias) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// StartCommands 返回默认的命令定义
func StartCommands() *Commands {
	c := &Commands{}
	c.cmds = []command{
		{aliases: []string{"help"}, cmdFn: c.help, helpMsg: "查看可以使用的命令"},
		{aliases: []string{"save"}, cmdFn: c.save, helpMsg: "以markdown 形式保存当前对话的所有信息"},
		{aliases: []string{"chat", "c"}, cmdFn: c.chat, helpMsg: "输入对话内容  eg: c `用go实现一个hello world的demo`"},
		{aliases: []string{"refresh"}, cmdFn: c.refresh, helpMsg: "清空之前对话中的上下文"},
	}

	sort.Sort(byFirstAlias(c.cmds))

	return c
}

// 帮助指令
func (c *Commands) help(t *Term, args string) error {
	if args != "" {
		for _, cmd := range c.cmds {
			for _, alias := range cmd.aliases {
				if alias == args {
					fmt.Fprintln(t.stdout, cmd.helpMsg)
					return nil
				}
			}
		}
	}

	fmt.Fprintln(t.stdout, colorize.FgHiGreen("The following commands are available:"))

	w := new(tabwriter.Writer)
	w.Init(t.stdout, 0, 8, 0, '-', 0)
	for _, cmd := range c.cmds {

		h := cmd.helpMsg
		if idx := strings.Index(h, "\n"); idx >= 0 {
			h = h[:idx]
		}
		if len(cmd.aliases) > 1 {
			fmt.Fprintf(w, colorize.FgHiGreen(fmt.Sprintf("    %s (alias: %s) \t %s\n", cmd.aliases[0], strings.Join(cmd.aliases[1:], " | "), h)))
		} else {
			fmt.Fprintf(w, colorize.FgHiGreen(fmt.Sprintf("    %s \t %s\n", cmd.aliases[0], h)))
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}

	fmt.Fprintln(t.stdout)
	return nil

}

// save 保存问题和答案到文件
func (c *Commands) save(t *Term, args string) error {
	return t.SaveConversion(args)

}

func (c *Commands) chat(t *Term, args string) error {
	stream, err := t.tc.Chat(t.ctx, &pb.ChatRequest{
		Question: args,
	})

	if err != nil {
		return err
	}
	answer := make([]string, 0, 1024)
	for {

		r, err := stream.Recv()
		if err != nil {
			break
		}
		answer = append(answer, r.Answer)
		fmt.Printf(colorize.FgHiCyan(r.Answer))
	}
	if err == nil {
		t.cacheAnswer(strings.Join(answer, ""))
		t.cacheQuestion(args)
	}

	return nil
}

func (c *Commands) refresh(t *Term, args string) error {
	_, err := c.ts.Refresh(t.ctx, &pb.RefreshRequest{})
	if err != nil {
		return err
	}
	return nil
}

// Call 执行一个指令
func (c *Commands) Call(cmdStr string, t *Term) error {
	vals := strings.SplitN(strings.TrimSpace(cmdStr), " ", 2)
	cmdname := vals[0]
	var args string
	if len(vals) > 1 {
		args = strings.TrimSpace(vals[1])
	}
	fc := c.Find(cmdname)
	if fc.match("generate_chat") {
		// 不处理空格切割 这样保证输入的空格不会被干扰
		return fc.cmdFn(t, cmdStr)
	}
	return c.Find(cmdname).cmdFn(t, args)
}

// Find 查找输入的指令
func (c *Commands) Find(cmdStr string) command {
	if cmdStr == "" {
		return command{
			aliases: []string{"nullcmd"}, cmdFn: nullCommand,
		}
	}

	for _, v := range c.cmds {
		if v.match(cmdStr) {
			return v
		}

	}

	return command{aliases: []string{"generate_chat"}, cmdFn: c.chat}
}

func nullCommand(t *Term, args string) error {
	return nil
}

var errNoCmd = errors.New("command not available")

func noCmdAvailable(t *Term, args string) error {
	return errNoCmd
}

func (c command) match(cmdstr string) bool {
	for _, v := range c.aliases {
		if v == cmdstr {
			return true
		}
	}
	return false
}
