package movie

import (
	"errors"
	"fmt"
	"math"
	"path"
	"sort"
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"
)

var NoResults = errors.New("no results")

// release_date field format.
// see: https://developers.themoviedb.org/3/search/search-movies
const releaseDateLayout = "2006-01-02"

type Movie struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Year         int    `json:"year"`
	Path         string `json:"path"`
	IsDir        bool   `json:"isDir"`
	Language     string `json:"language"`
	OriginalName string `json:"originalName"`
	Error        string `json:"error,omitempty"`
}

// Path computes the ideal path for a movie given its title and year.
func (m *Movie) ComputePath() bool {
	p := m.path()

	if m.IsDir {
		m.Path = p
	} else {
		m.Path = path.Join(p, m.OriginalName)
	}

	return m.Path != m.OriginalName
}

func (m *Movie) path() string {
	if m.Year > 0 {
		return fmt.Sprintf("%s (%d)", m.Title, m.Year)
	}

	return m.Title
}

// Best return the first movie from the results.
// First one is considered best because searchMovies.Results response from the api is ordered with best matches first.
func Best(searchMovies *gotmdb.SearchMovies) (*Movie, error) {
	if len(searchMovies.Results) <= 0 {
		return nil, NoResults
	}

	r := searchMovies.Results[0]

	releaseDate, err := time.Parse(releaseDateLayout, r.ReleaseDate)
	if err != nil {
		return nil, err
	}

	m := &Movie{
		ID:       r.ID,
		Title:    r.Title,
		Year:     releaseDate.Year(),
		Language: r.OriginalLanguage,
	}

	return m, nil
}

// BestByYear return the best match for year.
//
// Best match is computed by selecting the result with best "score".
// score = delta * index (the smaller the better)
// delta = math.Abs(movie.ReleaseDate.Year - year) (the smaller the better)
// index = index in the search results (the smaller the better)
//
// both delta and index are increased by 1 to avoid an entry being considered best because it is first (index=0) or matches exact year (delta=0, but index=499)
// searchMovies.Results entries are ordered with best matches first (0 is good, 499 is bad).
func BestByYear(searchMovies *gotmdb.SearchMovies, year int) (*Movie, error) {
	if len(searchMovies.Results) <= 0 {
		return nil, NoResults
	}

	// compute delta and index.
	type tmp struct {
		index int
		delta float64
	}
	var t []tmp
	for i := range searchMovies.Results {
		date, err := time.Parse(releaseDateLayout, searchMovies.Results[i].ReleaseDate)
		if err != nil {
			return nil, err
		}

		d := math.Abs(float64(year - date.Year()))
		ii := i + 1
		dd := d + 1
		//log.Debugf("delta: date=%s index=%d delta=%f\n", searchMovies.Results[i].ReleaseDate, ii, dd)
		t = append(t, tmp{index: ii, delta: dd})
	}

	// sort according to score.
	sort.SliceStable(t, func(i, j int) bool {
		a := float64(t[i].index) * t[i].delta
		b := float64(t[j].index) * t[j].delta
		//log.Debugf("sort: a=%#v => %f  b=%#v => %f\n", t[i], a, t[j], b)
		return a < b
	})

	//for i, e := range t {
	//	log.Debugf("list[%d] = %#v\n", i, e)
	//}

	// pick best score (first entry).
	r := searchMovies.Results[t[0].index-1]
	releaseDate, err := time.Parse(releaseDateLayout, r.ReleaseDate)
	if err != nil {
		return nil, err
	}

	m := &Movie{
		ID:       r.ID,
		Title:    r.Title,
		Year:     releaseDate.Year(),
		Language: r.OriginalLanguage,
	}

	return m, nil
}
