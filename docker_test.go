package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestExecOnCointainer(t *testing.T) {
	cases := []struct {
		container_image string
		container_cmd   []string
		commands        string
		expected        string
	}{
		{container_image: "centos", container_cmd: []string{"/bin/bash"}, commands: fmt.Sprintf("echo hello world!\n"), expected: fmt.Sprintf("hello world!\n")},
		//{container_image: "centos", container_cmd: "/bin/bash", commands: fmt.Sprintf("set\n")},
		//{commands: fmt.Sprintf("while : \ndo\n:\ndone\n"), expected: fmt.Sprintf("exit error: context deadline exceeded")},
		{container_image: "ankoro", container_cmd: []string{"/anko"}, commands: fmt.Sprintf("println(\"Hello Anko\")\n"), expected: fmt.Sprintf("Hello Anko")},
	}

	for _, c := range cases {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		//actual := execOnContainer(ctx, c.commands)
		//fmt.Println(actual)
		d, err := newDockerContainer(ctx, c.container_image, c.container_cmd)
		if err != nil {
			t.Errorf("create container error: %v", err)
			continue
		}
		defer d.remove(ctx)

		//fmt.Println("run")
		if err := d.run(ctx); err != nil {
			t.Errorf("run error: %v", err)
			continue
		}
		//fmt.Println("exec")
		if err := d.exec(c.commands); err != nil {
			t.Errorf("exec error: %v", err)
			continue
		}
		//time.Sleep(2 * time.Second)
		//fmt.Println("exit")
		actual, err := d.exit()
		if err != nil {
			t.Errorf("exit error: %v", err)
			continue
		}

		//fmt.Println("check")
		//if c.expected != "" && actual != c.expected {
		if !strings.Contains(actual, c.expected) {
			t.Errorf("got %v\nwant contains %v", actual, c.expected)
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
		defer d.remove(ctx)

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

		if err := d.commit(ctx, "sbox_commit_01"); err != nil {
			t.Errorf("container commit to image error: %v", err)
		}

		if err := d.remove(ctx); err != nil {
			t.Errorf("container remove error: %v", err)
		}

		if err := d.removeImage(ctx); err != nil {
			t.Errorf("remove image error: %v", err)
		}
	}
}
