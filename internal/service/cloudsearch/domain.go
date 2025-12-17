// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudsearch

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudsearch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudsearch/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudsearch_domain", name="Domain")
func newDomainResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &domainResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(20 * time.Minute)

	return r, nil
}

type domainResource struct {
	framework.ResourceWithModel[domainResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

var (
	indexNameRegex = regexache.MustCompile(`^(\*?[a-z][0-9a-z_]{2,63}|[a-z][0-9a-z_]{0,63}\*?)$`)
	nameRegex      = regexache.MustCompile(`^[a-z]([0-9a-z-]){2,27}$`)
)

func (r *domainResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	indexFieldSetOptions := []fwtypes.NestedObjectOfOption[indexFieldModel]{
		fwtypes.WithSemanticEqualityFunc(indexFieldSemanticEquality),
	}

	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"document_service_endpoint": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"multi_az": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(nameRegex, "Search domain names must start with a lowercase letter (a-z) and be at least 3 and no more than 28 lower-case letters, digits or hyphens"),
				},
			},
			"search_service_endpoint": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			// endpoint_options is Optional+Computed using ListAttribute (not ListNestedBlock)
			// to support true optional+computed semantics. This preserves drift detection
			// and allows users to reference computed values without explicitly configuring the block.
			// See: https://github.com/hashicorp/terraform-plugin-framework/issues/883
			"endpoint_options": framework.ResourceOptionalComputedListOfObjectsAttribute[endpointOptionsModel](ctx, 1, nil, listplanmodifier.UseStateForUnknown()),
			// scaling_parameters is Optional+Computed using ListAttribute (not ListNestedBlock)
			// for the same reasons as endpoint_options.
			"scaling_parameters": framework.ResourceOptionalComputedListOfObjectsAttribute[scalingParametersModel](ctx, 1, nil, listplanmodifier.UseStateForUnknown()),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"index_field": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[indexFieldModel](ctx, indexFieldSetOptions...),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"analysis_scheme": schema.StringAttribute{
							Optional: true,
						},
						names.AttrDefaultValue: schema.StringAttribute{
							Optional: true,
						},
						"facet": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"highlight": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						names.AttrName: schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(indexNameRegex, ""),
							},
						},
						"return": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"search": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"sort": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"source_fields": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(regexache.MustCompile(`^[^s]|s[^c]|sc[^o]|sco[^r]|scor[^e]`), "Cannot be set to reserved field score"),
							},
						},
						names.AttrType: schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.IndexFieldType](),
						},
					},
				},
			},
		},
	}
}

