// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_securityhub_aggregator_v2", name="Aggregator V2")
// @ArnIdentity
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/securityhub;securityhub;securityhub.GetAggregatorV2Output")
// @Testing(serialize=true)
// @Testing(tagsTest=false)
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
func newAggregatorV2Resource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &aggregatorV2Resource{}, nil
}

type aggregatorV2Resource struct {
	framework.ResourceWithModel[aggregatorV2ResourceModel]
	framework.WithImportByIdentity
}

func (r *aggregatorV2Resource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"aggregation_region": schema.StringAttribute{
				Computed:    true,
				Description: "The AWS Region where data is aggregated.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"linked_regions": schema.SetAttribute{
				ElementType: types.StringType,
				CustomType:  fwtypes.SetOfStringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(fwvalidators.AWSRegion()),
				},
				Description: "The list of Regions linked to the aggregation Region.",
			},
			"region_linking_mode": schema.StringAttribute{
				Required:    true,
				Description: "Determines how Regions are linked: ALL_REGIONS, ALL_REGIONS_EXCEPT_SPECIFIED, or SPECIFIED_REGIONS.",
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *aggregatorV2Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data aggregatorV2ResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	var input securityhub.CreateAggregatorV2Input
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateAggregatorV2(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating Security Hub V2 Aggregator", err.Error())
		return
	}

	// Set values for unknowns.
	data.ARN = fwflex.StringToFramework(ctx, output.AggregatorV2Arn)
	data.AggregationRegion = fwflex.StringToFramework(ctx, output.AggregationRegion)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *aggregatorV2Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data aggregatorV2ResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ARN)
	output, err := findAggregatorV2ByARN(ctx, conn, arn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Security Hub V2 Aggregator (%s)", arn), err.Error())
		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *aggregatorV2Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new aggregatorV2ResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	if !new.RegionLinkingMode.Equal(old.RegionLinkingMode) || !new.LinkedRegions.Equal(old.LinkedRegions) {
		arn := fwflex.StringValueFromFramework(ctx, new.ARN)
		var input securityhub.UpdateAggregatorV2Input
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.AggregatorV2Arn = aws.String(arn)

		_, err := conn.UpdateAggregatorV2(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Security Hub V2 Aggregator (%s)", arn), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *aggregatorV2Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data aggregatorV2ResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ARN)
	input := securityhub.DeleteAggregatorV2Input{
		AggregatorV2Arn: aws.String(arn),
	}
	_, err := conn.DeleteAggregatorV2(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Security Hub V2 Aggregator (%s)", arn), err.Error())
		return
	}
}

func findAggregatorV2ByARN(ctx context.Context, conn *securityhub.Client, arn string) (*securityhub.GetAggregatorV2Output, error) {
	input := securityhub.GetAggregatorV2Input{
		AggregatorV2Arn: aws.String(arn),
	}

	return findAggregatorV2(ctx, conn, &input)
}

func findAggregatorV2(ctx context.Context, conn *securityhub.Client, input *securityhub.GetAggregatorV2Input) (*securityhub.GetAggregatorV2Output, error) {
	output, err := conn.GetAggregatorV2(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.ConflictException](err, "Security Hub V2 is not enabled") {
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

type aggregatorV2ResourceModel struct {
	framework.WithRegionModel
	ARN               types.String        `tfsdk:"arn"`
	AggregationRegion types.String        `tfsdk:"aggregation_region"`
	LinkedRegions     fwtypes.SetOfString `tfsdk:"linked_regions"`
	RegionLinkingMode types.String        `tfsdk:"region_linking_mode"`
	Tags              tftags.Map          `tfsdk:"tags"`
	TagsAll           tftags.Map          `tfsdk:"tags_all"`
}
