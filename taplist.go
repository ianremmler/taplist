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

func recFind(node *html.Node, result *string, fn func(*html.Node, *string) bool) bool {
	for kid := node.FirstChild; kid != nil; kid = kid.NextSibling {
		if fn(kid, result) {
			return true
		}
	}
	return false
}

func findAttr(node *html.Node, keyRx, valRx string) *html.Attribute {
	for _, attr := range node.Attr {
		if ok, err := regexp.MatchString(keyRx, attr.Key); err == nil && ok {
			if ok, err := regexp.MatchString(valRx, attr.Val); err == nil && ok {
				return &attr
			}
		}
	}
	return nil
}

func findBeer(node *html.Node, beers *[]beerInfo) {
	if findAttr(node, "^id$", "^beer-\\d+$") != nil {
		brewery, brew := "", ""
		if findBrewery(node, &brewery) && findBrew(node, &brew) {
			*beers = append(*beers, beerInfo{brewery, brew})
		}
		return
	}
	for kid := node.FirstChild; kid != nil; kid = kid.NextSibling {
		findBeer(kid, beers)
	}
}

func findBrewery(node *html.Node, brewery *string) bool {
	if node.DataAtom == atom.H4 && node.FirstChild != nil {
		*brewery = node.FirstChild.Data
		return true
	}
	return recFind(node, brewery, findBrewery)
}

func findBrew(node *html.Node, brew *string) bool {
	if findAttr(node, "^class$", "^beer-name$") != nil && node.FirstChild != nil {
		*brew = node.FirstChild.Data
		return true
	}
	return recFind(node, brew, findBrew)
}

func findBarDesc(node *html.Node, desc *string) bool {
	if findAttr(node, "^name$", "^description$") != nil {
		if attr := findAttr(node, "^content$", ""); attr != nil {
			*desc = attr.Val
			return true
		}
	}
	return recFind(node, desc, findBarDesc)
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
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}
		idx := strings.IndexAny(line, " \t")
		if idx > 0 && idx < len(line)-1 {
			id, name := line[:idx], strings.TrimSpace(line[idx:])
			if checkId(id) && name != "" {
				barMap[id] = name
			}
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
	if resp.StatusCode != 200 {
		log.Fatalln("Sorry, couldn't find what's on tap at " + name)
	}

	page, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	desc, beers := "", []beerInfo{}
	findBarDesc(page, &desc)
	findBeer(page, &beers)
	if desc != "" {
		fmt.Println(desc + "\n")
	} else {
		fmt.Printf("%d beers on tap at "+name+"\n\n", len(beers))
	}
	for _, beer := range beers {
		fmt.Printf("%-38.38s  %s\n", beer.brewery, beer.brew)
	}
}