func (r *domainResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data domainResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Read config separately to determine what the user actually configured.
	// This is important because plan modifiers may create synthetic elements
	// (e.g., for endpoint_options and scaling_parameters) that shouldn't trigger
	// API calls unless the user explicitly configured them.
	var config domainResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudSearchClient(ctx)

	name := data.Name.ValueString()
	input := &cloudsearch.CreateDomainInput{
		DomainName: aws.String(name),
	}

	_, err := conn.CreateDomain(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating CloudSearch Domain (%s)", name), err.Error())
		return
	}

	data.ID = types.StringValue(name)

	// Update scaling parameters only if user explicitly configured them (check config, not plan).
	// The plan may contain synthetic elements from plan modifiers.
	if !config.ScalingParameters.IsNull() && len(config.ScalingParameters.Elements()) > 0 {
		scalingParams, diags := data.ScalingParameters.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		if scalingParams != nil {
			scalingInput := &cloudsearch.UpdateScalingParametersInput{
				DomainName:        aws.String(name),
				ScalingParameters: expandScalingParametersModel(scalingParams),
			}

			_, err := conn.UpdateScalingParameters(ctx, scalingInput)
			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("updating CloudSearch Domain (%s) scaling parameters", name), err.Error())
				return
			}
		}
	}

	// Update multi_az if specified (check config, not plan)
	if !config.MultiAZ.IsNull() {
		multiAZInput := &cloudsearch.UpdateAvailabilityOptionsInput{
			DomainName: aws.String(name),
			MultiAZ:    aws.Bool(data.MultiAZ.ValueBool()),
		}

		_, err := conn.UpdateAvailabilityOptions(ctx, multiAZInput)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating CloudSearch Domain (%s) availability options", name), err.Error())
			return
		}
	}

	// Update endpoint options only if user explicitly configured them (check config, not plan).
	// The plan may contain synthetic elements from plan modifiers.
	if !config.EndpointOptions.IsNull() && len(config.EndpointOptions.Elements()) > 0 {
		endpointOpts, diags := data.EndpointOptions.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		if endpointOpts != nil {
			endpointInput := &cloudsearch.UpdateDomainEndpointOptionsInput{
				DomainName:            aws.String(name),
				DomainEndpointOptions: expandEndpointOptionsModel(endpointOpts),
			}

			_, err := conn.UpdateDomainEndpointOptions(ctx, endpointInput)
			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("updating CloudSearch Domain (%s) endpoint options", name), err.Error())
				return
			}
		}
	}

	// Define index fields if specified
	if !data.IndexFields.IsNull() && !data.IndexFields.IsUnknown() {
		indexFields, diags := data.IndexFields.ToSlice(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		if len(indexFields) > 0 {
			for _, field := range indexFields {
				apiField, sourceFieldsConfigured, err := expandIndexFieldModel(field)
				if err != nil {
					response.Diagnostics.AddError(fmt.Sprintf("expanding index field (%s)", field.Name.ValueString()), err.Error())
					return
				}

				defineInput := &cloudsearch.DefineIndexFieldInput{
					DomainName: aws.String(name),
					IndexField: apiField,
				}

				_, err = conn.DefineIndexField(ctx, defineInput)
				if err != nil {
					response.Diagnostics.AddError(fmt.Sprintf("defining CloudSearch Domain (%s) index field (%s)", name, field.Name.ValueString()), err.Error())
					return
				}

				if sourceFieldsConfigured {
					tflog.Warn(ctx, "source_fields is configured for index field; if this is a new field, ensure the source field(s) exist before this field", map[string]any{
						"index_field": field.Name.ValueString(),
					})
				}
			}

			// Trigger indexing
			indexInput := &cloudsearch.IndexDocumentsInput{
				DomainName: aws.String(name),
			}

			_, err := conn.IndexDocuments(ctx, indexInput)
			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("indexing CloudSearch Domain (%s) documents", name), err.Error())
				return
			}
		}
	}

	// Wait for domain to become active
	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	if _, err := waitDomainActive(ctx, conn, name, createTimeout); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for CloudSearch Domain (%s) create", name), err.Error())
		return
	}

	// Read the domain to get computed values
	response.Diagnostics.Append(r.readDomain(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *domainResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data domainResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(r.readDomain(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *domainResource) readDomain(ctx context.Context, data *domainResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := r.Meta().CloudSearchClient(ctx)

	domainName := data.ID.ValueString()

	domain, err := findDomainByName(ctx, conn, domainName)
	if retry.NotFound(err) {
		diags.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		return diags
	}
	if err != nil {
		diags.AddError(fmt.Sprintf("reading CloudSearch Domain (%s)", domainName), err.Error())
		return diags
	}

	data.ARN = fwflex.StringToFramework(ctx, domain.ARN)
	if domain.DocService != nil {
		data.DocumentServiceEndpoint = fwflex.StringToFramework(ctx, domain.DocService.Endpoint)
	} else {
		data.DocumentServiceEndpoint = types.StringNull()
	}
	data.DomainID = fwflex.StringToFramework(ctx, domain.DomainId)
	data.Name = fwflex.StringToFramework(ctx, domain.DomainName)
	if domain.SearchService != nil {
		data.SearchServiceEndpoint = fwflex.StringToFramework(ctx, domain.SearchService.Endpoint)
	} else {
		data.SearchServiceEndpoint = types.StringNull()
	}

	// Read availability options
	availabilityOptionStatus, err := findAvailabilityOptionsStatusByName(ctx, conn, domainName)
	if err != nil {
		diags.AddError(fmt.Sprintf("reading CloudSearch Domain (%s) availability options", domainName), err.Error())
		return diags
	}
	data.MultiAZ = types.BoolValue(availabilityOptionStatus.Options)

	// Read endpoint options - always populate from API since this is now a true
	// Optional+Computed ListAttribute (not ListNestedBlock). The UseStateForUnknown
	// plan modifier handles preserving state during planning.
	endpointOptions, err := findDomainEndpointOptionsByName(ctx, conn, domainName)
	if err != nil {
		diags.AddError(fmt.Sprintf("reading CloudSearch Domain (%s) endpoint options", domainName), err.Error())
		return diags
	}
	data.EndpointOptions = flattenEndpointOptionsModel(ctx, endpointOptions)

	// Read scaling parameters - always populate from API for the same reason.
	scalingParameters, err := findScalingParametersByName(ctx, conn, domainName)
	if err != nil {
		diags.AddError(fmt.Sprintf("reading CloudSearch Domain (%s) scaling parameters", domainName), err.Error())
		return diags
	}
	data.ScalingParameters = flattenScalingParametersModel(ctx, scalingParameters)

	// Read index fields
	indexInput := &cloudsearch.DescribeIndexFieldsInput{
		DomainName: aws.String(domainName),
	}
	indexResults, err := conn.DescribeIndexFields(ctx, indexInput)
	if err != nil {
		diags.AddError(fmt.Sprintf("reading CloudSearch Domain (%s) index fields", domainName), err.Error())
		return diags
	}

	indexFields, d := flattenIndexFieldModels(ctx, indexResults.IndexFields)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	data.IndexFields = indexFields

	return diags
}

func (r *domainResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new domainResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Read config to determine what the user actually configured.
	// This is important for detecting when a user removes a block from config
	// (which should reset to defaults) vs when they never configured it.
	var config domainResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudSearchClient(ctx)
	domainName := new.ID.ValueString()
	requiresIndexDocuments := false

	// Update scaling parameters only if user configured them or is removing them.
	// Check config to determine user intent, not plan (which may have synthetic elements).
	configHasScalingParams := !config.ScalingParameters.IsNull() && len(config.ScalingParameters.Elements()) > 0
	if configHasScalingParams {
		// User configured scaling_parameters - apply their values
		scalingInput := &cloudsearch.UpdateScalingParametersInput{
			DomainName: aws.String(domainName),
		}

		scalingParams, diags := new.ScalingParameters.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		if scalingParams != nil {
			scalingInput.ScalingParameters = expandScalingParametersModel(scalingParams)
		} else {
			scalingInput.ScalingParameters = &awstypes.ScalingParameters{}
		}

		output, err := conn.UpdateScalingParameters(ctx, scalingInput)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating CloudSearch Domain (%s) scaling parameters", domainName), err.Error())
			return
		}

		if output != nil && output.ScalingParameters != nil && output.ScalingParameters.Status != nil && output.ScalingParameters.Status.State == awstypes.OptionStateRequiresIndexDocuments {
			requiresIndexDocuments = true
		}
	}
	// Note: If user removes scaling_parameters from config, we don't reset to defaults.
	// This matches the behavior where unconfigured blocks are just read from API.
	// The SDKv2 behavior of resetting to defaults when block is removed would require
	// tracking "was this previously managed" in private state.

	// Update multi_az only if user configured it
	if !config.MultiAZ.IsNull() && !new.MultiAZ.Equal(old.MultiAZ) {
		multiAZInput := &cloudsearch.UpdateAvailabilityOptionsInput{
			DomainName: aws.String(domainName),
			MultiAZ:    aws.Bool(new.MultiAZ.ValueBool()),
		}

		output, err := conn.UpdateAvailabilityOptions(ctx, multiAZInput)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating CloudSearch Domain (%s) availability options", domainName), err.Error())
			return
		}

		if output != nil && output.AvailabilityOptions != nil && output.AvailabilityOptions.Status != nil && output.AvailabilityOptions.Status.State == awstypes.OptionStateRequiresIndexDocuments {
			requiresIndexDocuments = true
		}
	}

	// Update endpoint options only if user configured them.
	// Check config to determine user intent, not plan (which may have synthetic elements).
	configHasEndpointOpts := !config.EndpointOptions.IsNull() && len(config.EndpointOptions.Elements()) > 0
	if configHasEndpointOpts {
		// User configured endpoint_options - apply their values
		endpointInput := &cloudsearch.UpdateDomainEndpointOptionsInput{
			DomainName: aws.String(domainName),
		}

		endpointOpts, diags := new.EndpointOptions.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		if endpointOpts != nil {
			endpointInput.DomainEndpointOptions = expandEndpointOptionsModel(endpointOpts)
		} else {
			endpointInput.DomainEndpointOptions = &awstypes.DomainEndpointOptions{}
		}

		output, err := conn.UpdateDomainEndpointOptions(ctx, endpointInput)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating CloudSearch Domain (%s) endpoint options", domainName), err.Error())
			return
		}

		if output != nil && output.DomainEndpointOptions != nil && output.DomainEndpointOptions.Status != nil && output.DomainEndpointOptions.Status.State == awstypes.OptionStateRequiresIndexDocuments {
			requiresIndexDocuments = true
		}
	}

	// Update index fields
	if !new.IndexFields.Equal(old.IndexFields) {
		oldFields, diags := old.IndexFields.ToSlice(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		newFields, diags := new.IndexFields.ToSlice(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		// Build maps by name for comparison
		oldByName := make(map[string]*indexFieldModel)
		for _, f := range oldFields {
			oldByName[f.Name.ValueString()] = f
		}

		newByName := make(map[string]*indexFieldModel)
		for _, f := range newFields {
			newByName[f.Name.ValueString()] = f
		}

		// Delete fields that are in old but not in new
		for name := range oldByName {
			if _, exists := newByName[name]; !exists {
				deleteInput := &cloudsearch.DeleteIndexFieldInput{
					DomainName:     aws.String(domainName),
					IndexFieldName: aws.String(name),
				}

				_, err := conn.DeleteIndexField(ctx, deleteInput)
				if err != nil {
					response.Diagnostics.AddError(fmt.Sprintf("deleting CloudSearch Domain (%s) index field (%s)", domainName, name), err.Error())
					return
				}

				requiresIndexDocuments = true
			}
		}

		// Add or update fields that are in new
		for name, newField := range newByName {
			oldField, exists := oldByName[name]

			// If field doesn't exist or has changed, define it
			if !exists || !indexFieldsEqual(oldField, newField) {
				apiField, sourceFieldsConfigured, err := expandIndexFieldModel(newField)
				if err != nil {
					response.Diagnostics.AddError(fmt.Sprintf("expanding index field (%s)", name), err.Error())
					return
				}

				defineInput := &cloudsearch.DefineIndexFieldInput{
					DomainName: aws.String(domainName),
					IndexField: apiField,
				}

				_, err = conn.DefineIndexField(ctx, defineInput)
				if err != nil {
					response.Diagnostics.AddError(fmt.Sprintf("defining CloudSearch Domain (%s) index field (%s)", domainName, name), err.Error())
					return
				}

				if sourceFieldsConfigured {
					tflog.Warn(ctx, "source_fields is configured for index field; if this is a new field, ensure the source field(s) exist before this field", map[string]any{
						"index_field": name,
					})
				}

				requiresIndexDocuments = true
			}
		}
	}

	if requiresIndexDocuments {
		indexInput := &cloudsearch.IndexDocumentsInput{
			DomainName: aws.String(domainName),
		}

		_, err := conn.IndexDocuments(ctx, indexInput)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("indexing CloudSearch Domain (%s) documents", domainName), err.Error())
			return
		}
	}

	// Wait for domain to become active
	updateTimeout := r.UpdateTimeout(ctx, new.Timeouts)
	if _, err := waitDomainActive(ctx, conn, domainName, updateTimeout); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for CloudSearch Domain (%s) update", domainName), err.Error())
		return
	}

	// Read the domain to get computed values
	response.Diagnostics.Append(r.readDomain(ctx, &new)...)
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

	conn := r.Meta().CloudSearchClient(ctx)
	domainName := data.ID.ValueString()

	tflog.Debug(ctx, "Deleting CloudSearch Domain", map[string]any{
		"domain_name": domainName,
	})

	input := &cloudsearch.DeleteDomainInput{
		DomainName: aws.String(domainName),
	}

	_, err := conn.DeleteDomain(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting CloudSearch Domain (%s)", domainName), err.Error())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, data.Timeouts)
	if _, err := waitDomainDeleted(ctx, conn, domainName, deleteTimeout); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for CloudSearch Domain (%s) delete", domainName), err.Error())
		return
	}
}

