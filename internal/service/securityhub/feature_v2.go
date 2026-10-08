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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @FrameworkResource("aws_securityhub_feature_v2", name="Feature V2")
// @IdentityAttribute("feature_name")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/securityhub/types;awstypes;awstypes.FeatureDetail")
// @Testing(checkDestroyNoop=true)
// @Testing(serialize=true)
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(importStateIdAttribute="feature_name")
func newFeatureV2Resource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &featureV2Resource{}, nil
}

type featureV2Resource struct {
	framework.ResourceWithModel[featureV2ResourceModel]
	framework.WithNoOpDelete
	framework.WithImportByIdentity
}

func (r *featureV2Resource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"feature_name": schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.FeatureName](),
				Required:    true,
				Description: "The name of the opt-in feature to enable. Valid values: NETWORK_SCANNING.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"feature_status": schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.FeatureStatus](),
				Required:    true,
				Description: "The current enablement status of the feature. Valid values: ENABLED, DISABLED.",
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

	var err error
	name := data.FeatureName.ValueEnum()
	switch data.FeatureStatus.ValueEnum() {
	case awstypes.FeatureStatusEnabled:
		err = enableFeatureV2(ctx, conn, name)
	case awstypes.FeatureStatusDisabled:
		err = disableFeatureV2(ctx, conn, name)
	}
	if err != nil {
		response.Diagnostics.AddError("", err.Error())
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *featureV2Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data featureV2ResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	name := data.FeatureName.ValueEnum()
	feature, err := findFeatureV2ByName(ctx, conn, name)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Security Hub V2 Feature (%s)", name), err.Error())
		return
	}

	data.FeatureStatus = fwtypes.StringEnumValue(feature.FeatureStatus)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *featureV2Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data featureV2ResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	var err error
	name := data.FeatureName.ValueEnum()
	switch data.FeatureStatus.ValueEnum() {
	case awstypes.FeatureStatusEnabled:
		err = enableFeatureV2(ctx, conn, name)
	case awstypes.FeatureStatusDisabled:
		err = disableFeatureV2(ctx, conn, name)
	}
	if err != nil {
		response.Diagnostics.AddError("", err.Error())
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func findFeatureV2ByName(ctx context.Context, conn *securityhub.Client, name awstypes.FeatureName) (*awstypes.FeatureDetail, error) {
	output, err := findAccountV2(ctx, conn)
	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResultMap(output.Features, string(name))
}

func enableFeatureV2(ctx context.Context, conn *securityhub.Client, name awstypes.FeatureName) error {
	input := securityhub.EnableSecurityHubFeatureV2Input{
		FeatureName: name,
	}
	_, err := conn.EnableSecurityHubFeatureV2(ctx, &input)

	if err != nil {
		return fmt.Errorf("enabling Security Hub V2 Feature (%s): %w", name, err)
	}

	return nil
}

func disableFeatureV2(ctx context.Context, conn *securityhub.Client, name awstypes.FeatureName) error {
	input := securityhub.DisableSecurityHubFeatureV2Input{
		FeatureName: name,
	}
	_, err := conn.DisableSecurityHubFeatureV2(ctx, &input)

	if err != nil {
		return fmt.Errorf("disabling Security Hub V2 Feature (%s): %w", name, err)
	}

	return nil
}

type featureV2ResourceModel struct {
	framework.WithRegionModel
	FeatureName   fwtypes.StringEnum[awstypes.FeatureName]   `tfsdk:"feature_name"`
	FeatureStatus fwtypes.StringEnum[awstypes.FeatureStatus] `tfsdk:"feature_status"`
}
