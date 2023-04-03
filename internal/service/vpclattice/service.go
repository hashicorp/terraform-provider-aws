package vpclattice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"

	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource
// @Tags(identifierAttribute="arn")
func newResourceService(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceService{}, nil
}

type resourceServiceData struct {
	ARN              types.String   `tfsdk:"arn"`
	AuthType         types.String   `tfsdk:"auth_type"`
	CertificateARN   types.String   `tfsdk:"certificate_arn"`
	CustomDomainName types.String   `tfsdk:"custom_domain_name"`
	DnsEntry         types.List     `tfsdk:"dns_entry"`
	ID               types.String   `tfsdk:"id"`
	Name             types.String   `tfsdk:"name"`
	Status           types.String   `tfsdk:"status"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

const (
	ResNameService       = "Service"
	serviceCreateTimeout = 30 * time.Minute
	serviceDeleteTimeout = 30 * time.Minute
)

type resourceService struct {
	framework.ResourceWithConfigure
}

func (r *resourceService) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_vpclattice_service"
}

func (r *resourceService) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"auth_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Type of IAM policy. Either `NONE` or `AWS_IAM`",
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.AuthType](),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"certificate_arn": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Amazon Resource Name (ARN) of the certificate",
			},
			"custom_domain_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Custom domain name of the service",
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 255),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the service. The name must be unique within the account. The valid characters are a-z, 0-9, and hyphens (-). You can't use a hyphen as the first or last character, or immediately after another hyphen.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 40),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Status of the service. If the status is `CREATE_FAILED`, you will have to delete and recreate the service.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"dns_entry": schema.ListNestedBlock{
				MarkdownDescription: "Public DNS name of the service",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"domain_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Domain name of the service",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"hosted_zone_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "ID of the hosted zone",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: false,
				Delete: true,
			}),
		},
	}
}

func (r *resourceService) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, state resourceServiceData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, serviceCreateTimeout) // nosemgrep:ci.semgrep.migrate.direct-CRUD-calls

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	conn := r.Meta().VPCLatticeClient()

	in := &vpclattice.CreateServiceInput{
		ClientToken: aws.String(id.UniqueId()),
		Name:        plan.Name.ValueStringPointer(),
	}

	if !plan.AuthType.IsNull() {
		in.AuthType = awstypes.AuthType(plan.AuthType.ValueString())
	}

	if !plan.CertificateARN.IsNull() {
		in.CertificateArn = plan.CertificateARN.ValueStringPointer()
	}

	if !plan.CustomDomainName.IsNull() {
		in.CustomDomainName = plan.CustomDomainName.ValueStringPointer()
	}

	out, err := conn.CreateService(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionCreating, ResNameService, plan.Name.String(), nil),
			err.Error(),
		)
		return
	}

	if _, err := waitServiceCreated(ctx, conn, *out.Id, createTimeout); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionWaitingForCreation, ResNameService, plan.Name.String(), nil),
			err.Error(),
		)
		return
	}

	state = plan
	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.AuthType = flex.StringToFramework(ctx, (*string)(&out.AuthType))
	state.CertificateARN = flex.StringToFramework(ctx, out.CertificateArn)
	state.CustomDomainName = flex.StringToFramework(ctx, out.CustomDomainName)
	// state.DnsEntry = flattenDNSEntry(ctx, out.DnsEntry)
	state.ID = flex.StringToFramework(ctx, out.Id)
	state.Name = flex.StringToFramework(ctx, out.Name)
	state.Status = flex.StringToFramework(ctx, (*string)(&out.Status))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceService) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().VPCLatticeClient()

	var state resourceServiceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findServiceByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.AuthType = flex.StringToFramework(ctx, (*string)(&out.AuthType))
	state.CertificateARN = flex.StringToFramework(ctx, out.CertificateArn)
	state.CustomDomainName = flex.StringToFramework(ctx, out.CustomDomainName)
	// state.DnsEntry = flattenDNSEntry(ctx, out.DnsEntry)
	state.ID = flex.StringToFramework(ctx, out.Id)
	state.Name = flex.StringToFramework(ctx, out.Name)
	state.Status = flex.StringToFramework(ctx, (*string)(&out.Status))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceService) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().VPCLatticeClient()

	var plan, state resourceServiceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.AuthType.Equal(state.AuthType) ||
		!plan.CertificateARN.Equal(state.CertificateARN) {
		input := &vpclattice.UpdateServiceInput{
			ServiceIdentifier: plan.ID.ValueStringPointer(),
		}

		if !plan.AuthType.Equal(state.AuthType) {
			input.AuthType = awstypes.AuthType(plan.AuthType.ValueString())
		}

		if !plan.CertificateARN.Equal(state.CertificateARN) {
			input.CertificateArn = plan.CertificateARN.ValueStringPointer()
		}

		out, err := conn.UpdateService(ctx, input)

		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("updating Security Policy (%s)", plan.Name.ValueString()), err.Error())
			return
		}

		state.AuthType = flex.StringToFramework(ctx, (*string)(&out.AuthType))
		state.CertificateARN = flex.StringToFramework(ctx, out.CertificateArn)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceService) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().VPCLatticeClient()

	var state resourceServiceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, serviceDeleteTimeout) // nosemgrep:ci.semgrep.migrate.direct-CRUD-calls

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	_, err := conn.DeleteService(ctx, &vpclattice.DeleteServiceInput{
		ServiceIdentifier: flex.StringFromFramework(ctx, state.ID),
	})
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionDeleting, ResNameService, state.Name.String(), nil),
			err.Error(),
		)
	}

	if _, err := waitServiceDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout); err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VPCLattice, create.ErrActionWaitingForDeletion, ResNameService, state.Name.String(), nil),
			err.Error(),
		)
		return
	}
}

func (r *resourceService) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

}

var dnsEntryAttrs = map[string]attr.Type{
	"domain_name":    types.StringType,
	"hosted_zone_id": types.StringType,
}

func flattenDNSEntry(ctx context.Context, dns *awstypes.DnsEntry) types.List {
	elemType := types.ObjectType{AttrTypes: dnsEntryAttrs}

	if dns == nil {
		return types.ListNull(elemType)
	}

	attrs := map[string]attr.Value{}
	attrs["domain_name"] = flex.StringToFramework(ctx, dns.DomainName)
	attrs["hosted_zone_id"] = flex.StringToFramework(ctx, dns.HostedZoneId)

	vals := types.ObjectValueMust(dnsEntryAttrs, attrs)

	return types.ListValueMust(elemType, []attr.Value{vals})
}
