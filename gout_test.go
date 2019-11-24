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
	var got []string
	for line := range MergeChan(ctx, o, e) {
		got = append(got, line)
	}
	checkSlice(t, "merged", got, []string{"ego sum abbas", "eee", "bar"})
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
	var gotOut []string
	var gotErr []string
	var wg sync.WaitGroup
	wg.Add(2)
	go fill(&wg, o, &gotOut)
	go fill(&wg, e, &gotErr)
	wg.Wait()
	checkSlice(t, "stdout", gotOut, []string{"foo", "bar"})
	checkSlice(t, "stderr", gotErr, []string{"o fortuna;"})
}

func checkSlice(t *testing.T, tag string, got, want []string) {
	for i := 0; i < len(got); i++ {
		t.Logf("%s: %q", tag, got[i])
		if got[i] != want[i] {
			t.Errorf("%s mismatch: got %s want %s", tag, got[i], want[i])
		}
	}
}

func fill(wg *sync.WaitGroup, ch <-chan string, b *[]string) {
	defer wg.Done()
	for line := range ch {
		*b = append(*b, line)
	}
}
