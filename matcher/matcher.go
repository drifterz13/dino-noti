package matcher

import (
	"sort"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

func MatchItem(itemDescription string, searchTerms []string) (bool, string) {
	matches := fuzzy.RankFindFold(itemDescription, searchTerms)
	sort.Sort(matches)

	if len(matches) == 0 || matches[0].Distance == -1 {
		return false, ""
	}

	return true, matches[0].Target
}
