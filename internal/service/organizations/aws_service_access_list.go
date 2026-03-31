// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package organizations
// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/organizations/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/names"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	// listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
)

// TIP: ==== FILE STRUCTURE ====
// All list resources should follow this basic outline. Improve this list resource's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main data source struct with schema method
// 4. Read method
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @FrameworkListResource("aws_organizations_aws_service_access")
func newAWSServiceAccessResourceAsListResource() list.ListResourceWithConfigure {
	return &awsServiceAccessListResource{}
}

var _ list.ListResource = &awsServiceAccessListResource{}

type awsServiceAccessListResource struct {
	awsServiceAccessResource
	framework.WithList
}
// TIP: ==== LIST RESOURCE SCHEMA ===
// This is only needed if the resource type requires any attributes for listing, such as a parent ID.
// Otherwise, it can be removed.
// func (l *awsServiceAccessListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
// 	response.Schema = listschema.Schema{
// 		Attributes: map[string]listschema.Attribute{
// 			"parent_id": listschema.StringAttribute{
// 				Required:    true,
// 				Description: "ID of the Parent to list AWSServiceAccesss from.",
// 			},
// 		},
// 	}
// }

func (l *awsServiceAccessListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	// TIP: ==== LIST RESOURCE LIST ====
	// Generally, the List function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the config
	// 3. Get information about a resource from AWS
	// 4. Set the ID, arguments, and attributes
	// 5. Set the tags
	// 6. Set the state
	// TIP: -- 1. Get a client connection to the relevant service
	conn := l.Meta().OrganizationsClient(ctx)
	
	// TIP: -- 2. Fetch the config
	var query listAWSServiceAccessModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	
	// TIP: -- 3. Retrieve required attributes
	// If the resource type requires any attributes for listing, such as a parent ID, retrieve them here.
	parentID := query.ParentID.ValueString()

	tflog.Info(ctx, "Listing Resources", map[string]any{
		logging.ResourceAttributeKey("parent_id"): parentID,
	})
	
	// TIP: -- 4. Get information about a resource from AWS
	stream.Results = func(yield func(list.ListResult) bool) {
		input := organizations.ListAWSServiceAccesssInput{
			ParentId: parentID,
		}
		for item, err := range listAWSServiceAccesss(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}
			// TIP: -- 5. Set identifying attributes for logging
			// Set one or more logging fields with attributes that will identify the resource.
			// Typically, these will be the attributes used in the Resource Identity
			arn := aws.ToString(item.AWSServiceAccessArn)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrARN), arn)

			result := request.NewListResult(ctx)
			
			var data awsServiceAccessResourceModel
			// TIP: -- 6. Set the ID, arguments, and attributes
			// Using a field name prefix allows mapping fields such as `AWSServiceAccessId` to `ID`
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(l.flatten(ctx, &item, &data)...)
				if result.Diagnostics.HasError() {
					return
				}

				// TIP: -- 7. Set the display name
				result.DisplayName = aws.ToString(item.AWSServiceAccessName)
			})

			if !yield(result) {
				return
			}
		}
	}
}

// TIP: ==== DATA STRUCTURES ====
// With Terraform Plugin-Framework configurations are deserialized into
// Go types, providing type safety without the need for type assertions.
// These structs should match the schema definition exactly, and the `tfsdk`
// tag value should match the attribute name.
//
// Nested objects are represented in their own data struct. These will
// also have a corresponding attribute type mapping for use inside flex
// functions.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values
type listAWSServiceAccessModel struct {
	framework.WithRegionModel
	// TIP: -- 1. Include required attributes
	// If the resource type requires any attributes for listing, such as a parent ID, include them here.
	ParentID types.String `tfsdk:"parent_id"`
}

// TIP: ==== LISTING FUNCTION ====
// This listing function is written using an iterator pattern to handle pagination
func listAWSServiceAccesss(ctx context.Context, conn *organizations.Client, input *organizations.ListAWSServiceAccesssInput) iter.Seq2[awstypes.AWSServiceAccess, error] {
	return func(yield func(awstypes.AWSServiceAccess, error) bool) {
		pages := organizations.NewListAWSServiceAccesssPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.AWSServiceAccess{}, fmt.Errorf("listing Organizations AWS Service Access resources: %w", err))
				return
			}

			for _, item := range page.AWSServiceAccesss {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}

// TIP: ==== RESOURCE FLATTENING FUNCTION ====
// This function should be placed in the resource type's source file ("aws_service_access.go"). It may already be present.
// It is intended to perform the flattening of the results of the API call or calls used to populate a resource's values.
// It should replace the flattening portion of the resource type's Read function (`awsServiceAccessResource.Read`) and take the API results
// as parameters.
// The replaced section of the Read function should be
//	response.Diagnostics.Append(r.flatten(ctx, output, &data)...)
//	if response.Diagnostics.HasError() {
//		return
//	}
// func (r *awsServiceAccessResource) flatten(ctx context.Context, awsServiceAccess *awstypes.AWSServiceAccess, data *awsServiceAccessResourceModel) (diags diag.Diagnostics) {
// 	diags.Append(fwflex.Flatten(ctx, awsServiceAccess, data)...)
// 	return diags
// }
