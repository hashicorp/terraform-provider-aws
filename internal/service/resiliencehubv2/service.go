// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_resiliencehubv2_service", name="Service")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types;awstypes;awstypes.Service")
// @Testing(hasNoPreExistingResource=true)
func newServiceResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &serviceResource{}, nil
}

type serviceResource struct {
	framework.ResourceWithModel[serviceResourceModel]
	framework.WithImportByIdentity
}

func (r *serviceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			// TODO associated_system, dependency_discovery, report_configuration.
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: fwschema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 615),
				},
			},
			names.AttrKMSKeyID: fwschema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: fwschema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					validResourceName,
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_arn": fwschema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			"regions": fwschema.SetAttribute{
				CustomType: fwtypes.SetOfStringType,
				Required:   true,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 5),
					setvalidator.ValueStringsAre(fwvalidators.AWSRegion()),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]fwschema.Block{
			"permission_model": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[permissionModelModel](ctx),
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						// TODO cross_account_role.
						"invoker_role_name": fwschema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

func (r *serviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	var input resiliencehubv2.CreateServiceInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	output, err := tfresource.RetryWhenIsAErrorMessageContains[*resiliencehubv2.CreateServiceOutput, *awstypes.ValidationException](ctx, propagationTimeout, func(ctx context.Context) (*resiliencehubv2.CreateServiceOutput, error) {
		return conn.CreateService(ctx, &input)
	}, "Ensure the role exists and its trust policy allows access")
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, output.Service, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *serviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ServiceARN)
	svc, err := findServiceByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, svc, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, state))
}

func (r *serviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state serviceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		arn := fwflex.StringValueFromFramework(ctx, plan.ServiceARN)
		var input resiliencehubv2.UpdateServiceInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.ValidationException](ctx, propagationTimeout, func(ctx context.Context) (any, error) {
			return conn.UpdateService(ctx, &input)
		}, "Ensure the role exists and its trust policy allows access")
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *serviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ServiceARN)
	input := resiliencehubv2.DeleteServiceInput{
		ServiceArn: aws.String(arn),
	}
	_, err := conn.DeleteService(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
	}
}

func (r *serviceResource) flatten(ctx context.Context, svc *awstypes.Service, data *serviceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, svc, data)...)
	if diags.HasError() {
		return diags
	}

	setTagsOut(ctx, svc.Tags)

	return diags
}

func findServiceByARN(ctx context.Context, conn *resiliencehubv2.Client, arn string) (*awstypes.Service, error) {
	input := resiliencehubv2.GetServiceInput{
		ServiceArn: aws.String(arn),
	}

	return findService(ctx, conn, &input)
}

func findService(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.GetServiceInput) (*awstypes.Service, error) {
	output, err := conn.GetService(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil || output.Service == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output.Service, nil
}

type serviceResourceModel struct {
	framework.WithRegionModel
	Description     types.String                                          `tfsdk:"description"`
	KMSKeyID        fwtypes.ARN                                           `tfsdk:"kms_key_id"`
	Name            types.String                                          `tfsdk:"name"`
	PermissionModel fwtypes.ListNestedObjectValueOf[permissionModelModel] `tfsdk:"permission_model"`
	PolicyARN       fwtypes.ARN                                           `tfsdk:"policy_arn"`
	Regions         fwtypes.SetOfString                                   `tfsdk:"regions"`
	ServiceARN      types.String                                          `tfsdk:"arn"`
	Tags            tftags.Map                                            `tfsdk:"tags"`
	TagsAll         tftags.Map                                            `tfsdk:"tags_all"`
}

type permissionModelModel struct {
	InvokerRoleName types.String `tfsdk:"invoker_role_name"`
}
