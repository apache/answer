package tag_common

import "testing"

// The bug: the search term was formatted into LOWER(%s), which put the function
// name into the value rather than applying it to the column, so the query became
// slug_name LIKE '%LOWER(coco)%' and matched nothing. Only display_name did any
// work, and that is case-sensitive on Postgres -- so typing a tag the way tags
// are actually written returned "no such tag".
func TestSearchTermIsLoweredNotWrapped(t *testing.T) {
	for _, in := range []string{"Coco", "COCO", "coco"} {
		got := searchTermForTag(in)
		if got != "coco" {
			t.Errorf("searchTermForTag(%q) = %q, want %q", in, got, "coco")
		}
	}
}
