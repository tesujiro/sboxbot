package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

//func main() {
func execOnContainer(ctx context.Context, cmd string) string {

	cli, err := client.NewClientWithOpts()
	if err != nil {
		//panic(err)
		fmt.Printf("Container New Client ERROR: %v\n", err)
		return fmt.Sprintf("%v", err)
	}

	reader, err := cli.ImagePull(ctx, "docker.io/library/alpine", types.ImagePullOptions{})
	if err != nil {
		//panic(err)
		fmt.Printf("Container Image Pull ERROR: %v\n", err)
		return fmt.Sprintf("%v", err)
	}
	io.Copy(os.Stdout, reader)

	// Create Container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "alpine",
		Cmd:   []string{"/bin/ash"},
		//Cmd:   []string{"echo", "hello world"},
		//Cmd: strings.Split(cmd, " "),
		//Cmd: []string{"sleep", "10"},
		OpenStdin:    true,
		StdinOnce:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}, nil, nil, "")
	if err != nil {
		//panic(err)
		fmt.Printf("Container Create ERROR: %v\n", err)
		return fmt.Sprintf("%v", err)
	}

	defer func() {
		if err := cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{}); err != nil {
			fmt.Printf("ContainerRemove ERROR: %v", err)
		}
	}()

	// Start Container
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		fmt.Printf("Container Start ERROR: %v\n", err)
		return fmt.Sprintf("%v", err)
	}
	defer func() {
		if err := cli.ContainerStop(ctx, resp.ID, nil); err != nil {
			fmt.Printf("ContainerStop ERROR: %v", err)
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
		return fmt.Sprintf("%v", err)
	}
	defer hjConn.Close()
	fmt.Printf("Container Attached\n")

	// Exec Commands
	//io.WriteString(os.Stdin, cmd)
	//os.Stdin.Close()
	hjConn.Conn.Write([]byte(cmd))

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			//panic(err)
			fmt.Printf("Container errCh ERROR: %v\n", err)
			return fmt.Sprintf("%v", err)
		}
	case <-statusCh:
	}
	fmt.Printf("Container Wait Finished\n")

	b, err := ioutil.ReadAll(hjConn.Conn)
	if err != nil {
		fmt.Printf("Container Read ERROR: %v\n", err)
		return fmt.Sprintf("%v", err)
	}
	result := string(b)
	return result

	// get log
	/*
		out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
		if err != nil {
			panic(err)
		}
		//io.Copy(os.Stdout, out)

		buf := new(bytes.Buffer)
		buf.ReadFrom(out)
		result := buf.String()
	*/

}
