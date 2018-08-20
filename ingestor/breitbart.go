package ingestor

import (
	"fmt"
	"strings"
	"time"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

type (
	breitbart struct {
		*simple
	}

	breitbartCrawler struct {
		host string
	}
)

const (
	breitbartSource = "breitbart"
	breitbartHost   = "https://www.breitbart.com"
)

func Breitbart() Ingestor {
	return breitbart{
		simple: newSimple(breitbartSource, breitbartHost, newBreitbartCrawler()),
	}
}

func newBreitbartCrawler() crawler {
	return &breitbartCrawler{
		host: breitbartHost,
	}
}

func (c *breitbartCrawler) CleanURL(url string) string {
	url = strings.Split(url, "?")[0]
	url = strings.Split(url, "#")[0]
	return fmt.Sprintf("%s%s", c.host, url)[:len(c.host)+len(url)-1]
}

func (c *breitbartCrawler) Headline(n *html.Node) bool {
	return scrape.Attr(n, "property") == "og:title"
}

func (c *breitbartCrawler) ParseHeadline(n *html.Node) string {
	return scrape.Attr(n, "content")
}

func (c *breitbartCrawler) Text(n *html.Node) bool {
	return scrape.Attr(n, "class") == "entry-content"
}

func (c *breitbartCrawler) Author(n *html.Node) bool {
	return scrape.Attr(n, "data-aname") != ""
}

func (c *breitbartCrawler) ParseAuthor(n *html.Node) string {
	fullName := scrape.Attr(n, "data-aname")

	parsedName := ""
	for _, name := range strings.Split(fullName, " ") {
		if len(name) <= 1 {
			parsedName = fmt.Sprintf("%s %s ", parsedName, name)
			continue
		}

		firstLetter, theRest := string(name[0]), name[1:]
		parsedName = fmt.Sprintf("%s %s%s", parsedName, firstLetter, strings.ToLower(theRest))
	}

	return parsedName[:len(parsedName)-1]
}

func (c *breitbartCrawler) Date(n *html.Node) bool {
	return scrape.Attr(n, "property") == "article:published_time"
}

func (c *breitbartCrawler) ParseDate(n *html.Node) (time.Time, error) {
	str := scrape.Attr(n, "content")
	return time.Parse("2006-01-02T15:04:05", str[:len(str)-len("-07:00")])
}

func (c *breitbartCrawler) ArticleLinks(n *html.Node) bool {
	if n.Data != "a" {
		return false
	}
	link := scrape.Attr(n, "href")

	_, err := dateFromURL(c.host, link, 1)
	if err != nil {
		return false
	}

	return true
}
