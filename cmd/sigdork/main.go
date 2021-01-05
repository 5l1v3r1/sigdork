package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/drsigned/gos"
	"github.com/logrusorgru/aurora/v3"
)

type options struct {
	engine string
	pages  int
	query  string
}

var o options

func banner() {
	fmt.Fprintln(os.Stderr, aurora.BrightBlue(`
     _           _            _    
 ___(_) __ _  __| | ___  _ __| | __
/ __| |/ _`+"`"+` |/ _`+"`"+` |/ _ \| '__| |/ /
\__ \ | (_| | (_| | (_) | |  |   < 
|___/_|\__, |\__,_|\___/|_|  |_|\_\ v1.2.0
       |___/
`).Bold())
}

func init() {
	flag.StringVar(&o.engine, "e", "google", "")
	flag.IntVar(&o.pages, "p", 1, "")
	flag.StringVar(&o.query, "q", "", "")

	flag.Usage = func() {
		banner()

		h := "USAGE:\n"
		h += "  sigdork [OPTIONS]\n"

		h += "\nOPTIONS:\n"
		h += "  -e              search engine (default: google)\n"
		h += "  -p              number of pages (default: 1)\n"
		h += "  -q              search query (use `-q -` to read from stdin)\n"

		fmt.Fprintf(os.Stderr, h)
	}

	flag.Parse()
}

// get html
func getHTML(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	return string(body)
}

// parse html : extract links
func parseHTML(html string, pattern string) [][]string {
	regex := regexp.MustCompile(pattern)
	match := regex.FindAllStringSubmatch(html, -1)[0:]
	return match

}

// search : execute dorks
func search(engine string, query string, pages int) {
	var params, engineURL, urlExtractRegex string

	queryEscaped := url.QueryEscape(query)

	switch strings.ToLower(engine) {
	case "google":
		urlExtractRegex = `"><a href="\/url\?q=(.*?)&amp;sa=U&amp;`
		engineURL = "https://www.google.com/search"
		params = ("q=" + queryEscaped + "&gws_rd=cr,ssl&client=ubuntu&ie=UTF-8&start=")
	default:
		fmt.Println("engine not supported yet")
	}

	for p := 1; p <= pages; p++ {
		page := strconv.Itoa(p)

		html := getHTML(engineURL + "?" + params + page)
		result := parseHTML(html, urlExtractRegex)

		for i := range result {
			URL, err := url.QueryUnescape(result[i][1])
			if err != nil {
				log.Fatalln(err)
			}

			fmt.Println(URL)
		}
	}
}

func main() {
	queries := getQueries(o.query)

	for query := range queries {
		search(o.engine, query, o.pages)
	}
}

func getQueries(query string) chan string {
	queries := make(chan string)

	go func() {
		defer close(queries)

		if query == "-" {
			if !gos.HasStdin() {
				log.Fatalln(errors.New("no stdin"))
			}

			scanner := bufio.NewScanner(os.Stdin)

			for scanner.Scan() {
				if scanner.Text() != "" {
					queries <- scanner.Text()
				}
			}

			if scanner.Err() != nil {
				log.Fatalln(scanner.Err())
			}

		} else {
			if query != "" {
				queries <- query
			}
		}
	}()

	return queries
}
