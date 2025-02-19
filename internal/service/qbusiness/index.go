// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"errors"
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
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
			names.AttrApplicationID: schema.StringAttribute{
				Description: "Identifier of the Amazon Q application associated with the index",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid application ID"),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
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
			names.AttrID: framework.IDAttribute(),
			"index_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
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
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QBusiness, create.ErrActionCreating, ResNameIndex, data.DisplayName.String(), err),
			err.Error(),
		)
		return
	}

	data.IndexId = fwflex.StringToFramework(ctx, out.IndexId)
	data.IndexArn = fwflex.StringToFramework(ctx, out.IndexArn)

	resp.Diagnostics.Append(data.setID()...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID.String())

	if _, err := waitIndexActive(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QBusiness, create.ErrActionWaitingForCreation, ResNameIndex, data.ID.String(), err),
			err.Error(),
		)
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
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QBusiness, create.ErrActionCreating, ResNameIndex, data.ID.String(), err),
				err.Error(),
			)
			return
		}

		if _, err := waitIndexActive(ctx, conn, data.ID.ValueString(), r.UpdateTimeout(ctx, data.Timeouts)); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QBusiness, create.ErrActionWaitingForCreation, ResNameIndex, data.ID.String(), err),
				err.Error(),
			)
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
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QBusiness, create.ErrActionDeleting, ResNameIndex, data.ID.String(), err),
			err.Error(),
		)
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
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QBusiness, create.ErrActionDeleting, ResNameIndex, data.IndexId.ValueString(), err),
			err.Error(),
		)
		return
	}

	if _, err := waitIndexDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QBusiness, create.ErrActionWaitingForDeletion, ResNameIndex, data.IndexId.ValueString(), err),
			err.Error(),
		)
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

	out, err := findIndexByID(ctx, conn, data.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QBusiness, create.ErrActionReading, ResNameIndex, data.ID.ValueString(), err),
			err.Error(),
		)
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
	var state, plan resourceIndexData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := plan.initFromID(); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QBusiness, create.ErrActionUpdating, ResNameIndex, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	if !state.CapacityConfiguration.Equal(plan.CapacityConfiguration) ||
		!state.DisplayName.Equal(plan.DisplayName) ||
		!state.Description.Equal(plan.Description) ||
		!state.DocumentAttributeConfigurations.Equal(plan.DocumentAttributeConfigurations) {
		conn := r.Meta().QBusinessClient(ctx)

		input := &qbusiness.UpdateIndexInput{}

		resp.Diagnostics.Append(fwflex.Expand(ctx, plan, input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateIndex(ctx, input)

		id := plan.ID.ValueString()
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QBusiness, create.ErrActionUpdating, ResNameIndex, id, err),
				err.Error(),
			)
			return
		}

		if _, err := waitIndexActive(ctx, conn, id, r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.QBusiness, create.ErrActionWaitingForUpdate, ResNameIndex, id, err),
				err.Error(),
			)
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceIndex) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

type resourceIndexData struct {
	ApplicationId                   types.String                                                       `tfsdk:"application_id"`
	CapacityConfiguration           fwtypes.ListNestedObjectValueOf[capacityConfigurationData]         `tfsdk:"capacity_configuration"`
	Description                     types.String                                                       `tfsdk:"description"`
	DisplayName                     types.String                                                       `tfsdk:"display_name"`
	DocumentAttributeConfigurations fwtypes.SetNestedObjectValueOf[documentAttributeConfigurationData] `tfsdk:"document_attribute_configuration"`
	ID                              types.String                                                       `tfsdk:"id"`
	IndexArn                        types.String                                                       `tfsdk:"arn"`
	IndexId                         types.String                                                       `tfsdk:"index_id"`
	Tags                            tftags.Map                                                         `tfsdk:"tags"`
	TagsAll                         tftags.Map                                                         `tfsdk:"tags_all"`
	Timeouts                        timeouts.Value                                                     `tfsdk:"timeouts"`
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

func (data *resourceIndexData) setID() diag.Diagnostics {
	var diags diag.Diagnostics

	id, err := flex.FlattenResourceId([]string{data.ApplicationId.ValueString(), data.IndexId.ValueString()}, indexResourceIDPartCount, false)
	if err != nil {
		diags.AddError(
			create.ProblemStandardMessage(names.QBusiness, create.ErrActionFlatteningResourceId, ResNameIndex, id, err),
			err.Error())
		return diags
	}
	data.ID = types.StringValue(id)
	return diags
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

func findIndexByID(ctx context.Context, conn *qbusiness.Client, index_id string) (*qbusiness.GetIndexOutput, error) {
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

func statusIndex(ctx context.Context, conn *qbusiness.Client, index_id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findIndexByID(ctx, conn, index_id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitIndexActive(ctx context.Context, conn *qbusiness.Client, index_id string, timeout time.Duration) (*qbusiness.GetIndexOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.IndexStatusCreating, awstypes.IndexStatusUpdating),
		Target:     enum.Slice(awstypes.IndexStatusActive),
		Refresh:    statusIndex(ctx, conn, index_id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetIndexOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.Status)))

		return output, err
	}
	return nil, err
}

func waitIndexDeleted(ctx context.Context, conn *qbusiness.Client, index_id string, timeout time.Duration) (*qbusiness.GetIndexOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.IndexStatusActive, awstypes.IndexStatusDeleting),
		Target:     []string{},
		Refresh:    statusIndex(ctx, conn, index_id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetIndexOutput); ok {
		tfresource.SetLastError(err, errors.New(string(output.Status)))

		return output, err
	}
	return nil, err
}
