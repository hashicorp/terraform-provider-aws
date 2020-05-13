package filter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
)

// FromMultimap returns an array of RDS Filter objects to be used when listing resources.
//
// The keys of the specified map are the resource attributes names used in the filter - see the documentation
// for the relevant "Describe" action for a list of the valid names. The resource must match all the filters
// to be included in the result.
// The values of the specified map are lists of resource attribute values used in the filter. The resource can
// match any of the filter values to be included in the result.
// See https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_Filter.html for more details.
func FromMultimap(m map[string][]string) []*rds.Filter {
	if len(m) == 0 {
		return nil
	}

	filters := []*rds.Filter{}
	for k, v := range m {
		filters = append(filters, &rds.Filter{
			Name:   aws.String(k),
			Values: aws.StringSlice(v),
		})
	}

	return filters
}
