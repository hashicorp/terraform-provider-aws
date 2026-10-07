// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_iam_instance_profile")
func newInstanceProfileResourceAsListResource() inttypes.ListResourceForSDK {
	l := instanceProfileListResource{}
	l.SetResourceSchema(resourceInstanceProfile())

	return &l
}

var _ list.ListResource = &instanceProfileListResource{}

type instanceProfileListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type instanceProfileListResourceModel struct {
	PathPrefix types.String `tfsdk:"path_prefix"`
}

func (l *instanceProfileListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"path_prefix": listschema.StringAttribute{
				Optional:    true,
				Description: "Path prefix on which to filter instance profile names.",
				Validators: []validator.String{
					validPolicyPathFramework,
				},
			},
		},
	}
}

func (l *instanceProfileListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.IAMClient(ctx)

	var query instanceProfileListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input iam.ListInstanceProfilesInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	tflog.Info(ctx, "Listing IAM Instance Profiles")

	stream.Results = func(yield func(list.ListResult) bool) {
		for instanceProfile, err := range listInstanceProfiles(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			name := aws.ToString(instanceProfile.InstanceProfileName)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrName), name)
			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(name)
			rd.Set(names.AttrName, name)

			if request.IncludeResource {
				tflog.Info(ctx, "Reading IAM Instance Profile")
				fullInstanceProfile, err := findInstanceProfileByName(ctx, conn, name)
				if err != nil {
					yield(fwdiag.NewListResultErrorDiagnostic(err))
					return
				}

				resourceInstanceProfileFlatten(ctx, fullInstanceProfile, rd)
			}

			result.DisplayName = resourceInstanceProfileDisplayName(instanceProfile)

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

func resourceInstanceProfileDisplayName(instanceProfile awstypes.InstanceProfile) string {
	var buf strings.Builder

	path := aws.ToString(instanceProfile.Path)
	buf.WriteString(strings.TrimPrefix(path, "/"))

	buf.WriteString(aws.ToString(instanceProfile.InstanceProfileName))

	return buf.String()
}

func listInstanceProfiles(ctx context.Context, conn *iam.Client, input *iam.ListInstanceProfilesInput) iter.Seq2[awstypes.InstanceProfile, error] {
	return func(yield func(awstypes.InstanceProfile, error) bool) {
		pages := iam.NewListInstanceProfilesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.InstanceProfile{}, fmt.Errorf("listing IAM Instance Profiles: %w", err))
				return
			}

			for _, item := range page.InstanceProfiles {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
