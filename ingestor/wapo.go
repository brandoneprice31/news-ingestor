package ingestor

import (
	"strings"
	"time"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

type (
	wapo struct {
		*simple
	}

	wapoCrawler struct {
		host string
	}
)

const (
	wapoSource = "wapo"
	wapoHost   = "https://www.washingtonpost.com"
)

func Wapo() Ingestor {
	return wapo{
		simple: newSimple(wapoSource, wapoHost, newwapoCrawler()),
	}
}

func newwapoCrawler() crawler {
	return &wapoCrawler{
		host: wapoHost,
	}
}

func (c *wapoCrawler) CleanURL(url string) string {
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

func (c *wapoCrawler) Headline(n *html.Node) bool {
	return scrape.Attr(n, "itemprop") == "headline"
}

func (c *wapoCrawler) ParseHeadline(n *html.Node) string {
	return scrape.Text(n)
}

func (c *wapoCrawler) Text(n *html.Node) bool {
	return strings.Contains(scrape.Attr(n, "itemprop"), "articleBody")
}

func (c *wapoCrawler) Author(n *html.Node) bool {
	return scrape.Attr(n, "itemprop") == "author"
}

func (c *wapoCrawler) ParseAuthor(n *html.Node) string {
	strs := strings.Split(scrape.Text(n), " ")
	return strings.Join(strs[1:3], " ")
}

func (c *wapoCrawler) Date(n *html.Node) bool {
	return scrape.Attr(n, "itemprop") == "datePublished"
}

func (c *wapoCrawler) ParseDate(n *html.Node) (time.Time, error) {
	str := scrape.Attr(n, "content")
	t, err := time.Parse("2006-01-02T15:04", str[:len(str)-4])
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func (c *wapoCrawler) ArticleLinks(n *html.Node) bool {
	if n.Data != "a" {
		return false
	}
	link := scrape.Attr(n, "href")

	_, err := dateFromURL(c.host, link, 2)
	if err != nil {
		return false
	}

	return true
}
