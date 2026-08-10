package selenium

// Relative (a.k.a. "friendly") locators, introduced in Selenium 4, find
// elements by their spatial relationship to a known anchor element. The
// geometry below mirrors Selenium's own relative-locator atom
// (javascript/atoms/locators/relative.js): candidates are filtered by the
// bounding-client-rect predicates and then sorted by the proximity of their
// centers to the last anchor.

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

// relativeLocatorScript filters and sorts candidate elements by their spatial
// relationship to the anchors. arguments[0] is the candidate element array;
// arguments[1] is the filter array, each entry {kind, anchor, distance}.
const relativeLocatorScript = `
const candidates = arguments[0];
const filters = arguments[1];
function rect(e) { return e.getBoundingClientRect(); }
function intersects(a, b) {
  return a.left <= b.left + b.width && b.left <= a.left + a.width &&
         a.top <= b.top + b.height && b.top <= a.top + a.height;
}
function passes(kind, cand, anchor, distance) {
  const expected = rect(anchor);
  const toFind = rect(cand);
  switch (kind) {
    case 'above': return toFind.top + toFind.height <= expected.top;
    case 'below': return toFind.top >= expected.top + expected.height;
    case 'left':  return toFind.left + toFind.width <= expected.left;
    case 'right': return toFind.left >= expected.left + expected.width;
    case 'near':
      if (cand === anchor) return false;
      const d = distance || 50;
      const big = { left: expected.left - d, top: expected.top - d,
                    width: expected.width + d * 2, height: expected.height + d * 2 };
      return intersects(big, toFind);
    default: return false;
  }
}
const matches = candidates.filter(function (el) {
  return filters.every(function (f) { return passes(f.kind, el, f.anchor, f.distance); });
});
const last = filters[filters.length - 1];
if (last) {
  const ar = rect(last.anchor);
  const ac = { x: ar.left + Math.max(1, ar.width) / 2, y: ar.top + Math.max(1, ar.height) / 2 };
  const dist = function (e) {
    const r = rect(e);
    const c = { x: r.left + Math.max(1, r.width) / 2, y: r.top + Math.max(1, r.height) / 2 };
    return Math.sqrt(Math.pow(ac.x - c.x, 2) + Math.pow(ac.y - c.y, 2));
  };
  matches.sort(function (a, b) { return dist(a) - dist(b); });
}
return matches;
`

func (wd *remoteWD) findRelative(rel RelativeBy) ([]byte, error) {
	candidates, err := wd.FindElements(rel.by, rel.value)
	if err != nil {
		return nil, err
	}

	candArg := make([]interface{}, len(candidates))
	for i, c := range candidates {
		candArg[i] = c
	}
	filterArg := make([]interface{}, len(rel.filters))
	for i, f := range rel.filters {
		filterArg[i] = map[string]interface{}{
			"kind":     f.kind,
			"anchor":   f.anchor,
			"distance": f.distance,
		}
	}

	return wd.ExecuteScriptRaw(relativeLocatorScript, []interface{}{candArg, filterArg})
}

func (wd *remoteWD) FindElementRelative(rel RelativeBy) (WebElement, error) {
	response, err := wd.findRelative(rel)
	if err != nil {
		return nil, err
	}
	elems, err := wd.DecodeElements(response)
	if err != nil {
		return nil, err
	}
	if len(elems) == 0 {
		return nil, &Error{Err: "no such element", Message: "no element matched the relative locator"}
	}
	return elems[0], nil
}

func (wd *remoteWD) FindElementsRelative(rel RelativeBy) ([]WebElement, error) {
	response, err := wd.findRelative(rel)
	if err != nil {
		return nil, err
	}
	return wd.DecodeElements(response)
}
