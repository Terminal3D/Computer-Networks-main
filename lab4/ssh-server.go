package main

import (
	"fmt"
	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"
	"golang.org/x/term"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	Username = "root"
	Password = "123"
)

func SftpHandler(sess ssh.Session) {
	debugStream := ioutil.Discard
	serverOptions := []sftp.ServerOption{
		sftp.WithDebug(debugStream),
	}
	server, err := sftp.NewServer(
		sess,
		serverOptions...,
	)
	if err != nil {
		log.Printf("sftp server init error: %s\n", err)
		return
	}
	if err := server.Serve(); err == io.EOF {
		server.Close()
		fmt.Println("sftp client exited session.")
	} else if err != nil {
		fmt.Println("sftp server completed with error:", err)
	}
}

func handleInput(sess ssh.Session) (string, error) {

	term := term.NewTerminal(sess, "> ")
	line, err := term.ReadLine()
	if err != nil {
		return "", err
	}
	return line, nil
}

func SSHSessionHandler(sess ssh.Session) {
	io.WriteString(sess, "Welcome to the SSH server!\n")

	for {
		line, err := handleInput(sess)
		if err != nil {
			if err == io.EOF {
				log.Printf("Terminal closed/EOF")
				break
			}
			log.Printf("Error reading from terminal: %s\n", err)
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "mkdir":
			if len(parts) < 2 {
				io.WriteString(sess, "Usage: mkdir <directory>\n")
			} else {
				err := os.Mkdir(parts[1], 0755)
				if err != nil {
					fmt.Fprintf(sess, "Error: %s\n", err)
				}
			}
		case "rmdir":
			if len(parts) < 2 {
				io.WriteString(sess, "Usage: rmdir <directory>\n")
			} else {
				err := os.Remove(parts[1])
				if err != nil {
					fmt.Fprintf(sess, "Error: %s\n", err)
				}
			}
		case "ls":
			if len(parts) < 2 {
				parts = append(parts, ".")
			}
			files, err := os.ReadDir(parts[1])
			if err != nil {
				fmt.Fprintf(sess, "Error: %s\n", err)
				continue
			}
			for _, file := range files {
				fmt.Fprintf(sess, "%s\n", file.Name())
			}
		case "mv":
			if len(parts) < 3 {
				io.WriteString(sess, "Usage: mv <source> <destination>\n")
			} else {
				err := os.Rename(parts[1], parts[2])
				if err != nil {
					fmt.Fprintf(sess, "Error: %s\n", err)
				}
			}
		case "rm":
			if len(parts) < 2 {
				io.WriteString(sess, "Usage: rm <file>\n")
			} else {
				err := os.Remove(parts[1])
				if err != nil {
					fmt.Fprintf(sess, "Error: %s\n", err)
				}
			}
		case "ping":
			if len(parts) < 2 {
				io.WriteString(sess, "Usage: ping <address>\n")
			} else {
				cmd := exec.Command("ping", parts[1])
				cmd.Stdout = sess
				cmd.Stderr = sess
				err := cmd.Run()
				if err != nil {
					fmt.Fprintf(sess, "Error: %s\n", err)
				}
			}
		default:
			io.WriteString(sess, "Unknown command\n")
		}
	}
}

func main() {
	sshHandler := ssh.Server{
		Addr: "185.139.70.64:9092",
		SubsystemHandlers: map[string]ssh.SubsystemHandler{
			"sftp": SftpHandler,
		},
		Handler: SSHSessionHandler,
		PasswordHandler: func(ctx ssh.Context, pass string) bool {
			return ctx.User() == Username && pass == Password
		},
	}

	log.Fatal(sshHandler.ListenAndServe())
}