func (r *domainResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), request, response)
}

// Model types

type domainResourceModel struct {
	framework.WithRegionModel
	ARN                     types.String                                            `tfsdk:"arn"`
	DocumentServiceEndpoint types.String                                            `tfsdk:"document_service_endpoint"`
	DomainID                types.String                                            `tfsdk:"domain_id"`
	EndpointOptions         fwtypes.ListNestedObjectValueOf[endpointOptionsModel]   `tfsdk:"endpoint_options"`
	ID                      types.String                                            `tfsdk:"id"`
	IndexFields             fwtypes.SetNestedObjectValueOf[indexFieldModel]         `tfsdk:"index_field"`
	MultiAZ                 types.Bool                                              `tfsdk:"multi_az"`
	Name                    types.String                                            `tfsdk:"name"`
	ScalingParameters       fwtypes.ListNestedObjectValueOf[scalingParametersModel] `tfsdk:"scaling_parameters"`
	SearchServiceEndpoint   types.String                                            `tfsdk:"search_service_endpoint"`
	Timeouts                timeouts.Value                                          `tfsdk:"timeouts"`
}

type endpointOptionsModel struct {
	EnforceHTTPS      types.Bool                                     `tfsdk:"enforce_https"`
	TLSSecurityPolicy fwtypes.StringEnum[awstypes.TLSSecurityPolicy] `tfsdk:"tls_security_policy"`
}

