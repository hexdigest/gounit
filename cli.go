package gounit

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type CLI struct {
	r io.Reader
	w io.Writer
}

func NewCLI(r io.Reader, w io.Writer) *CLI {
	return &CLI{r: r, w: w}
}

func (c *CLI) Read(prompt string, dest interface{}) error {
	fmt.Fprintf(c.w, "%s:\n", prompt)
	_, err := fmt.Fscanln(c.r, dest)
	return err
}

func (c *CLI) ReadUntil(prompt, until string) ([]byte, error) {
	fmt.Fprintf(c.w, "%s followed by %s:\n", prompt, until)
	scanner := bufio.NewScanner(c.r)
	scanner.Split(splitByDelimiter(until))
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan failed: %v", err)
	}

	return scanner.Bytes(), nil
}

func splitByDelimiter(delim string) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.Index(data, []byte(delim)); i >= 0 {
			return i + len([]byte(delim)), data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	}
}
