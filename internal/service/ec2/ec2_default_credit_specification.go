// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_default_credit_specification", name="Default Credit Specification")
func newDefaultCreditSpecificationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &defaultCreditSpecificationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)

	return r, nil
}

type defaultCreditSpecificationResource struct {
	framework.ResourceWithModel[defaultCreditSpecificationResourceModel]
	framework.WithNoOpDelete
	framework.WithTimeouts
}

func (r *defaultCreditSpecificationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cpu_credits": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(cpuCredits_Values()...),
				},
			},
			"instance_family": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.UnlimitedSupportedInstanceFamily](),
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
			}),
		},
	}
}

func (r *defaultCreditSpecificationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data defaultCreditSpecificationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	instanceFamily := data.InstanceFamily.ValueEnum()
	var input ec2.ModifyDefaultCreditSpecificationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.ModifyDefaultCreditSpecification(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating EC2 Default Credit Specification (%s)", instanceFamily), err.Error())

		return
	}

	_, err = tfresource.RetryUntilEqual(ctx, r.CreateTimeout(ctx, data.Timeouts), data.CPUCredits.ValueString(), func() (string, error) {
		output, err := findDefaultCreditSpecificationByInstanceFamily(ctx, conn, instanceFamily)

		if err != nil {
			return "", err
		}

		return aws.ToString(output.CpuCredits), nil
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Default Credit Specification (%s) create", instanceFamily), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *defaultCreditSpecificationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data defaultCreditSpecificationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	output, err := findDefaultCreditSpecificationByInstanceFamily(ctx, conn, data.InstanceFamily.ValueEnum())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 Default Credit Specification (%s)", data.InstanceFamily.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *defaultCreditSpecificationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new defaultCreditSpecificationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	instanceFamily := new.InstanceFamily.ValueEnum()
	var input ec2.ModifyDefaultCreditSpecificationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.ModifyDefaultCreditSpecification(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating EC2 Default Credit Specification (%s)", instanceFamily), err.Error())

		return
	}

	_, err = tfresource.RetryUntilEqual(ctx, r.UpdateTimeout(ctx, new.Timeouts), new.CPUCredits.ValueString(), func() (string, error) {
		output, err := findDefaultCreditSpecificationByInstanceFamily(ctx, conn, instanceFamily)

		if err != nil {
			return "", err
		}

		return aws.ToString(output.CpuCredits), nil
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 Default Credit Specification (%s) update", instanceFamily), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *defaultCreditSpecificationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("instance_family"), request, response)
}

type defaultCreditSpecificationResourceModel struct {
	framework.WithRegionModel
	CPUCredits     types.String                                                  `tfsdk:"cpu_credits"`
	InstanceFamily fwtypes.StringEnum[awstypes.UnlimitedSupportedInstanceFamily] `tfsdk:"instance_family"`
	Timeouts       timeouts.Value                                                `tfsdk:"timeouts"`
}
