// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_securityhub_account_v2", name="Account V2")
// @ArnIdentity
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/securityhub;securityhub;securityhub.DescribeSecurityHubV2Output")
// @Testing(serialize=true)
// @Testing(tagsTest=false)
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(importStateIdAttribute="arn")
func newAccountV2Resource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &accountV2Resource{}, nil
}

type accountV2Resource struct {
	framework.ResourceWithModel[accountV2ResourceModel]
	framework.WithImportByIdentity
}

func (r *accountV2Resource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN:     framework.ARNAttributeComputedOnly(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *accountV2Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data accountV2ResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	input := securityhub.EnableSecurityHubV2Input{
		Tags: getTagsIn(ctx),
	}
	output, err := conn.EnableSecurityHubV2(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating Security Hub V2 Account", err.Error())
		return
	}

	data.ARN = fwflex.StringToFramework(ctx, output.HubV2Arn)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *accountV2Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data accountV2ResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	output, err := findAccountV2(ctx, conn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError("reading Security Hub V2 Account", err.Error())
		return
	}

	data.ARN = fwflex.StringToFramework(ctx, output.HubV2Arn)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accountV2Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data accountV2ResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	var input securityhub.DisableSecurityHubV2Input
	_, err := conn.DisableSecurityHubV2(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError("deleting Security Hub V2 Account", err.Error())
	}
}

func findAccountV2(ctx context.Context, conn *securityhub.Client) (*securityhub.DescribeSecurityHubV2Output, error) {
	var input securityhub.DescribeSecurityHubV2Input
	output, err := conn.DescribeSecurityHubV2(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type accountV2ResourceModel struct {
	framework.WithRegionModel
	ARN     types.String `tfsdk:"arn"`
	Tags    tftags.Map   `tfsdk:"tags"`
	TagsAll tftags.Map   `tfsdk:"tags_all"`
}
