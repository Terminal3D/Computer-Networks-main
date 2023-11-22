package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

type SSHCommand struct {
	Path   string
	Env    []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type SSHClient struct {
	Config *ssh.ClientConfig
	Host   string
	Port   int
}

func (client *SSHClient) RunCommand() error {
	input := bufio.NewReader(os.Stdin)
	fmt.Println("Enter command to run (type 'exit' to quit):")

	for {
		session, err := client.newSession()
		if err != nil {
			return err
		}
		fmt.Print("ssh> ")
		command, err := input.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Exiting session.")
				return nil
			}
			return fmt.Errorf("error reading command: %v", err)
		}

		command = strings.TrimSpace(command) // Remove newline character

		if command == "exit" {
			fmt.Println("Exiting session.")
			break
		}

		cmd := &SSHCommand{
			Path:   command,
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		if err = client.prepareCommand(session, cmd); err != nil {
			session.Close()
			return err
		}

		log.Println("Running command:", command)
		if err := session.Run(cmd.Path); err != nil {
			fmt.Fprintf(os.Stderr, "command run error: %s\n", err)
		}

		session.Close()
	}

	return nil
}

func (client *SSHClient) prepareCommand(session *ssh.Session, cmd *SSHCommand) error {

	if cmd.Stdin != nil {
		stdin, err := session.StdinPipe()
		if err != nil {
			return fmt.Errorf("Unable to setup stdin for session: %v", err)
		}
		go io.Copy(stdin, cmd.Stdin)
	}

	if cmd.Stdout != nil {
		stdout, err := session.StdoutPipe()
		if err != nil {
			return fmt.Errorf("Unable to setup stdout for session: %v", err)
		}
		go io.Copy(cmd.Stdout, stdout)
	}

	if cmd.Stderr != nil {
		stderr, err := session.StderrPipe()
		if err != nil {
			return fmt.Errorf("unable to setup stderr for session: %v", err)
		}
		go io.Copy(cmd.Stderr, stderr)
	}

	return nil
}

func (client *SSHClient) newSession() (*ssh.Session, error) {
	connection, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", client.Host, client.Port), client.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %s", err)
	}

	session, err := connection.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %s", err)
	}

	modes := ssh.TerminalModes{
		// ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		session.Close()
		return nil, fmt.Errorf("request for pseudo terminal failed: %s", err)
	}

	return session, nil
}

func main() {
	sshConfig := &ssh.ClientConfig{
		User:            "test",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password("SDHBCXdsedfs222"),
		},
	}

	client := &SSHClient{
		Config: sshConfig,
		Host:   "151.248.113.144",
		Port:   443,
	}

	if err := client.RunCommand(); err != nil {
		fmt.Fprintf(os.Stderr, "command run error: %s\n", err)
		os.Exit(1)
	}
}
