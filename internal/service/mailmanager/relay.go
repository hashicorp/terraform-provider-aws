// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mailmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mailmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mailmanager/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_mailmanager_relay", name="Relay")
// @IdentityAttribute("id")
// @Tags(identifierAttribute="arn")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccRelayPreCheck")
// @Testing(skipEmptyTags=true, skipNullTags=true)
func newRelayResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &relayResource{}
	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

const (
	ResNameRelay = "Relay"
)

type relayResource struct {
	framework.ResourceWithModel[relayResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *relayResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"created_timestamp": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"last_modified_timestamp": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[a-zA-Z0-9-_]+$`),
						"must contain only alphanumeric characters, hyphens, and underscores",
					),
				},
			},
			"server_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[a-zA-Z0-9-.]+$`),
						"must contain only alphanumeric characters, hyphens, and dots",
					),
				},
			},
			"server_port": schema.Int32Attribute{
				Required: true,
				Validators: []validator.Int32{
					int32validator.Between(1, 65535),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"authentication": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[relayAuthenticationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"no_authentication": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[relayNoAuthenticationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("secret_arn"),
								),
							},
							NestedObject: schema.NestedBlockObject{},
						},
					},
					Attributes: map[string]schema.Attribute{
						"secret_arn": schema.StringAttribute{
							Optional:   true,
							CustomType: fwtypes.ARNType,
							Validators: []validator.String{
								stringvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("no_authentication"),
								),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *relayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	conn := r.Meta().MailManagerClient(ctx)
	var plan relayResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input mailmanager.CreateRelayInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Relay")))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateRelay(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	relayID := aws.ToString(out.RelayId)
	relayOut, err := findRelayByID(ctx, conn, relayID)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, relayID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, relayOut, &plan, flex.WithFieldNamePrefix("Relay")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *relayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	conn := r.Meta().MailManagerClient(ctx)

	var state relayResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	relayID := state.ID.ValueString()
	out, err := findRelayByID(ctx, conn, relayID)

	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, relayID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("Relay")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *relayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	conn := r.Meta().MailManagerClient(ctx)

	var plan, state relayResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	relayID := state.ID.ValueString()

	if diff.HasChanges() {
		var input mailmanager.UpdateRelayInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Relay")))
		if resp.Diagnostics.HasError() {
			return
		}
		input.RelayId = state.ID.ValueStringPointer()

		_, err := conn.UpdateRelay(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, relayID)
			return
		}

		relayOut, err := findRelayByID(ctx, conn, relayID)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, relayID)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, relayOut, &plan, flex.WithFieldNamePrefix("Relay")))
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		plan.ARN = state.ARN
		plan.CreatedTimestamp = state.CreatedTimestamp
		plan.LastModifiedTimestamp = state.LastModifiedTimestamp
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *relayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	conn := r.Meta().MailManagerClient(ctx)

	var state relayResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	relayID := state.ID.ValueString()
	input := mailmanager.DeleteRelayInput{
		RelayId: aws.String(relayID),
	}

	_, err := conn.DeleteRelay(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, relayID)
	}
}

func findRelayByID(ctx context.Context, conn *mailmanager.Client, id string) (*mailmanager.GetRelayOutput, error) {
	input := mailmanager.GetRelayInput{
		RelayId: aws.String(id),
	}

	out, err := conn.GetRelay(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type relayResourceModel struct {
	framework.WithRegionModel
	ARN                   types.String                                              `tfsdk:"arn"`
	Authentication        fwtypes.ListNestedObjectValueOf[relayAuthenticationModel] `tfsdk:"authentication"`
	CreatedTimestamp      timetypes.RFC3339                                         `tfsdk:"created_timestamp"`
	ID                    types.String                                              `tfsdk:"id"`
	LastModifiedTimestamp timetypes.RFC3339                                         `tfsdk:"last_modified_timestamp"`
	Name                  types.String                                              `tfsdk:"name"`
	ServerName            types.String                                              `tfsdk:"server_name"`
	ServerPort            types.Int32                                               `tfsdk:"server_port"`
	Tags                  tftags.Map                                                `tfsdk:"tags"`
	TagsAll               tftags.Map                                                `tfsdk:"tags_all"`
	Timeouts              timeouts.Value                                            `tfsdk:"timeouts"`
}

type relayAuthenticationModel struct {
	NoAuthentication fwtypes.ListNestedObjectValueOf[relayNoAuthenticationModel] `tfsdk:"no_authentication"`
	SecretARN        fwtypes.ARN                                                 `tfsdk:"secret_arn"`
}

type relayNoAuthenticationModel struct{}

var (
	_ flex.Expander  = relayAuthenticationModel{}
	_ flex.Flattener = &relayAuthenticationModel{}
)

func (m relayAuthenticationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.NoAuthentication.IsNull():
		return &awstypes.RelayAuthenticationMemberNoAuthentication{
			Value: awstypes.NoAuthentication{},
		}, diags
	case !m.SecretARN.IsNull():
		return &awstypes.RelayAuthenticationMemberSecretArn{
			Value: m.SecretARN.ValueString(),
		}, diags
	}
	return nil, diags
}

func (m *relayAuthenticationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch v.(type) {
	case awstypes.RelayAuthenticationMemberNoAuthentication:
		m.NoAuthentication = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &relayNoAuthenticationModel{})
	case awstypes.RelayAuthenticationMemberSecretArn:
		m.SecretARN = fwtypes.ARNValue(v.(awstypes.RelayAuthenticationMemberSecretArn).Value)
	default:
		diags.AddError("Unexpected Authentication Type",
			fmt.Sprintf("relay authentication flatten: unexpected type %T", v))
	}
	return diags
}
