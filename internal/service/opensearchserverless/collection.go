// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package opensearchserverless

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_opensearchserverless_collection", name="Collection")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types;types.CollectionDetail")
// @Testing(preIdentityVersion="v6.28.0")
func newCollectionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := collectionResource{}

	r.SetDefaultCreateTimeout(20 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return &r, nil
}

type collectionResourceModel struct {
	framework.WithRegionModel
	ARN                types.String   `tfsdk:"arn"`
	CollectionEndpoint types.String   `tfsdk:"collection_endpoint"`
	DashboardEndpoint  types.String   `tfsdk:"dashboard_endpoint"`
	Description        types.String   `tfsdk:"description"`
	ID                 types.String   `tfsdk:"id"`
	KmsKeyARN          types.String   `tfsdk:"kms_key_arn"`
	Name               types.String   `tfsdk:"name"`
	StandbyReplicas    types.String   `tfsdk:"standby_replicas"`
	Tags               tftags.Map     `tfsdk:"tags"`
	TagsAll            tftags.Map     `tfsdk:"tags_all"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
	Type               types.String   `tfsdk:"type"`
}

const (
	ResNameCollection = "Collection"
)

type collectionResource struct {
	framework.ResourceWithModel[collectionResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *collectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"collection_endpoint": schema.StringAttribute{
				Description: "Collection-specific endpoint used to submit index, search, and data upload requests to an OpenSearch Serverless collection.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dashboard_endpoint": schema.StringAttribute{
				Description: "Collection-specific endpoint used to access OpenSearch Dashboards.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Description: "Description of the collection.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1000),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrKMSKeyARN: schema.StringAttribute{
				Description: "The ARN of the Amazon Web Services KMS key used to encrypt the collection.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Description: "Name of the collection.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 32),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z][0-9a-z-]+$`),
						`must start with any lower case letter and can can include any lower case letter, number, or "-"`),
				},
			},
			"standby_replicas": schema.StringAttribute{
				Description: "Indicates whether standby replicas should be used for a collection. One of `ENABLED` or `DISABLED`. Defaults to `ENABLED`.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.StandbyReplicas](),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrType: schema.StringAttribute{
				Description: "Type of collection. One of `SEARCH`, `TIMESERIES`, or `VECTORSEARCH`. Defaults to `TIMESERIES`.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.CollectionType](),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *collectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan collectionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchServerlessClient(ctx)

	in := &opensearchserverless.CreateCollectionInput{}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)

	if resp.Diagnostics.HasError() {
		return
	}

	in.ClientToken = aws.String(sdkid.UniqueId())
	in.Tags = getTagsIn(ctx)

	out, err := conn.CreateCollection(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionCreating, ResNameCollection, plan.Name.ValueString(), nil),
			err.Error(),
		)
		return
	}

	state := plan

	state.ID = flex.StringToFramework(ctx, out.CreateCollectionDetail.Id)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	collection, err := waitCollectionCreated(ctx, conn, aws.ToString(out.CreateCollectionDetail.Id), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionWaitingForCreation, ResNameCollection, plan.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, collection, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *collectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var state collectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findCollectionByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionReading, ResNameCollection, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *collectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var plan, state collectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) {
		input := &opensearchserverless.UpdateCollectionInput{}

		resp.Diagnostics.Append(flex.Expand(ctx, plan, input)...)

		if resp.Diagnostics.HasError() {
			return
		}

		input.ClientToken = aws.String(sdkid.UniqueId())

		out, err := conn.UpdateCollection(ctx, input)

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionUpdating, ResNameCollection, state.ID.ValueString(), err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)

		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *collectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var state collectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteCollection(ctx, &opensearchserverless.DeleteCollectionInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		Id:          state.ID.ValueStringPointer(),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionDeleting, ResNameCollection, state.Name.ValueString(), nil),
			err.Error(),
		)
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitCollectionDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionWaitingForCreation, ResNameCollection, state.Name.ValueString(), err),
			err.Error(),
		)
		return
	}
}

func waitCollectionCreated(ctx context.Context, conn *opensearchserverless.Client, id string, timeout time.Duration) (*awstypes.CollectionDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.CollectionStatusCreating),
		Target:     enum.Slice(awstypes.CollectionStatusActive),
		Refresh:    statusCollection(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CollectionDetail); ok {
		if output.Status == awstypes.CollectionStatusFailed {
			retry.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(output.FailureCode), aws.ToString(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitCollectionDeleted(ctx context.Context, conn *opensearchserverless.Client, id string, timeout time.Duration) (*awstypes.CollectionDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.CollectionStatusDeleting),
		Target:     []string{},
		Refresh:    statusCollection(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CollectionDetail); ok {
		if output.Status == awstypes.CollectionStatusFailed {
			retry.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(output.FailureCode), aws.ToString(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func statusCollection(conn *opensearchserverless.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findCollectionByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}
