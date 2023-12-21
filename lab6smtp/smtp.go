package main

import (
	"bufio"
	"crypto/tls"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"log"
	"net/smtp"
	"os"
	"time"
)

const (
	smtpServer   = "mail.nic.ru"
	smtpPort     = "465"
	smtpUsername = "dts21@dactyl.su"
	smtpPassword = "12345678990DactylSUDTS"
)

type Message struct {
	To          string
	Subject     string
	MessageBody string
}

type SMTPResponse struct {
	DialStatus           string
	ClientCreationStatus string
	AuthStatus           string
	MailStatus           string
	SendStatus           string
}

func main() {
	err := mysql.SetLogger(log.New(os.Stdout, "[mysql] ", log.LstdFlags))
	if err != nil {
		fmt.Println(err)
	}
	input := bufio.NewScanner(os.Stdin)

	input.Scan()
	command := input.Text()

	if command == "personal" {

		fmt.Print("To: ")
		input.Scan()
		to := input.Text()

		fmt.Print("\nSubject: ")
		input.Scan()
		subject := input.Text()

		fmt.Print("\nBody: ")
		input.Scan()
		body := input.Text()
		msg := Message{to, subject, body}

		_, err := sendMessage(msg)
		if err != nil {
			log.Println(err)
		}

	} else {
		err := sendToAllFromDB()
		if err != nil {
			log.Println(err)
		}
	}

}

func sendToAllFromDB() error {
	db, err := sql.Open("mysql", "iu9networkslabs:Je2dTYr6@tcp(students.yss.su:3306)/iu9networkslabs")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT name, mail, text FROM iu9vorobiovSMTP")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, email, message string

		if err := rows.Scan(&name, &email, &message); err != nil {
			log.Fatal(err)
		}

		log.Println("Name: "+name, " Email: "+email, "\nMessage: "+message)

		var msg Message
		msg.To = email
		msg.Subject = "From Vorobiov Vladislav"
		msg.MessageBody = "<html>\n\t<body style=\"background-color: #C3E7E8;\">\n\t\t<p><strong>Hello " + name + "!</strong></p>\n\t\t<p><i>" + message + "</i></p>\n\t</body>\n</html>"
		response, err := sendMessage(msg)
		if err != nil {
			log.Println(err)
		}
		_, err = db.Exec("INSERT INTO iu9vorobiovSMTPLogs (email, dial_status, client_creation_status, auth_status, mail_status, send_status) VALUES (?, ?, ?, ?, ?, ?)", email, response.DialStatus, response.ClientCreationStatus, response.AuthStatus, response.MailStatus, response.SendStatus)
		if err != nil {
			log.Println(err)
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func sendMessage(msg Message) (SMTPResponse, error) {
	var smtpResponse SMTPResponse
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpServer,
	}

	conn, err := tls.Dial("tcp", smtpServer+":"+smtpPort, tlsConfig)
	if err != nil {
		smtpResponse.DialStatus = "Failed: " + err.Error()
		return smtpResponse, err
	}
	smtpResponse.DialStatus = "Success"
	defer conn.Close()

	client, err := smtp.NewClient(conn, smtpServer)
	if err != nil {
		smtpResponse.ClientCreationStatus = "Failed: " + err.Error()
		return smtpResponse, err
	}
	smtpResponse.ClientCreationStatus = "Success"
	defer client.Close()

	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpServer)
	if err := client.Auth(auth); err != nil {
		smtpResponse.AuthStatus = "Failed: " + err.Error()
		return smtpResponse, err
	}
	smtpResponse.AuthStatus = "Success"

	if err := client.Mail(smtpUsername); err != nil {
		smtpResponse.MailStatus = "Failed: " + err.Error()
		return smtpResponse, err
	}
	smtpResponse.MailStatus = "Success"

	if err := client.Rcpt(msg.To); err != nil {
		smtpResponse.SendStatus = "Failed: " + err.Error()
		return smtpResponse, err
	}
	smtpResponse.SendStatus = "Success"

	w, err := client.Data()
	if err != nil {
		return smtpResponse, err
	}
	defer w.Close()

	message := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", msg.To, msg.Subject, msg.MessageBody)
	_, err = w.Write([]byte(message))
	if err != nil {
		return smtpResponse, err
	}

	if err := w.Close(); err != nil {
		return smtpResponse, err
	}

	return smtpResponse, client.Quit()
}
