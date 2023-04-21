package terminal

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"go-chat-terminal/server"
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
		{aliases: []string{"help"}, cmdFn: c.help, helpMsg: ""},
		{aliases: []string{"save"}, cmdFn: c.save, helpMsg: "以markdown 形式保存当前对话的所有信息"},
		{aliases: []string{"chat"}, cmdFn: c.save, helpMsg: "以markdown 形式保存当前对话的所有信息"},
	}

	sort.Sort(byFirstAlias(c.cmds))

	return c
}

// 帮助指令
func (c *Commands) help(t *Term, args string) error {
	fmt.Println("帮助操作...")
	return nil

}

func (c *Commands) save(t *Term, args string) error {
	fmt.Println("保存操作...")
	return nil
}

func (c *Commands) chat(t *Term, args error) error {
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

	return command{aliases: []string{"nocmd"}, cmdFn: noCmdAvailable}
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
