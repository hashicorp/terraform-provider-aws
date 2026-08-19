// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mailmanager

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mailmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mailmanager/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

// @FrameworkResource("aws_mailmanager_archive", name="Archive")
// @IdentityAttribute("id")
// @Tags(identifierAttribute="arn")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccArchivePreCheck")
// @Testing(skipEmptyTags=true, skipNullTags=true)
func newArchiveResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &archiveResource{}, nil
}

const (
	ResNameArchive = "Archive"
)

type archiveResource struct {
	framework.ResourceWithModel[archiveResourceModel]
	framework.WithImportByIdentity
}

func (r *archiveResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"last_updated_timestamp": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[a-zA-Z0-9-_]+$`),
						"must contain only alphanumeric characters, hyphens, and underscores",
					),
				},
			},
			names.AttrKMSKeyARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"archive_state": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ArchiveState](),
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"retention": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[archiveRetentionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"retention_period": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.RetentionPeriod](),
							Required:   true,
						},
					},
				},
			},
		},
	}
}

func (r *archiveResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().MailManagerClient(ctx)

	var plan archiveResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input mailmanager.CreateArchiveInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Archive")))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateArchive(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	archiveID := aws.ToString(out.ArchiveId)
	archiveOut, err := findArchiveByID(ctx, conn, archiveID)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, archiveID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, archiveOut, &plan, flex.WithFieldNamePrefix("Archive")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *archiveResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().MailManagerClient(ctx)

	var state archiveResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	archiveID := state.ID.ValueString()
	out, err := findArchiveByID(ctx, conn, archiveID)

	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("Archive")))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *archiveResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().MailManagerClient(ctx)

	var plan, state archiveResourceModel
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

	archiveID := state.ID.ValueString()

	if diff.HasChanges() {
		var input mailmanager.UpdateArchiveInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Archive")))
		if resp.Diagnostics.HasError() {
			return
		}

		input.ArchiveId = state.ID.ValueStringPointer()

		_, err := conn.UpdateArchive(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, archiveID)
			return
		}

		archiveOut, err := findArchiveByID(ctx, conn, archiveID)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, archiveID)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, archiveOut, &plan, flex.WithFieldNamePrefix("Archive")))
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		plan.LastUpdatedTimestamp = state.LastUpdatedTimestamp
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *archiveResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().MailManagerClient(ctx)

	var state archiveResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	archiveID := state.ID.ValueString()
	input := mailmanager.DeleteArchiveInput{
		ArchiveId: aws.String(archiveID),
	}

	_, err := conn.DeleteArchive(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, archiveID)
	}
}

func findArchiveByID(ctx context.Context, conn *mailmanager.Client, id string) (*mailmanager.GetArchiveOutput, error) {
	input := mailmanager.GetArchiveInput{
		ArchiveId: aws.String(id),
	}

	out, err := conn.GetArchive(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type archiveResourceModel struct {
	framework.WithRegionModel
	ARN                  types.String                                           `tfsdk:"arn"`
	Retention            fwtypes.ListNestedObjectValueOf[archiveRetentionModel] `tfsdk:"retention"`
	KmsKeyArn            fwtypes.ARN                                            `tfsdk:"kms_key_arn"`
	ID                   types.String                                           `tfsdk:"id"`
	Name                 types.String                                           `tfsdk:"name"`
	ArchiveState         fwtypes.StringEnum[awstypes.ArchiveState]              `tfsdk:"archive_state"`
	CreatedTimestamp     timetypes.RFC3339                                      `tfsdk:"created_timestamp"`
	LastUpdatedTimestamp timetypes.RFC3339                                      `tfsdk:"last_updated_timestamp"`
	Tags                 tftags.Map                                             `tfsdk:"tags"`
	TagsAll              tftags.Map                                             `tfsdk:"tags_all"`
}

type archiveRetentionModel struct {
	RetentionPeriod fwtypes.StringEnum[awstypes.RetentionPeriod] `tfsdk:"retention_period"`
}

var (
	_ flex.Expander  = archiveRetentionModel{}
	_ flex.Flattener = &archiveRetentionModel{}
)

func (m archiveRetentionModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	return &awstypes.ArchiveRetentionMemberRetentionPeriod{
		Value: awstypes.RetentionPeriod(m.RetentionPeriod.ValueString()),
	}, nil
}

func (m *archiveRetentionModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch v := v.(type) {
	case *awstypes.ArchiveRetentionMemberRetentionPeriod:
		m.RetentionPeriod = fwtypes.StringEnumValue(v.Value)
	default:
		diags.AddError("Unexpected Retention Type",
			fmt.Sprintf("archive retention flatten: unexpected type %T", v))
	}
	return diags
}