type indexFieldModel struct {
	AnalysisScheme types.String                                `tfsdk:"analysis_scheme"`
	DefaultValue   types.String                                `tfsdk:"default_value"`
	Facet          types.Bool                                  `tfsdk:"facet"`
	Highlight      types.Bool                                  `tfsdk:"highlight"`
	Name           types.String                                `tfsdk:"name"`
	Return         types.Bool                                  `tfsdk:"return"`
	Search         types.Bool                                  `tfsdk:"search"`
	Sort           types.Bool                                  `tfsdk:"sort"`
	SourceFields   types.String                                `tfsdk:"source_fields"`
	Type           fwtypes.StringEnum[awstypes.IndexFieldType] `tfsdk:"type"`
}

type scalingParametersModel struct {
	DesiredInstanceType     fwtypes.StringEnum[awstypes.PartitionInstanceType] `tfsdk:"desired_instance_type"`
	DesiredPartitionCount   types.Int64                                        `tfsdk:"desired_partition_count"`
	DesiredReplicationCount types.Int64                                        `tfsdk:"desired_replication_count"`
}

// Semantic equality function for index_field set
// This matches elements by name only for identity purposes
func indexFieldSemanticEquality(ctx context.Context, oldValue, newValue fwtypes.NestedCollectionValue[indexFieldModel]) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oldSlice, d := oldValue.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	newSlice, d := newValue.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	if len(oldSlice) != len(newSlice) {
		return false, diags
	}

	// Build map by name for stable matching
	oldByName := make(map[string]*indexFieldModel)
	for _, f := range oldSlice {
		oldByName[f.Name.ValueString()] = f
	}

	// Check that all new fields exist in old and are equal
	for _, newField := range newSlice {
		oldField, exists := oldByName[newField.Name.ValueString()]
		if !exists {
			return false, diags
		}

		if !indexFieldsSemanticEqual(oldField, newField) {
			return false, diags
		}
	}

	return true, diags
}

// indexFieldsSemanticEqual compares two index fields for equality.
// This is used for set membership comparison to prevent spurious remove/add diffs.
// With schema defaults properly set (Default: false for booleans), we can use
// direct comparison instead of treating null/false as equivalent.
func indexFieldsSemanticEqual(a, b *indexFieldModel) bool {
	return a.Name.Equal(b.Name) &&
		a.Type.Equal(b.Type) &&
		a.Facet.Equal(b.Facet) &&
		a.Highlight.Equal(b.Highlight) &&
		a.Return.Equal(b.Return) &&
		a.Search.Equal(b.Search) &&
		a.Sort.Equal(b.Sort) &&
		a.AnalysisScheme.Equal(b.AnalysisScheme) &&
		a.DefaultValue.Equal(b.DefaultValue) &&
		a.SourceFields.Equal(b.SourceFields)
}

// indexFieldsEqual compares two index fields for exact equality (used in Update)
func indexFieldsEqual(a, b *indexFieldModel) bool {
	return a.Name.Equal(b.Name) &&
		a.Type.Equal(b.Type) &&
		a.Facet.Equal(b.Facet) &&
		a.Highlight.Equal(b.Highlight) &&
		a.Return.Equal(b.Return) &&
		a.Search.Equal(b.Search) &&
		a.Sort.Equal(b.Sort) &&
		a.AnalysisScheme.Equal(b.AnalysisScheme) &&
		a.DefaultValue.Equal(b.DefaultValue) &&
		a.SourceFields.Equal(b.SourceFields)
}

// Expand functions - convert Framework models to AWS API types

func expandEndpointOptionsModel(m *endpointOptionsModel) *awstypes.DomainEndpointOptions {
	if m == nil {
		return nil
	}

	opts := &awstypes.DomainEndpointOptions{}

	if !m.EnforceHTTPS.IsNull() && !m.EnforceHTTPS.IsUnknown() {
		opts.EnforceHTTPS = aws.Bool(m.EnforceHTTPS.ValueBool())
	}

	if !m.TLSSecurityPolicy.IsNull() && !m.TLSSecurityPolicy.IsUnknown() {
		opts.TLSSecurityPolicy = m.TLSSecurityPolicy.ValueEnum()
	}

	return opts
}

func expandScalingParametersModel(m *scalingParametersModel) *awstypes.ScalingParameters {
	if m == nil {
		return nil
	}

	params := &awstypes.ScalingParameters{}

	if !m.DesiredInstanceType.IsNull() && !m.DesiredInstanceType.IsUnknown() {
		params.DesiredInstanceType = m.DesiredInstanceType.ValueEnum()
	}

	if !m.DesiredPartitionCount.IsNull() && !m.DesiredPartitionCount.IsUnknown() {
		params.DesiredPartitionCount = int32(m.DesiredPartitionCount.ValueInt64())
	}

	if !m.DesiredReplicationCount.IsNull() && !m.DesiredReplicationCount.IsUnknown() {
		params.DesiredReplicationCount = int32(m.DesiredReplicationCount.ValueInt64())
	}

	return params
}

