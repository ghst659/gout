package gout

import (
	"context"
	"sync"
	"testing"
)

func TestRunOutputs(t *testing.T) {
	ctx := context.Background()
	cline := []string{
		"/bin/bash",
		"-c",
		"echo foo ; sleep 2 ; echo eee >&2 ; sleep 1; echo bar",
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
