// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func newFilter(name string, values []string) awstypes.Filter {
	return awstypes.Filter{
		Name:   aws.String(name),
		Values: values,
	}
}

// newTagFilterList takes a []*ec2.Tag and produces a []*ec2.Filter that
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
//	tags {
//	  Name = "my-awesome-subnet"
//	}
func newTagFilterList(tags []awstypes.Tag) []awstypes.Filter {
	return tfslices.ApplyToAll(tags, func(tag awstypes.Tag) awstypes.Filter {
		return newFilter("tag:"+aws.ToString(tag.Key), []string{aws.ToString(tag.Value)})
	})
}

// attributeFiltersFromMultimap returns an array of EC2 Filter objects to be used when listing resources.
//
// The keys of the specified map are the resource attributes names used in the filter - see the documentation
// for the relevant "Describe" action for a list of the valid names. The resource must match all the filters
// to be included in the result.
// The values of the specified map are lists of resource attribute values used in the filter. The resource can
// match any of the filter values to be included in the result.
// See https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Filtering.html#Filtering_Resources_CLI for more details.
func attributeFiltersFromMultimap(m map[string][]string) []awstypes.Filter {
	if len(m) == 0 {
		return nil
	}

	filters := []awstypes.Filter{}

	for k, v := range m {
		filters = append(filters, newFilter(k, v))
	}

	return filters
}

// tagFilters returns an array of EC2 Filter objects to be used when listing resources by tag.
func tagFilters(ctx context.Context) []awstypes.Filter {
	return newTagFilterList(getTagsIn(ctx))
}

// customFiltersSchema returns a *schema.Schema that represents
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
//	filter {
//	  name   = "availabilityZone"
//	  values = ["us-west-2a", "us-west-2b"]
//	}
func customFiltersSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
				},
				names.AttrValues: {
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

func customRequiredFiltersSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
				},
				names.AttrValues: {
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

// customFiltersBlock is the Plugin Framework variant of customFiltersSchema.
func customFiltersBlock() datasourceschema.Block {
	return datasourceschema.SetNestedBlock{
		NestedObject: datasourceschema.NestedBlockObject{
			Attributes: map[string]datasourceschema.Attribute{
				names.AttrName: datasourceschema.StringAttribute{
					Required: true,
				},
				names.AttrValues: datasourceschema.SetAttribute{
					ElementType: types.StringType,
					Required:    true,
				},
			},
		},
	}
}

// customFilterModel represents a single configured filter.
type customFilterModel struct {
	Name   types.String `tfsdk:"name"`
	Values types.Set    `tfsdk:"values"`
}

// newCustomFilterList takes the set value extracted from a schema
// attribute conforming to the schema returned by CustomFiltersSchema,
// and transforms it into a []*ec2.Filter representing the same filter
// expressions which is ready to pass into the "Filters" attribute on most
// of the "Describe..." functions in the EC2 API.
//
// This function is intended only to be used in conjunction with
// CustomFiltersSchema. See the docs on that function for more details
// on the configuration pattern this is intended to support.
func newCustomFilterList(s *schema.Set) []awstypes.Filter {
	if s == nil {
		return []awstypes.Filter{}
	}

	return tfslices.ApplyToAll(s.List(), func(tfList interface{}) awstypes.Filter {
		tfMap := tfList.(map[string]interface{})
		return newFilter(tfMap[names.AttrName].(string), flex.ExpandStringValueEmptySet(tfMap[names.AttrValues].(*schema.Set)))
	})
}

func newCustomFilterListFramework(ctx context.Context, filterSet types.Set) []awstypes.Filter {
	if filterSet.IsNull() || filterSet.IsUnknown() {
		return nil
	}

	var filters []awstypes.Filter

	for _, v := range filterSet.Elements() {
		var data customFilterModel

		if tfsdk.ValueAs(ctx, v, &data).HasError() {
			continue
		}

		if data.Name.IsNull() || data.Name.IsUnknown() {
			continue
		}

		if v := fwflex.ExpandFrameworkStringValueSet(ctx, data.Values); v != nil {
			filters = append(filters, awstypes.Filter{
				Name:   fwflex.StringFromFramework(ctx, data.Name),
				Values: v,
			})
		}
	}

	return filters
}

// newAttributeFilterList takes a flat map of scalar attributes (most
// likely values extracted from a *schema.ResourceData on an EC2-querying
// data source) and produces a []*ec2.Filter representing an exact match
// for each of the given non-empty attributes.
//
// The keys of the given attributes map are the attribute names expected
// by the EC2 API, which are usually either in camelcase or with dash-separated
// words. We conventionally map these to underscore-separated identifiers
// with the same words when presenting these as data source query attributes
// in Terraform.
//
// It's the callers responsibility to transform any non-string values into
// the appropriate string serialization required by the AWS API when
// encoding the given filter. Any attributes given with empty string values
// are ignored, assuming that the user wishes to leave that attribute
// unconstrained while filtering.
//
// The purpose of this function is to create values to pass in
// for the "Filters" attribute on most of the "Describe..." API functions in
// the EC2 API, to aid in the implementation of Terraform data sources that
// retrieve data about EC2 objects.
func newAttributeFilterList(m map[string]string) []awstypes.Filter {
	var filters []awstypes.Filter

	// Sort the filters by name to make the output deterministic.
	names := tfmaps.Keys(m)
	sort.Strings(names)

	for _, name := range names {
		value := m[name]
		if value == "" {
			continue
		}

		filters = append(filters, newFilter(name, []string{value}))
	}

	return filters
}
