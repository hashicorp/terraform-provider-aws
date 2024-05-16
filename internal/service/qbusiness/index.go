// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	awstypes "github.com/aws/aws-sdk-go-v2/service/qbusiness/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_qbusiness_index", name="Index")
// @Tags(identifierAttribute="arn")
func newResourceIndex(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceIndex{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameIndex = "Index"
)

type resourceIndex struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceIndex) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID:      framework.IDAttribute(),
			names.AttrARN:     framework.ARNAttributeComputedOnly(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"application_id": schema.StringAttribute{
				Description: "Identifier of the Amazon Q application associated with the index",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid application ID"),
				},
			},
			"index_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Description: "A description of the Amazon Q application.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1000),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				},
			},
			names.AttrDisplayName: schema.StringAttribute{
				Description: "The display name of the Amazon Q application.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1000),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"capacity_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[capacityConfigurationData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"units": schema.Int64Attribute{
							Required:    true,
							Description: "The number of additional storage units for the Amazon Q index.",
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
					},
				},
			},
			"document_attribute_configuration": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[documentAttributeConfigurationData](ctx),
				Validators: []validator.Set{
					setvalidator.SizeAtMost(500),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrName: schema.StringAttribute{
							Required:    true,
							Description: "The name of the document attribute.",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 2048),
							},
						},
						"search": schema.StringAttribute{
							Required:    true,
							Description: "Information about whether the document attribute can be used by an end user to search for information on their web experience.",
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.Status](),
							},
						},
						names.AttrType: schema.StringAttribute{
							Required:    true,
							Description: "The type of document attribute.",
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.AttributeType](),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
				Update: true,
			}),
		},
	}
}

func (r *resourceIndex) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceIndexData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)

	input := &qbusiness.CreateIndexInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)

	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)
	input.ClientToken = aws.String(id.UniqueId())

	out, err := conn.CreateIndex(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError("failed to create Amazon Q index", err.Error())
		return
	}

	data.IndexId = fwflex.StringToFramework(ctx, out.IndexId)
	data.IndexArn = fwflex.StringToFramework(ctx, out.IndexArn)

	data.setID()

	if _, err := waitIndexCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError("failed to wait for Amazon Q index creation", err.Error())
		return
	}

	update := &qbusiness.UpdateIndexInput{}

	resp.Diagnostics.Append(fwflex.Expand(ctx, data, update)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(update.DocumentAttributeConfigurations) > 0 {
		_, err = conn.UpdateIndex(ctx, update)
		if err != nil {
			resp.Diagnostics.AddError("failed to update Amazon Q index", err.Error())
			return
		}

		if _, err := waitIndexUpdated(ctx, conn, data.ID.ValueString(), r.UpdateTimeout(ctx, data.Timeouts)); err != nil {
			resp.Diagnostics.AddError("failed to wait for Amazon Q index update", err.Error())
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceIndex) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resourceIndexData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)

	if err := data.initFromID(); err != nil {
		resp.Diagnostics.AddError("parsing resource ID", err.Error())
		return
	}

	indexId := data.IndexId.ValueString()
	appId := data.ApplicationId.ValueString()

	input := &qbusiness.DeleteIndexInput{
		IndexId:       aws.String(indexId),
		ApplicationId: aws.String(appId),
	}

	_, err := conn.DeleteIndex(ctx, input)

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("failed to delete Q Business index (%s)", data.IndexId.ValueString()), err.Error())
		return
	}

	if _, err := waitIndexDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError("failed to wait for Q Business application deletion", err.Error())
		return
	}
}

func (r *resourceIndex) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_qbusiness_index"
}