func expandIndexFieldModel(m *indexFieldModel) (*awstypes.IndexField, bool, error) {
	if m == nil {
		return nil, false, nil
	}

	apiObject := &awstypes.IndexField{
		IndexFieldName: aws.String(m.Name.ValueString()),
		IndexFieldType: m.Type.ValueEnum(),
	}

	analysisScheme := ""
	if !m.AnalysisScheme.IsNull() && !m.AnalysisScheme.IsUnknown() {
		analysisScheme = m.AnalysisScheme.ValueString()
	}

	facetEnabled := false
	if !m.Facet.IsNull() && !m.Facet.IsUnknown() {
		facetEnabled = m.Facet.ValueBool()
	}

	highlightEnabled := false
	if !m.Highlight.IsNull() && !m.Highlight.IsUnknown() {
		highlightEnabled = m.Highlight.ValueBool()
	}

	returnEnabled := false
	if !m.Return.IsNull() && !m.Return.IsUnknown() {
		returnEnabled = m.Return.ValueBool()
	}

	searchEnabled := false
	if !m.Search.IsNull() && !m.Search.IsUnknown() {
		searchEnabled = m.Search.ValueBool()
	}

	sortEnabled := false
	if !m.Sort.IsNull() && !m.Sort.IsUnknown() {
		sortEnabled = m.Sort.ValueBool()
	}

	var sourceFieldsConfigured bool

	switch fieldType := apiObject.IndexFieldType; fieldType {
	case awstypes.IndexFieldTypeDate:
		options := &awstypes.DateOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
			SortEnabled:   aws.Bool(sortEnabled),
		}

		if !m.DefaultValue.IsNull() && !m.DefaultValue.IsUnknown() && m.DefaultValue.ValueString() != "" {
			options.DefaultValue = aws.String(m.DefaultValue.ValueString())
		}

		if !m.SourceFields.IsNull() && !m.SourceFields.IsUnknown() && m.SourceFields.ValueString() != "" {
			options.SourceField = aws.String(m.SourceFields.ValueString())
			sourceFieldsConfigured = true
		}

		apiObject.DateOptions = options

	case awstypes.IndexFieldTypeDateArray:
		options := &awstypes.DateArrayOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
		}

		if !m.DefaultValue.IsNull() && !m.DefaultValue.IsUnknown() && m.DefaultValue.ValueString() != "" {
			options.DefaultValue = aws.String(m.DefaultValue.ValueString())
		}

		if !m.SourceFields.IsNull() && !m.SourceFields.IsUnknown() && m.SourceFields.ValueString() != "" {
			options.SourceFields = aws.String(m.SourceFields.ValueString())
			sourceFieldsConfigured = true
		}

		apiObject.DateArrayOptions = options

	case awstypes.IndexFieldTypeDouble:
		options := &awstypes.DoubleOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
			SortEnabled:   aws.Bool(sortEnabled),
		}

		if !m.DefaultValue.IsNull() && !m.DefaultValue.IsUnknown() && m.DefaultValue.ValueString() != "" {
			v, err := strconv.ParseFloat(m.DefaultValue.ValueString(), 64)
			if err != nil {
				return nil, false, err
			}
			options.DefaultValue = aws.Float64(v)
		}

		if !m.SourceFields.IsNull() && !m.SourceFields.IsUnknown() && m.SourceFields.ValueString() != "" {
			options.SourceField = aws.String(m.SourceFields.ValueString())
			sourceFieldsConfigured = true
		}

		apiObject.DoubleOptions = options

	case awstypes.IndexFieldTypeDoubleArray:
		options := &awstypes.DoubleArrayOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
		}

		if !m.DefaultValue.IsNull() && !m.DefaultValue.IsUnknown() && m.DefaultValue.ValueString() != "" {
			v, err := strconv.ParseFloat(m.DefaultValue.ValueString(), 64)
			if err != nil {
				return nil, false, err
			}
			options.DefaultValue = aws.Float64(v)
		}

		if !m.SourceFields.IsNull() && !m.SourceFields.IsUnknown() && m.SourceFields.ValueString() != "" {
			options.SourceFields = aws.String(m.SourceFields.ValueString())
			sourceFieldsConfigured = true
		}

		apiObject.DoubleArrayOptions = options

	case awstypes.IndexFieldTypeInt:
		options := &awstypes.IntOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
			SortEnabled:   aws.Bool(sortEnabled),
		}

		if !m.DefaultValue.IsNull() && !m.DefaultValue.IsUnknown() && m.DefaultValue.ValueString() != "" {
			v, err := strconv.Atoi(m.DefaultValue.ValueString())
			if err != nil {
				return nil, false, err
			}
			options.DefaultValue = aws.Int64(int64(v))
		}

		if !m.SourceFields.IsNull() && !m.SourceFields.IsUnknown() && m.SourceFields.ValueString() != "" {
			options.SourceField = aws.String(m.SourceFields.ValueString())
			sourceFieldsConfigured = true
		}

		apiObject.IntOptions = options

	case awstypes.IndexFieldTypeIntArray:
		options := &awstypes.IntArrayOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
		}

		if !m.DefaultValue.IsNull() && !m.DefaultValue.IsUnknown() && m.DefaultValue.ValueString() != "" {
			v, err := strconv.Atoi(m.DefaultValue.ValueString())
			if err != nil {
				return nil, false, err
			}
			options.DefaultValue = aws.Int64(int64(v))
		}

		if !m.SourceFields.IsNull() && !m.SourceFields.IsUnknown() && m.SourceFields.ValueString() != "" {
			options.SourceFields = aws.String(m.SourceFields.ValueString())
			sourceFieldsConfigured = true
		}

		apiObject.IntArrayOptions = options

	case awstypes.IndexFieldTypeLatlon:
		options := &awstypes.LatLonOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
			SortEnabled:   aws.Bool(sortEnabled),
		}

		if !m.DefaultValue.IsNull() && !m.DefaultValue.IsUnknown() && m.DefaultValue.ValueString() != "" {
			options.DefaultValue = aws.String(m.DefaultValue.ValueString())
		}

		if !m.SourceFields.IsNull() && !m.SourceFields.IsUnknown() && m.SourceFields.ValueString() != "" {
			options.SourceField = aws.String(m.SourceFields.ValueString())
			sourceFieldsConfigured = true
		}

		apiObject.LatLonOptions = options

	case awstypes.IndexFieldTypeLiteral:
		options := &awstypes.LiteralOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
			SortEnabled:   aws.Bool(sortEnabled),
		}

		if !m.DefaultValue.IsNull() && !m.DefaultValue.IsUnknown() && m.DefaultValue.ValueString() != "" {
			options.DefaultValue = aws.String(m.DefaultValue.ValueString())
		}

		if !m.SourceFields.IsNull() && !m.SourceFields.IsUnknown() && m.SourceFields.ValueString() != "" {
			options.SourceField = aws.String(m.SourceFields.ValueString())
			sourceFieldsConfigured = true
		}

		apiObject.LiteralOptions = options

	case awstypes.IndexFieldTypeLiteralArray:
		options := &awstypes.LiteralArrayOptions{
			FacetEnabled:  aws.Bool(facetEnabled),
			ReturnEnabled: aws.Bool(returnEnabled),
			SearchEnabled: aws.Bool(searchEnabled),
		}

		if !m.DefaultValue.IsNull() && !m.DefaultValue.IsUnknown() && m.DefaultValue.ValueString() != "" {
			options.DefaultValue = aws.String(m.DefaultValue.ValueString())
		}

		if !m.SourceFields.IsNull() && !m.SourceFields.IsUnknown() && m.SourceFields.ValueString() != "" {
			options.SourceFields = aws.String(m.SourceFields.ValueString())
			sourceFieldsConfigured = true
		}

		apiObject.LiteralArrayOptions = options

	case awstypes.IndexFieldTypeText:
		options := &awstypes.TextOptions{
			HighlightEnabled: aws.Bool(highlightEnabled),
			ReturnEnabled:    aws.Bool(returnEnabled),
			SortEnabled:      aws.Bool(sortEnabled),
		}

		if analysisScheme != "" {
			options.AnalysisScheme = aws.String(analysisScheme)
		}

		if !m.DefaultValue.IsNull() && !m.DefaultValue.IsUnknown() && m.DefaultValue.ValueString() != "" {
			options.DefaultValue = aws.String(m.DefaultValue.ValueString())
		}

		if !m.SourceFields.IsNull() && !m.SourceFields.IsUnknown() && m.SourceFields.ValueString() != "" {
			options.SourceField = aws.String(m.SourceFields.ValueString())
			sourceFieldsConfigured = true
		}

		apiObject.TextOptions = options

	case awstypes.IndexFieldTypeTextArray:
		options := &awstypes.TextArrayOptions{
			HighlightEnabled: aws.Bool(highlightEnabled),
			ReturnEnabled:    aws.Bool(returnEnabled),
		}

		if analysisScheme != "" {
			options.AnalysisScheme = aws.String(analysisScheme)
		}

		if !m.DefaultValue.IsNull() && !m.DefaultValue.IsUnknown() && m.DefaultValue.ValueString() != "" {
			options.DefaultValue = aws.String(m.DefaultValue.ValueString())
		}

		if !m.SourceFields.IsNull() && !m.SourceFields.IsUnknown() && m.SourceFields.ValueString() != "" {
			options.SourceFields = aws.String(m.SourceFields.ValueString())
			sourceFieldsConfigured = true
		}

		apiObject.TextArrayOptions = options

	default:
		return nil, false, fmt.Errorf("unsupported index_field type: %s", fieldType)
	}

	return apiObject, sourceFieldsConfigured, nil
}

