// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"errors"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_dms_instance_profile", name="Instance Profile")
// @Tags(identifierAttribute="arn")
// @ArnIdentity(identityDuplicateAttributes="id")
// @Testing(generator=false)
// @Testing(hasNoPreExistingResource=true)
func newInstanceProfileResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &instanceProfileResource{}, nil
}

type instanceProfileResource struct {
	framework.ResourceWithModel[instanceProfileResourceModel]
	framework.WithImportByIdentity
}

func (r *instanceProfileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrAvailabilityZone: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrKMSKeyARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf(instanceProfileNetworkType_Values()...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrPubliclyAccessible: schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"subnet_group_identifier": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrVPCSecurityGroupIDs: schema.SetAttribute{
				CustomType: fwtypes.SetOfStringType,
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *instanceProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DMSClient(ctx)

	var plan instanceProfileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input databasemigrationservice.CreateInstanceProfileInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("InstanceProfile")))
	if resp.Diagnostics.HasError() {
		return
	}
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateInstanceProfile(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.ValueString())
		return
	}
	if out == nil || out.InstanceProfile == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out.InstanceProfile, &plan, fwflex.WithFieldNamePrefix("InstanceProfile")))
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = fwflex.StringToFramework(ctx, out.InstanceProfile.InstanceProfileArn)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *instanceProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DMSClient(ctx)

	var state instanceProfileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findInstanceProfileByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state, fwflex.WithFieldNamePrefix("InstanceProfile")))
	if resp.Diagnostics.HasError() {
		return
	}
	state.ID = fwflex.StringToFramework(ctx, out.InstanceProfileArn)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *instanceProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DMSClient(ctx)

	var plan, state instanceProfileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input databasemigrationservice.ModifyInstanceProfileInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("InstanceProfile")))
		if resp.Diagnostics.HasError() {
			return
		}
		input.InstanceProfileIdentifier = state.ID.ValueStringPointer()

		out, err := conn.ModifyInstanceProfile(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.ValueString())
			return
		}
		if out == nil || out.InstanceProfile == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ID.ValueString())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out.InstanceProfile, &plan, fwflex.WithFieldNamePrefix("InstanceProfile")))
		if resp.Diagnostics.HasError() {
			return
		}
		plan.ID = fwflex.StringToFramework(ctx, out.InstanceProfile.InstanceProfileArn)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *instanceProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DMSClient(ctx)

	var state instanceProfileResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := databasemigrationservice.DeleteInstanceProfileInput{
		InstanceProfileIdentifier: state.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteInstanceProfile(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}
}

func findInstanceProfileByID(ctx context.Context, conn *databasemigrationservice.Client, id string) (*awstypes.InstanceProfile, error) {
	input := databasemigrationservice.DescribeInstanceProfilesInput{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("instance-profile-identifier"),
				Values: []string{id},
			},
		},
	}

	return findInstanceProfile(ctx, conn, &input)
}

func findInstanceProfile(ctx context.Context, conn *databasemigrationservice.Client, input *databasemigrationservice.DescribeInstanceProfilesInput) (*awstypes.InstanceProfile, error) {
	output, err := findInstanceProfiles(ctx, conn, input)
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output))
}

func findInstanceProfiles(ctx context.Context, conn *databasemigrationservice.Client, input *databasemigrationservice.DescribeInstanceProfilesInput) ([]awstypes.InstanceProfile, error) {
	var output []awstypes.InstanceProfile

	pages := databasemigrationservice.NewDescribeInstanceProfilesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if errs.IsA[*awstypes.ResourceNotFoundFault](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		output = append(output, page.InstanceProfiles...)
	}

	return output, nil
}

type instanceProfileResourceModel struct {
	framework.WithRegionModel
	ARN                   types.String        `tfsdk:"arn"`
	AvailabilityZone      types.String        `tfsdk:"availability_zone"`
	Description           types.String        `tfsdk:"description"`
	ID                    types.String        `tfsdk:"id"`
	KMSKeyARN             fwtypes.ARN         `tfsdk:"kms_key_arn"`
	Name                  types.String        `tfsdk:"name"`
	NetworkType           types.String        `tfsdk:"network_type"`
	PubliclyAccessible    types.Bool          `tfsdk:"publicly_accessible"`
	SubnetGroupIdentifier types.String        `tfsdk:"subnet_group_identifier"`
	Tags                  tftags.Map          `tfsdk:"tags"`
	TagsAll               tftags.Map          `tfsdk:"tags_all"`
	VpcSecurityGroups     fwtypes.SetOfString `tfsdk:"vpc_security_group_ids"`
}
