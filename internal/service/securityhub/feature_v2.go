// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

// @FrameworkResource("aws_securityhub_feature_v2", name="Feature V2")
// @IdentityAttribute("feature_name")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/securityhub/types;awstypes;awstypes.FeatureDetail")
// @Testing(serialize=true)
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(importStateIdAttribute="feature_name")
func newFeatureV2Resource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &featureV2Resource{}, nil
}

type featureV2Resource struct {
	framework.ResourceWithModel[featureV2ResourceModel]
	framework.WithImportByIdentity
}

func (r *featureV2Resource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"feature_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the opt-in feature to enable. Valid values: NETWORK_SCANNING.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.FeatureName](),
				},
			},
			"feature_status": schema.StringAttribute{
				Computed:    true,
				Description: "The current enablement status of the feature. Valid values: ENABLED, DISABLED.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *featureV2Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data featureV2ResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	featureName := awstypes.FeatureName(fwflex.StringValueFromFramework(ctx, data.FeatureName))
	input := securityhub.EnableSecurityHubFeatureV2Input{
		FeatureName: featureName,
	}
	_, err := conn.EnableSecurityHubFeatureV2(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("enabling Security Hub V2 Feature (%s)", featureName), err.Error())
		return
	}

	// Read back the current status of the feature.
	feature, err := findFeatureV2ByName(ctx, conn, featureName)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Security Hub V2 Feature (%s)", featureName), err.Error())
		return
	}

	data.FeatureStatus = fwflex.StringValueToFramework(ctx, feature.FeatureStatus)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *featureV2Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data featureV2ResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	featureName := awstypes.FeatureName(fwflex.StringValueFromFramework(ctx, data.FeatureName))
	feature, err := findFeatureV2ByName(ctx, conn, featureName)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Security Hub V2 Feature (%s)", featureName), err.Error())
		return
	}

	data.FeatureStatus = fwflex.StringValueToFramework(ctx, feature.FeatureStatus)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *featureV2Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data featureV2ResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	featureName := awstypes.FeatureName(fwflex.StringValueFromFramework(ctx, data.FeatureName))
	input := securityhub.DisableSecurityHubFeatureV2Input{
		FeatureName: featureName,
	}
	_, err := conn.DisableSecurityHubFeatureV2(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("disabling Security Hub V2 Feature (%s)", featureName), err.Error())
		return
	}
}

func findFeatureV2ByName(ctx context.Context, conn *securityhub.Client, featureName awstypes.FeatureName) (*awstypes.FeatureDetail, error) {
	output, err := findAccountV2(ctx, conn)

	if err != nil {
		return nil, err
	}

	feature, ok := output.Features[string(featureName)]

	// The feature is only considered to exist when it is present and enabled.
	// A missing entry or a DISABLED status means the resource no longer exists.
	if !ok || feature.FeatureStatus == awstypes.FeatureStatusDisabled {
		return nil, &retry.NotFoundError{
			Message: fmt.Sprintf("Security Hub V2 feature %s not enabled", featureName),
		}
	}

	return &feature, nil
}

type featureV2ResourceModel struct {
	framework.WithRegionModel
	FeatureName   types.String `tfsdk:"feature_name"`
	FeatureStatus types.String `tfsdk:"feature_status"`
}
