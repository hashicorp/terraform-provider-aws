package dms

import (
	"fmt"

	dmstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/aws/aws-sdk-go/aws"
)

// BuildTagFilterList takes a []*dms.Tag and produces a []*dms.Filter that
// represents exact matches for all of the tag key/value pairs given in
// the tag set.
//
// The purpose of this function is to create values to pass in for
// the "Filters" attribute on most of the "Describe..." API functions
// in the DMS API, to implement filtering by tag values e.g. in Terraform
// data sources that retrieve data about DMS objects.
//
// It is conventional for a data source to include an attribute called
// "tags" which conforms to the schema returned by the tftags.TagsSchema() function.
// The value of this can then be converted to a tags slice using tagsFromMap,
// and the result finally passed in to this function.
//
// In Terraform configuration this would then look like this, to constrain
// results by name:
//
//	tags {
//	  Name = "my-awesome-subnet"
//	}
func BuildTagFilterList(tags []*dmstypes.Tag) []*dmstypes.Filter {
	filters := make([]*dmstypes.Filter, len(tags))

	for i, tag := range tags {
		filters[i] = &dmstypes.Filter{
			Name:   aws.String(fmt.Sprintf("tag:%s", *tag.Key)),
			Values: []string{*tag.Value},
		}
	}
	return filters
}
