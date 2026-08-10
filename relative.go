package webdriver

// Relative (a.k.a. "friendly") locators, introduced in Selenium 4, find
// elements by their spatial relationship to a known anchor element. This file
// defines the builder and its accessors (the contract); the geometry is
// evaluated by the transport (see the remote package), which mirrors Selenium's
// own relative-locator atom.

// relativeFilter is a single spatial constraint in a RelativeBy query.
type relativeFilter struct {
	kind     string
	anchor   WebElement
	distance int
}

// RelativeBy describes a relative-locator query: a base locator that selects
// candidate elements, refined by one or more spatial filters. Build one with
// With and chain the Above/Below/ToLeftOf/ToRightOf/Near methods, then pass it
// to WebDriver.FindElementRelative or FindElementsRelative.
type RelativeBy struct {
	by, value string
	filters   []relativeFilter
}

// With starts a relative-locator query whose candidate elements are those
// matching the given base locator, e.g. With(ByTagName, "div").
func With(by, value string) RelativeBy {
	return RelativeBy{by: by, value: value}
}

func (r RelativeBy) add(kind string, anchor WebElement, distance int) RelativeBy {
	r.filters = append(r.filters, relativeFilter{kind: kind, anchor: anchor, distance: distance})
	return r
}

// Above restricts candidates to those positioned above the anchor element.
func (r RelativeBy) Above(anchor WebElement) RelativeBy {
	return r.add("above", anchor, 0)
}

// Below restricts candidates to those positioned below the anchor element.
func (r RelativeBy) Below(anchor WebElement) RelativeBy {
	return r.add("below", anchor, 0)
}

// ToLeftOf restricts candidates to those positioned to the left of the anchor.
func (r RelativeBy) ToLeftOf(anchor WebElement) RelativeBy {
	return r.add("left", anchor, 0)
}

// ToRightOf restricts candidates to those positioned to the right of the anchor.
func (r RelativeBy) ToRightOf(anchor WebElement) RelativeBy {
	return r.add("right", anchor, 0)
}

// Near restricts candidates to those within 50 pixels of the anchor, in any
// direction. An element is never near itself.
func (r RelativeBy) Near(anchor WebElement) RelativeBy {
	return r.add("near", anchor, 0)
}

// NearWithin behaves like Near but uses the given pixel radius instead of the
// 50-pixel default.
func (r RelativeBy) NearWithin(anchor WebElement, distance int) RelativeBy {
	return r.add("near", anchor, distance)
}

// RelativeFilter is a single spatial filter, exposed for WebDriver
// implementations that evaluate a relative locator.
type RelativeFilter struct {
	// Kind is the filter direction: "above", "below", "left", "right" or "near".
	Kind string
	// Anchor is the reference element the candidate is positioned relative to.
	Anchor WebElement
	// Distance is the radius in pixels for the "near" filter (0 means default).
	Distance int
}

// Root returns the base locator strategy and value used to select candidate
// elements before the spatial filters are applied.
func (r RelativeBy) Root() (by, value string) {
	return r.by, r.value
}

// Filters returns the ordered spatial filters that refine the candidates.
func (r RelativeBy) Filters() []RelativeFilter {
	out := make([]RelativeFilter, len(r.filters))
	for i, f := range r.filters {
		out[i] = RelativeFilter{Kind: f.kind, Anchor: f.anchor, Distance: f.distance}
	}
	return out
}
