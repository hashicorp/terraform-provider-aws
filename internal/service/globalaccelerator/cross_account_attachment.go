// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Cross Account Attachment")
func newResourceCrossAccountAttachment(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCrossAccountAttachment{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameCrossAccountAttachment = "Cross Account Attachment"
)

type resourceCrossAccountAttachment struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceCrossAccountAttachment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_globalaccelerator_cross_account_attachment"
}

func (r *resourceCrossAccountAttachment) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"attachment_arn": framework.ARNAttributeComputedOnly(),
			"id":             framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"principals": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"resources": schema.ListAttribute{
				Optional:    true,
				ElementType: ResourceDataElementType,
			},
			"created_time": schema.StringAttribute{
				Computed: true,
			},
			"last_modified_time": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *resourceCrossAccountAttachment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().GlobalAcceleratorConn(ctx)

	var plan resourceCrossAccountAttachmentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &globalaccelerator.CreateCrossAccountAttachmentInput{
		IdempotencyToken: aws.String(id.UniqueId()),
		Name:             aws.String(plan.Name.ValueString()),
		Tags:             getTagsIn(ctx),
	}

	if !plan.Principals.IsNull() {
		input.Principals = flex.ExpandFrameworkStringList(ctx, plan.Principals)
	}

	if !plan.Resources.IsNull() {
		var tfResources []ResourceData
		diags := plan.Resources.ElementsAs(ctx, &tfResources, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		input.Resources = expandResources(tfResources)
	}

	out, err := conn.CreateCrossAccountAttachmentWithContext(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.GlobalAccelerator, create.ErrActionCreating, ResNameCrossAccountAttachment, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.CrossAccountAttachment == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.GlobalAccelerator, create.ErrActionCreating, ResNameCrossAccountAttachment, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	state := plan
	state.ID = flex.StringToFramework(ctx, out.CrossAccountAttachment.AttachmentArn)
	state.ARN = flex.StringToFramework(ctx, out.CrossAccountAttachment.AttachmentArn)

	state.CreatedTime = types.StringValue(out.CrossAccountAttachment.CreatedTime.Format(time.RFC3339))
	if out.CrossAccountAttachment.LastModifiedTime != nil {
		state.LastModifiedTime = types.StringValue(out.CrossAccountAttachment.LastModifiedTime.Format(time.RFC3339))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceCrossAccountAttachment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().GlobalAcceleratorConn(ctx)

	var state resourceCrossAccountAttachmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &globalaccelerator.DescribeCrossAccountAttachmentInput{
		AttachmentArn: aws.String(state.ARN.ValueString()),
	}
	out, err := conn.DescribeCrossAccountAttachmentWithContext(ctx, input)

	var nfe *globalaccelerator.AttachmentNotFoundException
	if errors.As(err, &nfe) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.GlobalAccelerator, create.ErrActionSetting, ResNameCrossAccountAttachment, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.CrossAccountAttachment.AttachmentArn)
	state.Name = flex.StringToFramework(ctx, out.CrossAccountAttachment.Name)
	state.CreatedTime = types.StringValue(out.CrossAccountAttachment.CreatedTime.Format(time.RFC3339))
	if out.CrossAccountAttachment.LastModifiedTime != nil {
		state.LastModifiedTime = types.StringValue(out.CrossAccountAttachment.LastModifiedTime.Format(time.RFC3339))
	}

	state.Principals = flex.FlattenFrameworkStringList(ctx, out.CrossAccountAttachment.Principals)

	resources, errDiags := flattenResources(ctx, out.CrossAccountAttachment.Resources)
	resp.Diagnostics.Append(errDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Resources = resources
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCrossAccountAttachment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().GlobalAcceleratorConn(ctx)

	var plan, state resourceCrossAccountAttachmentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &globalaccelerator.UpdateCrossAccountAttachmentInput{
		AttachmentArn: aws.String(state.ARN.ValueString()),
	}

	var diags diag.Diagnostics
	if !plan.Principals.Equal(state.Principals) {
		input.AddPrincipals, input.RemovePrincipals, diags = diffPrincipals(ctx, state.Principals, plan.Principals)
		resp.Diagnostics.Append(diags...)
	}

	if !plan.Resources.Equal(state.Resources) {
		input.AddResources, input.RemoveResources, diags = diffResources(ctx, state.Resources, plan.Resources)
		resp.Diagnostics.Append(diags...)
	}

	if !plan.Name.Equal(state.Name) {
		input.Name = aws.String(plan.Name.ValueString())
	}

	if input.Name != nil || input.AddPrincipals != nil || input.RemovePrincipals != nil || input.AddResources != nil || input.RemoveResources != nil {
		out, err := conn.UpdateCrossAccountAttachmentWithContext(ctx, input)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating CrossAccountAttachment",
				fmt.Sprintf("Could not update CrossAccountAttachment %s: %s", state.ARN.ValueString(), err),
			)
			return
		}

		state.CreatedTime = types.StringValue(out.CrossAccountAttachment.CreatedTime.Format(time.RFC3339))
		if out.CrossAccountAttachment.LastModifiedTime != nil {
			state.LastModifiedTime = types.StringValue(out.CrossAccountAttachment.LastModifiedTime.Format(time.RFC3339))
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceCrossAccountAttachment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().GlobalAcceleratorConn(ctx)

	var state resourceCrossAccountAttachmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &globalaccelerator.DeleteCrossAccountAttachmentInput{
		AttachmentArn: aws.String(state.ARN.ValueString()),
	}

	_, err := conn.DeleteCrossAccountAttachmentWithContext(ctx, input)

	if err != nil {
		var nfe *globalaccelerator.AttachmentNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting Global Accelerator CrossAccountAttachment",
			fmt.Sprintf("Could not delete CrossAccountAttachment %s: %s", state.ARN.ValueString(), err),
		)
		return
	}
}

func (r *resourceCrossAccountAttachment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	conn := r.Meta().GlobalAcceleratorConn(ctx)
	attachmentArn := req.ID

	output, err := conn.DescribeCrossAccountAttachmentWithContext(ctx, &globalaccelerator.DescribeCrossAccountAttachmentInput{
		AttachmentArn: aws.String(attachmentArn),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error describing CrossAccountAttachment", fmt.Sprintf("Could not describe CrossAccountAttachment with ARN %s: %s", attachmentArn, err))
		return
	}

	if output == nil || output.CrossAccountAttachment == nil {
		resp.Diagnostics.AddError("Error describing CrossAccountAttachment", fmt.Sprintf("CrossAccountAttachment with ARN %s not found", attachmentArn))
		return
	}

	var plan resourceCrossAccountAttachmentData
	plan.ARN = flex.StringToFramework(ctx, output.CrossAccountAttachment.AttachmentArn)
	plan.ID = flex.StringToFramework(ctx, output.CrossAccountAttachment.AttachmentArn)
	plan.Name = flex.StringToFramework(ctx, output.CrossAccountAttachment.Name)

	if output.CrossAccountAttachment.Principals != nil {
		plan.Principals = flex.FlattenFrameworkStringList(ctx, output.CrossAccountAttachment.Principals)
	}
	if output.CrossAccountAttachment.Resources != nil {
		resources, errDiags := flattenResources(ctx, output.CrossAccountAttachment.Resources)
		if errDiags.HasError() {
			resp.Diagnostics.Append(errDiags...)
			return
		}
		plan.Resources = resources
	}

	diags := resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
}

var ResourceDataElementType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"endpoint_id": types.StringType,
		"region":      types.StringType,
	},
}