// Flatten functions - convert AWS API types to Framework models

func flattenEndpointOptionsModel(ctx context.Context, apiObject *awstypes.DomainEndpointOptions) fwtypes.ListNestedObjectValueOf[endpointOptionsModel] {
	if apiObject == nil {
		return fwtypes.NewListNestedObjectValueOfNull[endpointOptionsModel](ctx)
	}

	m := &endpointOptionsModel{
		EnforceHTTPS:      types.BoolValue(aws.ToBool(apiObject.EnforceHTTPS)),
		TLSSecurityPolicy: fwtypes.StringEnumValue(apiObject.TLSSecurityPolicy),
	}

	return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, m)
}

func flattenScalingParametersModel(ctx context.Context, apiObject *awstypes.ScalingParameters) fwtypes.ListNestedObjectValueOf[scalingParametersModel] {
	if apiObject == nil {
		return fwtypes.NewListNestedObjectValueOfNull[scalingParametersModel](ctx)
	}

	m := &scalingParametersModel{
		DesiredInstanceType:     fwtypes.StringEnumValue(apiObject.DesiredInstanceType),
		DesiredPartitionCount:   types.Int64Value(int64(apiObject.DesiredPartitionCount)),
		DesiredReplicationCount: types.Int64Value(int64(apiObject.DesiredReplicationCount)),
	}

	return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, m)
}

func flattenIndexFieldModels(ctx context.Context, apiObjects []awstypes.IndexFieldStatus) (fwtypes.SetNestedObjectValueOf[indexFieldModel], diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(apiObjects) == 0 {
		return fwtypes.NewSetNestedObjectValueOfNull[indexFieldModel](ctx), diags
	}

	var models []*indexFieldModel

	for _, apiObject := range apiObjects {
		m, err := flattenIndexFieldModel(apiObject)
		if err != nil {
			diags.AddError("flattening index field", err.Error())
			return fwtypes.NewSetNestedObjectValueOfNull[indexFieldModel](ctx), diags
		}
		if m != nil {
			models = append(models, m)
		}
	}

	// Use the semantic equality function when creating the set
	result, d := fwtypes.NewSetNestedObjectValueOfSlice(ctx, models, indexFieldSemanticEquality)
	diags.Append(d...)

	return result, diags
}

