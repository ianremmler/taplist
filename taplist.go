package main

import (
	"code.google.com/p/go.net/html"
	"code.google.com/p/go.net/html/atom"

	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func findBeer(node *html.Node, beers *[]string) {
	for _, attr := range node.Attr {
		if attr.Key == "id" && strings.HasPrefix(attr.Val, "beer-") {
			brewery, brew := "", ""
			findBrewery(node, &brewery)
			findBrew(node, &brew)
			beer := fmt.Sprintf("%-38.38s  %s", brewery, brew)
			*beers = append(*beers, beer)
		}
	}
	for kid := node.FirstChild; kid != nil; kid = kid.NextSibling {
		findBeer(kid, beers)
	}
}

func findBrewery(node *html.Node, brewery *string) bool {
	if node.DataAtom == atom.H4 {
		if content := node.FirstChild; content != nil {
			*brewery = content.Data
			return true
		}
	}
	for kid := node.FirstChild; kid != nil; kid = kid.NextSibling {
		if findBrewery(kid, brewery) {
			return true
		}
	}
	return false
}

func findBrew(node *html.Node, brew *string) bool {
	for _, attr := range node.Attr {
		if attr.Key == "class" && attr.Val == "beer-name" {
			if content := node.FirstChild; content != nil {
				*brew = content.Data
				return true
			}
		}
	}
	for kid := node.FirstChild; kid != nil; kid = kid.NextSibling {
		if findBrew(kid, brew) {
			return true
		}
	}
	return false
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