func expandResources(tfList []ResourceData) []*globalaccelerator.Resource {
	if len(tfList) == 0 {
		return nil
	}

	apiResources := make([]*globalaccelerator.Resource, len(tfList))

	for i, tfResource := range tfList {
		apiResource := &globalaccelerator.Resource{
			EndpointId: aws.String(tfResource.EndpointID.ValueString()),
		}

		if !tfResource.Region.IsNull() && tfResource.Region.ValueString() != "" {
			apiResource.Region = aws.String(tfResource.Region.ValueString())
		}

		apiResources[i] = apiResource
	}

	return apiResources
}

func flattenResources(ctx context.Context, resources []*globalaccelerator.Resource) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(resources) == 0 {
		return types.ListNull(ResourceDataElementType), diags
	}

	elems := []attr.Value{}
	for _, resource := range resources {
		endpointID := aws.StringValue(resource.EndpointId)
		region := ""
		// Extract the region from the ARN if the endpoint ID is an ARN
		if arn.IsARN(endpointID) {
			parsedARN, err := arn.Parse(endpointID)
			if err != nil {
				diags.AddError("Error parsing ARN", err.Error())
				continue
			}
			region = parsedARN.Region
		}

		obj := map[string]attr.Value{
			"endpoint_id": types.StringValue(endpointID),
			"region":      types.StringValue(region),
		}
		objVal, d := types.ObjectValue(ResourceDataElementType.AttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(ResourceDataElementType, elems)
	diags.Append(d...)

	return listVal, diags
}

func diffResources(ctx context.Context, oldList, newList types.List) (toAdd, toRemove []*globalaccelerator.Resource, diags diag.Diagnostics) {
	toAdd = []*globalaccelerator.Resource{}
	toRemove = []*globalaccelerator.Resource{}
	var oldSlice, newSlice []ResourceData

	oldSlice, diags = convertListToResourceDataSlice(ctx, oldList)
	if diags.HasError() {
		return toAdd, toRemove, diags
	}

	newSlice, diags = convertListToResourceDataSlice(ctx, newList)
	if diags.HasError() {
		return toAdd, toRemove, diags
	}

	addSet, removeSet := diffResourceDataSlices(oldSlice, newSlice)

	for _, r := range addSet {
		toAdd = append(toAdd, &globalaccelerator.Resource{
			EndpointId: r.EndpointId,
			Region:     r.Region,
		})
	}
	for _, r := range removeSet {
		toRemove = append(toRemove, &globalaccelerator.Resource{
			EndpointId: r.EndpointId,
			Region:     r.Region,
		})
	}

	return toAdd, toRemove, diags
}

func convertListToResourceDataSlice(ctx context.Context, resourceList types.List) ([]ResourceData, diag.Diagnostics) {
	var diags diag.Diagnostics
	var resourceDataSlice []ResourceData

	if !resourceList.IsNull() {
		diags := resourceList.ElementsAs(ctx, &resourceDataSlice, false)
		if diags.HasError() {
			return nil, diags
		}
	}

	return resourceDataSlice, diags
}

func diffResourceDataSlices(oldSlice, newSlice []ResourceData) (toAdd, toRemove []*globalaccelerator.Resource) {
	toRemoveMap := make(map[string]*globalaccelerator.Resource)
	toAddMap := make(map[string]*globalaccelerator.Resource)

	for _, oldResource := range oldSlice {
		key := generateCompositeKey(oldResource.EndpointID.ValueString(), oldResource.Region.ValueString())
		apiResource := &globalaccelerator.Resource{
			EndpointId: aws.String(oldResource.EndpointID.ValueString()),
		}
		if !oldResource.Region.IsNull() && oldResource.Region.ValueString() != "" {
			apiResource.Region = aws.String(oldResource.Region.ValueString())
		}
		toRemoveMap[key] = apiResource
	}

	for _, newResource := range newSlice {
		key := generateCompositeKey(newResource.EndpointID.ValueString(), newResource.Region.ValueString())
		apiResource := &globalaccelerator.Resource{
			EndpointId: aws.String(newResource.EndpointID.ValueString()),
		}
		if !newResource.Region.IsNull() && newResource.Region.ValueString() != "" {
			apiResource.Region = aws.String(newResource.Region.ValueString())
		}
		if _, found := toRemoveMap[key]; found {
			delete(toRemoveMap, key)
		} else {
			toAddMap[key] = apiResource
		}
	}

	for _, resource := range toRemoveMap {
		toRemove = append(toRemove, resource)
	}
	for _, resource := range toAddMap {
		toAdd = append(toAdd, resource)
	}

	return toAdd, toRemove
}

func generateCompositeKey(endpointID, region string) string {
	if region == "" {
		region = "NO_REGION" // Special placeholder for resources without region
	}
	return endpointID + ":" + region
}

func diffPrincipals(ctx context.Context, oldList, newList types.List) (toAdd, toRemove []*string, diags diag.Diagnostics) {
	toAdd = []*string{}
	toRemove = []*string{}
	var oldSlice, newSlice []string

	if !oldList.IsNull() {
		var oldElements []types.String
		d := oldList.ElementsAs(ctx, &oldElements, false)
		diags = append(diags, d...)
		for _, element := range oldElements {
			oldSlice = append(oldSlice, element.ValueString())
		}
	}

	if !newList.IsNull() {
		var newElements []types.String
		d := newList.ElementsAs(ctx, &newElements, false)
		diags = append(diags, d...)
		for _, element := range newElements {
			newSlice = append(newSlice, element.ValueString())
		}
	}

	addSet, removeSet := diffSlices(oldSlice, newSlice)

	for elem := range addSet {
		toAdd = append(toAdd, aws.String(elem))
	}

	for elem := range removeSet {
		toRemove = append(toRemove, aws.String(elem))
	}

	return toAdd, toRemove, diags
}

func diffSlices(oldSlice, newSlice []string) (toAdd, toRemove map[string]struct{}) {
	toAdd = make(map[string]struct{})
	toRemove = make(map[string]struct{})

	for _, s := range oldSlice {
		toRemove[s] = struct{}{}
	}

	for _, s := range newSlice {
		if _, found := toRemove[s]; found {
			delete(toRemove, s)
			continue
		}
		toAdd[s] = struct{}{}
	}

	return toAdd, toRemove
}

type resourceCrossAccountAttachmentData struct {
	ID               types.String `tfsdk:"id"`
	ARN              types.String `tfsdk:"attachment_arn"`
	Name             types.String `tfsdk:"name"`
	Principals       types.List   `tfsdk:"principals"`
	Resources        types.List   `tfsdk:"resources"`
	CreatedTime      types.String `tfsdk:"created_time"`
	LastModifiedTime types.String `tfsdk:"last_modified_time"`
}

type ResourceData struct {
	EndpointID types.String `tfsdk:"endpoint_id"`
	Region     types.String `tfsdk:"region"`
}
