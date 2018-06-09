package main

import (
	"bytes"
	"context"
	"fmt"
	"net"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type instance struct {
	client     client.Client
	imageId    string
	id         string
	cmdCh      chan string
	resultCh   chan string
	runErrCh   chan error
	execErrCh  chan error
	exitErrCh  chan error
	stdinConn  net.Conn
	stdoutConn net.Conn
}

func newDockerContainer(ctx context.Context, image string, cmd []string) (*instance, error) {
	c := instance{
		cmdCh:     make(chan string),
		resultCh:  make(chan string),
		runErrCh:  make(chan error),
		execErrCh: make(chan error),
		exitErrCh: make(chan error),
	}

	if cli, err := client.NewClientWithOpts(); err != nil {
		fmt.Printf("Container New Client ERROR: %v\n", err)
		return &c, err
	} else {
		c.client = *cli
	}

	/*
		reader, err := c.client.ImagePull(ctx, "docker.io/library/alpine", types.ImagePullOptions{})
		if err != nil {
		}
		io.Copy(os.Stdout, reader)
	*/

	resp, err := c.client.ContainerCreate(ctx, &container.Config{
		Image:        image,
		Cmd:          cmd,
		OpenStdin:    true,
		StdinOnce:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	}, nil, nil, "")
	if err != nil {
		fmt.Printf("Container Create ERROR: %v\n", err)
		panic(err)
	}
	c.id = resp.ID

	return &c, nil
}

func (c *instance) commit(ctx context.Context, img string) error {
	fmt.Printf("ContainerCommit: \n")
	cco := types.ContainerCommitOptions{
		Reference: img,
		Comment:   "sbox bot",
		Config:    &container.Config{},
	}
	if resp, err := c.client.ContainerCommit(ctx, c.id, cco); err != nil {
		fmt.Printf("ContainerCommit ERROR: %v\n", err)
		return err
	} else {
		c.imageId = resp.ID
	}
	return nil
}

func (c *instance) remove(ctx context.Context) error {
	if err := c.client.ContainerRemove(ctx, c.id, types.ContainerRemoveOptions{}); err != nil {
		fmt.Printf("ContainerRemove ERROR: %v\n", err)
		return err
	}
	return nil
}

func (c *instance) removeImage(ctx context.Context) error {
	if _, err := c.client.ImageRemove(ctx, c.imageId, types.ImageRemoveOptions{}); err != nil {
		fmt.Printf("ImageRemove ERROR: %v\n", err)
		return err
	}
	return nil
}

func (c *instance) run(ctx context.Context) error {
	go c.doRun(ctx)
	err := <-c.runErrCh
	return err
}

func (c *instance) exec(cmd string) error {
	c.cmdCh <- cmd
	err := <-c.execErrCh
	return err
}

func (c *instance) exit() (string, error) {
	close(c.cmdCh)
	return <-c.resultCh, <-c.exitErrCh
}

const BUFSIZE = 1024

func (c *instance) result() (string, error) {
	var err error
	buf := make([]byte, BUFSIZE)
	_, err = c.stdoutConn.Read(buf)
	//fmt.Printf("read buffer \n%v\n", hex.Dump(buf))
	buf = bytes.TrimRight(buf, "\x00")
	result := string(buf)
	if len(result) > 8 {
		result = result[8:] // remove header bytes
	}
	return result, err
}

func (c *instance) doRun(ctx context.Context) {

	// Start Container
	if err := c.client.ContainerStart(ctx, c.id, types.ContainerStartOptions{}); err != nil {
		fmt.Printf("Container Start ERROR: %v\n", err)
		c.runErrCh <- err
		return
	}
	defer func() {
		if err := c.client.ContainerStop(context.Background(), c.id, nil); err != nil {
			fmt.Printf("ContainerStop ERROR: %v\n", err)
		}
	}()

	// Attach Container for Stdin
	readConn, err := c.client.ContainerAttach(ctx, c.id, types.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: false,
		Stderr: false,
		Logs:   false,
	})
	if err != nil {
		fmt.Printf("Container Attach ERROR: %v\n", err)
		c.runErrCh <- err
		return
	}
	defer readConn.Close()
	c.stdinConn = readConn.Conn

	// Attach Container for Stdout,Stderr
	writeConn, err := c.client.ContainerAttach(ctx, c.id, types.ContainerAttachOptions{
		Stream: true,
		Stdin:  false,
		Stdout: true,
		Stderr: true,
		Logs:   false,
	})
	if err != nil {
		fmt.Printf("Container Attach ERROR: %v\n", err)
		c.runErrCh <- err
		return
	}
	defer writeConn.Close()
	c.stdoutConn = writeConn.Conn

	c.runErrCh <- nil // no error while attaching

	for cmd := range c.cmdCh {
		fmt.Printf("cmd received:%v", cmd)
		//cmd = fmt.Sprintf("%s\nexit\n", cmd)
		c.stdinConn.Write([]byte(cmd))
		c.execErrCh <- nil
	}

	c.stdinConn.Close() // finish command
	statusCh, errCh := c.client.ContainerWait(ctx, c.id, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			fmt.Printf("Container errCh ERROR: %v\n", err)
			result, _ := c.result()
			c.resultCh <- result
			c.exitErrCh <- err
		}
	case <-statusCh:
	case <-ctx.Done():
	}

	result, _ := c.result()
	c.resultCh <- result
	c.exitErrCh <- nil
	return
}
