// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrock_account_data_retention", name="Account Data Retention")
// @SingletonIdentity
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(checkDestroyNoop=true)
// @Testing(serialize=true)
func newAccountDataRetentionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &accountDataRetentionResource{}, nil
}

const (
	ResNameAccountDataRetention = "Account Data Retention"
)

type accountDataRetentionResource struct {
	framework.ResourceWithModel[accountDataRetentionResourceModel]
	framework.WithNoOpDelete
	framework.WithImportByIdentity
}

func (r *accountDataRetentionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrMode: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DataRetentionMode](),
				Required:   true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
	}
}

func (r *accountDataRetentionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan accountDataRetentionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.put(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *accountDataRetentionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var state accountDataRetentionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrock.GetAccountDataRetentionInput{}
	out, err := conn.GetAccountDataRetention(ctx, &input)
	if retry.NotFound(err) || out == nil {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	state.Mode = fwtypes.StringEnumValue(out.Mode)
	state.UpdatedAt = timetypes.NewRFC3339TimePointerValue(out.UpdatedAt)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *accountDataRetentionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan accountDataRetentionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.put(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

// put applies the planned mode and reads back the server-set updated_at, so the
// attribute is never left unknown after apply.
func (r *accountDataRetentionResource) put(ctx context.Context, data *accountDataRetentionResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := r.Meta().BedrockClient(ctx)

	input := bedrock.PutAccountDataRetentionInput{
		Mode: data.Mode.ValueEnum(),
	}
	if _, err := conn.PutAccountDataRetention(ctx, &input); err != nil {
		smerr.AddError(ctx, &diags, err)
		return diags
	}

	out, err := conn.GetAccountDataRetention(ctx, &bedrock.GetAccountDataRetentionInput{})
	if err != nil {
		smerr.AddError(ctx, &diags, err)
		return diags
	}

	data.UpdatedAt = timetypes.NewRFC3339TimePointerValue(out.UpdatedAt)

	return diags
}

type accountDataRetentionResourceModel struct {
	framework.WithRegionModel

	Mode      fwtypes.StringEnum[awstypes.DataRetentionMode] `tfsdk:"mode"`
	UpdatedAt timetypes.RFC3339                              `tfsdk:"updated_at"`
}
