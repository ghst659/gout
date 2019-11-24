package gout

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestMergeChan(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cline := []string{
		"dash",
		"-c",
		"echo \"ego sum abbas\"; sleep 1 ; echo eee >&2 ; sleep 1; echo bar; sleep 2 ; echo mumble",
	}
	o, e, err := RunOutputs(ctx, cline)
	if err != nil {
		t.Fatalf("fatal error: %v", err)
	}
	for mLine := range MergeChan(ctx, o, e) {
		t.Logf("merged: %q", mLine)
	}
}

func TestRunOutputs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cline := []string{
		"dash",
		"-c",
		"echo foo ; sleep 1 ; echo \"o fortuna;\" >&2 ; sleep 1; echo bar; sleep 2 ; echo mumble",
	}
	o, e, err := RunOutputs(ctx, cline)
	if err != nil {
		t.Fatalf("fatal error: %v", err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go printStream(t, &wg, "stdout", o)
	go printStream(t, &wg, "stderr", e)
	wg.Wait()
}

func printStream(t *testing.T, wg *sync.WaitGroup, tag string, ch <-chan string) {
	defer wg.Done()
	for line := range ch {
		t.Logf("%s: %q", tag, line)
	}
	t.Logf("%s closed", tag)
}
