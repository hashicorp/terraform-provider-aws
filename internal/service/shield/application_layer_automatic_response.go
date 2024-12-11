// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	awstypes "github.com/aws/aws-sdk-go-v2/service/shield/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type applicationLayerAutomaticResponseAction string

const (
	applicationLayerAutomaticResponseActionBlock applicationLayerAutomaticResponseAction = "BLOCK"
	applicationLayerAutomaticResponseActionCount applicationLayerAutomaticResponseAction = "COUNT"
)

func (applicationLayerAutomaticResponseAction) Values() []applicationLayerAutomaticResponseAction {
	return []applicationLayerAutomaticResponseAction{
		applicationLayerAutomaticResponseActionBlock,
		applicationLayerAutomaticResponseActionCount,
	}
}

// @FrameworkResource(name="Application Layer Automatic Response")
func newApplicationLayerAutomaticResponseResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &applicationLayerAutomaticResponseResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type applicationLayerAutomaticResponseResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *applicationLayerAutomaticResponseResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_shield_application_layer_automatic_response"
}

func (r *applicationLayerAutomaticResponseResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAction: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[applicationLayerAutomaticResponseAction](),
				Required:   true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrResourceARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *applicationLayerAutomaticResponseResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data applicationLayerAutomaticResponseResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ShieldClient(ctx)

	action := &awstypes.ResponseAction{}
	switch data.Action.ValueEnum() {
	case applicationLayerAutomaticResponseActionBlock:
		action.Block = &awstypes.BlockAction{}
	case applicationLayerAutomaticResponseActionCount:
		action.Count = &awstypes.CountAction{}
	}

	resourceARN := data.ResourceARN.ValueString()
	input := &shield.EnableApplicationLayerAutomaticResponseInput{
		Action:      action,
		ResourceArn: aws.String(resourceARN),
	}

	_, err := conn.EnableApplicationLayerAutomaticResponse(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("enabling Shield Application Layer Automatic Response (%s)", resourceARN), err.Error())

		return
	}

	// Set values for unknowns.
	data.setID()

	if _, err := waitApplicationLayerAutomaticResponseEnabled(ctx, conn, resourceARN, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Shield Application Layer Automatic Response (%s) create", resourceARN), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *applicationLayerAutomaticResponseResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data applicationLayerAutomaticResponseResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().ShieldClient(ctx)

	resourceARN := data.ID.ValueString()
	output, err := findApplicationLayerAutomaticResponseByResourceARN(ctx, conn, resourceARN)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Shield Application Layer Automatic Response (%s)", resourceARN), err.Error())

		return
	}

	if output.Action.Block != nil {
		data.Action = fwtypes.StringEnumValue(applicationLayerAutomaticResponseActionBlock)
	} else if output.Action.Count != nil {
		data.Action = fwtypes.StringEnumValue(applicationLayerAutomaticResponseActionCount)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *applicationLayerAutomaticResponseResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new applicationLayerAutomaticResponseResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ShieldClient(ctx)

	if !new.Action.Equal(old.Action) {
		action := &awstypes.ResponseAction{}
		switch new.Action.ValueEnum() {
		case applicationLayerAutomaticResponseActionBlock:
			action.Block = &awstypes.BlockAction{}
		case applicationLayerAutomaticResponseActionCount:
			action.Count = &awstypes.CountAction{}
		}

		resourceARN := new.ResourceARN.ValueString()
		input := &shield.UpdateApplicationLayerAutomaticResponseInput{
			Action:      action,
			ResourceArn: aws.String(resourceARN),
		}

		_, err := conn.UpdateApplicationLayerAutomaticResponse(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Shield Application Layer Automatic Response (%s)", resourceARN), err.Error())

			return
		}

		if _, err := waitApplicationLayerAutomaticResponseEnabled(ctx, conn, resourceARN, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Shield Application Layer Automatic Response (%s) update", resourceARN), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *applicationLayerAutomaticResponseResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data applicationLayerAutomaticResponseResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ShieldClient(ctx)

	resourceARN := data.ID.ValueString()
	input := &shield.DisableApplicationLayerAutomaticResponseInput{
		ResourceArn: aws.String(resourceARN),
	}

	_, err := conn.DisableApplicationLayerAutomaticResponse(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("disabling Application Layer Automatic Response (%s)", resourceARN), err.Error())

		return
	}

	if _, err := waitApplicationLayerAutomaticResponseDeleted(ctx, conn, resourceARN, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Shield Application Layer Automatic Response (%s) delete", resourceARN), err.Error())

		return
	}
}

func findApplicationLayerAutomaticResponseByResourceARN(ctx context.Context, conn *shield.Client, arn string) (*awstypes.ApplicationLayerAutomaticResponseConfiguration, error) {
	output, err := findProtectionByResourceARN(ctx, conn, arn)

	if err != nil {
		return nil, err
	}

	if output.ApplicationLayerAutomaticResponseConfiguration == nil || output.ApplicationLayerAutomaticResponseConfiguration.Action == nil {
		return nil, tfresource.NewEmptyResultError(arn)
	}

	if status := output.ApplicationLayerAutomaticResponseConfiguration.Status; status == awstypes.ApplicationLayerAutomaticResponseStatusDisabled {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: arn,
		}
	}

	return output.ApplicationLayerAutomaticResponseConfiguration, nil
}

func statusApplicationLayerAutomaticResponse(ctx context.Context, conn *shield.Client, resourceARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findApplicationLayerAutomaticResponseByResourceARN(ctx, conn, resourceARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitApplicationLayerAutomaticResponseEnabled(ctx context.Context, conn *shield.Client, resourceARN string, timeout time.Duration) (*awstypes.ApplicationLayerAutomaticResponseConfiguration, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.ApplicationLayerAutomaticResponseStatusEnabled),
		Refresh:                   statusApplicationLayerAutomaticResponse(ctx, conn, resourceARN),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ApplicationLayerAutomaticResponseConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitApplicationLayerAutomaticResponseDeleted(ctx context.Context, conn *shield.Client, resourceARN string, timeout time.Duration) (*awstypes.ApplicationLayerAutomaticResponseConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationLayerAutomaticResponseStatusEnabled),
		Target:  []string{},
		Refresh: statusApplicationLayerAutomaticResponse(ctx, conn, resourceARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ApplicationLayerAutomaticResponseConfiguration); ok {
		return output, err
	}

	return nil, err
}

type applicationLayerAutomaticResponseResourceModel struct {
	Action      fwtypes.StringEnum[applicationLayerAutomaticResponseAction] `tfsdk:"action"`
	ID          types.String                                                `tfsdk:"id"`
	ResourceARN fwtypes.ARN                                                 `tfsdk:"resource_arn"`
	Timeouts    timeouts.Value                                              `tfsdk:"timeouts"`
}

func (data *applicationLayerAutomaticResponseResourceModel) InitFromID() error {
	_, err := arn.Parse(data.ID.ValueString())
	if err != nil {
		return err
	}

	data.ResourceARN = fwtypes.ARNValue(data.ID.ValueString())

	return nil
}

func (data *applicationLayerAutomaticResponseResourceModel) setID() {
	data.ID = data.ResourceARN.StringValue
}
