// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_datazone_domain", name="Domain")
// @Tags(identifierAttribute="arn")
func newDomainResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &domainResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type domainResource struct {
	framework.ResourceWithModel[domainResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *domainResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	authType := fwtypes.StringEnumType[awstypes.AuthType]()

	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"domain_execution_role": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrID: framework.IDAttribute(),
			"kms_key_identifier": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"portal_url": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"skip_deletion_check": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"single_sign_on": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[singleSignOnModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrType: schema.StringAttribute{
							CustomType: authType,
							Optional:   true,
							Computed:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Default: authType.AttributeDefault(awstypes.AuthTypeDisabled),
						},
						"user_assignment": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.UserAssignment](),
							Optional:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *domainResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data domainResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DataZoneClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input datazone.CreateDomainInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	const (
		timeout = 30 * time.Second
	)
	outputRaw, err := tfresource.RetryWhenAWSErrCodeContains(ctx, timeout, func() (any, error) {
		return conn.CreateDomain(ctx, &input)
	}, ErrorCodeAccessDenied)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating DataZone Domain (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	output := outputRaw.(*datazone.CreateDomainOutput)
	data.ARN = fwflex.StringToFramework(ctx, output.Arn)
	data.ID = fwflex.StringToFramework(ctx, output.Id)
	data.PortalURL = fwflex.StringToFramework(ctx, output.PortalUrl)

	if _, err := waitDomainCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for DataZone Domain (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *domainResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data domainResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DataZoneClient(ctx)

	output, err := findDomainByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading DataZone Domain (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Do not set single sign on in state if it was null and response is DISABLED as this is equivalent.
	if output.SingleSignOn.Type == awstypes.AuthTypeDisabled && data.SingleSignOn.IsNull() {
		output.SingleSignOn = nil
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *domainResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old domainResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DataZoneClient(ctx)

	if !new.Description.Equal(old.Description) ||
		!new.DomainExecutionRole.Equal(old.DomainExecutionRole) ||
		!new.Name.Equal(old.Name) ||
		!new.SingleSignOn.Equal(old.SingleSignOn) {
		var input datazone.UpdateDomainInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(sdkid.UniqueId())
		input.Identifier = fwflex.StringFromFramework(ctx, new.ID)

		_, err := conn.UpdateDomain(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating DataZone Domain (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *domainResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data domainResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DataZoneClient(ctx)

	input := datazone.DeleteDomainInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		Identifier:  data.ID.ValueStringPointer(),
	}
	if !data.SkipDeletionCheck.IsNull() {
		input.SkipDeletionCheck = data.SkipDeletionCheck.ValueBoolPointer()
	}
	_, err := conn.DeleteDomain(ctx, &input)

	if isResourceMissing(err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting DataZone Domain (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitDomainDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for DataZone Domain (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

func findDomainByID(ctx context.Context, conn *datazone.Client, id string) (*datazone.GetDomainOutput, error) {
	input := datazone.GetDomainInput{
		Identifier: aws.String(id),
	}
	output, err := conn.GetDomain(ctx, &input)

	if isResourceMissing(err) {
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

func statusDomain(ctx context.Context, conn *datazone.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findDomainByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitDomainCreated(ctx context.Context, conn *datazone.Client, id string, timeout time.Duration) (*datazone.GetDomainOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainStatusCreating),
		Target:  enum.Slice(awstypes.DomainStatusAvailable),
		Refresh: statusDomain(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*datazone.GetDomainOutput); ok {
		return output, err
	}

	return nil, err
}

func waitDomainDeleted(ctx context.Context, conn *datazone.Client, id string, timeout time.Duration) (*datazone.GetDomainOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainStatusAvailable, awstypes.DomainStatusDeleting),
		Target:  []string{},
		Refresh: statusDomain(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*datazone.GetDomainOutput); ok {
		return output, err
	}

	return nil, err
}

type domainResourceModel struct {
	framework.WithRegionModel
	ARN                 types.String                                       `tfsdk:"arn"`
	Description         types.String                                       `tfsdk:"description"`
	DomainExecutionRole fwtypes.ARN                                        `tfsdk:"domain_execution_role"`
	ID                  types.String                                       `tfsdk:"id"`
	KMSKeyIdentifier    fwtypes.ARN                                        `tfsdk:"kms_key_identifier"`
	Name                types.String                                       `tfsdk:"name"`
	PortalURL           types.String                                       `tfsdk:"portal_url"`
	SkipDeletionCheck   types.Bool                                         `tfsdk:"skip_deletion_check"`
	SingleSignOn        fwtypes.ListNestedObjectValueOf[singleSignOnModel] `tfsdk:"single_sign_on"`
	Tags                tftags.Map                                         `tfsdk:"tags"`
	TagsAll             tftags.Map                                         `tfsdk:"tags_all"`
	Timeouts            timeouts.Value                                     `tfsdk:"timeouts"`
}

type singleSignOnModel struct {
	Type           fwtypes.StringEnum[awstypes.AuthType]       `tfsdk:"type"`
	UserAssignment fwtypes.StringEnum[awstypes.UserAssignment] `tfsdk:"user_assignment"`
}
