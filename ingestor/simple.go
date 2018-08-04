package ingestor

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/brandoneprice31/news-ingestor/article"
	"github.com/rs/zerolog"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

type (
	simple struct {
		host    string
		source  string
		crawler crawler
	}

	crawler interface {
		ArticleLinks(n *html.Node) bool
		CleanURL(string) string
		Headline(n *html.Node) bool
		ParseHeadline(n *html.Node) string
		Text(n *html.Node) bool
		Author(n *html.Node) bool
		ParseAuthor(n *html.Node) string
		Date(n *html.Node) bool
		ParseDate(n *html.Node) (time.Time, error)
	}
)

var (
	log = zerolog.New(os.Stderr).With().Timestamp().Logger()
)

func newSimple(source, host string, crawler crawler) *simple {
	return &simple{
		source:  source,
		host:    host,
		crawler: crawler,
	}
}

func (c *simple) Source() string {
	return c.source
}

func (c *simple) Ingest() ([]article.Article, error) {
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
			log.Error().Str("source", c.Source()).Err(err).Msg("could not get node from link")
			continue
		}

		a, err := c.parseArticle(articleNode)
		if err != nil {
			log.Error().Str("source", c.Source()).Err(err).Msg("could not parse node")
			continue
		}

		a.Headline = true
		a.URL = c.crawler.CleanURL(l)
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

func (c *simple) parseArticle(n *html.Node) (*article.Article, error) {
	headline, ok := scrape.Find(n, c.crawler.Headline)
	if !ok {
		return nil, errors.New("can't find headline")
	}
	title := c.crawler.ParseHeadline(headline)

	tt := scrape.FindAll(n, c.crawler.Text)
	if len(tt) == 0 {
		return nil, errors.New("can't find text")
	}
	text := ""
	for _, t := range tt {
		text += scrape.Text(t)
	}

	authorNode, ok := scrape.Find(n, c.crawler.Author)
	if !ok {
		return nil, errors.New("can't find author")
	}
	author := c.crawler.ParseAuthor(authorNode)

	dateNode, ok := scrape.Find(n, c.crawler.Date)
	if !ok {
		return nil, errors.New("can't find date")
	}
	date, err := c.crawler.ParseDate(dateNode)
	if err != nil {
		return nil, err
	}

	return &article.Article{
		Source: c.Source(),
		Title:  title,
		Author: author,
		Text:   text,
		Date:   date,
	}, nil
}

func (c *simple) articleLinks(node *html.Node) ([]string, error) {
	// grab all articles and print them
	nodes := scrape.FindAll(node, c.crawler.ArticleLinks)

	links := make([]string, len(nodes))
	for i, n := range nodes {
		links[i] = scrape.Attr(n, "href")
	}

	return links, nil
}

// Helper that looks for a date in a url given the start index of the date.
func dateFromURL(host, url string, startIndex int) (*time.Time, error) {
	lengthOfHostAndSlash := len(host) + 1
	if len(url) <= lengthOfHostAndSlash {
		return nil, errors.New("could not parse url")
	}
	path := url[lengthOfHostAndSlash:]

	// now check that this path is prefixed with a date
	pathEntries := strings.Split(path, "/")
	if len(pathEntries) < startIndex+4 {
		return nil, errors.New("could not parse url")
	}

	yearStr, monthStr, dayStr := pathEntries[startIndex], pathEntries[startIndex+1], pathEntries[startIndex+2]

	year, yErr := strconv.Atoi(yearStr)
	month, mErr := strconv.Atoi(monthStr)
	day, dErr := strconv.Atoi(dayStr)
	if yErr != nil || mErr != nil || dErr != nil {
		return nil, errors.New("could not parse url")
	}

	d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return &d, nil
}
