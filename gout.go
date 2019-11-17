// Package gout runs a process and channels outputs to the caller.
package gout

import (
	"context"
	"io"
	"os/exec"
	"text/scanner"
)

// RunOutputs runs a command-line and returns channels that stream out its stdout and stderr.
func RunOutputs(ctx context.Context, cline []string) (outs, errs <-chan string, err error) {
	program, err := exec.LookPath(cline[0])
	if err != nil {
		return nil, nil, err
	}
	cmd := exec.CommandContext(ctx, program, cline[1:]...)
	oPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	ePipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}
	outs = makeChan(ctx, oPipe)
	errs = makeChan(ctx, ePipe)

	go func() {
		cmd.Wait()
	}()
	return
}

func makeChan(ctx context.Context, stream io.ReadCloser) <-chan string {
	ch := make(chan string)
	var s scanner.Scanner
	go func(ctx context.Context, s *scanner.Scanner, ch chan<- string) {
		defer close(ch)
		for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
			select {
			case ch <- s.TokenText():
			case <-ctx.Done():
				return
			}
		}
	}(ctx, s.Init(stream), ch)
	return ch
}
