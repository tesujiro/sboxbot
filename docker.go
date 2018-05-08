package main

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type instance struct {
	cmdCh    chan string
	resultCh chan string
	errCh    chan error
}

func newDockerContainer() *instance {
	c := instance{
		cmdCh:    make(chan string),
		resultCh: make(chan string),
		errCh:    make(chan error),
	}
	return &c
}

func (c *instance) exec(cmd string) error {
	c.cmdCh <- cmd
	return nil
}

func (c *instance) exit() (string, error) {
	close(c.cmdCh)
	//close(c.resultCh)
	//TODO:
	return <-c.resultCh, nil
}

//func execOnContainer(ctx context.Context, cmd string) string {
func (c *instance) run(ctx context.Context) error {

	go func() {
		cli, err := client.NewClientWithOpts()
		if err != nil {
			//panic(err)
			fmt.Printf("Container New Client ERROR: %v\n", err)
			//return err
		}

		/*
			reader, err := cli.ImagePull(ctx, "docker.io/library/alpine", types.ImagePullOptions{})
			if err != nil {
				//panic(err)
				fmt.Printf("Container Image Pull ERROR: %v\n", err)
				return fmt.Sprintf("%v", err)
			}
			io.Copy(os.Stdout, reader)
		*/

		// Create Container
		resp, err := cli.ContainerCreate(ctx, &container.Config{
			//Image:        "alpine",
			//Cmd:          []string{"/bin/ash"},
			Image:        "centos",
			Cmd:          []string{"/bin/bash"},
			OpenStdin:    true,
			StdinOnce:    true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          false,
		}, nil, nil, "")
		if err != nil {
			//panic(err)
			fmt.Printf("Container Create ERROR: %v\n", err)
			//return err
		}

		defer func() {
			if err := cli.ContainerRemove(context.Background(), resp.ID, types.ContainerRemoveOptions{}); err != nil {
				fmt.Printf("ContainerRemove ERROR: %v\n", err)
			}
		}()

		// Start Container
		if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			fmt.Printf("Container Start ERROR: %v\n", err)
			//return err
		}
		defer func() {
			if err := cli.ContainerStop(context.Background(), resp.ID, nil); err != nil {
				fmt.Printf("ContainerStop ERROR: %v\n", err)
			}
		}()

		fmt.Printf("Container Started\n")

		// Attach Container
		hjConn, err := cli.ContainerAttach(ctx, resp.ID, types.ContainerAttachOptions{
			Stream: true,
			Stdin:  true,
			Stdout: true,
			Stderr: true,
			Logs:   false,
		})
		if err != nil {
			fmt.Printf("Container Attach ERROR: %v\n", err)
			//return err
		}
		defer hjConn.Close()
		fmt.Printf("Container Attached\n")

		// Exec Commands
		//io.WriteString(os.Stdin, cmd)
		//os.Stdin.Closefalse()
		for cmd := range c.cmdCh {
			fmt.Printf("cmd received:%v\n", cmd)
			//cmd = fmt.Sprintf("%s\nexit\n", cmd)
			hjConn.Conn.Write([]byte(cmd))
		}

		hjConn.Conn.Write([]byte(fmt.Sprintln("exit")))
		statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
		select {
		case err := <-errCh:
			if err != nil {
				//panic(err)
				fmt.Printf("Container errCh ERROR: %v\n", err)
				//return err
			}
		case <-statusCh:
		}
		fmt.Printf("Container Wait Finished\n")
		b, err := ioutil.ReadAll(hjConn.Conn)
		if err != nil {
			fmt.Printf("Container Read ERROR: %v\n", err)
		}
		//result := string(b)[8:]
		result := string(b)
		fmt.Println("result:" + result)
		c.resultCh <- result
	}()

	return nil
}
