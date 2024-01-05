package networkfirewall

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/networkfirewall/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	// "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	// "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)
// TIP: ==== FILE STRUCTURE ====
// All resources should follow this basic outline. Improve this resource's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main resource struct with schema method
// 4. Create, read, update, delete methods (in that order)
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="TLS Inspection Configuration")
func newResourceTLSInspectionConfiguration(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceTLSInspectionConfiguration{}
	
	// TIP: ==== CONFIGURABLE TIMEOUTS ====
	// Users can configure timeout lengths but you need to use the times they
	// provide. Access the timeout they configure (or the defaults) using,
	// e.g., r.CreateTimeout(ctx, plan.Timeouts) (see below). The times here are
	// the defaults if they don't configure timeouts.
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameTLSInspectionConfiguration = "TLS Inspection Configuration"
)

type resourceTLSInspectionConfiguration struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceTLSInspectionConfiguration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_networkfirewall_tls_inspection_configuration"
}

// TIP: ==== SCHEMA ====
// In the schema, add each of the attributes in snake case (e.g.,
// delete_automated_backups).
//
// Formatting rules:
// * Alphabetize attributes to make them easier to find.
// * Do not add a blank line between attributes.
//
// Attribute basics:
// * If a user can provide a value ("configure a value") for an
//   attribute (e.g., instances = 5), we call the attribute an
//   "argument."
// * You change the way users interact with attributes using:
//     - Required
//     - Optional
//     - Computed
// * There are only four valid combinations:
//
// 1. Required only - the user must provide a value
// Required: true,
//
// 2. Optional only - the user can configure or omit a value; do not
//    use Default or DefaultFunc
// Optional: true,
//
// 3. Computed only - the provider can provide a value but the user
//    cannot, i.e., read-only
// Computed: true,
//
// 4. Optional AND Computed - the provider or user can provide a value;
//    use this combination if you are using Default
// Optional: true,
// Computed: true,
//
// You will typically find arguments in the input struct
// (e.g., CreateDBInstanceInput) for the create operation. Sometimes
// they are only in the input struct (e.g., ModifyDBInstanceInput) for
// the modify operation.
//
// For more about schema options, visit
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas?page=schemas
func (r *resourceTLSInspectionConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"description": schema.StringAttribute{
				Optional: true,
			},
			"id": framework.IDAttribute(),
			// Map name to TLSInspectionConfigurationName
			"name": schema.StringAttribute{
				Required: true,
				// TIP: ==== PLAN MODIFIERS ====
				// Plan modifiers were introduced with Plugin-Framework to provide a mechanism
				// for adjusting planned changes prior to apply. The planmodifier subpackage
				// provides built-in modifiers for many common use cases such as 
				// requiring replacement on a value change ("ForceNew: true" in Plugin-SDK 
				// resources).
				//
				// See more:
				// https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"last_modified_time": schema.StringAttribute{
				Computed: true,
			},
			"number_of_associations": schema.Int64Attribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"update_token": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"encryption_configuration": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key_id": schema.StringAttribute{
							Optional: true,
						},
						"type": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"tls_inspection_configuration": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"server_certificate_configurations": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"certificate_authority_arn": schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"check_certificate_revocation_status": schema.ListNestedBlock{
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"revoked_status_action": schema.StringAttribute{
													Optional: true,
												},
												"unknown_status_action": schema.StringAttribute{
													Optional: true,
												},
											},
										},
									},
									"server_certificates": schema.ListNestedBlock{
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"resource_arn": schema.StringAttribute{
													Optional: true,
													// TODO: Add string validation with regex
												},
											},
										},
									},
									"scope": schema.ListNestedBlock{
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"protocols": schema.ListAttribute{
													ElementType: types.Int64Type,
													Required:    true,
												},
											},
											Blocks: map[string]schema.Block{
												"destination_ports": schema.ListNestedBlock{
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"from_port": schema.Int64Attribute{
																Required: true,
															},
															"to_port": schema.Int64Attribute{
																Required: true,
															},
														},
													},
												},
												"destinations": schema.ListNestedBlock{
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"address_definition": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
												"source_ports": schema.ListNestedBlock{
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"from_port": schema.Int64Attribute{
																Required: true,
															},
															"to_port": schema.Int64Attribute{
																Required: true,
															},
														},
													},
												},
												"sources": schema.ListNestedBlock{
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"address_definition": schema.StringAttribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									},
								},
							},
						},
				},
			},
			// "complex_argument": schema.ListNestedBlock{
			// 	// TIP: ==== LIST VALIDATORS ====
			// 	// List and set validators take the place of MaxItems and MinItems in 
			// 	// Plugin-Framework based resources. Use listvalidator.SizeAtLeast(1) to
			// 	// make a nested object required. Similar to Plugin-SDK, complex objects 
			// 	// can be represented as lists or sets with listvalidator.SizeAtMost(1).
			// 	//
			// 	// For a complete mapping of Plugin-SDK to Plugin-Framework schema fields, 
			// 	// see:
			// 	// https://developer.hashicorp.com/terraform/plugin/framework/migrating/attributes-blocks/blocks
			// 	Validators: []validator.List{
			// 		listvalidator.SizeAtMost(1),
			// 	},
			// 	NestedObject: schema.NestedBlockObject{
			// 		Attributes: map[string]schema.Attribute{
			// 			"nested_required": schema.StringAttribute{
			// 				Required: true,
			// 			},
			// 			"nested_computed": schema.StringAttribute{
			// 				Computed: true,
			// 				PlanModifiers: []planmodifier.String{
			// 					stringplanmodifier.UseStateForUnknown(),
			// 				},
			// 			},
			// 		},
			// 	},
			// }, 
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceTLSInspectionConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// TIP: ==== RESOURCE CREATE ====
	// Generally, the Create function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the plan
	// 3. Populate a create input structure
	// 4. Call the AWS create/put function
	// 5. Using the output from the create function, set the minimum arguments
	//    and attributes for the Read function to work, as well as any computed
	//    only attributes. 
	// 6. Use a waiter to wait for create to complete
	// 7. Save the request plan to response state

	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().NetworkFirewallConn(ctx)
	
	// TIP: -- 2. Fetch the plan
	var plan resourceTLSInspectionConfigurationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	// TIP: -- 3. Populate a create input structure
	in := &networkfirewall.CreateTLSInspectionConfigurationInput{
		// NOTE: Name is mandatory
		TLSInspectionConfigurationName: aws.String(plan.Name.ValueString()),
	}

	if !plan.Description.IsNull() {
		// NOTE: Description is optional
		in.Description = aws.String(plan.Description.ValueString())
	}

	// Complex arguments
	if !plan.TLSInspectionConfiguration.IsNull() {
		// TIP: Use an expander to assign a complex argument. The elements must be
		// deserialized into the appropriate struct before being passed to the expander.
		var tfList []tlsInspectionConfigurationData
		resp.Diagnostics.Append(plan.TLSInspectionConfiguration.ElementsAs(ctx, &tfList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		// in.TLSInspectionConfiguration = expandComplexArgument(tfList)
		in.TLSInspectionConfiguration = expandTLSInspectionConfiguration(tfList)
	}

	if !plan.EncryptionConfiguration.IsNull() {
		// TIP: Use an expander to assign a complex argument. The elements must be
		// deserialized into the appropriate struct before being passed to the expander.
		var tfList []encryptionConfigurationData
		resp.Diagnostics.Append(plan.EncryptionConfiguration.ElementsAs(ctx, &tfList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		in.EncryptionConfiguration = expandTLSEncryptionConfiguration(tfList)
	}
	
	// TIP: -- 4. Call the AWS create function
	out, err := conn.CreateTLSInspectionConfiguration(in)
	if err != nil {
		// TIP: Since ID has not been set yet, you cannot use plan.ID.String()
		// in error messages at this point.
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionCreating, ResNameTLSInspectionConfiguration, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.TLSInspectionConfigurationResponse == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionCreating, ResNameTLSInspectionConfiguration, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}
	
	// TIP: -- 5. Using the output from the create function, set the minimum attributes
	// Output consists only of TLSInspectionConfigurationResponse
	plan.ARN = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationArn)
	plan.ID = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationId)
	
	// TIP: -- 6. Use a waiter to wait for create to complete
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitTLSInspectionConfigurationCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionWaitingForCreation, ResNameTLSInspectionConfiguration, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	
	// TIP: -- 7. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceTLSInspectionConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// TIP: ==== RESOURCE READ ====
	// Generally, the Read function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the state
	// 3. Get the resource from AWS
	// 4. Remove resource from state if it is not found
	// 5. Set the arguments and attributes
	// 6. Set the state

	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().NetworkFirewallConn(ctx)
	
	// TIP: -- 2. Fetch the state
	var state resourceTLSInspectionConfigurationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findTLSInspectionConfigurationByNameAndARN(ctx, conn, state.ARN.ValueString())
	// TIP: -- 4. Remove resource from state if it is not found
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionSetting, ResNameTLSInspectionConfiguration, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	
	// TIP: -- 5. Set the arguments and attributes
	//
	// For simple data types (i.e., schema.StringAttribute, schema.BoolAttribute,
	// schema.Int64Attribute, and schema.Float64Attribue), simply setting the  
	// appropriate data struct field is sufficient. The flex package implements
	// helpers for converting between Go and Plugin-Framework types seamlessly. No 
	// error or nil checking is necessary.
	//
	// However, there are some situations where more handling is needed such as
	// complex data types (e.g., schema.ListAttribute, schema.SetAttribute). In 
	// these cases the flatten function may have a diagnostics return value, which
	// should be appended to resp.Diagnostics.
	state.ARN = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationArn)
	state.Description = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.Description)
	state.ID = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationId)
	state.Name = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationName)
	
	state.LastModifiedTime = flex.StringValueToFramework(ctx, out.TLSInspectionConfigurationResponse.LastModifiedTime.Format(time.RFC3339))
	state.NumberOfAssociations = flex.Int64ToFramework(ctx, out.TLSInspectionConfigurationResponse.NumberOfAssociations)
	state.UpdateToken = flex.StringToFramework(ctx, out.UpdateToken)
	state.Status = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationStatus)

	// Complex types
	encryptionConfiguration, d := flattenTLSEncryptionConfiguration(ctx, out.TLSInspectionConfigurationResponse.EncryptionConfiguration)
	resp.Diagnostics.Append(d...)
	state.EncryptionConfiguration = encryptionConfiguration
	
	certificateAuthority, d := flattenTLSCertificate(ctx, out.TLSInspectionConfigurationResponse.CertificateAuthority)
	resp.Diagnostics.Append(d...)
	state.CertificateAuthority = certificateAuthority
	
	certificates, d := flattenCertificates(ctx, out.TLSInspectionConfigurationResponse.Certificates)
	resp.Diagnostics.Append(d...)
	state.Certificates = certificates
	
	// TIP: Setting a complex type.
	// complexArgument, d := flattenComplexArgument(ctx, out.ComplexArgument)
	// resp.Diagnostics.Append(d...)
	// state.ComplexArgument = complexArgument
	
	// TIP: -- 6. Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceTLSInspectionConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TIP: ==== RESOURCE UPDATE ====
	// Not all resources have Update functions. There are a few reasons:
	// a. The AWS API does not support changing a resource
	// b. All arguments have RequiresReplace() plan modifiers
	// c. The AWS API uses a create call to modify an existing resource
	//
	// In the cases of a. and b., the resource will not have an update method
	// defined. In the case of c., Update and Create can be refactored to call
	// the same underlying function.
	//
	// The rest of the time, there should be an Update function and it should
	// do the following things. Make sure there is a good reason if you don't
	// do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the plan and state
	// 3. Populate a modify input structure and check for changes
	// 4. Call the AWS modify/update function
	// 5. Use a waiter to wait for update to complete
	// 6. Save the request plan to response state
	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().NetworkFirewallConn(ctx)
	
	// TIP: -- 2. Fetch the plan
	var plan, state resourceTLSInspectionConfigurationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	// TIP: -- 3. Populate a modify input structure and check for changes
	if !plan.Description.Equal(state.Description) ||
		!plan.TLSInspectionConfiguration.Equal(state.TLSInspectionConfiguration) ||
		!plan.EncryptionConfiguration.Equal(state.EncryptionConfiguration) ||
		!plan.Type.Equal(state.Type) {

		in := &networkfirewall.UpdateTLSInspectionConfigurationInput{
			// TIP: Mandatory or fields that will always be present can be set when
			// you create the Input structure. (Replace these with real fields.)
			TLSInspectionConfigurationArn:   aws.String(plan.ARN.ValueString()),
			TLSInspectionConfigurationName: aws.String(plan.Name.ValueString()),
			UpdateToken: aws.String(plan.UpdateToken.ValueString()),
		}

		if !plan.Description.IsNull() {
			// TIP: Optional fields should be set based on whether or not they are
			// used.
			in.Description = aws.String(plan.Description.ValueString())
		}
		// if !plan.ComplexArgument.IsNull() {
		// 	// TIP: Use an expander to assign a complex argument. The elements must be
		// 	// deserialized into the appropriate struct before being passed to the expander.
		// 	var tfList []complexArgumentData
		// 	resp.Diagnostics.Append(plan.ComplexArgument.ElementsAs(ctx, &tfList, false)...)
		// 	if resp.Diagnostics.HasError() {
		// 		return
		// 	}

		// 	in.ComplexArgument = expandComplexArgument(tfList)
		// }

		if !plan.TLSInspectionConfiguration.IsNull() {
			var tfList []tlsInspectionConfigurationData
			resp.Diagnostics.Append(plan.TLSInspectionConfiguration.ElementsAs(ctx, &tfList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			in.TLSInspectionConfiguration = expandTLSInspectionConfiguration(tfList)
		
		}
		
		if !plan.EncryptionConfiguration.IsNull() {
			var tfList []encryptionConfigurationData
			resp.Diagnostics.Append(plan.EncryptionConfiguration.ElementsAs(ctx, &tfList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			in.EncryptionConfiguration = expandTLSEncryptionConfiguration(tfList)
		}

		// TIP: -- 4. Call the AWS modify/update function
		out, err := conn.UpdateTLSInspectionConfigurationWithContext(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionUpdating, ResNameTLSInspectionConfiguration, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.TLSInspectionConfigurationResponse == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionUpdating, ResNameTLSInspectionConfiguration, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		
		// TIP: Using the output from the update function, re-set any computed attributes
		plan.ARN = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationArn)
		plan.ID = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationId)

		plan.LastModifiedTime = flex.StringValueToFramework(ctx, out.TLSInspectionConfigurationResponse.LastModifiedTime.Format(time.RFC3339))
		plan.NumberOfAssociations = flex.Int64ToFramework(ctx, out.TLSInspectionConfigurationResponse.NumberOfAssociations)
		plan.UpdateToken = flex.StringToFramework(ctx, out.UpdateToken)
		plan.Status = flex.StringToFramework(ctx, out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationStatus)
	}

	
	// TIP: -- 5. Use a waiter to wait for update to complete
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitTLSInspectionConfigurationUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionWaitingForUpdate, ResNameTLSInspectionConfiguration, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	
	// TIP: -- 6. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceTLSInspectionConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().NetworkFirewallConn(ctx)
	
	// TIP: -- 2. Fetch the state
	var state resourceTLSInspectionConfigurationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	// TIP: -- 3. Populate a delete input structure
	in := &networkfirewall.DeleteTLSInspectionConfigurationInput{
		TLSInspectionConfigurationArn: aws.String(state.ARN.ValueString()),
	}
	
	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DeleteTLSInspectionConfigurationWithContext(ctx, in)
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionDeleting, ResNameTLSInspectionConfiguration, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	
	// TIP: -- 5. Use a waiter to wait for delete to complete
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitTLSInspectionConfigurationDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NetworkFirewall, create.ErrActionWaitingForDeletion, ResNameTLSInspectionConfiguration, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

// TIP: ==== TERRAFORM IMPORTING ====
// If Read can get all the information it needs from the Identifier
// (i.e., path.Root("id")), you can use the PassthroughID importer. Otherwise,
// you'll need a custom import function.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/resources/import
func (r *resourceTLSInspectionConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}


// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., awstypes.StatusInProgress).
const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

// TIP: ==== WAITERS ====
// Some resources of some services have waiters provided by the AWS API.
// Unless they do not work properly, use them rather than defining new ones
// here.
//
// Sometimes we define the wait, status, and find functions in separate
// files, wait.go, status.go, and find.go. Follow the pattern set out in the
// service and define these where it makes the most sense.
//
// If these functions are used in the _test.go file, they will need to be
// exported (i.e., capitalized).
//
// You will need to adjust the parameters and names to fit the service.
func waitTLSInspectionConfigurationCreated(ctx context.Context, conn *networkfirewall.NetworkFirewall, id string, timeout time.Duration) (*networkfirewall.TLSInspectionConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusTLSInspectionConfiguration(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*networkfirewall.TLSInspectionConfiguration); ok {
		return out, err
	}

	return nil, err
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitTLSInspectionConfigurationUpdated(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string, timeout time.Duration) (*networkfirewall.TLSInspectionConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusTLSInspectionConfiguration(ctx, conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*networkfirewall.TLSInspectionConfiguration); ok {
		return out, err
	}

	return nil, err
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitTLSInspectionConfigurationDeleted(ctx context.Context, conn *networkfirewall.NetworkFirewall, id string, timeout time.Duration) (*networkfirewall.TLSInspectionConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusDeleting, statusNormal},
		Target:                    []string{},
		Refresh:                   statusTLSInspectionConfiguration(ctx, conn, id),
		Timeout:                   timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*networkfirewall.TLSInspectionConfiguration); ok {
		return out, err
	}

	return nil, err
}

// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.
func statusTLSInspectionConfiguration(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findTLSInspectionConfigurationByNameAndARN(ctx, conn, arn)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.TLSInspectionConfigurationResponse.TLSInspectionConfigurationStatus), nil
	}
}

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
func findTLSInspectionConfigurationByNameAndARN(ctx context.Context, conn *networkfirewall.NetworkFirewall, arn string) (*networkfirewall.DescribeTLSInspectionConfigurationOutput, error) {
	in := &networkfirewall.DescribeTLSInspectionConfigurationInput{
		TLSInspectionConfigurationArn: aws.String(arn),
	}
	
	out, err := conn.DescribeTLSInspectionConfigurationWithContext(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.TLSInspectionConfigurationResponse == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

// TIP: ==== FLEX ====
// Flatteners and expanders ("flex" functions) help handle complex data
// types. Flatteners take an API data type and return the equivalent Plugin-Framework 
// type. In other words, flatteners translate from AWS -> Terraform.
//
// On the other hand, expanders take a Terraform data structure and return
// something that you can send to the AWS API. In other words, expanders
// translate from Terraform -> AWS.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/
// func flattenComplexArgument(ctx context.Context, apiObject *awstypes.ComplexArgument) (types.List, diag.Diagnostics) {
// 	var diags diag.Diagnostics
// 	elemType := types.ObjectType{AttrTypes: complexArgumentAttrTypes}

// 	if apiObject == nil {
// 		return types.ListNull(elemType), diags
// 	}

// 	obj := map[string]attr.Value{
// 		"nested_required": flex.StringValueToFramework(ctx, apiObject.NestedRequired),
// 		"nested_optional": flex.StringValueToFramework(ctx, apiObject.NestedOptional),
// 	}
// 	objVal, d := types.ObjectValue(complexArgumentAttrTypes, obj)
// 	diags.Append(d...)

// 	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
// 	diags.Append(d...)

// 	return listVal, diags
// }


func flattenCertificates(ctx context.Context, certificateList []*networkfirewall.TlsCertificateData)(types.List, diag.Diagnostics){
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: certificatesAttrTypes}

	if len(certificateList) == 0 {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, certificate := range certificateList {
		if certificate == nil {
			continue
		}
		flattenedCertificate, d := flattenTLSCertificate(ctx, certificate)
		diags.Append(d...)
		elems = append(elems, flattenedCertificate)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func flattenTLSCertificate(ctx context.Context, certificate *networkfirewall.TlsCertificateData)(types.List, diag.Diagnostics){
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: certificatesAttrTypes}

	if certificate == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"certificate_arn": flex.StringToFramework(ctx, certificate.CertificateArn),
		"certificate_serial": flex.StringToFramework(ctx, certificate.CertificateSerial),
		"status": flex.StringToFramework(ctx, certificate.Status),
		"status_message": flex.StringToFramework(ctx, certificate.StatusMessage),
	}
	objVal, d := types.ObjectValue(certificatesAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags

}


func flattenScopes(ctx context.Context, scopes []*networkfirewall.ServerCertificateScope)(types.List, diag.Diagnostics){
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: scopeAttrTypes}

	if len(scopes) == 0 {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, scope := range scopes {
		if scope == nil {
			continue
		}

		destinationPorts, d := flattenPortRange(ctx, scope.DestinationPorts)
		destinations, d := flattenSourceDestinations(ctx, scope.Destinations)
		sourcePorts, d := flattenPortRange(ctx, scope.SourcePorts)
		sources, d := flattenSourceDestinations(ctx, scope.Sources)


		obj := map[string]attr.Value{
			"destination_ports": destinationPorts,
			"destinations": destinations,
			// "protocols": flex.
			"source_ports": sourcePorts,
			"sources": sources,
		}
		objVal, d := types.ObjectValue(scopeAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags

}


func flattenSourceDestinations(ctx context.Context, destinations []*networkfirewall.Address) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: sourceDestinationAttrTypes}

	if len(destinations) == 0 {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, destination := range destinations {
		if destination == nil {
			continue
		}

		obj := map[string]attr.Value{
			"address_definition": flex.StringToFramework(ctx, destination.AddressDefinition),
		}
		objVal, d := types.ObjectValue(sourceDestinationAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func flattenPortRange(ctx context.Context, ranges []*networkfirewall.PortRange) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: portRangeAttrTypes}

	if len(ranges) == 0 {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, portRange := range ranges {
		if portRange == nil {
			continue
		}

		obj := map[string]attr.Value{
			"from_port": flex.Int64ToFramework(ctx, portRange.FromPort),
			"to_port": flex.Int64ToFramework(ctx, portRange.ToPort),
		}
		objVal, d := types.ObjectValue(portRangeAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func flattenCheckCertificateRevocationStatus(){}

func flattenTLSEncryptionConfiguration(ctx context.Context, encryptionConfiguration *networkfirewall.EncryptionConfiguration) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: encryptionConfigurationAttrTypes}

	if encryptionConfiguration == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"key_id": flex.StringToFramework(ctx, encryptionConfiguration.KeyId),
		"type": flex.StringToFramework(ctx, encryptionConfiguration.Type),
	}
	objVal, d := types.ObjectValue(encryptionConfigurationAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags

}

// TIP: Often the AWS API will return a slice of structures in response to a
// request for information. Sometimes you will have set criteria (e.g., the ID)
// that means you'll get back a one-length slice. This plural function works
// brilliantly for that situation too.
// func flattenComplexArguments(ctx context.Context, apiObjects []*awstypes.ComplexArgument) (types.List, diag.Diagnostics) {
// 	var diags diag.Diagnostics
// 	elemType := types.ObjectType{AttrTypes: complexArgumentAttrTypes}

// 	if len(apiObjects) == 0 {
// 		return types.ListNull(elemType), diags
// 	}

// 	elems := []attr.Value{}
// 	for _, apiObject := range apiObjects {
// 		if apiObject == nil {
// 			continue
// 		}

// 		obj := map[string]attr.Value{
// 			"nested_required": flex.StringValueToFramework(ctx, apiObject.NestedRequired),
// 			"nested_optional": flex.StringValueToFramework(ctx, apiObject.NestedOptional),
// 		}
// 		objVal, d := types.ObjectValue(complexArgumentAttrTypes, obj)
// 		diags.Append(d...)

// 		elems = append(elems, objVal)
// 	}

// 	listVal, d := types.ListValue(elemType, elems)
// 	diags.Append(d...)

// 	return listVal, diags
// }

// TIP: Remember, as mentioned above, expanders take a Terraform data structure
// and return something that you can send to the AWS API. In other words,
// expanders translate from Terraform -> AWS.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/
// func expandComplexArgument(tfList []complexArgumentData) *awstypes.ComplexArgument {
// 	if len(tfList) == 0 {
// 		return nil
// 	}

// 	tfObj := tfList[0]
// 	apiObject := &awstypes.ComplexArgument{
// 		NestedRequired: aws.String(tfObj.NestedRequired.ValueString()),
// 	}
// 	if !tfObj.NestedOptional.IsNull() {
// 		apiObject.NestedOptional = aws.String(tfObj.NestedOptional.ValueString())
// 	}

// 	return apiObject
// }

// TODO: add note explaining why not using existing expandEncryptionConfiguration()
func expandTLSEncryptionConfiguration(tfList []encryptionConfigurationData) *networkfirewall.EncryptionConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	apiObject := &networkfirewall.EncryptionConfiguration{
		KeyId: aws.String(tfObj.KeyId.ValueString()),
		Type: aws.String(tfObj.Type.ValueString()),
	}

	return apiObject
}

func expandTLSInspectionConfiguration(tfList []tlsInspectionConfigurationData) *networkfirewall.TLSInspectionConfiguration {
	if len(tfList) == 0 {
		return nil
	}
	
	// var apiObject []networkfirewall.TLSInspectionConfiguration

	tfObj := tfList[0]

	// for _, item := range tfList {
	apiObject := &networkfirewall.TLSInspectionConfiguration{}
	if !tfObj.ServerCertificateConfiguration.IsNull() {
		var serverConfig []serverCertificateConfigurationsData
		apiObject.ServerCertificateConfigurations = expandServerCertificateConfigurations(serverConfig)
	}
		
	// apiObject = append(apiObject, conf)
	
	return apiObject
}

func expandServerCertificateConfigurations(tfList []serverCertificateConfigurationsData) []*networkfirewall.ServerCertificateConfiguration {
	var apiObject []*networkfirewall.ServerCertificateConfiguration

	for _, item := range tfList {
		conf := &networkfirewall.ServerCertificateConfiguration{
			CertificateAuthorityArn: aws.String(item.CertificateAuthorityArn.ValueString()),
		}
		if !item.CheckCertificateRevocationsStatus.IsNull() {
			var certificateRevocationStatus []checkCertificateRevocationStatusData
			conf.CheckCertificateRevocationStatus = expandCheckCertificateRevocationStatus(certificateRevocationStatus)
		}
		if !item.Scope.IsNull() {
			var scopesList []scopeData
			conf.Scopes = expandScopes(scopesList)
		}
		if !item.ServerCertificates.IsNull() {
			var serverCertificates []serverCertificatesData
			conf.ServerCertificates = expandServerCertificates(serverCertificates)
		}

		apiObject = append(apiObject, conf)
	}
	return apiObject
}

func expandCheckCertificateRevocationStatus(tfList []checkCertificateRevocationStatusData) *networkfirewall.CheckCertificateRevocationStatusActions{
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	apiObject := &networkfirewall.CheckCertificateRevocationStatusActions{
		RevokedStatusAction: aws.String(tfObj.RevokedStatusAction.ValueString()),
		UnknownStatusAction: aws.String(tfObj.UnknownStatusAction.ValueString()),
	}
	return apiObject
}

func expandServerCertificates(tfList []serverCertificatesData) []*networkfirewall.ServerCertificate {
	var apiObject []*networkfirewall.ServerCertificate

	for _, item := range tfList {
		conf := &networkfirewall.ServerCertificate{
			ResourceArn: aws.String(item.ResourceARN.ValueString()),
		}

		apiObject = append(apiObject, conf)
	}
	return apiObject
}

func expandScopes(tfList []scopeData) []*networkfirewall.ServerCertificateScope {
	var apiObject []*networkfirewall.ServerCertificateScope

	for _, tfObj := range tfList {
		item := &networkfirewall.ServerCertificateScope{
			// Protocols: aws.Int64Slice(),
		}
		if !tfObj.DestinationPorts.IsNull() {
			var destinationPorts []portRangeData
			item.DestinationPorts = expandPortRange(destinationPorts)
		}
		if !tfObj.Destinations.IsNull() {
			var destinations []sourceDestinationData
			item.Destinations = expandSourceDestinations(destinations)
		}
		if !tfObj.SourcePorts.IsNull() {
			var sourcePorts []portRangeData
			item.SourcePorts = expandPortRange(sourcePorts)
		}
		if !tfObj.Sources.IsNull() {
			var sources []sourceDestinationData
			item.Sources = expandSourceDestinations(sources)
		}
		apiObject = append(apiObject, item)
	}

	return apiObject
}

func expandPortRange(tfList []portRangeData) []*networkfirewall.PortRange {
	var apiObject []*networkfirewall.PortRange

	for _, tfObj := range tfList {
		item := &networkfirewall.PortRange{
			FromPort: aws.Int64(tfObj.FromPort.ValueInt64()),
			ToPort: aws.Int64(tfObj.ToPort.ValueInt64()),
		}
		apiObject = append(apiObject, item)
	}

	return apiObject
}

func expandSourceDestinations(tfList []sourceDestinationData) []*networkfirewall.Address {
	var apiObject []*networkfirewall.Address

	for _, tfObj := range tfList {
		item := &networkfirewall.Address{
			AddressDefinition: aws.String(tfObj.AddressDefinition.ValueString()),
		}
		apiObject = append(apiObject, item)
	}

	return apiObject
}




// TIP: Even when you have a list with max length of 1, this plural function
// works brilliantly. However, if the AWS API takes a structure rather than a
// slice of structures, you will not need it.
// func expandComplexArguments(tfList []complexArgumentData) []*networkfirewall.ComplexArgument {
// 	// TIP: The AWS API can be picky about whether you send a nil or zero-
// 	// length for an argument that should be cleared. For example, in some
// 	// cases, if you send a nil value, the AWS API interprets that as "make no
// 	// changes" when what you want to say is "remove everything." Sometimes
// 	// using a zero-length list will cause an error.
// 	//
// 	// As a result, here are two options. Usually, option 1, nil, will work as
// 	// expected, clearing the field. But, test going from something to nothing
// 	// to make sure it works. If not, try the second option.
// 	// TIP: Option 1: Returning nil for zero-length list
//     if len(tfList) == 0 {
//         return nil
//     }
//     var apiObject []*awstypes.ComplexArgument
// 	// TIP: Option 2: Return zero-length list for zero-length list. If option 1 does
// 	// not work, after testing going from something to nothing (if that is
// 	// possible), uncomment out the next line and remove option 1.
// 	//
// 	// apiObject := make([]*networkfirewall.ComplexArgument, 0)

// 	for _, tfObj := range tfList {
// 		item := &networkfirewall.ComplexArgument{
// 			NestedRequired: aws.String(tfObj.NestedRequired.ValueString()),
// 		}
// 		if !tfObj.NestedOptional.IsNull() {
// 			item.NestedOptional = aws.String(tfObj.NestedOptional.ValueString())
// 		}

// 		apiObject = append(apiObject, item)
// 	}

// 	return apiObject
// }

// TIP: ==== DATA STRUCTURES ====
// With Terraform Plugin-Framework configurations are deserialized into
// Go types, providing type safety without the need for type assertions.
// These structs should match the schema definition exactly, and the `tfsdk`
// tag value should match the attribute name. 
//
// Nested objects are represented in their own data struct. These will 
// also have a corresponding attribute type mapping for use inside flex
// functions.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values
type resourceTLSInspectionConfigurationData struct {
	ARN                        types.String   `tfsdk:"arn"`
	EncryptionConfiguration    types.List     `tfsdk:"encryption_configuration"`
	Certificates               types.List     `tfsdk:"certificates"`
	CertificateAuthority       types.List     `tfsdk:"certificate_authority"`
	// ComplexArgument            types.List     `tfsdk:"complex_argument"`
	Description                types.String   `tfsdk:"description"`
	ID                         types.String   `tfsdk:"id"`
	LastModifiedTime           types.String   `tfsdk:"last_modified_time"`
	Name                       types.String   `tfsdk:"name"`
	NumberOfAssociations       types.Int64    `tfsdk:"number_of_associations"`
	Status                     types.String   `tfsdk:"status"`
	TLSInspectionConfiguration types.List     `tfsdk:"tls_inspection_configuration"`
	Timeouts                   timeouts.Value `tfsdk:"timeouts"`
	Type                       types.String   `tfsdk:"type"`
	UpdateToken                types.String   `tfsdk:"update_token"`
}

type encryptionConfigurationData struct {
	Type  types.String `tfsdk:"type"`
	KeyId types.String `tfsdk:"key_id"`
}

type certificatesData struct {
	CertificateArn types.String `tfsdk:"certificate_arn"`
	CertificateSerial types.String `tfsdk:"certificate_serial"`
	Status types.String `tfsdk:"status"`
	StatusMessage types.String `tfsdk:"status_message"`
}

type tlsInspectionConfigurationData struct {
	ServerCertificateConfiguration types.List `tfsdk:"server_certificate_configuration"`
}

type serverCertificateConfigurationsData struct {
	CertificateAuthorityArn types.String `tfsdk:"certificate_authority_arn"`
	CheckCertificateRevocationsStatus types.List `tfsdk:"check_certificate_revocations_status"`
	Scope types.List `tfsdk:"scope"`
	ServerCertificates types.List `tfsdk:"server_certificates"`
}

type complexArgumentData struct {
	NestedRequired types.String `tfsdk:"nested_required"`
	NestedOptional types.String `tfsdk:"nested_optional"`
}

type scopeData struct {
	DestinationPorts types.List `tfsdk:"destination_ports"`
	Destinations types.List `tfsdk:"destinations"`
	Protocols types.List `tfsdk:"protocols"`
	SourcePorts types.List `tfsdk:"source_ports"`
	Sources types.List `tfsdk:"sources"`
}

type sourceDestinationData struct {
	AddressDefinition types.String `tfsdk:"address_definition"`
}

type portRangeData struct {
	FromPort types.Int64 `tfsdk:"from_port"`
	ToPort types.Int64 `tfsdk:"to_port"`
}

type checkCertificateRevocationStatusData struct {
	RevokedStatusAction types.String `tfsdk:"revoked_status_action"`
	UnknownStatusAction types.String `tfsdk:"unknown_status_action"`
}

type serverCertificatesData struct {
	ResourceARN types.String `tfsdk:"resource_arn"`
}

//////////////

var certificatesAttrTypes = map[string]attr.Type{
	"certificate_arn": types.StringType,
	"certificate_serial": types.StringType,
	"status": types.StringType,
	"status_message": types.StringType,
}

// var complexArgumentAttrTypes = map[string]attr.Type{
// 	"nested_required": types.StringType,
// 	"nested_optional": types.StringType,
// }

var encryptionConfigurationAttrTypes = map[string]attr.Type{
	"type": types.StringType,
	"key_id": types.StringType,
}

var tlsInspectionConfigurationAttrTypes = map[string]attr.Type{
	"server_certificate_configuration": types.ListType{ElemType: types.ObjectType{AttrTypes: serverCertificateConfigurationAttrTypes}},
}

var	serverCertificateConfigurationAttrTypes = map[string]attr.Type{
	"certificate_authority_arn": types.StringType,
	"check_certificate_revocations_status": types.ListType{ElemType: types.ObjectType{AttrTypes: checkCertificateRevocationStatusAttrTypes}},
	"scope": types.ListType{ElemType: types.ObjectType{AttrTypes: scopeAttrTypes}},
	"server_certificates": types.ListType{ElemType: types.ObjectType{AttrTypes: serverCertificatesAttrTypes}},
}

var checkCertificateRevocationStatusAttrTypes = map[string]attr.Type{
	"revoked_status_action": types.StringType,
	"unknown_status_action": types.StringType,
}

var (
	scopeAttrTypes = map[string]attr.Type{
		"destination_ports": types.ListType{ElemType: types.ObjectType{AttrTypes: portRangeAttrTypes}},
		"destinations": types.ListType{ElemType: types.ObjectType{AttrTypes: sourceDestinationAttrTypes}},
		"protocols": types.ListType{ElemType: types.Int64Type},
		"source_ports": types.ListType{ElemType: types.ObjectType{AttrTypes: portRangeAttrTypes}},
		"sources": types.ListType{ElemType: types.ObjectType{AttrTypes: sourceDestinationAttrTypes}},
	}

	sourceDestinationAttrTypes = map[string]attr.Type{
		"address_definition": types.StringType,
	}

	portRangeAttrTypes = map[string]attr.Type{
		"from_port": types.Int64Type,
		"to_port": types.Int64Type,
	}
)

var serverCertificatesAttrTypes = map[string]attr.Type{
	"resource_arn": types.StringType,
}