package metric

type RawScores struct {
	Production       map[string]float64
	Quality          map[string]float64
	Survival         map[string]float64
	RawSurvival      map[string]float64 // non-decayed blame line count
	Design           map[string]float64
	Breadth          map[string]float64
	DebtCleanup      map[string]float64
	Indispensability map[string]float64
	TotalCommits     map[string]int // total commit count per author (across all repos in domain)
}

func NewRawScores() *RawScores {
	return &RawScores{
		Production:       make(map[string]float64),
		Quality:          make(map[string]float64),
		Survival:         make(map[string]float64),
		RawSurvival:      make(map[string]float64),
		Design:           make(map[string]float64),
		Breadth:          make(map[string]float64),
		DebtCleanup:      make(map[string]float64),
		Indispensability: make(map[string]float64),
		TotalCommits:     make(map[string]int),
	}
}

func (r *RawScores) Authors() []string {
	seen := make(map[string]bool)
	var authors []string

	for _, m := range []map[string]float64{
		r.Production, r.Quality, r.Survival, r.Design,
		r.Breadth, r.DebtCleanup, r.Indispensability,
	} {
		for author := range m {
			if !seen[author] {
				seen[author] = true
				authors = append(authors, author)
			}
		}
	}
	return authors
}
