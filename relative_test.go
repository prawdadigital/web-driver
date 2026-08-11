package webdriver

import "testing"

func TestRelativeByBuilder(t *testing.T) {
	rel := With(ByTagName, "td").
		Above(nil).
		Below(nil).
		ToLeftOf(nil).
		ToRightOf(nil).
		Near(nil).
		NearWithin(nil, 25)

	by, value := rel.Root()
	if by != ByTagName || value != "td" {
		t.Errorf("Root() = %q, %q, want tag name/td", by, value)
	}

	filters := rel.Filters()
	wantKinds := []string{"above", "below", "left", "right", "near", "near"}
	if len(filters) != len(wantKinds) {
		t.Fatalf("Filters() length = %d, want %d", len(filters), len(wantKinds))
	}
	for i, f := range filters {
		if f.Kind != wantKinds[i] {
			t.Errorf("filter[%d].Kind = %q, want %q", i, f.Kind, wantKinds[i])
		}
	}
	// Near uses the default radius (0); NearWithin carries the explicit radius.
	if filters[4].Distance != 0 {
		t.Errorf("Near distance = %d, want 0", filters[4].Distance)
	}
	if filters[5].Distance != 25 {
		t.Errorf("NearWithin distance = %d, want 25", filters[5].Distance)
	}
}

func TestRelativeByImmutability(t *testing.T) {
	base := With(ByCSSSelector, "div")
	withFilter := base.Above(nil)
	if len(base.Filters()) != 0 {
		t.Errorf("base query was mutated: %d filters", len(base.Filters()))
	}
	if len(withFilter.Filters()) != 1 {
		t.Errorf("derived query has %d filters, want 1", len(withFilter.Filters()))
	}
}
