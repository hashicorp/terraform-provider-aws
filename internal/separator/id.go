package separator

import "strings"

const (
	ResourceIdSeparator = ","
)

func ExpandResourceId(id string) []string {
	return strings.Split(id, ResourceIdSeparator)
}

func FlattenResourceId(idParts []string) string {
	return strings.Join(idParts, ResourceIdSeparator)
}
