package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"golang.org/x/net/html"
)

func GetPageCnt(r io.Reader) (int, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return 0, err
	}
	var lastCnt int
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "option" {
			lastCnt, err = strconv.Atoi(n.Attr[0].Val)
			if err != nil {
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return lastCnt, nil
}

func main() {

	file, err := os.Open("index.html")
	if err != nil {
		fmt.Println(err)
		return
	}

	doc, err := html.Parse(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	var lastCnt int
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "select" {
			fmt.Println(n.Data) // print tag name
		}
		if n.Type == html.ElementNode && n.Data == "option" {
			lastCnt, err = strconv.Atoi(n.Attr[0].Val)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	fmt.Println(lastCnt)
}
