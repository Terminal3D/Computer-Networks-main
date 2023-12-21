package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type Address struct {
	IPv4 string
	Port string
}

type Node struct {
	Next    Address
	Address Address
}

type Pixel struct {
	X     string `json:"X"`
	Y     string `json:"Y"`
	Color string `json:"Color"`
}

type Package struct {
	From Address
	Data Pixel
}

var pixels []Pixel

func handleServer(node *Node) {
	listen, err := net.Listen("tcp", "185.139.70.64:"+node.Address.Port)
	if err != nil {
		panic("listen error")
	}
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			break
		}
		go handleConnection(node, conn)
	}
}

func handleConnection(node *Node, conn net.Conn) {
	defer conn.Close()
	var (
		buffer  = make([]byte, 512)
		message string
		pack    Package
	)
	for {
		length, err := conn.Read(buffer)
		if err != nil {
			break
		}
		message += string(buffer[:length])
	}

	err := json.Unmarshal([]byte(message), &pack)
	if err != nil {
		return
	}
	if pack.From != node.Address {
		fmt.Println(pack.Data)
		pixels = append(pixels, pack.Data)
		node.Send(&pack)
	}
}

func handleClient(node *Node) {
	for {
		var message string
		message, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		message = strings.Replace(message, "\r\n", "", -1)
		splited := strings.Split(message, " ")
		switch splited[0] {
		case "/exit":
			os.Exit(0)
		case "/drawn":
			for _, pixel := range pixels {
				fmt.Println(pixel)
			}
		case "/refresh":
			displayPixels()
		default:
			pixel := strings.Split(message, " ")
			pixels = append(pixels, Pixel{X: pixel[0], Y: pixel[1], Color: pixel[2]})
			node.sendToAllMessage(pixel)
		}
	}
}

func displayPixels() {
	tmpl := `
        <!DOCTYPE html>
        <html>
        <head>
            <style>
                .pixel {
                    position: absolute;
                    width: 10px;
                    height: 10px;
                }
            </style>
        </head>
        <body>
            {{range .}}
                <div class="pixel" style="left: {{.X}}px; top: {{.Y}}px; background-color: {{.Color}};"></div>
            {{end}}
        </body>
        </html>
    `
	t, err := template.New("webpage").Parse(tmpl)
	if err != nil {
		fmt.Println("Error creating template:", err)
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}
	fileName := filepath.Join(cwd, "lab3/p2p/pixels.html")

	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	err = t.Execute(file, pixels)
	if err != nil {
		fmt.Println("Error executing template:", err)
	} else {
		fmt.Println("Pixels displayed in", fileName)
	}
}

func (node *Node) Run(handleServer func(node *Node), handleClient func(node *Node)) {
	go handleServer(node)
	handleClient(node)
}

func (node *Node) sendToAllMessage(message []string) {
	pixel := message

	node.Send(&Package{
		From: node.Address,
		Data: Pixel{X: pixel[0], Y: pixel[1], Color: pixel[2]},
	})
}

func (node *Node) Send(pack *Package) {
	conn, err := net.Dial("tcp", node.Next.IPv4+":"+node.Next.Port)
	if err != nil {
		return
	}
	defer conn.Close()
	jsonPack, err := json.Marshal(*pack)
	if err != nil {
		panic(err)
	}
	_, err = conn.Write(jsonPack)
	if err != nil {
		panic(err)
	}
}

func main() {
	pixels = make([]Pixel, 0)
	node := Node{Address: Address{IPv4: "185.139.70.64", Port: "9093"}, Next: Address{IPv4: "185.104.249.105", Port: "9094"}}
	node.Run(handleServer, handleClient)
}
