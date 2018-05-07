package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestExecOnCointainer(t *testing.T) {
	cases := []struct {
		commands string
		expected string
	}{
		{commands: fmt.Sprintf("echo hello world!\n"), expected: fmt.Sprintf("hello world!\n")},
		{commands: fmt.Sprintf("set\n")},
		{commands: fmt.Sprintf("while : \ndo\n:\ndone\n"), expected: fmt.Sprintf("context deadline exceeded")},
	}

	for _, c := range cases {
		//ctx, cancel := context.WithCancel(context.Background())
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		actual := execOnContainer(ctx, c.commands)
		//fmt.Println(actual)
		if c.expected != "" && actual != c.expected {
			t.Errorf("got %v\nwant %v", actual, c.expected)
		}
	}
}
