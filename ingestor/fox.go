package ingestor

import (
	"errors"
	"strconv"
	"strings"
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

func FoxIngestor() Ingestor {
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

	_, err := c.dateFromURL(link)
	if err != nil {
		return false
	}

	return true
}

func (c foxCrawler) dateFromURL(url string) (*time.Time, error) {
	lengthOfHostAndSlash := len(c.host) + 1
	if len(url) <= lengthOfHostAndSlash {
		return nil, errors.New("could not parse url")
	}
	path := url[lengthOfHostAndSlash:]

	// now check that this path is prefixed with a date
	pathEntries := strings.Split(path, "/")
	if len(pathEntries) < 5 {
		return nil, errors.New("could not parse url")
	}

	yearStr, monthStr, dayStr := pathEntries[1], pathEntries[2], pathEntries[3]

	year, yErr := strconv.Atoi(yearStr)
	month, mErr := strconv.Atoi(monthStr)
	day, dErr := strconv.Atoi(dayStr)
	if yErr != nil || mErr != nil || dErr != nil {
		return nil, errors.New("could not parse url")
	}

	d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return &d, nil
}
