package ingestor

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"

	"github.com/brandoneprice31/news-ingestor/article"
)

type (
	fox struct {
		host string
	}
)

func FoxIngestor() Ingestor {

	return fox{
		host: "http://www.foxnews.com",
	}
}

func (c fox) Source() string {
	return "fox"
}

func (c fox) Ingest() ([]article.Article, error) {
	// request and parse the front page
	rootNode, err := htmlNode(c.host)
	if err != nil {
		return nil, err
	}

	links, err := c.articleLinks(rootNode)
	if err != nil {
		return nil, err
	}

	articles := []article.Article{}
	for _, l := range links {
		articleNode, err := htmlNode(l)
		if err != nil {
			continue
		}

		a, err := c.parseArticle(articleNode)
		if err != nil {
			continue
		}

		t, err := c.dateFromURL(l)
		if err != nil {
			continue
		}

		a.Date = *t
		a.Headline = true
		a.URL = l
		articles = append(articles, *a)
	}

	return articles, nil
}

func htmlNode(link string) (*html.Node, error) {
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}

	n, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (c fox) parseArticle(n *html.Node) (*article.Article, error) {
	headline, ok := scrape.Find(n, func(n *html.Node) bool {
		return scrape.Attr(n, "class") == "headline head1"
	})
	if !ok {
		return nil, errors.New("not ok")
	}

	title := scrape.Text(headline)

	body, ok := scrape.Find(n, func(n *html.Node) bool {
		return scrape.Attr(n, "class") == "article-body"
	})
	if !ok {
		return nil, errors.New("not ok")
	}

	text := scrape.Text(body)

	authorNode, ok := scrape.Find(n, func(n *html.Node) bool {
		return scrape.Attr(n, "rel") == "author"
	})
	if !ok {
		return nil, errors.New("not ok")
	}

	author := scrape.Text(authorNode)

	return &article.Article{
		Source: c.Source(),
		Title:  title,
		Author: author,
		Text:   text,
	}, nil
}

func (c fox) articleLinks(node *html.Node) ([]string, error) {
	// grab all articles and print them
	nodes := scrape.FindAll(node, c.articleFinder)

	links := make([]string, len(nodes))
	for i, n := range nodes {
		links[i] = scrape.Attr(n, "href")
	}

	return links, nil
}

func (c fox) articleFinder(n *html.Node) bool {
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

func (c fox) dateFromURL(url string) (*time.Time, error) {
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
