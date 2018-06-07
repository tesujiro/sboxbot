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
		d, err := newDockerContainer(ctx, "centos", []string{"/bin/bash"})
		if err != nil {
			t.Errorf("create container error: %v", err)
			continue
		}
		defer d.finalize()

		if err := d.run(ctx); err != nil {
			t.Errorf("run error: %v", err)
			continue
		}
		if err := d.exec(c.commands); err != nil {
			t.Errorf("exec error: %v", err)
			continue
		}
		actual, err := d.exit()
		if err != nil {
			t.Errorf("exit error: %v", err)
			continue
		}

		if c.expected != "" && actual != c.expected {
			t.Errorf("got %v\nwant %v", actual, c.expected)
		}
	}
}

func TestCommitContainer(t *testing.T) {
	cases := []struct {
		commands string
		expected string
		image    string
	}{
		{commands: fmt.Sprintf("echo hello world!\n"), expected: fmt.Sprintf("hello world!\n"), image: "sbox_test_image"},
		//{commands: fmt.Sprintf("while : \ndo\n:\ndone\n"), expected: fmt.Sprintf("exit error: context deadline exceeded")},
	}

	for _, c := range cases {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		d, err := newDockerContainer(ctx, "centos", []string{"/bin/bash"})
		if err != nil {
			t.Errorf("create container error: %v", err)
			continue
		}
		defer d.finalize()

		if err := d.run(ctx); err != nil {
			t.Errorf("run error: %v", err)
			continue
		}

		if err := d.exec(c.commands); err != nil {
			t.Errorf("exec error: %v", err)
			continue
		}
		actual, err := d.exit()
		if err != nil {
			t.Errorf("exit error: %v", err)
			continue
		}
		if c.expected != "" && actual != c.expected {
			t.Errorf("got %v\nwant %v", actual, c.expected)
			continue
		}

		if err := d.commit("sbox_commit_01"); err != nil {
			t.Errorf("container commit to image error: %v", err)
		}

		if err := d.removeImage(); err != nil {
			t.Errorf("remove image error: %v", err)
		}
	}
}
