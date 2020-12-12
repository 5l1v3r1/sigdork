package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/logrusorgru/aurora/v3"
)

type options struct {
	engine   string
	pages    int
	query    string
	template bool
}

type dork struct {
	Query string `json:"query,omitempty"`
}

var o options

func banner() {
	fmt.Fprintln(os.Stderr, aurora.BrightBlue(`
     _           _            _    
 ___(_) __ _  __| | ___  _ __| | __
/ __| |/ _`+"`"+` |/ _`+"`"+` |/ _ \| '__| |/ /
\__ \ | (_| | (_| | (_) | |  |   < 
|___/_|\__, |\__,_|\___/|_|  |_|\_\ v1.0.0
       |___/
`).Bold())
}

func init() {
	flag.StringVar(&o.engine, "e", "google", "")

	flag.IntVar(&o.pages, "p", 1, "")

	flag.StringVar(&o.query, "q", "", "")

	flag.BoolVar(&o.template, "t", false, "")

	flag.Usage = func() {
		banner()

		h := "USAGE:\n"
		h += "  sigdork [OPTIONS]\n"

		h += "\nOPTIONS:\n"
		h += "  -e              search engine (default: google)\n"
		h += "  -p              number of pages (default: 1)\n"
		h += "  -q              search query\n"
		h += "  -t              template query mode (default: false)\n"

		fmt.Fprintf(os.Stderr, h)
	}

	flag.Parse()
}

// get dorks dir
func getDorksDir(engine string) (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	path := filepath.Join(currentUser.HomeDir, ".config/sigdork/"+engine)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return path, nil
	}

	return filepath.Join(currentUser.HomeDir, ".sigdork/"+engine), nil
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
	if o.template {
		dorksDir, err := getDorksDir(o.engine)
		if err != nil {
			fmt.Fprintln(os.Stderr, "unable to open user's dorks directory")
			return
		}

		filename := filepath.Join(dorksDir, o.query+".json")
		f, err := os.Open(filename)
		if err != nil {
			fmt.Fprintln(os.Stderr, "no such dork")
			return
		}

		defer f.Close()

		d := dork{}
		dec := json.NewDecoder(f)
		err = dec.Decode(&d)

		if err != nil {
			fmt.Fprintf(os.Stderr, "pattern file '%s' is malformed: %s\n", filename, err)
			return
		}

		o.query = d.Query
	}

	search(o.engine, o.query, o.pages)
}
