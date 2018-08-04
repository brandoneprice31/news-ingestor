package ingestor

import (
	"strings"
	"time"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

type (
	nyt struct {
		*simple
	}

	nytCrawler struct {
		host string
	}
)

const (
	nytSource = "nyt"
	nytHost   = "https://www.nytimes.com"
)

func NYT() Ingestor {
	return nyt{
		simple: newSimple(nytSource, nytHost, newNYTCrawler()),
	}
}

func newNYTCrawler() crawler {
	return &nytCrawler{
		host: nytHost,
	}
}

func (c *nytCrawler) CleanURL(url string) string {
	splitQuestionMark := strings.Split(url, "?")
	if len(splitQuestionMark) > 1 {
		return splitQuestionMark[0]
	}

	splitHashtag := strings.Split(url, "#")
	if len(splitHashtag) > 1 {
		return splitHashtag[0]
	}
	return url
}

func (c *nytCrawler) Headline(n *html.Node) bool {
	return scrape.Attr(n, "itemprop") == "headline"
}

func (c *nytCrawler) ParseHeadline(n *html.Node) string {
	return scrape.Text(n)
}

func (c *nytCrawler) Text(n *html.Node) bool {
	return strings.Contains(scrape.Attr(n, "class"), "StoryBodyCompanionColumn")
}

func (c *nytCrawler) Author(n *html.Node) bool {
	return scrape.Attr(n, "itemprop") == "author creator"
}

func (c *nytCrawler) ParseAuthor(n *html.Node) string {
	return strings.Replace(scrape.Text(n), "By ", "", 1)
}

func (c *nytCrawler) Date(n *html.Node) bool {
	return scrape.Attr(n, "class") == "css-pnci9c eqgapgq0"
}

func (c *nytCrawler) ParseDate(n *html.Node) (time.Time, error) {
	t, err := time.Parse("2006-01-02", scrape.Attr(n, "datetime"))
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func (c *nytCrawler) ArticleLinks(n *html.Node) bool {
	if n.Data != "a" {
		return false
	}
	link := scrape.Attr(n, "href")

	_, err := dateFromURL(c.host, link, 0)
	if err != nil {
		return false
	}

	return true
}
