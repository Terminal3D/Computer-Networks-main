package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mgutz/logxi/v1"
	"image"
	"image/color"
	_ "image/jpeg"
	"math/big"
	"net"
	"os"
	"strconv"
)

import "lab1/sample/src/proto"

// Client - состояние клиента.
type Client struct {
	logger log.Logger    // Объект для печати логов
	conn   *net.TCPConn  // Объект TCP-соединения
	enc    *json.Encoder // Объект для кодирования и отправки сообщений
	sum    *big.Rat      // Текущая сумма полученных от клиента дробей
	count  int64         // Количество полученных от клиента дробей
}

// NewClient - конструктор клиента, принимает в качестве параметра
// объект TCP-соединения.
func NewClient(conn *net.TCPConn) *Client {
	return &Client{
		logger: log.New(fmt.Sprintf("client %s", conn.RemoteAddr().String())),
		conn:   conn,
		enc:    json.NewEncoder(conn),
		sum:    big.NewRat(0, 1),
		count:  0,
	}
}

// serve - метод, в котором реализован цикл взаимодействия с клиентом.
// Подразумевается, что метод serve будет вызаваться в отдельной go-программе.
func (client *Client) serve() {
	defer client.conn.Close()
	decoder := json.NewDecoder(client.conn)
	for {
		var req proto.Request
		if err := decoder.Decode(&req); err != nil {
			client.logger.Error("cannot decode message", "reason", err)
			break
		} else {
			client.logger.Info("received command", "command", req.Command)
			if client.handleRequest(&req) {
				client.logger.Info("shutting down connection")
				break
			}
		}
	}
}

var images map[string]*image.Image

// handleRequest - метод обработки запроса от клиента. Он возвращает true,
// если клиент передал команду "quit" и хочет завершить общение.
func (client *Client) handleRequest(req *proto.Request) bool {
	switch req.Command {
	case "quit":
		client.respond("ok", nil)
		return true
	case "get_image_list":

		imageList := make([]string, 0)
		for name, _ := range images {
			imageList = append(imageList, name)
			client.logger.Info(name+" added to images name list", "command", req.Command)
		}

		errorMsg := ""
		if len(imageList) == 0 {
			errorMsg = "no image found"
		}
		if errorMsg == "" {
			client.respond("list found", imageList)
		} else {
			client.logger.Error("getting failed", "reason", errorMsg)
			client.respond("failed", errorMsg)
		}
	case "get_color":
		var colorRequest proto.ImagePoint
		json.Unmarshal(*req.Data, &colorRequest)
		request := "Request: "
		request += "PictureName: " + colorRequest.ImageName + ", "
		request += "X: " + strconv.Itoa(colorRequest.X) + ", "
		request += "Y: " + strconv.Itoa(colorRequest.Y) + "\n"

		log.Info(request)
		currentPicture := images[colorRequest.ImageName]
		colorFound := (*currentPicture).At(colorRequest.X, colorRequest.Y).(color.YCbCr)
		red, green, blue := color.YCbCrToRGB(colorFound.Y, colorFound.Cb, colorFound.Cr)

		client.respond("pixel found", "Red: "+strconv.Itoa(int(red))+" Green: "+
			strconv.Itoa(int(green))+" Blue: "+strconv.Itoa(int(blue)))
	default:
		client.logger.Error("unknown command")
		client.respond("failed", "unknown command")
	}
	return false
}

// respond - вспомогательный метод для передачи ответа с указанным статусом
// и данными. Данные могут быть пустыми (data == nil).
func (client *Client) respond(status string, data interface{}) {
	var raw json.RawMessage
	raw, _ = json.Marshal(data)
	client.enc.Encode(&proto.Response{status, &raw})
}

func main() {
	// Работа с командной строкой, в которой может указываться необязательный ключ -addr.
	var addrStr string
	flag.StringVar(&addrStr, "addr", "127.0.0.1:6000", "specify ip address and port")
	flag.Parse()

	reader1, err := os.Open("C:\\Users\\vvlad\\Documents\\Univer\\3sem\\networks\\lab1\\sample\\src\\server\\images\\picture1.jpeg")
	if err != nil {
		log.Fatal("error parsing image")
	}
	defer reader1.Close()

	images = make(map[string]*image.Image)
	image1, _, err := image.Decode(reader1)

	reader2, err := os.Open("C:\\Users\\vvlad\\Documents\\Univer\\3sem\\networks\\lab1\\sample\\src\\server\\images\\picture2.jpg")
	if err != nil {
		log.Fatal("error parsing image")
	}
	defer reader2.Close()

	image2, _, err := image.Decode(reader2)
	images["picture1"] = &image1
	images["picture2"] = &image2

	// Разбор адреса, строковое представление которого находится в переменной addrStr.
	if addr, err := net.ResolveTCPAddr("tcp", addrStr); err != nil {
		log.Error("address resolution failed", "address", addrStr)
	} else {
		log.Info("resolved TCP address", "address", addr.String())

		// Инициация слушания сети на заданном адресе.
		if listener, err := net.ListenTCP("tcp", addr); err != nil {
			log.Error("listening failed", "reason", err)
		} else {
			// Цикл приёма входящих соединений.
			for {
				if conn, err := listener.AcceptTCP(); err != nil {
					log.Error("cannot accept connection", "reason", err)
				} else {
					log.Info("accepted connection", "address", conn.RemoteAddr().String())

					// Запуск go-программы для обслуживания клиентов.
					go NewClient(conn).serve()
				}
			}
		}
	}
}
