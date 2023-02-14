package route53

// Exports for use in tests only.
var (
	CIDRLocationParseResourceID  = cidrLocationParseResourceID
	FindCIDRCollectionByID       = findCIDRCollectionByID
	FindCIDRLocationByTwoPartKey = findCIDRLocationByTwoPartKey
	ResourceCIDRCollection       = newResourceCIDRCollection
	ResourceCIDRLocation         = newResourceCIDRLocation
)
