// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package neptunegraph

import (
	"context"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptunegraph"
	awstypes "github.com/aws/aws-sdk-go-v2/service/neptunegraph/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_neptunegraph_graph_snapshot", name="Graph Snapshot")
// @Tags(identifierAttribute="arn")
func newGraphSnapshotResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &graphSnapshotResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type graphSnapshotResource struct {
	framework.ResourceWithModel[graphSnapshotResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *graphSnapshotResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"graph_identifier": schema.StringAttribute{
				Description: "The unique identifier of the Neptune Analytics graph to snapshot.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^g-[a-z0-9]{10}$`), ""),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"kms_key_identifier": schema.StringAttribute{
				Description: "The ID of the KMS key used to encrypt and decrypt the snapshot.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"snapshot_create_time": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Description: "The time when the snapshot was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"snapshot_name": schema.StringAttribute{
				Description: "The snapshot name. Must contain from 1 to 63 lowercase letters, digits, or hyphens, begin with a letter, not end with a hyphen, not contain two consecutive hyphens, and not begin with the reserved prefix \"gs-\".",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`), ""),
					snapshotNameValidator{},
				},
			},
			"source_graph_id": schema.StringAttribute{
				Description: "The ID of the Neptune Analytics graph from which the snapshot was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Description: "The current status of the snapshot.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *graphSnapshotResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data graphSnapshotResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NeptuneGraphClient(ctx)

	var input neptunegraph.CreateGraphSnapshotInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	// NeptuneGraph SDK GetGraphSnapshotOutput param for Name differs from CreateGraphSnapshotInput param SnapshotName.
	input.SnapshotName = fwflex.StringFromFramework(ctx, data.Name)
	input.Tags = getTagsIn(ctx)

	// A ConflictException is expected while the source graph is mid-operation (e.g. SNAPSHOTTING).
	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	output, err := tfresource.RetryWhenIsA[*neptunegraph.CreateGraphSnapshotOutput, *awstypes.ConflictException](ctx, createTimeout, func(ctx context.Context) (*neptunegraph.CreateGraphSnapshotOutput, error) {
		return conn.CreateGraphSnapshot(ctx, &input)
	})
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.Name.ValueString())
		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringToFramework(ctx, output.Id)

	snapshot, err := waitGraphSnapshotCreated(ctx, conn, data.ID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, snapshot, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *graphSnapshotResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data graphSnapshotResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NeptuneGraphClient(ctx)

	output, err := findGraphSnapshotByID(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	// GetGraphSnapshot does not return a graph_identifier; the source graph ID is the
	// graph the snapshot was taken from, so populate graph_identifier from it. This also
	// makes import (which starts from only the snapshot ID) produce a complete state.
	data.GraphIdentifier = fwflex.StringToFramework(ctx, output.SourceGraphId)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *graphSnapshotResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data graphSnapshotResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NeptuneGraphClient(ctx)

	input := neptunegraph.DeleteGraphSnapshotInput{
		SnapshotIdentifier: data.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteGraphSnapshot(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	if _, err := waitGraphSnapshotDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}
}

func findGraphSnapshotByID(ctx context.Context, conn *neptunegraph.Client, id string) (*neptunegraph.GetGraphSnapshotOutput, error) {
	input := neptunegraph.GetGraphSnapshotInput{
		SnapshotIdentifier: aws.String(id),
	}

	return findGraphSnapshot(ctx, conn, &input)
}

func findGraphSnapshot(ctx context.Context, conn *neptunegraph.Client, input *neptunegraph.GetGraphSnapshotInput) (*neptunegraph.GetGraphSnapshotOutput, error) {
	output, err := conn.GetGraphSnapshot(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, nil
}

func statusGraphSnapshot(conn *neptunegraph.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findGraphSnapshotByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitGraphSnapshotCreated(ctx context.Context, conn *neptunegraph.Client, id string, timeout time.Duration) (*neptunegraph.GetGraphSnapshotOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.SnapshotStatusCreating),
		Target:                    enum.Slice(awstypes.SnapshotStatusAvailable),
		Refresh:                   statusGraphSnapshot(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*neptunegraph.GetGraphSnapshotOutput); ok {
		return output, err
	}

	return nil, err
}

func waitGraphSnapshotDeleted(ctx context.Context, conn *neptunegraph.Client, id string, timeout time.Duration) (*neptunegraph.GetGraphSnapshotOutput, error) {
	const (
		delay = 10 * time.Second
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SnapshotStatusDeleting),
		Target:  []string{},
		Refresh: statusGraphSnapshot(conn, id),
		Delay:   delay,
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*neptunegraph.GetGraphSnapshotOutput); ok {
		return output, err
	}

	return nil, err
}

// snapshotNameValidator rejects a snapshot name beginning with the reserved "gs-"
// prefix. The service pattern expresses this with a negative lookahead, which RE2
// (used by regexache) does not support, so it is checked separately.
type snapshotNameValidator struct{}

func (snapshotNameValidator) Description(_ context.Context) string {
	return `must not begin with the reserved prefix "gs-"`
}

func (v snapshotNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v snapshotNameValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	if strings.HasPrefix(request.ConfigValue.ValueString(), "gs-") {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid Snapshot Name",
			v.Description(ctx),
		)
	}
}

type graphSnapshotResourceModel struct {
	framework.WithRegionModel
	ARN                types.String      `tfsdk:"arn"`
	GraphIdentifier    types.String      `tfsdk:"graph_identifier"`
	ID                 types.String      `tfsdk:"id"`
	KMSKeyIdentifier   types.String      `tfsdk:"kms_key_identifier"`
	Name               types.String      `tfsdk:"snapshot_name"`
	SnapshotCreateTime timetypes.RFC3339 `tfsdk:"snapshot_create_time"`
	SourceGraphID      types.String      `tfsdk:"source_graph_id"`
	Status             types.String      `tfsdk:"status"`
	Tags               tftags.Map        `tfsdk:"tags"`
	TagsAll            tftags.Map        `tfsdk:"tags_all"`
	Timeouts           timeouts.Value    `tfsdk:"timeouts"`
}
