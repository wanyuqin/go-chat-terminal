package terminal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/derekparker/trie"
	"github.com/peterh/liner"
	"google.golang.org/grpc"

	pb "go-chat-terminal/gen/proto/v1"
)

func New() *Term {
	cmds := StartCommands()

	t := &Term{
		cmds:     cmds,
		stdout:   &transcriptWriter{pw: &pagingWriter{w: os.Stdout}},
		line:     liner.NewLiner(),
		prompt:   "(gchat) ",
		question: make([]string, 0, 1024),
		answer:   make([]string, 0, 1024),
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

	question []string
	answer   []string

	questionMutex sync.Mutex
	answerMutex   sync.Mutex
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

	// ch := make(chan os.Signal, 1)
	// signal.Notify(ch, syscall.SIGINT)
	// go t.signalGuard(ch)
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
		case "nullcmd":
			// 模糊搜索
			commands := cmds.FuzzySearch(strings.ToLower(line))
			c = append(c, commands...)
		case "nocmd":
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
				// TODO 处理退出信号
				t.quittingMutex.Lock()
				quitting := t.quitting
				t.quittingMutex.Unlock()
				if quitting {
					return t.handleExit()
				}
				fmt.Fprintf(os.Stderr, "Command failed: %s\n", err)
			}
		}
		t.stdout.Write([]byte("\n"))

	}

}

// 监听sign
func (t *Term) signalGuard(ch <-chan os.Signal) {
	for range ch {
		fmt.Fprintf(t.stdout, "received SIGINT, stopping process (will not forward signal)\n")
		t.quittingMutex.Lock()
		t.quitting = true
		t.quittingMutex.Unlock()

		return
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

func (t *Term) cacheQuestion(q string) {
	t.questionMutex.Lock()
	t.question = append(t.question, q)
	t.questionMutex.Unlock()
}

func (t *Term) cacheAnswer(a string) {
	t.answerMutex.Lock()
	t.answer = append(t.answer, a)
	t.answerMutex.Unlock()
}

func (t *Term) SaveConversion(args string) error {
	_, err := os.Stat(args)

	var file *os.File
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if errors.Is(err, os.ErrNotExist) {
		file, err = os.OpenFile(args, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
	} else {
		file, err = os.OpenFile(args, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		_, err = file.Seek(0, 2)
		if err != nil {
			return err
		}
	}

	defer file.Close()

	file.WriteString("\n")
	file.WriteString("\n")
	file.WriteString(fmt.Sprintf("# %s", time.Now().Format("2006-01-02 15:04:05")))
	file.WriteString("\n")
	file.WriteString("\n")
	for i := 0; i < len(t.question); i++ {
		file.WriteString(fmt.Sprintf("**问题:%d** %s", i+1, t.question[i]))
		file.WriteString("\n")
		file.WriteString("\n")
		file.WriteString(fmt.Sprintf("**回答:%d** %s", i+1, t.answer[i]))
		if strings.HasSuffix(t.answer[i], "```") {
			continue
		}
		file.WriteString("\n")
		file.WriteString("\n")

	}

	err = file.Sync()
	if err != nil {
		return err
	}

	// 刷新存储内容
	t.questionMutex.Lock()
	t.question = make([]string, 0, 1024)
	t.questionMutex.Unlock()

	t.answerMutex.Lock()
	t.answer = make([]string, 0, 1024)
	t.answerMutex.Unlock()

	fmt.Printf("\033[1;31;40m%s\033[0m", "save file successfully")

	return nil
}

func (t *Term) handleExit() (int, error) {
	err := t.conn.Close()
	if err != nil {
		fmt.Printf("grpc conn close failed:%v\n", err)
		return 1, err
	}
	return 0, errors.New("exited")
}
