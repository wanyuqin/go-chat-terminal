package terminal

import (
	"bufio"
	"io"
	"os/exec"
)

type transcriptWriter struct {
	pw   *pagingWriter
	file *bufio.Writer
}

func (t *transcriptWriter) Write(p []byte) (nn int, err error) {
	return t.pw.Write(p)
}

func (t *transcriptWriter) CloseTranscript() error {
	if t.file == nil {
		return nil
	}
	t.file.Flush()
	t.file = nil

	return nil
}

func (t *transcriptWriter) Echo(str string) {
	if t.file != nil {
		t.file.WriteString(str)
	}
}

type pagingWriter struct {
	mode     int
	w        io.Writer
	buf      []byte
	cmd      *exec.Cmd
	cmdStdin io.WriteCloser

	cancel func()
}

func (w *pagingWriter) Write(p []byte) (n int, err error) {
	switch w.mode {
	default:
		fallthrough
	case 0:
		return w.w.Write(p)

	}
}
