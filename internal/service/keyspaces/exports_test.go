package keyspaces

// Exports for use in tests only.
var (
	ResourceKeyspace = resourceKeyspace
	ResourceTable    = resourceTable

	FindKeyspaceByName    = findKeyspaceByName
	FindTableByTwoPartKey = findTableByTwoPartKey

	TableParseResourceID = tableParseResourceID
)
