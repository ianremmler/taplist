package main

import (
	"code.google.com/p/go.net/html"

	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func findBeer(node *html.Node, beers *[]string) {
	for _, attr := range node.Attr {
		if attr.Key == "class" && attr.Val == "beer-name" {
			if content := node.FirstChild; content != nil {
				*beers = append(*beers, content.Data)
				return
			}
		}
	}
	for kid := node.FirstChild; kid != nil; kid = kid.NextSibling {
		findBeer(kid, beers)
	}
}

func checkId(id string) bool {
	ok, err := regexp.MatchString("^[[:xdigit:]]{24}$", id)
	if err != nil {
		return false
	}
	return ok
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("taplist: ")

	if len(os.Args) != 2 {
		log.Fatalln("usage: taplist <id>")
	}
	id := strings.ToLower(os.Args[1])
	if !checkId(id) {
		log.Fatalln(id + " doesn't look like a valid taplister bar id")
	}
	resp, err := http.Get("http://www.taplister.com/bars/" + id)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	beers := []string{}
	findBeer(doc, &beers)
	for _, beer := range beers {
		fmt.Println(beer)
	}
}
