package article

import (
	"github.com/brandoneprice31/async"
	runner "gopkg.in/mgutz/dat.v1/sqlx-runner"
)

type (
	Service interface {
		Save([]Article) (int, int, error)
	}

	service struct {
		db DB
	}
)

func NewService(conn *runner.DB) Service {
	return &service{db: NewDB(conn)}
}

func (s *service) Save(aa []Article) (int, int, error) {
	if len(aa) == 0 {
		return 0, 0, nil
	}

	// extract urls
	uu := urls(aa)

	// find any overlapping articles
	aaOld, err := s.db.FindByURLs(uu)
	if err != nil {
		return 0, 0, err
	}
	overlapping := make(map[string]bool)
	for _, a := range aaOld {
		overlapping[a.URL] = true
	}

	// find the diff between the old and new
	diff := diff(aaOld, aa)

	// figure out which ones need to be updated and which need to be inserted
	urlMap, update, insert := make(map[string]bool), []Article{}, []Article{}
	for _, a := range aa {
		if urlMap[a.URL] {
			continue
		}

		if diff[a.URL] {
			update = append(update, a)
		} else if !overlapping[a.URL] {
			insert = append(insert, a)
		}

		urlMap[a.URL] = true
	}

	return len(insert), len(update),
		async.Parallel(
			func() error {
				if len(insert) > 0 {
					return s.db.Insert(insert)
				}
				return nil
			},
			func() error {
				if len(update) > 0 {
					return s.db.Update(update)
				}
				return nil
			},
		).ToError()
}

// extracts the urls from a slice of articles
func urls(aa []Article) []string {
	uu := make([]string, len(aa))
	for i, a := range aa {
		uu[i] = a.URL
	}
	return uu
}

// calculates the diff between two slices of articles
func diff(aaOld, aaNew []Article) map[string]bool {
	n := len(aaOld)
	if len(aaNew) > n {
		n = len(aaNew)
	}

	diffs := make(map[string]bool)
	for i := 0; i < n; i++ {
		if i >= len(aaOld) || i >= len(aaNew) {
			break
		}
		aOld, aNew := aaOld[i], aaNew[i]

		if aOld.Author != aNew.Author || aOld.Date != aNew.Date || aOld.Headline != aNew.Headline ||
			aOld.Source != aNew.Source || aOld.Text != aNew.Text || aOld.Title != aNew.Title ||
			aOld.URL != aNew.URL {
			diffs[aNew.URL] = true
		}
	}

	return diffs
}
