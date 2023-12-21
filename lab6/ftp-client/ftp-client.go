package main

import (
	"bufio"
	"fmt"
	"github.com/jlaffaye/ftp"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {

	c, err := ftp.Dial("students.yss.su:21", ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	err = c.Login("ftpiu8", "3Ru7yOTA")
	if err != nil {
		log.Fatal(err)
	}

	defer c.Quit()
	input := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		input.Scan()
		command := strings.Fields(input.Text())
		switch command[0] {
		case "upload":
			uploadFile(c, command[1], command[2])
		case "download":
			downloadFile(c, command[1])
		case "delete":
			deleteFile(c, command[1])
		case "ls":
			listDirectory(c, command[1])
		case "mkdir":
			createDirectory(c, command[1])
		case "quit":
			break
		}
		fmt.Println()
	}
}

func uploadFile(c *ftp.ServerConn, fileName string, remotePath string) {
	uploadDir := "C:\\Users\\vvlad\\Documents\\Univer\\3sem\\Computer-Networks-main\\lab6\\ftp-client\\uploads"
	localPath := filepath.Join(uploadDir, fileName)
	file, err := os.Open(localPath)
	if err != nil {
		log.Println("Error finding local file")
		return
	}

	defer file.Close()
	err = c.Stor(remotePath, file)
	if err != nil {
		log.Println(err)
		return
	}
}

func downloadFile(c *ftp.ServerConn, remotePath string) {

	downloadDir := "C:\\Users\\vvlad\\Documents\\Univer\\3sem\\Computer-Networks-main\\lab6\\ftp-client\\downloads"
	fileName := filepath.Base(remotePath)
	localFilePath := filepath.Join(downloadDir, fileName)

	reader, err := c.Retr(remotePath)
	if err != nil {
		log.Println("Error downloading file")
		return
	}
	defer reader.Close()

	localFile, err := os.Create(localFilePath)
	if err != nil {
		return
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, reader)
	if err != nil {
		log.Println("Error copying file")
	}
	return
}

func deleteFile(c *ftp.ServerConn, remotePath string) {
	err := c.Delete(remotePath)
	if err != nil {
		log.Println("Error deleting file")
		return
	}
}

func createDirectory(c *ftp.ServerConn, dirPath string) {
	err := c.MakeDir(dirPath)
	if err != nil {
		log.Println("Error creating directory")
		return
	}
	log.Println("Directory at", dirPath, "created")
}

func listDirectory(c *ftp.ServerConn, dirPath string) {
	entries, err := c.List(dirPath)
	if err != nil {
		log.Println("Error listing directory")
		return
	}

	for _, entry := range entries {
		fmt.Println(entry.Name)
	}

}