func (r *resourceIndex) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceIndexData
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QBusinessClient(ctx)

	out, err := FindIndexByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("failed to retrieve Q Business index (%s)", data.ID.ValueString()), err.Error())
		return
	}

	out.DocumentAttributeConfigurations = filterDefaultDocumentAttributeConfigurations(out.DocumentAttributeConfigurations)

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceIndex) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, new resourceIndexData

	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := new.initFromID(); err != nil {
		resp.Diagnostics.AddError("parsing resource ID", err.Error())
		return
	}

	if !old.CapacityConfiguration.Equal(new.CapacityConfiguration) ||
		!old.DisplayName.Equal(new.DisplayName) ||
		!old.Description.Equal(new.Description) ||
		!old.DocumentAttributeConfigurations.Equal(new.DocumentAttributeConfigurations) {
		conn := r.Meta().QBusinessClient(ctx)

		input := &qbusiness.UpdateIndexInput{}

		resp.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateIndex(ctx, input)

		if err != nil {
			resp.Diagnostics.AddError("failed to update Amazon Q index", err.Error())
			return
		}

		if _, err := waitIndexUpdated(ctx, conn, new.ID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			resp.Diagnostics.AddError("failed to wait for Amazon Q index update", err.Error())
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *resourceIndex) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

type resourceIndexData struct {
	ApplicationId                   types.String                                                       `tfsdk:"application_id"`
	CapacityConfiguration           fwtypes.ListNestedObjectValueOf[capacityConfigurationData]         `tfsdk:"capacity_configuration"`
	DisplayName                     types.String                                                       `tfsdk:"display_name"`
	Description                     types.String                                                       `tfsdk:"description"`
	ID                              types.String                                                       `tfsdk:"id"`
	IndexId                         types.String                                                       `tfsdk:"index_id"`
	IndexArn                        types.String                                                       `tfsdk:"arn"`
	Tags                            types.Map                                                          `tfsdk:"tags"`
	TagsAll                         types.Map                                                          `tfsdk:"tags_all"`
	Timeouts                        timeouts.Value                                                     `tfsdk:"timeouts"`
	DocumentAttributeConfigurations fwtypes.SetNestedObjectValueOf[documentAttributeConfigurationData] `tfsdk:"document_attribute_configuration"`
}

type capacityConfigurationData struct {
	Units types.Int64 `tfsdk:"units"`
}

type documentAttributeConfigurationData struct {
	Name   types.String                               `tfsdk:"name"`
	Search fwtypes.StringEnum[awstypes.Status]        `tfsdk:"search"`
	Type   fwtypes.StringEnum[awstypes.AttributeType] `tfsdk:"type"`
}

const (
	indexResourceIDPartCount = 2
)

func (data *resourceIndexData) setID() {
	data.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{data.ApplicationId.ValueString(), data.IndexId.ValueString()}, indexResourceIDPartCount, false)))
}

func (data *resourceIndexData) initFromID() error {
	parts, err := flex.ExpandResourceId(data.ID.ValueString(), indexResourceIDPartCount, false)
	if err != nil {
		return err
	}

	data.ApplicationId = types.StringValue(parts[0])
	data.IndexId = types.StringValue(parts[1])
	return nil
}

func FindIndexByID(ctx context.Context, conn *qbusiness.Client, index_id string) (*qbusiness.GetIndexOutput, error) {
	parts, err := flex.ExpandResourceId(index_id, indexResourceIDPartCount, false)

	if err != nil {
		return nil, err
	}

	input := &qbusiness.GetIndexInput{
		ApplicationId: aws.String(parts[0]),
		IndexId:       aws.String(parts[1]),
	}

	output, err := conn.GetIndex(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func filterDefaultDocumentAttributeConfigurations(conf []awstypes.DocumentAttributeConfiguration) []awstypes.DocumentAttributeConfiguration {
	var attributes []awstypes.DocumentAttributeConfiguration
	for _, attribute := range conf {
		filter := false
		if strings.HasPrefix(aws.ToString(attribute.Name), "_") {
			filter = true
		}
		if aws.ToString(attribute.Name) == "_document_title" && attribute.Search == awstypes.StatusDisabled {
			filter = false
		}
		if filter {
			continue
		}
		attributes = append(attributes, attribute)
	}
	return attributes
}
