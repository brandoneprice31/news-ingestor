package ingestor

import (
	"time"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

type (
	fox struct {
		*simple
	}

	foxCrawler struct {
		host string
	}
)

const (
	foxSource = "fox"
	foxHost   = "http://www.foxnews.com"
)

func Fox() Ingestor {
	return fox{
		simple: newSimple(foxSource, foxHost, newFoxCrawler()),
	}
}

func newFoxCrawler() crawler {
	return &foxCrawler{
		host: foxHost,
	}
}

func (c *foxCrawler) CleanURL(url string) string {
	return url
}

func (c *foxCrawler) Headline(n *html.Node) bool {
	return scrape.Attr(n, "class") == "headline head1"
}

func (c *foxCrawler) ParseHeadline(n *html.Node) string {
	return scrape.Text(n)
}

func (c *foxCrawler) Text(n *html.Node) bool {
	return scrape.Attr(n, "class") == "article-body"
}

func (c *foxCrawler) Author(n *html.Node) bool {
	return scrape.Attr(n, "rel") == "author"
}

func (c *foxCrawler) ParseAuthor(n *html.Node) string {
	return scrape.Text(n)
}

func (c *foxCrawler) Date(n *html.Node) bool {
	return scrape.Attr(n, "data-time-published") != ""
}

func (c *foxCrawler) ParseDate(n *html.Node) (time.Time, error) {
	t, err := time.Parse("2006-01-02T15:04:03Z-0400", scrape.Attr(n, "data-time-published"))
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func (c *foxCrawler) ArticleLinks(n *html.Node) bool {
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
