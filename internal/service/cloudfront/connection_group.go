// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	ret "github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudfront_connection_group", name="Connection Group")
// @Tags(identifierAttribute="arn")
func newConnectionGroupResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &connectionGroupResource{}
	return r, nil
}

const (
	ResNameConnectionGroup = "Connection Group"
)

type connectionGroupResource struct {
	framework.ResourceWithModel[connectionGroupResourceModel]
	framework.WithImportByID
}

func (r *connectionGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"anycast_ip_list_id": schema.StringAttribute{
				Optional: true,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrEnabled: schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"etag": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			"ipv6_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"is_default": schema.BoolAttribute{
				Computed: true,
			},
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"routing_endpoint": schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			"wait_for_deployment": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

type connectionGroupResourceModel struct {
	AnycastIPListID   types.String      `tfsdk:"anycast_ip_list_id"`
	ARN               types.String      `tfsdk:"arn"`
	Enabled           types.Bool        `tfsdk:"enabled"`
	ETag              types.String      `tfsdk:"etag"`
	ID                types.String      `tfsdk:"id"`
	IPv6Enabled       types.Bool        `tfsdk:"ipv6_enabled"`
	IsDefault         types.Bool        `tfsdk:"is_default"`
	LastModifiedTime  timetypes.RFC3339 `tfsdk:"last_modified_time"`
	Name              types.String      `tfsdk:"name"`
	RoutingEndpoint   types.String      `tfsdk:"routing_endpoint"`
	Status            types.String      `tfsdk:"status"`
	WaitForDeployment types.Bool        `tfsdk:"wait_for_deployment"`
	Tags              tftags.Map        `tfsdk:"tags"`
	TagsAll           tftags.Map        `tfsdk:"tags_all"`
}

func (r *connectionGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data connectionGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	input := &cloudfront.CreateConnectionGroupInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if tags := getTagsIn(ctx); len(tags) > 0 {
		input.Tags = &awstypes.Tags{
			Items: tags,
		}
	}

	output, err := conn.CreateConnectionGroup(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionCreating, ResNameConnectionGroup, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	// Use create response directly - no extra read needed
	resp.Diagnostics.Append(fwflex.Flatten(ctx, output.ConnectionGroup, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set fields that fwflex.Flatten might not handle correctly
	data.ID = fwflex.StringToFramework(ctx, output.ConnectionGroup.Id)
	data.LastModifiedTime = fwflex.TimeToFramework(ctx, output.ConnectionGroup.LastModifiedTime)
	data.ARN = fwflex.StringToFramework(ctx, output.ConnectionGroup.Arn)
	data.ETag = fwflex.StringToFramework(ctx, output.ETag)

	if data.WaitForDeployment.ValueBool() {
		if _, err := waitConnectionGroupDeployed(ctx, conn, data.ID.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudFront, create.ErrActionWaitingForCreation, ResNameConnectionGroup, data.ID.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data connectionGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	output, err := findConnectionGroupByID(ctx, conn, data.ID.ValueString())
	if ret.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionReading, ResNameConnectionGroup, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	connectionGroup := output.ConnectionGroup

	// Flatten the connection group data into the model
	resp.Diagnostics.Append(fwflex.Flatten(ctx, connectionGroup, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set computed fields that need special handling
	data.ETag = fwflex.StringToFramework(ctx, output.ETag)
	data.LastModifiedTime = fwflex.TimeToFramework(ctx, output.ConnectionGroup.LastModifiedTime)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *connectionGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new connectionGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)

	var updateOutput *cloudfront.UpdateConnectionGroupOutput

	// Remove tags from main update condition since they're handled separately
	if !new.AnycastIPListID.Equal(old.AnycastIPListID) || !new.IPv6Enabled.Equal(old.IPv6Enabled) || !new.Enabled.Equal(old.Enabled) {
		input := &cloudfront.UpdateConnectionGroupInput{}
		resp.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Handle special fields manually
		input.Id = fwflex.StringFromFramework(ctx, new.ID)
		input.IfMatch = fwflex.StringFromFramework(ctx, old.ETag)

		var err error
		updateOutput, err = conn.UpdateConnectionGroup(ctx, input)

		// Refresh our ETag if it is out of date and attempt update again.
		if errs.IsA[*awstypes.PreconditionFailed](err) {
			var etag string
			etag, err = connectionGroupETag(ctx, conn, new.ID.ValueString())

			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.CloudFront, create.ErrActionUpdating, ResNameConnectionGroup, new.ID.String(), err),
					err.Error(),
				)
				return
			}

			input.IfMatch = aws.String(etag)
			updateOutput, err = conn.UpdateConnectionGroup(ctx, input)
		}

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudFront, create.ErrActionUpdating, ResNameConnectionGroup, new.ID.String(), err),
				err.Error(),
			)
			return
		}

		if new.WaitForDeployment.ValueBool() {
			if _, err := waitConnectionGroupDeployed(ctx, conn, new.ID.ValueString()); err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.CloudFront, create.ErrActionWaitingForUpdate, ResNameConnectionGroup, new.ID.String(), err),
					err.Error(),
				)
				return
			}
		}

		new.LastModifiedTime = fwflex.TimeToFramework(ctx, updateOutput.ConnectionGroup.LastModifiedTime)
	} else {
		new.LastModifiedTime = old.LastModifiedTime
	}

	// Flatten the connection group data into the model
	resp.Diagnostics.Append(fwflex.Flatten(ctx, updateOutput.ConnectionGroup, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	new.ETag = fwflex.StringToFramework(ctx, updateOutput.ETag)

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *connectionGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data connectionGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudFrontClient(ctx)
	id := data.ID.ValueString()

	if err := disableConnectionGroup(ctx, conn, id); err != nil {
		if ret.NotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionDeleting, ResNameConnectionGroup, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	err := deleteConnectionGroup(ctx, conn, id)

	if err == nil || ret.NotFound(err) || errs.IsA[*awstypes.EntityNotFound](err) {
		return
	}

	// Disable connection group if it is not yet disabled and attempt deletion again.
	if errs.IsA[*awstypes.ResourceNotDisabled](err) {
		if err := disableConnectionGroup(ctx, conn, id); err != nil {
			if ret.NotFound(err) {
				return
			}
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudFront, create.ErrActionDeleting, ResNameConnectionGroup, data.ID.String(), err),
				err.Error(),
			)
			return
		}

		const timeout = 30 * time.Second
		_, err = tfresource.RetryWhenIsA[any, *awstypes.ResourceNotDisabled](ctx, timeout, func(ctx context.Context) (any, error) {
			return nil, deleteConnectionGroup(ctx, conn, id)
		})
	}

	if errs.IsA[*awstypes.PreconditionFailed](err) || errs.IsA[*awstypes.InvalidIfMatchVersion](err) {
		const timeout = 10 * time.Second
		_, err = tfresource.RetryWhenIsOneOf2[any, *awstypes.PreconditionFailed, *awstypes.InvalidIfMatchVersion](ctx, timeout, func(ctx context.Context) (any, error) {
			return nil, deleteConnectionGroup(ctx, conn, id)
		})
	}

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionDeleting, ResNameConnectionGroup, data.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findConnectionGroupByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetConnectionGroupOutput, error) {
	input := &cloudfront.GetConnectionGroupInput{
		Identifier: aws.String(id),
	}

	output, err := conn.GetConnectionGroup(ctx, input)

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ConnectionGroup == nil || output.ConnectionGroup.Name == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findConnectionGroupByRoutingEndpoint(ctx context.Context, conn *cloudfront.Client, endpoint string) (*cloudfront.GetConnectionGroupByRoutingEndpointOutput, error) {
	input := cloudfront.GetConnectionGroupByRoutingEndpointInput{
		RoutingEndpoint: aws.String(endpoint),
	}

	output, err := conn.GetConnectionGroupByRoutingEndpoint(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ConnectionGroup == nil || output.ConnectionGroup.Name == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func disableConnectionGroup(ctx context.Context, conn *cloudfront.Client, id string) error {
	output, err := findConnectionGroupByID(ctx, conn, id)

	if err != nil {
		return fmt.Errorf("reading CloudFront Connection Group (%s): %w", id, err)
	}

	if aws.ToString(output.ConnectionGroup.Status) == connectionGroupStatusInProgress {
		output, err = waitConnectionGroupDeployed(ctx, conn, id)

		if err != nil {
			return fmt.Errorf("waiting for CloudFront Connection Group (%s) deploy: %w", id, err)
		}
	}

	if !aws.ToBool(output.ConnectionGroup.Enabled) {
		return nil
	}

	input := cloudfront.UpdateConnectionGroupInput{
		Id:      aws.String(id),
		IfMatch: output.ETag,
	}

	input.Enabled = aws.Bool(false)

	_, err = conn.UpdateConnectionGroup(ctx, &input)

	if err != nil {
		return fmt.Errorf("updating CloudFront Connection Group (%s): %w", id, err)
	}

	if _, err := waitConnectionGroupDeployed(ctx, conn, id); err != nil {
		return fmt.Errorf("waiting for CloudFront Connection Group (%s) deploy: %w", id, err)
	}

	return nil
}

func deleteConnectionGroup(ctx context.Context, conn *cloudfront.Client, id string) error {
	etag, err := connectionGroupETag(ctx, conn, id)

	if err != nil {
		return err
	}

	input := cloudfront.DeleteConnectionGroupInput{
		Id:      aws.String(id),
		IfMatch: aws.String(etag),
	}

	_, err = conn.DeleteConnectionGroup(ctx, &input)

	if err != nil {
		return fmt.Errorf("deleting CloudFront Connection Group (%s): %w", id, err)
	}

	if _, err := waitConnectionGroupDeleted(ctx, conn, id); err != nil {
		return fmt.Errorf("waiting for CloudFront Connection Group (%s) delete: %w", id, err)
	}

	return nil
}

func waitConnectionGroupDeployed(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetConnectionGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{connectionGroupStatusInProgress},
		Target:     []string{connectionGroupStatusDeployed},
		Refresh:    statusConnectionGroup(ctx, conn, id),
		Timeout:    30 * time.Minute,
		MinTimeout: 5 * time.Second,
		Delay:      10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudfront.GetConnectionGroupOutput); ok {
		return output, err
	}

	return nil, err
}

func waitConnectionGroupDeleted(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetConnectionGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{connectionGroupStatusInProgress, connectionGroupStatusDeployed},
		Target:     []string{},
		Refresh:    statusConnectionGroup(ctx, conn, id),
		Timeout:    30 * time.Minute,
		MinTimeout: 5 * time.Second,
		Delay:      10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudfront.GetConnectionGroupOutput); ok {
		return output, err
	}

	return nil, err
}

func statusConnectionGroup(ctx context.Context, conn *cloudfront.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findConnectionGroupByID(ctx, conn, id)

		if ret.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, aws.ToString(output.ConnectionGroup.Status), nil
	}
}

func connectionGroupETag(ctx context.Context, conn *cloudfront.Client, id string) (string, error) {
	output, err := findConnectionGroupByID(ctx, conn, id)

	if err != nil {
		return "", fmt.Errorf("reading CloudFront Connection Group (%s): %w", id, err)
	}

	return aws.ToString(output.ETag), nil
}
