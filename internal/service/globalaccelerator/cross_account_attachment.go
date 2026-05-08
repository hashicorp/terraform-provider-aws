// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package globalaccelerator

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/globalaccelerator"
	awstypes "github.com/aws/aws-sdk-go-v2/service/globalaccelerator/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_globalaccelerator_cross_account_attachment", name="Cross-Account Attachment")
// @Tags(identifierAttribute="arn")
// @ArnIdentity(identityDuplicateAttributes="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/globalaccelerator/types;awstypes;awstypes.Attachment")
// @Testing(preIdentityVersion="v5.100.0")
func newCrossAccountAttachmentResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &crossAccountAttachmentResource{}

	return r, nil
}

type crossAccountAttachmentResource struct {
	framework.ResourceWithModel[crossAccountAttachmentResourceModel]
	framework.WithImportByIdentity
}

func (r *crossAccountAttachmentResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"principals": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Optional:    true,
				ElementType: types.StringType,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"resource": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[resourceModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"endpoint_id": schema.StringAttribute{
							Optional: true,
						},
						names.AttrRegion: schema.StringAttribute{
							Optional: true,
						},
						names.AttrCIDRBlock: schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func (r *crossAccountAttachmentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data crossAccountAttachmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GlobalAcceleratorClient(ctx)

	var input globalaccelerator.CreateCrossAccountAttachmentInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.IdempotencyToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateCrossAccountAttachment(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating Global Accelerator Cross-account Attachment", err.Error())

		return
	}

	// Set values for unknowns.
	attachment := output.CrossAccountAttachment
	data.AttachmentARN = fwflex.StringToFramework(ctx, attachment.AttachmentArn)
	data.CreatedTime = fwflex.TimeToFramework(ctx, attachment.CreatedTime)
	data.ID = data.AttachmentARN
	data.LastModifiedTime = fwflex.TimeToFramework(ctx, attachment.LastModifiedTime)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *crossAccountAttachmentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data crossAccountAttachmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GlobalAcceleratorClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ID)
	attachment, err := findCrossAccountAttachmentByARN(ctx, conn, arn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Global Accelerator Cross-account Attachment (%s)", arn), err.Error())

		return
	}

	// Normalize return value.
	if data.Principals.IsNull() && len(attachment.Principals) == 0 {
		attachment.Principals = nil
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, attachment, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *crossAccountAttachmentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new crossAccountAttachmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GlobalAcceleratorClient(ctx)

	if !new.Name.Equal(old.Name) ||
		!new.Principals.Equal(old.Principals) ||
		!new.Resources.Equal(old.Resources) {
		arn := fwflex.StringValueFromFramework(ctx, new.ID)
		input := &globalaccelerator.UpdateCrossAccountAttachmentInput{
			AttachmentArn: aws.String(arn),
		}

		if !new.Name.Equal(old.Name) {
			input.Name = fwflex.StringFromFramework(ctx, new.Name)
		}

		if !new.Principals.Equal(old.Principals) {
			oldPrincipals, newPrincipals := fwflex.ExpandFrameworkStringValueSet(ctx, old.Principals), fwflex.ExpandFrameworkStringValueSet(ctx, new.Principals)
			input.AddPrincipals, input.RemovePrincipals = newPrincipals.Difference(oldPrincipals), oldPrincipals.Difference(newPrincipals)
		}

		if !new.Resources.Equal(old.Resources) {
			oldResources, diags := old.Resources.ToSlice(ctx)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			newResources, diags := new.Resources.ToSlice(ctx)
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}

			add, remove, _ := flex.DiffSlices(oldResources, newResources, func(v1, v2 *resourceModel) bool {
				return v1.Cidr.Equal(v2.Cidr) && v1.EndpointID.Equal(v2.EndpointID) && v1.Region.Equal(v2.Region)
			})

			input.AddResources = tfslices.ApplyToAll(add, func(v *resourceModel) awstypes.Resource {
				return awstypes.Resource{
					Cidr:       fwflex.StringFromFramework(ctx, v.Cidr),
					EndpointId: fwflex.StringFromFramework(ctx, v.EndpointID),
					Region:     fwflex.StringFromFramework(ctx, v.Region),
				}
			})
			input.RemoveResources = tfslices.ApplyToAll(remove, func(v *resourceModel) awstypes.Resource {
				return awstypes.Resource{
					Cidr:       fwflex.StringFromFramework(ctx, v.Cidr),
					EndpointId: fwflex.StringFromFramework(ctx, v.EndpointID),
					Region:     fwflex.StringFromFramework(ctx, v.Region),
				}
			})
		}

		output, err := conn.UpdateCrossAccountAttachment(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Global Accelerator Cross-account Attachment (%s)", arn), err.Error())

			return
		}

		new.LastModifiedTime = fwflex.TimeToFramework(ctx, output.CrossAccountAttachment.LastModifiedTime)
	} else {
		new.LastModifiedTime = old.LastModifiedTime
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *crossAccountAttachmentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data crossAccountAttachmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().GlobalAcceleratorClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, data.ID)
	input := globalaccelerator.DeleteCrossAccountAttachmentInput{
		AttachmentArn: aws.String(arn),
	}
	_, err := conn.DeleteCrossAccountAttachment(ctx, &input)

	if errs.IsA[*awstypes.AttachmentNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Global Accelerator Cross-account Attachment (%s)", arn), err.Error())

		return
	}
}

func findCrossAccountAttachmentByARN(ctx context.Context, conn *globalaccelerator.Client, arn string) (*awstypes.Attachment, error) {
	input := globalaccelerator.DescribeCrossAccountAttachmentInput{
		AttachmentArn: aws.String(arn),
	}

	return findCrossAccountAttachment(ctx, conn, &input)
}

func findCrossAccountAttachment(ctx context.Context, conn *globalaccelerator.Client, input *globalaccelerator.DescribeCrossAccountAttachmentInput) (*awstypes.Attachment, error) {
	output, err := conn.DescribeCrossAccountAttachment(ctx, input)

	if errs.IsA[*awstypes.AttachmentNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.CrossAccountAttachment == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.CrossAccountAttachment, nil
}

type crossAccountAttachmentResourceModel struct {
	AttachmentARN    types.String                                  `tfsdk:"arn"`
	CreatedTime      timetypes.RFC3339                             `tfsdk:"created_time"`
	ID               types.String                                  `tfsdk:"id"`
	LastModifiedTime timetypes.RFC3339                             `tfsdk:"last_modified_time"`
	Name             types.String                                  `tfsdk:"name"`
	Principals       fwtypes.SetValueOf[types.String]              `tfsdk:"principals"`
	Resources        fwtypes.SetNestedObjectValueOf[resourceModel] `tfsdk:"resource"`
	Tags             tftags.Map                                    `tfsdk:"tags"`
	TagsAll          tftags.Map                                    `tfsdk:"tags_all"`
}

type resourceModel struct {
	Cidr       types.String `tfsdk:"cidr_block"`
	EndpointID types.String `tfsdk:"endpoint_id"`
	Region     types.String `tfsdk:"region"`
}
