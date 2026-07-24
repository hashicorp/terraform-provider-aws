// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package quicksight

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_ip_restriction", name="IP Restriction")
func newIPRestrictionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &ipRestrictionResource{}

	return r, nil
}

type ipRestrictionResource struct {
	framework.ResourceWithModel[ipRestrictionResourceModel]
}

func (r *ipRestrictionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAWSAccountID: quicksightschema.AWSAccountIDAttribute(),
			names.AttrEnabled: schema.BoolAttribute{
				Required: true,
			},
			"ip_restriction_rule_map": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Map{
					mapvalidator.KeysAre(fwvalidators.IPv4CIDRNetworkAddress()),
					mapvalidator.ValueStringsAre(stringvalidator.LengthBetween(0, 150)),
				},
			},
			"vpc_endpoint_id_restriction_rule_map": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Map{
					mapvalidator.KeysAre(stringvalidator.RegexMatches(regexache.MustCompile(`^vpce-[0-9a-z]*$`), "value must be a VPC endpoint ID")),
					mapvalidator.ValueStringsAre(stringvalidator.LengthBetween(0, 150)),
				},
			},
			"vpc_id_restriction_rule_map": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Map{
					mapvalidator.KeysAre(stringvalidator.RegexMatches(regexache.MustCompile(`^vpc-[0-9a-z]*$`), "value must be a VPC ID")),
					mapvalidator.ValueStringsAre(stringvalidator.LengthBetween(0, 150)),
				},
			},
		},
	}
}

func (r *ipRestrictionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data ipRestrictionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	if data.AWSAccountID.IsUnknown() {
		data.AWSAccountID = fwflex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	}

	conn := r.Meta().QuickSightClient(ctx)

	accountID := fwflex.StringValueFromFramework(ctx, data.AWSAccountID)
	var input quicksight.UpdateIpRestrictionInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Send empty maps, not nil.
	if data.IPRestrictionRuleMap.IsNull() {
		input.IpRestrictionRuleMap = map[string]string{}
	}
	if data.VPCEndpointIDRestrictionRuleMap.IsNull() {
		input.VpcEndpointIdRestrictionRuleMap = map[string]string{}
	}
	if data.VPCIDRestrictionRuleMap.IsNull() {
		input.VpcIdRestrictionRuleMap = map[string]string{}
	}

	_, err := conn.UpdateIpRestriction(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Quicksight IP Restriction (%s)", accountID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *ipRestrictionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data ipRestrictionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	accountID := fwflex.StringValueFromFramework(ctx, data.AWSAccountID)
	output, err := findIPRestrictionByID(ctx, conn, accountID)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Quicksight IP Restriction (%s)", accountID), err.Error())

		return
	}

	// Set attributes for import.
	// API returns empty maps, not nil.
	if data.IPRestrictionRuleMap.IsNull() && len(output.IpRestrictionRuleMap) == 0 {
		output.IpRestrictionRuleMap = nil
	}
	if data.VPCEndpointIDRestrictionRuleMap.IsNull() && len(output.VpcEndpointIdRestrictionRuleMap) == 0 {
		output.VpcEndpointIdRestrictionRuleMap = nil
	}
	if data.VPCIDRestrictionRuleMap.IsNull() && len(output.VpcIdRestrictionRuleMap) == 0 {
		output.VpcIdRestrictionRuleMap = nil
	}
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *ipRestrictionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old ipRestrictionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	accountID := fwflex.StringValueFromFramework(ctx, new.AWSAccountID)
	var input quicksight.UpdateIpRestrictionInput
	response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Send empty maps, not nil.
	if new.IPRestrictionRuleMap.IsNull() {
		input.IpRestrictionRuleMap = map[string]string{}
	}
	if new.VPCEndpointIDRestrictionRuleMap.IsNull() {
		input.VpcEndpointIdRestrictionRuleMap = map[string]string{}
	}
	if new.VPCIDRestrictionRuleMap.IsNull() {
		input.VpcIdRestrictionRuleMap = map[string]string{}
	}

	_, err := conn.UpdateIpRestriction(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Quicksight IP Restriction (%s)", accountID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *ipRestrictionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data ipRestrictionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	accountID := fwflex.StringValueFromFramework(ctx, data.AWSAccountID)
	input := quicksight.UpdateIpRestrictionInput{
		AwsAccountId:                    aws.String(accountID),
		Enabled:                         aws.Bool(false),
		IpRestrictionRuleMap:            map[string]string{},
		VpcEndpointIdRestrictionRuleMap: map[string]string{},
		VpcIdRestrictionRuleMap:         map[string]string{},
	}
	_, err := conn.UpdateIpRestriction(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Quicksight IP Restriction (%s)", accountID), err.Error())

		return
	}
}

func (r *ipRestrictionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrAWSAccountID), request, response)
}

func findIPRestrictionByID(ctx context.Context, conn *quicksight.Client, id string) (*quicksight.DescribeIpRestrictionOutput, error) {
	input := quicksight.DescribeIpRestrictionInput{
		AwsAccountId: aws.String(id),
	}
	output, err := conn.DescribeIpRestriction(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || (!aws.ToBool(output.Enabled) && len(output.IpRestrictionRuleMap) == 0 && len(output.VpcEndpointIdRestrictionRuleMap) == 0 && len(output.VpcIdRestrictionRuleMap) == 0) {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type ipRestrictionResourceModel struct {
	framework.WithRegionModel
	AWSAccountID                    types.String        `tfsdk:"aws_account_id"`
	Enabled                         types.Bool          `tfsdk:"enabled"`
	IPRestrictionRuleMap            fwtypes.MapOfString `tfsdk:"ip_restriction_rule_map"`
	VPCEndpointIDRestrictionRuleMap fwtypes.MapOfString `tfsdk:"vpc_endpoint_id_restriction_rule_map"`
	VPCIDRestrictionRuleMap         fwtypes.MapOfString `tfsdk:"vpc_id_restriction_rule_map"`
}
