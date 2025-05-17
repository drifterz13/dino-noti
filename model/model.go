package model

type MatchedItem struct {
	Index        int
	URL          string
	OriginalName string
	MatchedName  string
	Price        string
	ImageURL     string
}

type ScrapeItem struct {
	URL      string
	Name     string
	Price    string
	ImageURL string
}
