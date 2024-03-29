// Package gout runs a process and channels outputs to the caller.
package gout

import (
	"bufio"
	"context"
	"io"
	"log"
	"os/exec"
	"sync"
	"text/scanner"
)

// MergeChan merges strings from two channels into one.
func MergeChan(ctx context.Context, inchans ...<-chan string) <-chan string {
	out := make(chan string)
	var wg sync.WaitGroup
	for _, in := range inchans {
		wg.Add(1)
		go chanToChan(ctx, &wg, in, out)
	}
	go func() {
		defer close(out)
		wg.Wait()
	}()
	return out
}

func chanToChan(ctx context.Context, wg *sync.WaitGroup, in <-chan string, out chan<- string) {
	defer wg.Done()
	for line := range in {
		select {
		case out <- line:
		case <-ctx.Done():
			return
		}
	}
}

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
		if err := cmd.Wait(); err != nil {
			log.Printf("%s: %s", program, err.Error())
		}
	}()
	return
}

func makeChan(ctx context.Context, stream io.ReadCloser) <-chan string {
	ch := make(chan string)
	s := bufio.NewScanner(stream)
	go func(ctx context.Context, s *bufio.Scanner, ch chan<- string) {
		defer close(ch)
		for s.Scan() {
			select {
			case ch <- s.Text():
			case <-ctx.Done():
				return
			}
		}
	}(ctx, s, ch)
	return ch
}

func makeChanOld(ctx context.Context, stream io.ReadCloser) <-chan string {
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
