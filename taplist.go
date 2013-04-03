package main

import (
	"code.google.com/p/go.net/html"
	"code.google.com/p/go.net/html/atom"

	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"regexp"
	"strings"
)

var barMap = map[string]string{}

type beerInfo struct {
	brewery string
	brew    string
}

func findBeer(node *html.Node, beers *[]beerInfo) {
	if node.DataAtom == atom.Div {
		for _, attr := range node.Attr {
			if attr.Key == "id" && strings.HasPrefix(attr.Val, "beer-") {
				brewery, brew := "", ""
				findBrewery(node, &brewery)
				findBrew(node, &brew)
				*beers = append(*beers, beerInfo{brewery, brew})
			}
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

func findBarDesc(node *html.Node, desc *string) bool {
	if node.DataAtom == atom.Meta {
		isDesc := false
		for _, attr := range node.Attr {
			if attr.Key == "name" && attr.Val == "description" {
				isDesc = true
				break
			}
		}
		if isDesc {
			for _, attr := range node.Attr {
				if attr.Key == "content" {
					*desc = attr.Val
					break
				}
			}
		}
	}
	for kid := node.FirstChild; kid != nil; kid = kid.NextSibling {
		if findBarDesc(kid, desc) {
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

func readRc() {
	usr, err := user.Current()
	if err != nil {
		return
	}
	data, err := ioutil.ReadFile(usr.HomeDir + "/.taplistrc")
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		idx := strings.IndexAny(line, " \t")
		if idx < len(line)-1 {
			id, name := line[:idx], strings.TrimSpace(line[idx:])
			barMap[id] = name
		}
	}
}

func lookupBar(arg string) (string, string) {
	for id, name := range barMap {
		if strings.Contains(strings.ToLower(name), strings.ToLower(arg)) {
			return id, name
		}
	}
	return "", ""
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("taplist: ")

	if len(os.Args) != 2 {
		log.Fatalln("usage: taplist <id> | <name>")
	}
	readRc()
	arg := strings.ToLower(os.Args[1])
	id, name := "", ""
	if checkId(arg) {
		id, name = arg, arg
	} else {
		id, name = lookupBar(arg)
	}
	if id == "" {
		log.Fatalln(arg + " doesn't look like a valid name or taplister bar id")
	}

	resp, err := http.Get("http://www.taplister.com/bars/" + id)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	desc, beers := "", []beerInfo{}
	findBarDesc(doc, &desc)
	findBeer(doc, &beers)
	if desc != "" {
		fmt.Println(desc + "\n")
	} else {
		fmt.Printf("%d beers on tap at "+name+"\n\n", len(beers))
	}
	for _, beer := range beers {
		fmt.Printf("%-38.38s  %s\n", beer.brewery, beer.brew)
	}
}
