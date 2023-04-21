package terminal

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/derekparker/trie"
	"github.com/peterh/liner"
	"google.golang.org/grpc"

	pb "go-chat-terminal/gen/proto/v1"
)

func New() *Term {
	cmds := StartCommands()

	t := &Term{
		cmds:   cmds,
		stdout: &transcriptWriter{pw: &pagingWriter{w: os.Stdout}},
		line:   liner.NewLiner(),
	}
	t.line.SetCtrlCAborts(true)

	return t
}

type Term struct {
	ctx context.Context

	conn *grpc.ClientConn
	tc   pb.TerminalClient

	cmds   *Commands
	stdout *transcriptWriter
	line   *liner.State
	prompt string

	quittingMutex sync.Mutex
	quitting      bool
}

func (t *Term) Close() {
	t.stdout.CloseTranscript()
}

func (t *Term) SetTerminalClient(tc pb.TerminalClient) *Term {
	t.tc = tc
	return t
}

func (t *Term) SetContext(ctx context.Context) *Term {
	t.ctx = ctx
	return t
}

func (t *Term) SetConn(conn *grpc.ClientConn) *Term {
	t.conn = conn
	return t
}

// Run 终端运行gchat
func (t *Term) Run() (int, error) {
	defer t.Close()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT)
	go t.signalGuard(ch)
	// 用于极速前缀/模糊字符串搜索的数据结构和相关算法
	cmds := trie.New()

	for _, cmd := range t.cmds.cmds {
		for _, alias := range cmd.aliases {
			cmds.Add(alias, nil)
		}
	}

	t.line.SetCompleter(func(line string) (c []string) {
		cmd := t.cmds.Find(strings.Split(line, " ")[0])
		switch cmd.aliases[0] {
		case "nullcmd", "nocmd":
			// 模糊搜索
			commands := cmds.FuzzySearch(strings.ToLower(line))
			c = append(c, commands...)

		}
		return c
	})

	fmt.Println("Type 'help' for list of commands.")

	var lastCmd string

	_ = t.conn.GetState()

	for {

		cmdstr, err := t.promptForInput()
		if err != nil {
			if err == io.EOF {
				fmt.Fprintln(t.stdout, "exit")
			}
			return 1, fmt.Errorf("Prompt for input failed.\n")

		}

		if strings.TrimSpace(cmdstr) == "" {
			cmdstr = lastCmd
		}

		lastCmd = cmdstr

		err = t.cmds.Call(cmdstr, t)
		if err != nil {
			if strings.Contains(err.Error(), "exited") {
				fmt.Fprintln(os.Stderr, err.Error())
			} else {
				t.quittingMutex.Lock()
				quitting := t.quitting
				t.quittingMutex.Unlock()
				if quitting {
					// return t.handleExit()
				}
				fmt.Fprintf(os.Stderr, "Command failed: %s\n", err)
			}
		}
	}

}

// 监听sign
func (t *Term) signalGuard(ch <-chan os.Signal) {
	for range ch {
		fmt.Fprintf(t.stdout, "received SIGINT, stopping process (will not forward signal)\n")
	}
}

func (t *Term) promptForInput() (string, error) {
	l, err := t.line.Prompt(t.prompt)
	if err != nil {
		return "", err
	}

	l = strings.TrimSuffix(l, "\n")
	if l != "" {
		t.line.AppendHistory(l)
	}

	return l, nil
}
