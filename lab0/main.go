package main

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func PostRouteHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("post.html")
	tmpl.Execute(w, nil)
	r.ParseForm()

	for k, v := range r.Form {

		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
		fmt.Fprintf(w, "%s", r.Form.Get("entered_name"))
	}
}

func HomeRouterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //анализ аргументов,

	fmt.Println(r.Form) // ввод информации о форме на стороне сервера

	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])

	tmpl, _ := template.ParseFiles("main.html")
	tmpl.Execute(w, nil)

	for k, v := range r.Form {

		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
		fmt.Fprintf(w, "%s: %s", k, v)
	}
}

func RSSRouteHandler(w http.ResponseWriter, r *http.Request) {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL("https://www.mos.ru/rss")

	rssHTML := "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n    <meta charset=\"UTF-8\">" +
		"<title>" + feed.Title + "</title>"
	rssHTML += "<h1>" + feed.Description + "</h1>"
	rssHTML += "</head>"
	rssHTML += "<body> <ul> \n"
	for _, v := range feed.Items {
		rssHTML += "<li>"

		if len(v.Enclosures) != 0 {
			rssHTML += "<img src = \"" + v.Enclosures[0].URL + "\"" +
				"width = 200" +
				"/>"
		}
		rssHTML += "<a href = \"" + v.Link + "\">" + v.Title
		rssHTML += "</li>"

	}
	rssHTML += "</ul> </body> </html>"
	fmt.Fprintf(w, "%s", rssHTML)
}

func main() {
	http.HandleFunc("/", HomeRouterHandler)
	http.HandleFunc("/rss", RSSRouteHandler)
	http.HandleFunc("/post", PostRouteHandler)

	err := http.ListenAndServe(":9000", nil) // задаем слушать порт
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
