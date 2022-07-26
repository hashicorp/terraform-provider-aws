package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// BuildTagFilterList takes a []*ec2.Tag and produces a []*ec2.Filter that
// represents exact matches for all of the tag key/value pairs given in
// the tag set.
//
// The purpose of this function is to create values to pass in for
// the "Filters" attribute on most of the "Describe..." API functions
// in the EC2 API, to implement filtering by tag values e.g. in Terraform
// data sources that retrieve data about EC2 objects.
//
// It is conventional for an EC2 data source to include an attribute called
// "tags" which conforms to the schema returned by the tftags.TagsSchema() function.
// The value of this can then be converted to a tags slice using tagsFromMap,
// and the result finally passed in to this function.
//
// In Terraform configuration this would then look like this, to constrain
// results by name:
//
// tags {
//   Name = "my-awesome-subnet"
// }
func BuildTagFilterList(tags []*ec2.Tag) []*ec2.Filter {
	filters := make([]*ec2.Filter, len(tags))

	for i, tag := range tags {
		filters[i] = &ec2.Filter{
			Name:   aws.String(fmt.Sprintf("tag:%s", *tag.Key)),
			Values: []*string{tag.Value},
		}
	}

	return filters
}

// attributeFiltersFromMultimap returns an array of EC2 Filter objects to be used when listing resources.
//
// The keys of the specified map are the resource attributes names used in the filter - see the documentation
// for the relevant "Describe" action for a list of the valid names. The resource must match all the filters
// to be included in the result.
// The values of the specified map are lists of resource attribute values used in the filter. The resource can
// match any of the filter values to be included in the result.
// See https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Filtering.html#Filtering_Resources_CLI for more details.
func attributeFiltersFromMultimap(m map[string][]string) []*ec2.Filter {
	if len(m) == 0 {
		return nil
	}

	filters := []*ec2.Filter{}
	for k, v := range m {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String(k),
			Values: aws.StringSlice(v),
		})
	}

	return filters
}

// tagFiltersFromMap returns an array of EC2 Filter objects to be used when listing resources.
//
// The filters represent exact matches for all the resource tags in the given key/value map.
func tagFiltersFromMap(m map[string]interface{}) []*ec2.Filter {
	if len(m) == 0 {
		return nil
	}

	filters := []*ec2.Filter{}
	for _, tag := range Tags(tftags.New(m).IgnoreAWS()) {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String(fmt.Sprintf("tag:%s", aws.StringValue(tag.Key))),
			Values: []*string{tag.Value},
		})
	}

	return filters
}

// CustomFiltersSchema returns a *schema.Schema that represents
// a set of custom filtering criteria that a user can specify as input
// to a data source that wraps one of the many "Describe..." API calls
// in the EC2 API.
//
// It is conventional for an attribute of this type to be included
// as a top-level attribute called "filter". This is the "catch all" for
// filter combinations that are not possible to express using scalar
// attributes or tags. In Terraform configuration, the custom filter blocks
// then look like this:
//
// filter {
//   name   = "availabilityZone"
//   values = ["us-west-2a", "us-west-2b"]
// }
func CustomFiltersSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"values": {
					Type:     schema.TypeSet,
					Required: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}

// BuildCustomFilterList takes the set value extracted from a schema
// attribute conforming to the schema returned by CustomFiltersSchema,
// and transforms it into a []*ec2.Filter representing the same filter
// expressions which is ready to pass into the "Filters" attribute on most
// of the "Describe..." functions in the EC2 API.
//
// This function is intended only to be used in conjunction with
// ec2CustomFitlersSchema. See the docs on that function for more details
// on the configuration pattern this is intended to support.
func BuildCustomFilterList(filterSet *schema.Set) []*ec2.Filter {
	if filterSet == nil {
		return []*ec2.Filter{}
	}

	customFilters := filterSet.List()
	filters := make([]*ec2.Filter, len(customFilters))

	for filterIdx, customFilterI := range customFilters {
		customFilterMapI := customFilterI.(map[string]interface{})
		name := customFilterMapI["name"].(string)
		valuesI := customFilterMapI["values"].(*schema.Set).List()
		values := make([]*string, len(valuesI))
		for valueIdx, valueI := range valuesI {
			values[valueIdx] = aws.String(valueI.(string))
		}

		filters[filterIdx] = &ec2.Filter{
			Name:   &name,
			Values: values,
		}
	}

	return filters
}