func flattenIndexFieldModel(apiObject awstypes.IndexFieldStatus) (*indexFieldModel, error) {
	if apiObject.Options == nil || apiObject.Status == nil {
		return nil, nil
	}

	// Don't read in any fields that are pending deletion
	if aws.ToBool(apiObject.Status.PendingDeletion) {
		return nil, nil
	}

	field := apiObject.Options
	// Initialize boolean fields to false to match SDKv2's Default: false behavior.
	// String fields are initialized to null since they don't have defaults.
	// This ensures consistency between schema defaults and flatten output.
	m := &indexFieldModel{
		Name:           types.StringValue(aws.ToString(field.IndexFieldName)),
		Type:           fwtypes.StringEnumValue(field.IndexFieldType),
		AnalysisScheme: types.StringNull(),
		DefaultValue:   types.StringNull(),
		Facet:          types.BoolValue(false),
		Highlight:      types.BoolValue(false),
		Return:         types.BoolValue(false),
		Search:         types.BoolValue(false),
		Sort:           types.BoolValue(false),
		SourceFields:   types.StringNull(),
	}

	switch fieldType := field.IndexFieldType; fieldType {
	case awstypes.IndexFieldTypeDate:
		options := field.DateOptions
		if options == nil {
			break
		}

		if v := options.DefaultValue; v != nil {
			m.DefaultValue = types.StringValue(aws.ToString(v))
		}

		if v := options.FacetEnabled; v != nil {
			m.Facet = types.BoolValue(aws.ToBool(v))
		}

		if v := options.ReturnEnabled; v != nil {
			m.Return = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SearchEnabled; v != nil {
			m.Search = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SortEnabled; v != nil {
			m.Sort = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SourceField; v != nil {
			m.SourceFields = types.StringValue(aws.ToString(v))
		}

	case awstypes.IndexFieldTypeDateArray:
		options := field.DateArrayOptions
		if options == nil {
			break
		}

		if v := options.DefaultValue; v != nil {
			m.DefaultValue = types.StringValue(aws.ToString(v))
		}

		if v := options.FacetEnabled; v != nil {
			m.Facet = types.BoolValue(aws.ToBool(v))
		}

		if v := options.ReturnEnabled; v != nil {
			m.Return = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SearchEnabled; v != nil {
			m.Search = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SourceFields; v != nil {
			m.SourceFields = types.StringValue(aws.ToString(v))
		}

	case awstypes.IndexFieldTypeDouble:
		options := field.DoubleOptions
		if options == nil {
			break
		}

		if v := options.DefaultValue; v != nil {
			m.DefaultValue = types.StringValue(strconv.FormatFloat(aws.ToFloat64(v), 'f', -1, 64))
		}

		if v := options.FacetEnabled; v != nil {
			m.Facet = types.BoolValue(aws.ToBool(v))
		}

		if v := options.ReturnEnabled; v != nil {
			m.Return = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SearchEnabled; v != nil {
			m.Search = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SortEnabled; v != nil {
			m.Sort = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SourceField; v != nil {
			m.SourceFields = types.StringValue(aws.ToString(v))
		}

	case awstypes.IndexFieldTypeDoubleArray:
		options := field.DoubleArrayOptions
		if options == nil {
			break
		}

		if v := options.DefaultValue; v != nil {
			m.DefaultValue = types.StringValue(strconv.FormatFloat(aws.ToFloat64(v), 'f', -1, 64))
		}

		if v := options.FacetEnabled; v != nil {
			m.Facet = types.BoolValue(aws.ToBool(v))
		}

		if v := options.ReturnEnabled; v != nil {
			m.Return = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SearchEnabled; v != nil {
			m.Search = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SourceFields; v != nil {
			m.SourceFields = types.StringValue(aws.ToString(v))
		}

	case awstypes.IndexFieldTypeInt:
		options := field.IntOptions
		if options == nil {
			break
		}

		if v := options.DefaultValue; v != nil {
			m.DefaultValue = types.StringValue(strconv.FormatInt(aws.ToInt64(v), 10))
		}

		if v := options.FacetEnabled; v != nil {
			m.Facet = types.BoolValue(aws.ToBool(v))
		}

		if v := options.ReturnEnabled; v != nil {
			m.Return = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SearchEnabled; v != nil {
			m.Search = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SortEnabled; v != nil {
			m.Sort = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SourceField; v != nil {
			m.SourceFields = types.StringValue(aws.ToString(v))
		}

	case awstypes.IndexFieldTypeIntArray:
		options := field.IntArrayOptions
		if options == nil {
			break
		}

		if v := options.DefaultValue; v != nil {
			m.DefaultValue = types.StringValue(strconv.FormatInt(aws.ToInt64(v), 10))
		}

		if v := options.FacetEnabled; v != nil {
			m.Facet = types.BoolValue(aws.ToBool(v))
		}

		if v := options.ReturnEnabled; v != nil {
			m.Return = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SearchEnabled; v != nil {
			m.Search = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SourceFields; v != nil {
			m.SourceFields = types.StringValue(aws.ToString(v))
		}

	case awstypes.IndexFieldTypeLatlon:
		options := field.LatLonOptions
		if options == nil {
			break
		}

		if v := options.DefaultValue; v != nil {
			m.DefaultValue = types.StringValue(aws.ToString(v))
		}

		if v := options.FacetEnabled; v != nil {
			m.Facet = types.BoolValue(aws.ToBool(v))
		}

		if v := options.ReturnEnabled; v != nil {
			m.Return = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SearchEnabled; v != nil {
			m.Search = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SortEnabled; v != nil {
			m.Sort = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SourceField; v != nil {
			m.SourceFields = types.StringValue(aws.ToString(v))
		}

	case awstypes.IndexFieldTypeLiteral:
		options := field.LiteralOptions
		if options == nil {
			break
		}

		if v := options.DefaultValue; v != nil {
			m.DefaultValue = types.StringValue(aws.ToString(v))
		}

		if v := options.FacetEnabled; v != nil {
			m.Facet = types.BoolValue(aws.ToBool(v))
		}

		if v := options.ReturnEnabled; v != nil {
			m.Return = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SearchEnabled; v != nil {
			m.Search = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SortEnabled; v != nil {
			m.Sort = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SourceField; v != nil {
			m.SourceFields = types.StringValue(aws.ToString(v))
		}

	case awstypes.IndexFieldTypeLiteralArray:
		options := field.LiteralArrayOptions
		if options == nil {
			break
		}

		if v := options.DefaultValue; v != nil {
			m.DefaultValue = types.StringValue(aws.ToString(v))
		}

		if v := options.FacetEnabled; v != nil {
			m.Facet = types.BoolValue(aws.ToBool(v))
		}

		if v := options.ReturnEnabled; v != nil {
			m.Return = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SearchEnabled; v != nil {
			m.Search = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SourceFields; v != nil {
			m.SourceFields = types.StringValue(aws.ToString(v))
		}

	case awstypes.IndexFieldTypeText:
		options := field.TextOptions
		if options == nil {
			break
		}

		if v := options.AnalysisScheme; v != nil {
			m.AnalysisScheme = types.StringValue(aws.ToString(v))
		}

		if v := options.DefaultValue; v != nil {
			m.DefaultValue = types.StringValue(aws.ToString(v))
		}

		if v := options.HighlightEnabled; v != nil {
			m.Highlight = types.BoolValue(aws.ToBool(v))
		}

		if v := options.ReturnEnabled; v != nil {
			m.Return = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SortEnabled; v != nil {
			m.Sort = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SourceField; v != nil {
			m.SourceFields = types.StringValue(aws.ToString(v))
		}

	case awstypes.IndexFieldTypeTextArray:
		options := field.TextArrayOptions
		if options == nil {
			break
		}

		if v := options.AnalysisScheme; v != nil {
			m.AnalysisScheme = types.StringValue(aws.ToString(v))
		}

		if v := options.DefaultValue; v != nil {
			m.DefaultValue = types.StringValue(aws.ToString(v))
		}

		if v := options.HighlightEnabled; v != nil {
			m.Highlight = types.BoolValue(aws.ToBool(v))
		}

		if v := options.ReturnEnabled; v != nil {
			m.Return = types.BoolValue(aws.ToBool(v))
		}

		if v := options.SourceFields; v != nil {
			m.SourceFields = types.StringValue(aws.ToString(v))
		}

	default:
		return nil, fmt.Errorf("unsupported index_field type: %s", fieldType)
	}

	return m, nil
}

// Helper functions - shared between Framework and SDKv2 implementations

func findDomainByName(ctx context.Context, conn *cloudsearch.Client, name string) (*awstypes.DomainStatus, error) {
	input := &cloudsearch.DescribeDomainsInput{
		DomainNames: []string{name},
	}

	output, err := conn.DescribeDomains(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfresource.AssertSingleValueResult(output.DomainStatusList)
}

func findAvailabilityOptionsStatusByName(ctx context.Context, conn *cloudsearch.Client, name string) (*awstypes.AvailabilityOptionsStatus, error) {
	input := &cloudsearch.DescribeAvailabilityOptionsInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DescribeAvailabilityOptions(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AvailabilityOptions == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AvailabilityOptions, nil
}

func findDomainEndpointOptionsByName(ctx context.Context, conn *cloudsearch.Client, name string) (*awstypes.DomainEndpointOptions, error) {
	output, err := findDomainEndpointOptionsStatusByName(ctx, conn, name)

	if err != nil {
		return nil, err
	}

	if output.Options == nil {
		return nil, tfresource.NewEmptyResultError(name)
	}

	return output.Options, nil
}

func findDomainEndpointOptionsStatusByName(ctx context.Context, conn *cloudsearch.Client, name string) (*awstypes.DomainEndpointOptionsStatus, error) {
	input := &cloudsearch.DescribeDomainEndpointOptionsInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DescribeDomainEndpointOptions(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DomainEndpointOptions == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DomainEndpointOptions, nil
}

func findScalingParametersByName(ctx context.Context, conn *cloudsearch.Client, name string) (*awstypes.ScalingParameters, error) {
	output, err := findScalingParametersStatusByName(ctx, conn, name)

	if err != nil {
		return nil, err
	}

	if output.Options == nil {
		return nil, tfresource.NewEmptyResultError(name)
	}

	return output.Options, nil
}

func findScalingParametersStatusByName(ctx context.Context, conn *cloudsearch.Client, name string) (*awstypes.ScalingParametersStatus, error) {
	input := &cloudsearch.DescribeScalingParametersInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DescribeScalingParameters(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ScalingParameters == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ScalingParameters, nil
}

func statusDomainDeleting(conn *cloudsearch.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findDomainByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, flex.BoolToStringValue(output.Deleted), nil
	}
}

func statusDomainProcessing(conn *cloudsearch.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findDomainByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, flex.BoolToStringValue(output.Processing), nil
	}
}

func waitDomainActive(ctx context.Context, conn *cloudsearch.Client, name string, timeout time.Duration) (*awstypes.DomainStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"true"},
		Target:  []string{"false"},
		Refresh: statusDomainProcessing(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DomainStatus); ok {
		return output, err
	}

	return nil, err
}

func waitDomainDeleted(ctx context.Context, conn *cloudsearch.Client, name string, timeout time.Duration) (*awstypes.DomainStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"true"},
		Target:  []string{},
		Refresh: statusDomainDeleting(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DomainStatus); ok {
		return output, err
	}

	return nil, err
}
