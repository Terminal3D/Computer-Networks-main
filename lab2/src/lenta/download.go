package main

import (
	"github.com/mgutz/logxi/v1"
	"golang.org/x/net/html"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func getAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func isElem(node *html.Node, tag string) bool {
	return node != nil && node.Type == html.ElementNode && node.Data == tag
}

func isDiv(node *html.Node, class string) bool {
	return isElem(node, "div") && getAttr(node, "class") == class
}

type Item struct {
	Name, Volume string
}

func parseTableRow(node *html.Node, blockName string) *html.Node {
	if node == nil {
		return nil
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if isDiv(c, blockName) {
			return c
		}
		t := parseTableRow(c, blockName)
		if isDiv(t, blockName) {
			return t
		}
	}
	return nil
}

func parseBlock(node *html.Node, pClass string) *html.Node {
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if isElem(c, "p") && getAttr(c, "class") == pClass {
			return c
		} else {
			t := parseBlock(c, pClass)
			if isElem(t, "p") && getAttr(t, "class") == pClass {
				return t
			}
		}
	}
	return nil
}

func search(node *html.Node) []*Item {
	//fmt.Println(node.Data)

	if isDiv(node, "sc-66133f36-2 cgmess") {
		var items []*Item
		c := node.FirstChild
		for ; c != nil; c = c.NextSibling {
			if isElem(c, "table") {
				break
			}
		}
		for c := c.FirstChild; c != nil; c = c.NextSibling {
			if isElem(c, "tbody") {
				for tr := c.FirstChild; tr != nil; tr = tr.NextSibling {
					if isElem(tr, "tr") {
						volumeBlock := parseTableRow(tr, "sc-aef7b723-0 sc-97d9abce-0 eGkfri")
						nameBlock := parseTableRow(tr, "sc-aef7b723-0 LCOyB")
						if nameBlock != nil && volumeBlock != nil {
							volume := parseBlock(volumeBlock, "sc-4984dd93-0 jZrMxO font_weight_500").FirstChild.Data
							name := parseBlock(nameBlock, "sc-4984dd93-0 kKpPOn").FirstChild.Data
							items = append(items, &Item{
								Name:   name,
								Volume: volume,
							})
						}
					}
				}
			}
		}
		sort.Slice(items, func(i, j int) bool {
			volume1, _ := strconv.ParseInt(strings.Replace(items[i].Volume[1:], ",", "", -1), 10, 64)
			volume2, _ := strconv.ParseInt(strings.Replace(items[j].Volume[1:], ",", "", -1), 10, 64)
			return volume1 > volume2
		})
		return items
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if items := search(c); items != nil {
			return items
		}
	}
	return nil
}

func downloadNews() []*Item {
	log.Info("sending request to coinmarketcap.com")
	if response, err := http.Get("https://coinmarketcap.com"); err != nil {
		log.Error("request to coinmarketcap.com failed", "error", err)
	} else {
		defer response.Body.Close()
		status := response.StatusCode
		log.Info("got response from coinmarketcap.com", "status", status)
		if status == http.StatusOK {
			if doc, err := html.Parse(response.Body); err != nil {
				log.Error("invalid HTML from coinmarketcap.com", "error", err)
			} else {
				log.Info("HTML from coinmarketcap.com parsed successfully")
				return search(doc)
			}
		}
	}
	return nil
}
