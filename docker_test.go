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
		//{commands: fmt.Sprintf("while : \ndo\n:\ndone\n"), expected: fmt.Sprintf("exit error: context deadline exceeded")},
	}

	for _, c := range cases {
		//ctx, cancel := context.WithCancel(context.Background())
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		//actual := execOnContainer(ctx, c.commands)
		//fmt.Println(actual)
		d := newDockerContainer()
		if err := d.run(ctx); err != nil {
			t.Errorf("run error: %v", err)
		}
		if err := d.exec(c.commands); err != nil {
			t.Errorf("exec error: %v", err)
		}
		actual, err := d.exit()
		if err != nil {
			t.Errorf("exit error: %v", err)
		}

		if c.expected != "" && actual != c.expected {
			t.Errorf("got %v\nwant %v", actual, c.expected)
		}
	}
}
