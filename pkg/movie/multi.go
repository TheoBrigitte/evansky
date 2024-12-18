package movie

import (
	"fmt"
	"math"
	"path"
	"sort"
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"
	log "github.com/sirupsen/logrus"
)

type Multi struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Year         int    `json:"year"`
	Path         string `json:"path"`
	IsDir        bool   `json:"isDir"`
	Language     string `json:"language"`
	OriginalName string `json:"originalName"`
	MediaType    string `json:"mediaType"`
	Error        string `json:"error,omitempty"`
}

// Path computes the ideal path for a movie given its title and year.
func (m *Multi) ComputePath() bool {
	p := m.path()

	if m.IsDir {
		m.Path = p
	} else {
		m.Path = path.Join(p, m.OriginalName)
	}

	return m.Path != m.OriginalName
}

func (m *Multi) path() string {
	if m.Year > 0 {
		return fmt.Sprintf("%s (%d)", m.Title, m.Year)
	}

	return m.Title
}

// Best return the first movie from the results.
// First one is considered best because searchMovies.Results response from the api is ordered with best matches first.
func BestMulti(searchMovies *gotmdb.SearchMulti) (*Multi, error) {
	if len(searchMovies.Results) <= 0 {
		return nil, NoResults
	}

	r := searchMovies.Results[0]
	log.Debugf("best: %#v\n", r)

	date := ""
	if r.ReleaseDate != "" {
		date = r.ReleaseDate
	}
	if r.FirstAirDate != "" {
		date = r.FirstAirDate
	}

	var releaseDate time.Time
	var err error
	if date != "" {
		releaseDate, err = time.Parse(releaseDateLayout, date)
		if err != nil {
			return nil, err
		}
	}

	m := &Multi{
		ID:        r.ID,
		Title:     r.Title,
		Year:      releaseDate.Year(),
		Language:  r.OriginalLanguage,
		MediaType: r.MediaType,
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
func BestByYearMulti(searchMovies *gotmdb.SearchMulti, year int) (*Multi, error) {
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
		dateInput := ""
		if searchMovies.Results[i].ReleaseDate != "" {
			dateInput = searchMovies.Results[i].ReleaseDate
		}
		if searchMovies.Results[i].FirstAirDate != "" {
			dateInput = searchMovies.Results[i].FirstAirDate
		}
		if dateInput == "" {
			continue
		}

		date, err := time.Parse(releaseDateLayout, dateInput)
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
	log.Debugf("bestByYear: %#v\n", r)
	releaseDate, err := time.Parse(releaseDateLayout, r.ReleaseDate)
	if err != nil {
		return nil, err
	}

	m := &Multi{
		ID:        r.ID,
		Title:     r.Title,
		Year:      releaseDate.Year(),
		Language:  r.OriginalLanguage,
		MediaType: r.MediaType,
	}

	return m, nil
}
