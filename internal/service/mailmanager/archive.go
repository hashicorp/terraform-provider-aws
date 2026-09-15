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
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
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
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tflistplanmodifier "github.com/hashicorp/terraform-provider-aws/internal/framework/planmodifiers/listplanmodifier"
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
// @Testing(importIgnore="retention", plannableImportAction="NoOp")
// @Testing(preCheck="testAccArchivePreCheck")
// @Testing(serialize=true)
// @Testing(skipEmptyTags=true, skipNullTags=true)
func newArchiveResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &archiveResource{}, nil
}

const (
	archiveCreateTimeout = 2 * time.Minute
	archiveDeleteTimeout = 2 * time.Minute
)

type archiveResource struct {
	framework.ResourceWithModel[archiveResourceModel]
	framework.WithImportByIdentity
}

func (r *archiveResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"archive_state": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ArchiveState](),
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"created_timestamp": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrKMSKeyARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
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
			"retention_actual": framework.ResourceComputedListOfObjectsAttribute[archiveRetentionModel](ctx, tflistplanmodifier.UnknownWhenOtherValueChanges(path.Root("retention"))),
			names.AttrTags:     tftags.TagsAttribute(),
			names.AttrTagsAll:  tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"retention": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[archiveRetentionModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrRetentionPeriod: schema.StringAttribute{
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
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	out, err := tfresource.RetryWhenIsA[*mailmanager.CreateArchiveOutput, *awstypes.ConflictException](
		ctx, archiveCreateTimeout, func(ctx context.Context) (*mailmanager.CreateArchiveOutput, error) {
			return conn.CreateArchive(ctx, &input)
		},
	)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ArchiveName.String())
		return
	}

	archiveID := aws.ToString(out.ArchiveId)
	archiveOut, err := findArchiveByID(ctx, conn, archiveID)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, archiveID)
		return
	}

	retentionIsConfigured := plan.Retention.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0
	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, archiveOut, &plan, retentionIsConfigured))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *archiveResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().MailManagerClient(ctx)

	var data archiveResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	// When importing, all attributes other than `id` will be null.
	// During a read of an existing resource, `name` will be set as it is a required attribute.
	retentionIsConfigured := data.Retention.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0
	isImport := data.ArchiveName.IsNull()

	archiveID := data.ArchiveId.ValueString()
	out, err := findArchiveByID(ctx, conn, archiveID)

	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ArchiveId.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &data, retentionIsConfigured || isImport))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
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

	archiveID := state.ArchiveId.ValueString()

	if diff.HasChanges() {
		var input mailmanager.UpdateArchiveInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		input.ArchiveId = state.ArchiveId.ValueStringPointer()

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

		retentionIsConfigured := plan.Retention.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0
		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, archiveOut, &plan, retentionIsConfigured))
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

	archiveID := state.ArchiveId.ValueString()
	input := mailmanager.DeleteArchiveInput{
		ArchiveId: aws.String(archiveID),
	}

	_, err := conn.DeleteArchive(ctx, &input)
	// ConflictException means the archive is already in PENDING_DELETION (deleted but waiting 30 days).
	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.ConflictException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, archiveID)
		return
	}

	// Wait until the archive goes to PENDING_DELETION (or is gone)
	_, err = (&retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ArchiveStateActive),
		Target:  []string{},
		Refresh: func(ctx context.Context) (any, string, error) {
			out, err := conn.GetArchive(ctx, &mailmanager.GetArchiveInput{
				ArchiveId: aws.String(archiveID),
			})
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil, "", nil
			}
			if err != nil {
				return nil, "", err
			}
			if out.ArchiveState == awstypes.ArchiveStatePendingDeletion {
				return nil, "", nil
			}
			return out, string(out.ArchiveState), nil
		},
		Timeout:    archiveDeleteTimeout,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}).WaitForStateContext(ctx)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, archiveID)
	}
}

func (r *archiveResource) flatten(ctx context.Context, apiObject *mailmanager.GetArchiveOutput, data *archiveResourceModel, populateRetention bool) diag.Diagnostics {
	diags := flex.Flatten(ctx, apiObject, data)
	if diags.HasError() {
		return diags
	}
	r.flattenRetention(ctx, apiObject.Retention, data, populateRetention)
	return diags
}

func (r *archiveResource) flattenRetention(ctx context.Context, apiRetention awstypes.ArchiveRetention, data *archiveResourceModel, populateRetention bool) {
	// Always populate retention_actual from the current API state.
	if apiRetention != nil {
		var m archiveRetentionModel
		var mDiags diag.Diagnostics
		mDiags.Append(m.Flatten(ctx, apiRetention)...)
		if !mDiags.HasError() {
			data.RetentionActual = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &m)
		}
	} else {
		data.RetentionActual = fwtypes.NewListNestedObjectValueOfNull[archiveRetentionModel](ctx)
	}

	if !populateRetention {
		data.Retention = fwtypes.NewListNestedObjectValueOfNull[archiveRetentionModel](ctx)
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
	if out.ArchiveState == awstypes.ArchiveStatePendingDeletion {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: fmt.Errorf("archive is in PENDING_DELETION state"),
		})
	}

	return out, nil
}

type archiveResourceModel struct {
	framework.WithRegionModel
	ArchiveArn           types.String                                           `tfsdk:"arn"`
	Retention            fwtypes.ListNestedObjectValueOf[archiveRetentionModel] `tfsdk:"retention" autoflex:",noflatten"`
	RetentionActual      fwtypes.ListNestedObjectValueOf[archiveRetentionModel] `tfsdk:"retention_actual" autoflex:"-"`
	KmsKeyArn            fwtypes.ARN                                            `tfsdk:"kms_key_arn"`
	ArchiveId            types.String                                           `tfsdk:"id"`
	ArchiveName          types.String                                           `tfsdk:"name"`
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
