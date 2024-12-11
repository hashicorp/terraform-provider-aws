// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53domains

import (
	"context"
	"strings"
	"time"

	awstypes "github.com/aws/aws-sdk-go-v2/service/route53domains/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Domain")
// @Tags(identifierAttribute="id")
func newDomainResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &domainResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type domainResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithImportByID
}

func (*domainResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_route53domains_domain"
}

func (r *domainResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"abuse_contact_email": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"abuse_contact_phone": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"admin_privacy": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"auto_renew": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"billing_privacy": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"creation_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDomainName: schema.StringAttribute{
				CustomType: fwtypes.CaseInsensitiveStringType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"duration_in_years": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.Between(1, 10),
				},
			},
			"expiration_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"name_server": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[nameserverModel](ctx),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(6),
				},
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[nameserverModel](ctx),
				},
			},
			"registrant_privacy": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"registrar_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"registrar_url": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"reseller": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status_list": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Computed:    true,
				ElementType: types.StringType,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"tech_privacy": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"transfer_lock": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"updated_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"whois_server": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"admin_contact":      contactDetailBlock(ctx, true),
			"billing_contact":    contactDetailBlock(ctx, false),
			"registrant_contact": contactDetailBlock(ctx, true),
			"tech_contact":       contactDetailBlock(ctx, true),
		},
	}
}

func contactDetailBlock(ctx context.Context, required bool) schema.Block {
	block := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[contactDetailModel](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"address_line_1": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(255),
					},
				},
				"address_line_2": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(255),
					},
				},
				"city": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(255),
					},
				},
				"contact_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.ContactType](),
					Optional:   true,
				},
				"country_code": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.CountryCode](),
					Optional:   true,
				},
				names.AttrEmail: schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(254),
					},
				},
				"extra_params": schema.MapAttribute{
					CustomType:  fwtypes.MapOfStringType,
					Optional:    true,
					ElementType: types.StringType,
				},
				"fax": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(30),
					},
				},
				"first_name": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(255),
					},
				},
				"last_name": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(255),
					},
				},
				"organization_name": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(255),
					},
				},
				"phone_number": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(30),
					},
				},
				"state": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(255),
					},
				},
				"zip_code": schema.StringAttribute{
					Optional: true,
					Validators: []validator.String{
						stringvalidator.LengthAtMost(255),
					},
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
	if required {
		block.Validators = append(block.Validators, listvalidator.IsRequired(), listvalidator.SizeAtLeast(1))
	}

	return block
}

func (r *domainResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data domainResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	domainName := fwflex.StringValueFromFramework(ctx, data.DomainName)

	// Set values for unknowns.
	data.ID = fwflex.StringValueToFramework(ctx, strings.ToLower(domainName))

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *domainResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data domainResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().Route53DomainsClient(ctx)

	// Set attributes for import.

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *domainResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new domainResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *domainResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data domainResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
}

func (r *domainResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

type domainResourceModel struct {
	AbuseContactEmail types.String                                        `tfsdk:"abuse_contact_email"`
	AbuseContactPhone types.String                                        `tfsdk:"abuse_contact_phone"`
	AdminContact      fwtypes.ListNestedObjectValueOf[contactDetailModel] `tfsdk:"admin_contact"`
	AdminPrivacy      types.Bool                                          `tfsdk:"admin_privacy"`
	AutoRenew         types.Bool                                          `tfsdk:"auto_renew"`
	BillingContact    fwtypes.ListNestedObjectValueOf[contactDetailModel] `tfsdk:"billing_contact"`
	BillingPrivacy    types.Bool                                          `tfsdk:"billing_privacy"`
	CreationDate      timetypes.RFC3339                                   `tfsdk:"creation_date"`
	DomainName        fwtypes.CaseInsensitiveString                       `tfsdk:"domain_name"`
	DurationInYears   types.Int64                                         `tfsdk:"duration_in_years"`
	ExpirationDate    timetypes.RFC3339                                   `tfsdk:"expiration_date"`
	ID                types.String                                        `tfsdk:"id"`
	NameServers       fwtypes.ListNestedObjectValueOf[nameserverModel]    `tfsdk:"name_server"`
	RegistrantContact fwtypes.ListNestedObjectValueOf[contactDetailModel] `tfsdk:"registrant_contact"`
	RegistrantPrivacy types.Bool                                          `tfsdk:"registrant_privacy"`
	RegistrarName     types.String                                        `tfsdk:"registrar_name"`
	RegistrarURL      types.String                                        `tfsdk:"registrar_url"`
	Reseller          types.String                                        `tfsdk:"reseller"`
	StatusList        fwtypes.ListOfString                                `tfsdk:"status_list"`
	TechContact       fwtypes.ListNestedObjectValueOf[contactDetailModel] `tfsdk:"tech_contact"`
	TechPrivacy       types.Bool                                          `tfsdk:"tech_privacy"`
	TransferLock      types.Bool                                          `tfsdk:"transfer_lock"`
	UpdatedDate       timetypes.RFC3339                                   `tfsdk:"updated_date"`
	WhoIsServer       types.String                                        `tfsdk:"whois_server"`
}

func (data *domainResourceModel) InitFromID() error {
	data.DomainName = fwtypes.CaseInsensitiveStringValue(data.ID.ValueString())

	return nil
}

type contactDetailModel struct {
	AddressLine1     types.String                             `tfsdk:"address_line_1"`
	AddressLine2     types.String                             `tfsdk:"address_line_2"`
	City             types.String                             `tfsdk:"city"`
	ContactType      fwtypes.StringEnum[awstypes.ContactType] `tfsdk:"contact_type"`
	CountryCode      fwtypes.StringEnum[awstypes.CountryCode] `tfsdk:"country_code"`
	Email            types.String                             `tfsdk:"email"`
	ExtraParams      fwtypes.MapOfString                      `tfsdk:"extra_params"`
	Fax              types.String                             `tfsdk:"fax"`
	FirstName        types.String                             `tfsdk:"first_name"`
	LastName         types.String                             `tfsdk:"last_name"`
	OrganizationName types.String                             `tfsdk:"organization_name"`
	PhoneNumber      types.String                             `tfsdk:"phone_number"`
	State            types.String                             `tfsdk:"state"`
	ZipCode          types.String                             `tfsdk:"zip_code"`
}

type nameserverModel struct {
	GlueIPs fwtypes.SetOfString `tfsdk:"glue_ips"`
	Name    types.String        `tfsdk:"name"`
}
