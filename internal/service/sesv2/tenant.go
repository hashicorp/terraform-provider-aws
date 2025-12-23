// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sesv2_tenant", name="Tenant")
// @Tags(identifierAttribute="arn")
func newResourceTenant(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceTenant{}
	return r, nil
}

const (
	ResNameTenant = "Tenant"
)

type resourceTenant struct {
	framework.ResourceWithModel[resourceTenantModel]
	client *sesv2.Client
}

func (r *resourceTenant) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*conns.AWSClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *conns.AWSClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client.SESV2Client(ctx)
}

func (r *resourceTenant) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"created_timestamp": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp of when the Tenant was created",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"tenant_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Name of the Tenant",
			},
			"sending_status": schema.StringAttribute{
				Computed:    true,
				Description: "The sending status of the tenant. ENABLED, DISABLED, or REINSTATED",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

var tenantFlattenIgnoredFields = []string{
	"CreatedTimestamp",
	"TagsAll",
	"Tags",
}

func (r *resourceTenant) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.client

	var plan resourceTenantModel

	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	var input sesv2.CreateTenantInput
	input.Tags = getTagsIn(ctx)

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Tenant")))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateTenant(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.TenantName.String())
		return
	}
	if out == nil || out.TenantId == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.TenantName.String())
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan, flex.WithFieldNamePrefix("Tenant"), flex.WithIgnoredFieldNames(tenantFlattenIgnoredFields)))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(aws.ToString(out.TenantName))
	plan.CreatedTimestamp = types.StringValue(aws.ToTime(out.CreatedTimestamp).Format(time.RFC3339))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceTenant) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.client

	var state resourceTenantModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindTenantByName(ctx, conn, state.TenantName.ValueString())

	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("Tenant"), flex.WithIgnoredFieldNames(tenantFlattenIgnoredFields)))
	if resp.Diagnostics.HasError() {
		return
	}

	kvTags := keyValueTags(ctx, getTagsIn(ctx))
	plainMap := kvTags.Map()
	state.TagsAll = tftags.FlattenStringValueMap(ctx, plainMap)
	state.ID = types.StringValue(aws.ToString(out.TenantName))
	state.CreatedTimestamp = types.StringValue(aws.ToTime(out.CreatedTimestamp).Format(time.RFC3339))
	state.ARN = types.StringValue(aws.ToString(out.TenantArn))
	state.SendingStatus = types.StringValue(string(out.SendingStatus))

	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceTenant) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.client

	var state resourceTenantModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := sesv2.DeleteTenantInput{
		TenantName: state.TenantName.ValueStringPointer(),
	}

	_, err := conn.DeleteTenant(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

}

func (r *resourceTenant) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("tenant_name"), req, resp)
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

func FindTenantByName(ctx context.Context, conn *sesv2.Client, name string) (*awstypes.Tenant, error) {
	input := sesv2.GetTenantInput{
		TenantName: aws.String(name),
	}

	out, err := conn.GetTenant(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Tenant == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out.Tenant, nil
}

type resourceTenantModel struct {
	framework.WithRegionModel
	ARN              types.String `tfsdk:"arn"`
	CreatedTimestamp types.String `tfsdk:"created_timestamp"`
	ID               types.String `tfsdk:"id"`
	SendingStatus    types.String `tfsdk:"sending_status"`
	Tags             tftags.Map   `tfsdk:"tags"`
	TagsAll          tftags.Map   `tfsdk:"tags_all"`
	TenantName       types.String `tfsdk:"tenant_name"`
}

func sweepTenants(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := sesv2.ListTenantsInput{}
	conn := client.SESV2Client(ctx)
	var sweepResources []sweep.Sweepable

	pages := sesv2.NewListTenantsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Tenants {
			sweepResources = append(
				sweepResources,
				sweepfw.NewSweepResource(
					newResourceTenant, client,
					sweepfw.NewAttribute(names.AttrID, aws.ToString(v.TenantId))),
			)
		}
	}

	return sweepResources, nil
}
