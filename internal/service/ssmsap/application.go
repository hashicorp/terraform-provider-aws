// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmsap

import (
	"context"
	"fmt"
	"regexp"
	"time"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmsap"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssmsap/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
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
// @FrameworkResource(name="Application")
// @Tags(identifierAttribute="arn")
func newResourceApplication(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceApplication{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type resourceApplicationData struct {
	ARN               types.String `tfsdk:"arn"`
	ID                types.String `tfsdk:"id"`
	ApplicationType   types.String `tfsdk:"application_type"`
	SapInstanceNumber types.String `tfsdk:"sap_instance_number"`
	Sid               types.String `tfsdk:"sap_system_id"`
	Instances         types.List   `tfsdk:"instances"`
	//DatabaseARN         types.String   `tfsdk:"database_arn"`
	Credentials types.List     `tfsdk:"credentials"`
	Tags        types.Map      `tfsdk:"tags"`
	TagsAll     types.Map      `tfsdk:"tags_all"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

type credentialsData struct {
	CredentialType types.String `tfsdk:"credential_type"`
	DatabaseName   types.String `tfsdk:"database_name"`
	SecretId       types.String `tfsdk:"secret_id"`
}

var credentialsAttrTypes = map[string]attr.Type{
	"credential_type": types.StringType,
	"database_name":   types.StringType,
	"secret_id":       types.StringType,
}

const (
	ResNameApplication = "Application"
)

type resourceApplication struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceApplication) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_ssmsap_application"
}

func (r *resourceApplication) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(50),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[\w\d]{1,50}$`),
						"Only letters and digits are allowed. The length must be between 1 and 50 characters.",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "ID of the application",
			},
			"application_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.ApplicationType](),
				},
				Description: "Type of the application",
			},
			"sap_instance_number": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(2),
					stringvalidator.LengthAtMost(2),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9]{2}$`),
						"must be a two digit number",
					),
				},
				Description: "SAP instance number of the application",
			},
			"sap_system_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				//three letter or digits
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(3),
					stringvalidator.LengthAtMost(3),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[A-Z0-9]{3}$`),
						"must be a three letter or digit uppercase string",
					),
				},
				Description: "System ID of the application",
			},
			"instances": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				Description: "Amazon EC2 instances on which the SAP application is running",
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"credentials": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(20),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"credential_type": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.CredentialType](),
							},
							Description: "The type of the application credentials.",
						},
						"database_name": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
								stringvalidator.LengthAtMost(100),
							},
							Description: "The name of the SAP HANA database.",
						},
						"secret_id": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
								stringvalidator.LengthAtMost(100),
							},
							Description: "The secret ID created in AWS Secrets Manager to store the credentials of the SAP application. ",
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceApplication) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceApplicationData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSMSAPClient(ctx)

	//SAP_ABAP not yet supported
	if plan.ApplicationType.ValueString() == "SAP_ABAP" {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMSAP, create.ErrActionCreating, ResNameApplication, plan.ID.String(), nil),
			"SAP_ABAP application type is not yet supported by the provider",
		)
		return
	}

	in := &ssmsap.RegisterApplicationInput{
		ApplicationId:     aws.String(plan.ID.ValueString()),
		ApplicationType:   awstypes.ApplicationType(plan.ApplicationType.ValueString()),
		SapInstanceNumber: aws.String(plan.SapInstanceNumber.ValueString()),
		Sid:               aws.String(plan.Sid.ValueString()),
		Instances:         flex.ExpandFrameworkStringValueList(ctx, plan.Instances),
		Tags:              getTagsIn(ctx),
	}

	if !plan.Credentials.IsNull() {
		var tfList []credentialsData
		resp.Diagnostics.Append(plan.Credentials.ElementsAs(ctx, &tfList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.Credentials = expandCredentials(tfList, plan.Sid.ValueString())
	}

	out, err := conn.RegisterApplication(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMSAP, create.ErrActionCreating, ResNameApplication, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	state := plan

	state.ID = flex.StringToFramework(ctx, out.Application.Id)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitApplicationCreated(ctx, conn, aws.ToString(out.Application.Id), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMSAP, create.ErrActionWaitingForCreation, ResNameApplication, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	application, tags, err := findApplicationByID(ctx, conn, aws.ToString(out.Application.Id))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMSAP, create.ErrActionSetting, ResNameApplication, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	err = state.refreshFromOutput(ctx, conn, application, tags)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMSAP, create.ErrActionSetting, ResNameApplication, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceApplication) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SSMSAPClient(ctx)
	var state resourceApplicationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	application, tags, err := findApplicationByID(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMSAP, create.ErrActionSetting, ResNameApplication, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	err = state.refreshFromOutput(ctx, conn, application, tags)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMSAP, create.ErrActionSetting, ResNameApplication, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceApplication) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	//call delete and create
	// delete(ctx, req, resp)

	// // TIP: ==== RESOURCE UPDATE ====
	// // Not all resources have Update functions. There are a few reasons:
	// // a. The AWS API does not support changing a resource
	// // b. All arguments have RequiresReplace() plan modifiers
	// // c. The AWS API uses a create call to modify an existing resource
	// //
	// // In the cases of a. and b., the resource will not have an update method
	// // defined. In the case of c., Update and Create can be refactored to call
	// // the same underlying function.
	// //
	// // The rest of the time, there should be an Update function and it should
	// // do the following things. Make sure there is a good reason if you don't
	// // do one of these.
	// //
	// // 1. Get a client connection to the relevant service
	// // 2. Fetch the plan and state
	// // 3. Populate a modify input structure and check for changes
	// // 4. Call the AWS modify/update function
	// // 5. Use a waiter to wait for update to complete
	// // 6. Save the request plan to response state
	// // TIP: -- 1. Get a client connection to the relevant service
	// conn := r.Meta().SSMSAPClient(ctx)

	// // TIP: -- 2. Fetch the plan
	// var plan, state resourceApplicationData
	// resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// // TIP: -- 3. Populate a modify input structure and check for changes
	// if !plan.Name.Equal(state.Name) ||
	// 	!plan.Description.Equal(state.Description) ||
	// 	!plan.Credential.Equal(state.Credential) ||
	// 	!plan.Type.Equal(state.Type) {

	// 	in := &ssmsap.UpdateApplicationInput{
	// 		// TIP: Mandatory or fields that will always be present can be set when
	// 		// you create the Input structure. (Replace these with real fields.)
	// 		ID:   aws.String(plan.ID.ValueString()),
	// 		ApplicationName: aws.String(plan.Name.ValueString()),
	// 		ApplicationType: aws.String(plan.Type.ValueString()),
	// 	}

	// 	if !plan.Description.IsNull() {
	// 		// TIP: Optional fields should be set based on whether or not they are
	// 		// used.
	// 		in.Description = aws.String(plan.Description.ValueString())
	// 	}
	// 	if !plan.Credential.IsNull() {
	// 		// TIP: Use an expander to assign a complex argument. The elements must be
	// 		// deserialized into the appropriate struct before being passed to the expander.
	// 		var tfList []CredentialData
	// 		resp.Diagnostics.Append(plan.Credential.ElementsAs(ctx, &tfList, false)...)
	// 		if resp.Diagnostics.HasError() {
	// 			return
	// 		}

	// 		in.Credential = expandCredential(tfList)
	// 	}

	// 	// TIP: -- 4. Call the AWS modify/update function
	// 	out, err := conn.UpdateApplication(ctx, in)
	// 	if err != nil {
	// 		resp.Diagnostics.AddError(
	// 			create.ProblemStandardMessage(names.SSMSAP, create.ErrActionUpdating, ResNameApplication, plan.ID.String(), err),
	// 			err.Error(),
	// 		)
	// 		return
	// 	}
	// 	if out == nil || out.Application == nil {
	// 		resp.Diagnostics.AddError(
	// 			create.ProblemStandardMessage(names.SSMSAP, create.ErrActionUpdating, ResNameApplication, plan.ID.String(), nil),
	// 			errors.New("empty output").Error(),
	// 		)
	// 		return
	// 	}

	// 	// TIP: Using the output from the update function, re-set any computed attributes
	// 	plan.ARN = flex.StringToFramework(ctx, out.Application.Arn)
	// 	plan.ID = flex.StringToFramework(ctx, out.Application.ID)
	// }

	// // TIP: -- 5. Use a waiter to wait for update to complete
	// updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	// _, err := waitApplicationUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.SSMSAP, create.ErrActionWaitingForUpdate, ResNameApplication, plan.ID.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }

	// // TIP: -- 6. Save the request plan to response state
	// resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceApplication) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SSMSAPClient(ctx)

	var state resourceApplicationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ssmsap.DeregisterApplicationInput{
		ApplicationId: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeregisterApplication(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMSAP, create.ErrActionDeleting, ResNameApplication, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitApplicationDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSMSAP, create.ErrActionWaitingForDeletion, ResNameApplication, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}


func (r *resourceApplication) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitApplicationCreated(ctx context.Context, conn *ssmsap.Client, id string, timeout time.Duration) (*awstypes.Application, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(awstypes.ApplicationStatusRegistering), string(awstypes.ApplicationStatusStarting)},
		Target:                    []string{string(awstypes.ApplicationStatusActivated)},
		Refresh:                   statusApplication(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Application); ok {
		return out, err
	}

	return nil, err
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitApplicationUpdated(ctx context.Context, conn *ssmsap.Client, id string, timeout time.Duration) (*awstypes.Application, error) {
	// stateConf := &retry.StateChangeConf{
	// 	Pending:                   []string{statusChangePending},
	// 	Target:                    []string{statusUpdated},
	// 	Refresh:                   statusApplication(ctx, conn, id),
	// 	Timeout:                   timeout,
	// 	NotFoundChecks:            20,
	// 	ContinuousTargetOccurence: 2,
	// }

	// outputRaw, err := stateConf.WaitForStateContext(ctx)
	// if out, ok := outputRaw.(*ssmsap.Application); ok {
	// 	return out.Application, err //TODO: get all application data and return
	// }

	return nil, nil
}

func waitApplicationDeleted(ctx context.Context, conn *ssmsap.Client, id string, timeout time.Duration) (*awstypes.Application, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(awstypes.ApplicationStatusDeleting), string(awstypes.ApplicationStatusStopping)},
		Target:                    []string{},
		Refresh:                   statusApplication(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func statusApplication(ctx context.Context, conn *ssmsap.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, _, err := findApplicationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}


func flattenCredential(ctx context.Context, apiObject *awstypes.ApplicationCredential) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: credentialsAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"credential_type": flex.StringValueToFramework(ctx, apiObject.CredentialType),
		"database_name":   flex.StringValueToFramework(ctx, *apiObject.DatabaseName),
		"secret_id":       flex.StringValueToFramework(ctx, *apiObject.SecretId),
	}
	objVal, d := types.ObjectValue(credentialsAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenCredentials(ctx context.Context, apiObjects []*awstypes.ApplicationCredential) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: credentialsAttrTypes}

	if len(apiObjects) == 0 {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		obj := map[string]attr.Value{
			"credential_type": flex.StringValueToFramework(ctx, apiObject.CredentialType),
			"database_name":   flex.StringValueToFramework(ctx, *apiObject.DatabaseName),
			"secret_id":       flex.StringValueToFramework(ctx, *apiObject.SecretId),
		}
		objVal, d := types.ObjectValue(credentialsAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}

func expandCredential(tfList []credentialsData, sapSystemId string) *awstypes.ApplicationCredential {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	apiObject := &awstypes.ApplicationCredential{
		CredentialType: awstypes.CredentialType(tfObj.CredentialType.ValueString()),
		DatabaseName:   aws.String(sapSystemId + "/" + tfObj.DatabaseName.ValueString()),
		SecretId:       aws.String(tfObj.SecretId.ValueString()),
	}

	return apiObject
}

func expandCredentials(tfList []credentialsData, sapSystemId string) []awstypes.ApplicationCredential {
	if len(tfList) == 0 {
		return nil
	}
	var apiObject []awstypes.ApplicationCredential
	
	for _, tfObj := range tfList {
		item := awstypes.ApplicationCredential{
			CredentialType: awstypes.CredentialType(tfObj.CredentialType.ValueString()),
			DatabaseName:   aws.String(sapSystemId + "/" + tfObj.DatabaseName.ValueString()),
			SecretId:       aws.String(tfObj.SecretId.ValueString()),
		}

		apiObject = append(apiObject, item)
	}

	return apiObject
}

func (r *resourceApplication) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

// refreshFromOutput writes state data from an AWS response object
func (state *resourceApplicationData) refreshFromOutput(ctx context.Context, conn *ssmsap.Client, application *awstypes.Application, tags map[string]string) error {
	if application == nil {
		return fmt.Errorf("application is not set")
	}

	state.ARN = flex.StringToFramework(ctx, application.Arn)
	state.ID = flex.StringToFramework(ctx, application.Id)
	state.ApplicationType = flex.StringToFramework(ctx, (*string)(&application.Type))

	if len(application.Components) > 1 { //only one component is expected at the moment to keep it simple first
		return fmt.Errorf("more than one component is found for application")

	}

	//check if component name match the following format "<SID>-<SID><InstanceNumber>"
	component_name := application.Components[0]
	if !regexp.MustCompile(`^[A-Z0-9]{3}-[A-Z0-9]{3}[0-9]{2}$`).MatchString(component_name) {
		return fmt.Errorf("component name does not match the following format <SID>-<SID><InstanceNumber>")
	}
	sapInstanceNumber := component_name[7:9]
	sid := component_name[0:3]

	state.SapInstanceNumber = flex.StringToFramework(ctx, &sapInstanceNumber)
	state.Sid = flex.StringToFramework(ctx, &sid)

	component, tags, err := findComponentByID(ctx, conn, *application.Id, component_name)

	if err != nil {
		return err
	}

	//extract credentials from databases
	credentials := []*awstypes.ApplicationCredential{}

	for _, database := range component.Databases {
		in := &ssmsap.GetDatabaseInput{
			ApplicationId: aws.String(*application.Id),
			ComponentId:   aws.String(*component.ComponentId),
			DatabaseId:    aws.String(database),
		}
		db_out, err := conn.GetDatabase(ctx, in)
		if err != nil {
			return err
		}
		if len(db_out.Database.Credentials) > 1 {
			return fmt.Errorf("more than one credential is per db returned")
		}
		cred := db_out.Database.Credentials[0]
		credentials = append(credentials, &cred)
	}

	state.Credentials, _ = flattenCredentials(ctx, credentials)

	//get instances from component
	state.Instances = flex.FlattenFrameworkStringList(ctx, []*string{component.PrimaryHost})
	
	setTagsOut(ctx, tags)

	return nil
}

