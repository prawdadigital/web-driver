package appium

// Appium element location strategies, for use as the "by" argument to
// FindElement/FindElements in addition to the standard webdriver By* strategies
// (which also work: ByID, ByXPATH, ByClassName, ...). Each strategy is supported
// by a subset of drivers, noted below.
const (
	// ByAccessibilityID matches the accessibility identifier / content
	// description. Supported by all native drivers.
	ByAccessibilityID = "accessibility id"

	// ByAndroidUIAutomator runs a UiSelector expression (UiAutomator2 driver).
	ByAndroidUIAutomator = "-android uiautomator"
	// ByAndroidViewTag matches an Espresso view tag (Espresso driver).
	ByAndroidViewTag = "-android viewtag"
	// ByAndroidDataMatcher matches via an Espresso data matcher (Espresso driver).
	ByAndroidDataMatcher = "-android datamatcher"

	// ByIOSClassChain matches an iOS class chain query (XCUITest driver).
	ByIOSClassChain = "-ios class chain"
	// ByIOSPredicateString matches an NSPredicate string (XCUITest driver).
	ByIOSPredicateString = "-ios predicate string"

	// ByImage matches by template image comparison (all drivers, via the images
	// plugin).
	ByImage = "-image"
	// ByCustom delegates to a registered custom locator plugin.
	ByCustom = "-custom"
)
