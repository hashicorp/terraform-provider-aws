// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_glue_integration_table_properties", name="Integration Table Properties")
// @IdentityAttribute("resource_arn")
// @IdentityAttribute("table_name")
// @Testing(tagsTest=false)
// @Testing(hasNoPreExistingResource=true)
func newIntegrationTablePropertiesResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &integrationTablePropertiesResource{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameIntegrationTableProperties = "Integration Table Properties"
)

type integrationTablePropertiesResource struct {
	framework.ResourceWithModel[integrationTablePropertiesResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *integrationTablePropertiesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"resource_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"table_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"source_table_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[sourceTableConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"fields": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
						},
						"filter_predicate": schema.StringAttribute{
							Optional: true,
						},
						"primary_key": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
						},
						"record_update_field": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"target_table_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[targetTableConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"target_table_name": schema.StringAttribute{
							Optional: true,
						},
						"unnest_spec": schema.StringAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]schema.Block{
						"partition_spec": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[integrationPartitionModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"conversion_spec": schema.StringAttribute{
										Optional: true,
									},
									"field_name": schema.StringAttribute{
										Optional: true,
									},
									"function_spec": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *integrationTablePropertiesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().GlueClient(ctx)

	var plan integrationTablePropertiesResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	var input glue.CreateIntegrationTablePropertiesInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.CreateIntegrationTableProperties(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ResourceArn.String(), plan.TableName.String())
		return
	}

	// CreateIntegrationTableProperties returns empty body, so we construct ID from ResourceArn and TableName
	id := fmt.Sprintf("%s%s%s", plan.ResourceArn.ValueString(), intflex.ResourceIdSeparator, plan.TableName.ValueString())
	plan.ID = types.StringValue(id)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitIntegrationTablePropertiesCreated(ctx, conn, plan.ResourceArn.ValueString(), plan.TableName.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *integrationTablePropertiesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().GlueClient(ctx)

	var state integrationTablePropertiesResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	resourceArn := state.ResourceArn.ValueString()
	tableName := state.TableName.ValueString()

	output, err := findIntegrationTableProperties(ctx, conn, resourceArn, tableName)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	// Construct the ID if not already set (needed for import and normal operations)
	if state.ID.IsNull() || state.ID.ValueString() == "" {
		id := fmt.Sprintf("%s%s%s", resourceArn, intflex.ResourceIdSeparator, tableName)
		state.ID = types.StringValue(id)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *integrationTablePropertiesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().GlueClient(ctx)

	var plan, state integrationTablePropertiesResourceModel
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

	if diff.HasChanges() {
		var input glue.UpdateIntegrationTablePropertiesInput
		input.ResourceArn = aws.String(plan.ResourceArn.ValueString())
		input.TableName = aws.String(plan.TableName.ValueString())

		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateIntegrationTableProperties(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}

		// Note: UpdateIntegrationTableProperties returns empty response body
		// No need to flatten response data as there are no computed attributes to update
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitIntegrationTablePropertiesUpdated(ctx, conn, plan.ResourceArn.ValueString(), plan.TableName.ValueString(), updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *integrationTablePropertiesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().GlueClient(ctx)

	var state integrationTablePropertiesResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := glue.DeleteIntegrationTablePropertiesInput{
		ResourceArn: aws.String(state.ResourceArn.ValueString()),
		TableName:   aws.String(state.TableName.ValueString()),
	}

	_, err := conn.DeleteIntegrationTableProperties(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitIntegrationTablePropertiesDeleted(ctx, conn, state.ResourceArn.ValueString(), state.TableName.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

func waitIntegrationTablePropertiesCreated(ctx context.Context, conn *glue.Client, resourceArn, tableName string, timeout time.Duration) (*glue.GetIntegrationTablePropertiesOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusIntegrationTableProperties(conn, resourceArn, tableName),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*glue.GetIntegrationTablePropertiesOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitIntegrationTablePropertiesUpdated(ctx context.Context, conn *glue.Client, resourceArn, tableName string, timeout time.Duration) (*glue.GetIntegrationTablePropertiesOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusIntegrationTableProperties(conn, resourceArn, tableName),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*glue.GetIntegrationTablePropertiesOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitIntegrationTablePropertiesDeleted(ctx context.Context, conn *glue.Client, resourceArn, tableName string, timeout time.Duration) (*glue.GetIntegrationTablePropertiesOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusIntegrationTableProperties(conn, resourceArn, tableName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*glue.GetIntegrationTablePropertiesOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusIntegrationTableProperties(conn *glue.Client, resourceArn, tableName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		input := &glue.GetIntegrationTablePropertiesInput{
			ResourceArn: aws.String(resourceArn),
			TableName:   aws.String(tableName),
		}

		output, err := conn.GetIntegrationTableProperties(ctx, input)
		if err != nil {
			if errs.IsA[*awstypes.EntityNotFoundException](err) {
				return nil, "", nil
			}
			return nil, "", smarterr.NewError(err)
		}

		// Since GetIntegrationTableProperties doesn't return status,
		// we assume it exists if we get a response
		return output, statusNormal, nil
	}
}

type integrationTablePropertiesResourceModel struct {
	framework.WithRegionModel
	ID                types.String                                            `tfsdk:"id"`
	ResourceArn       types.String                                            `tfsdk:"resource_arn"`
	SourceTableConfig fwtypes.ListNestedObjectValueOf[sourceTableConfigModel] `tfsdk:"source_table_config"`
	TableName         types.String                                            `tfsdk:"table_name"`
	TargetTableConfig fwtypes.ListNestedObjectValueOf[targetTableConfigModel] `tfsdk:"target_table_config"`
	Timeouts          timeouts.Value                                          `tfsdk:"timeouts"`
}

type sourceTableConfigModel struct {
	Fields            fwtypes.ListValueOf[types.String] `tfsdk:"fields"`
	FilterPredicate   types.String                      `tfsdk:"filter_predicate"`
	PrimaryKey        fwtypes.ListValueOf[types.String] `tfsdk:"primary_key"`
	RecordUpdateField types.String                      `tfsdk:"record_update_field"`
}

type targetTableConfigModel struct {
	PartitionSpec   fwtypes.ListNestedObjectValueOf[integrationPartitionModel] `tfsdk:"partition_spec"`
	TargetTableName types.String                                               `tfsdk:"target_table_name"`
	UnnestSpec      types.String                                               `tfsdk:"unnest_spec"`
}

type integrationPartitionModel struct {
	ConversionSpec types.String `tfsdk:"conversion_spec"`
	FieldName      types.String `tfsdk:"field_name"`
	FunctionSpec   types.String `tfsdk:"function_spec"`
}

var (
	_ inttypes.ImportIDParser = integrationTablePropertiesImportID{}
)

type integrationTablePropertiesImportID struct{}

func (integrationTablePropertiesImportID) Parse(id string) (string, map[string]any, error) {
	resourceArn, tableName, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id \"%s\" should be in the format <resource_arn>"+intflex.ResourceIdSeparator+"<table_name>", id)
	}

	result := map[string]any{
		names.AttrResourceARN: resourceArn,
		names.AttrTableName:   tableName,
	}

	return id, result, nil
}

func findIntegrationTableProperties(ctx context.Context, conn *glue.Client, resourceArn, tableName string) (*glue.GetIntegrationTablePropertiesOutput, error) {
	input := &glue.GetIntegrationTablePropertiesInput{
		ResourceArn: aws.String(resourceArn),
		TableName:   aws.String(tableName),
	}

	output, err := conn.GetIntegrationTableProperties(ctx, input)
	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
