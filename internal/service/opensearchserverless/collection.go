// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Collection")
// @Tags(identifierAttribute="arn")
func newResourceCollection(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := resourceCollection{}
	r.SetDefaultCreateTimeout(20 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return &r, nil
}

type resourceCollectionData struct {
	ARN                types.String   `tfsdk:"arn"`
	CollectionEndpoint types.String   `tfsdk:"collection_endpoint"`
	DashboardEndpoint  types.String   `tfsdk:"dashboard_endpoint"`
	Description        types.String   `tfsdk:"description"`
	ID                 types.String   `tfsdk:"id"`
	KmsKeyARN          types.String   `tfsdk:"kms_key_arn"`
	Name               types.String   `tfsdk:"name"`
	Tags               types.Map      `tfsdk:"tags"`
	TagsAll            types.Map      `tfsdk:"tags_all"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
	Type               types.String   `tfsdk:"type"`
}

const (
	ResNameCollection = "Collection"
)

type resourceCollection struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceCollection) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_opensearchserverless_collection"
}

func (r *resourceCollection) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"collection_endpoint": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dashboard_endpoint": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1000),
				},
			},
			"id": framework.IDAttribute(),
			"kms_key_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 32),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-z][a-z0-9-]+$`),
						`must start with any lower case letter and can can include any lower case letter, number, or "-"`),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"type": schema.StringAttribute{
				Optional: true,
				Computed: true,
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
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceCollection) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceCollectionData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchServerlessClient(ctx)

	in := &opensearchserverless.CreateCollectionInput{
		ClientToken: aws.String(id.UniqueId()),
		Name:        aws.String(plan.Name.ValueString()),
		Tags:        getTagsIn(ctx),
	}

	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}

	if !plan.Type.IsNull() {
		in.Type = awstypes.CollectionType(plan.Type.ValueString())
	}

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
	waitOut, err := waitCollectionCreated(ctx, conn, aws.ToString(out.CreateCollectionDetail.Id), createTimeout)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionWaitingForCreation, ResNameCollection, plan.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, waitOut.Arn)
	state.CollectionEndpoint = flex.StringToFramework(ctx, waitOut.CollectionEndpoint)
	state.DashboardEndpoint = flex.StringToFramework(ctx, waitOut.DashboardEndpoint)
	state.Description = flex.StringToFramework(ctx, waitOut.Description)
	state.KmsKeyARN = flex.StringToFramework(ctx, waitOut.KmsKeyArn)
	state.Name = flex.StringToFramework(ctx, waitOut.Name)
	state.Type = flex.StringValueToFramework(ctx, waitOut.Type)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCollection) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var state resourceCollectionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findCollectionByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.CollectionEndpoint = flex.StringToFramework(ctx, out.CollectionEndpoint)
	state.DashboardEndpoint = flex.StringToFramework(ctx, out.DashboardEndpoint)
	state.Description = flex.StringToFramework(ctx, out.Description)
	state.ID = flex.StringToFramework(ctx, out.Id)
	state.KmsKeyARN = flex.StringToFramework(ctx, out.KmsKeyArn)
	state.Name = flex.StringToFramework(ctx, out.Name)
	state.Type = flex.StringValueToFramework(ctx, out.Type)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCollection) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var plan, state resourceCollectionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) {
		input := &opensearchserverless.UpdateCollectionInput{
			ClientToken: aws.String(id.UniqueId()),
			Id:          flex.StringFromFramework(ctx, plan.ID),
			Description: flex.StringFromFramework(ctx, plan.Description),
		}

		out, err := conn.UpdateCollection(ctx, input)

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionUpdating, ResNameCollection, state.ID.ValueString(), err),
				err.Error(),
			)
			return
		}

		plan.ARN = flex.StringToFramework(ctx, out.UpdateCollectionDetail.Arn)
		plan.Description = flex.StringToFramework(ctx, out.UpdateCollectionDetail.Description)
		plan.ID = flex.StringToFramework(ctx, out.UpdateCollectionDetail.Id)
		plan.Name = flex.StringToFramework(ctx, out.UpdateCollectionDetail.Name)
		plan.Type = flex.StringValueToFramework(ctx, out.UpdateCollectionDetail.Type)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceCollection) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var state resourceCollectionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteCollection(ctx, &opensearchserverless.DeleteCollectionInput{
		ClientToken: aws.String(id.UniqueId()),
		Id:          aws.String(state.ID.ValueString()),
	})

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
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

func (r *resourceCollection) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func (r *resourceCollection) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitCollectionCreated(ctx context.Context, conn *opensearchserverless.Client, id string, timeout time.Duration) (*awstypes.CollectionDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.CollectionStatusCreating),
		Target:     enum.Slice(awstypes.CollectionStatusActive),
		Refresh:    statusCollection(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CollectionDetail); ok {
		return output, err
	}

	return nil, err
}

func waitCollectionDeleted(ctx context.Context, conn *opensearchserverless.Client, id string, timeout time.Duration) (*awstypes.CollectionDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.CollectionStatusDeleting),
		Target:     []string{},
		Refresh:    statusCollection(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CollectionDetail); ok {
		return output, err
	}

	return nil, err
}

func statusCollection(ctx context.Context, conn *opensearchserverless.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findCollectionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}
