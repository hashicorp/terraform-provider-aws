// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fis

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

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
	// using the services/fis/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fis"
	"github.com/aws/aws-sdk-go-v2/service/fis/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// TIP: ==== FILE STRUCTURE ====
// All resources should follow this basic outline. Improve this resource's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main resource function with schema
// 4. Create, read, update, delete functions (in that order)
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_fis_target_account_configuration", name="Target Account Configuration")
func ResourceTargetAccountConfiguration() *schema.Resource {
	return &schema.Resource{
		// TIP: ==== ASSIGN CRUD FUNCTIONS ====
		// These 4 functions handle CRUD responsibilities below.
		CreateWithoutTimeout: resourceTargetAccountConfigurationCreate,
		ReadWithoutTimeout:   resourceTargetAccountConfigurationRead,
		UpdateWithoutTimeout: resourceTargetAccountConfigurationUpdate,
		DeleteWithoutTimeout: resourceTargetAccountConfigurationDelete,

		// TIP: ==== TERRAFORM IMPORTING ====
		// If Read can get all the information it needs from the Identifier
		// (i.e., d.Id()), you can use the Passthrough importer. Otherwise,
		// you'll need a custom import function.
		//
		// See more:
		// https://hashicorp.github.io/terraform-provider-aws/add-import-support/
		// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#implicit-state-passthrough
		// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#virtual-attributes
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// TIP: ==== CONFIGURABLE TIMEOUTS ====
		// Users can configure timeout lengths but you need to use the times they
		// provide. Access the timeout they configure (or the defaults) using,
		// e.g., d.Timeout(schema.TimeoutCreate) (see below). The times here are
		// the defaults if they don't configure timeouts.
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

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
		//    use this combination if you are using Default or DefaultFunc
		// Optional: true,
		// Computed: true,
		//
		// You will typically find arguments in the input struct
		// (e.g., CreateDBInstanceInput) for the create operation. Sometimes
		// they are only in the input struct (e.g., ModifyDBInstanceInput) for
		// the modify operation.
		//
		// For more about schema options, visit
		// https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema#Schema
		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidAwsAccountID,
				ForceNew:     true,
			},
			"experiment_template_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
				ForceNew:     true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
				ForceNew:     true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
		},
	}
}

const (
	ResNameTargetAccountConfiguration = "Target Account Configuration"
)

func resourceTargetAccountConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).FISClient(ctx)

	in := &fis.CreateTargetAccountConfigurationInput{
		AccountId:            aws.String(d.Get("account_id").(string)),
		ExperimentTemplateId: aws.String(d.Get("experiment_template_id").(string)),
		RoleArn:              aws.String(d.Get("role_arn").(string)),
		ClientToken:          aws.String(sdkid.UniqueId()),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	out, err := conn.CreateTargetAccountConfiguration(ctx, in)
	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Get("experiment_template_id").(string))
	}

	if out == nil || out.TargetAccountConfiguration == nil {
		return smerr.Append(ctx, diags, tfresource.NewEmptyResultError(in), smerr.ID, d.Get("experiment_template_id").(string))
	}

	tac := out.TargetAccountConfiguration
	templateID := aws.ToString(tac.ExperimentTemplateId)
	if templateID == "" {
		templateID = d.Get("experiment_template_id").(string)
	}
	accountID := aws.ToString(tac.AccountId)
	if accountID == "" {
		accountID = d.Get("account_id").(string)
	}

	d.SetId(fmt.Sprintf("%s/%s", templateID, accountID))
	d.Set("description", aws.ToString(tac.Description))
	d.Set("role_arn", aws.ToString(tac.RoleArn))

	return append(diags, resourceTargetAccountConfigurationRead(ctx, d, meta)...)
}

func resourceTargetAccountConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).FISClient(ctx)

	experimentTemplateId, accountId, err := parseTargetAccountConfigurationID(d.Id())
	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	in := &fis.GetTargetAccountConfigurationInput{
		AccountId:            aws.String(accountId),
		ExperimentTemplateId: aws.String(experimentTemplateId),
	}

	out, err := conn.GetTargetAccountConfiguration(ctx, in)
	if errs.IsA[*types.ResourceNotFoundException](err) {
		if !d.IsNewResource() {
			log.Printf("[WARN] FIS TargetAccountConfiguration (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}
	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	if out == nil || out.TargetAccountConfiguration == nil {
		return smerr.Append(ctx, diags, tfresource.NewEmptyResultError(in), smerr.ID, d.Id())
	}

	tac := out.TargetAccountConfiguration
	accountValue := accountId
	if v := aws.ToString(tac.AccountId); v != "" {
		accountValue = v
	}

	roleARN := d.Get("role_arn").(string)
	if v := aws.ToString(tac.RoleArn); v != "" {
		roleARN = v
	}

	d.Set("account_id", accountValue)
	d.Set("experiment_template_id", experimentTemplateId)
	d.Set("role_arn", roleARN)
	d.Set("description", aws.ToString(tac.Description))

	return diags
}

func resourceTargetAccountConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	if !d.HasChange("description") {
		return diags
	}

	conn := meta.(*conns.AWSClient).FISClient(ctx)

	experimentTemplateId, accountId, err := parseTargetAccountConfigurationID(d.Id())
	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	in := &fis.CreateTargetAccountConfigurationInput{
		AccountId:            aws.String(accountId),
		ExperimentTemplateId: aws.String(experimentTemplateId),
		RoleArn:              aws.String(d.Get("role_arn").(string)),
		ClientToken:          aws.String(sdkid.UniqueId()),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	_, err = conn.CreateTargetAccountConfiguration(ctx, in)
	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	return append(diags, resourceTargetAccountConfigurationRead(ctx, d, meta)...)
}

func resourceTargetAccountConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).FISClient(ctx)

	experimentTemplateId, accountId, err := parseTargetAccountConfigurationID(d.Id())
	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	log.Printf("[INFO] Deleting FIS TargetAccountConfiguration %s", d.Id())

	_, err = conn.DeleteTargetAccountConfiguration(ctx, &fis.DeleteTargetAccountConfigurationInput{
		AccountId:            aws.String(accountId),
		ExperimentTemplateId: aws.String(experimentTemplateId),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		d.SetId("")
		return diags
	}
	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	d.SetId("")

	return diags
}

// parseTargetAccountConfigurationID parses the composite ID into experiment template ID and account ID
func parseTargetAccountConfigurationID(id string) (experimentTemplateId, accountId string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		err = fmt.Errorf("invalid ID format: expected experiment_template_id/account_id, got %s", id)
		return
	}
	experimentTemplateId = parts[0]
	accountId = parts[1]
	return
}

// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., amp.WorkspaceStatusCodeActive).
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
func waitTargetAccountConfigurationCreated(ctx context.Context, conn *fis.Client, id string, timeout time.Duration) (*types.TargetAccountConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusTargetAccountConfiguration(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.TargetAccountConfiguration); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitTargetAccountConfigurationUpdated(ctx context.Context, conn *fis.Client, id string, timeout time.Duration) (*types.TargetAccountConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusTargetAccountConfiguration(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.TargetAccountConfiguration); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitTargetAccountConfigurationDeleted(ctx context.Context, conn *fis.Client, id string, timeout time.Duration) (*types.TargetAccountConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusTargetAccountConfiguration(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.TargetAccountConfiguration); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.
func statusTargetAccountConfiguration(ctx context.Context, conn *fis.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findTargetAccountConfigurationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		status := ""
		if out != nil {
			status = string(out.Status)
		}

		return out, status, nil
	}
}

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
func findTargetAccountConfigurationByID(ctx context.Context, conn *fis.Client, id string) (*types.TargetAccountConfiguration, error) {
	experimentTemplateID, accountID, err := parseTargetAccountConfigurationID(id)
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	in := &fis.GetTargetAccountConfigurationInput{
		AccountId:            aws.String(accountID),
		ExperimentTemplateId: aws.String(experimentTemplateID),
	}

	out, err := conn.GetTargetAccountConfiguration(ctx, in)
	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.TargetAccountConfiguration == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(in))
	}

	return out.TargetAccountConfiguration, nil
}

// TIP: ==== FLEX ====
// Flatteners and expanders ("flex" functions) help handle complex data
// types. Flatteners take an API data type and return something you can use in
// a d.Set() call. In other words, flatteners translate from AWS -> Terraform.
//
// On the other hand, expanders take a Terraform data structure and return
// something that you can send to the AWS API. In other words, expanders
// translate from Terraform -> AWS.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/
func flattenComplexArgument(apiObject *fis.ComplexArgument) map[string]any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{}

	if v := apiObject.SubFieldOne; v != nil {
		m["sub_field_one"] = aws.ToString(v)
	}

	if v := apiObject.SubFieldTwo; v != nil {
		m["sub_field_two"] = aws.ToString(v)
	}

	return m
}

// TIP: Often the AWS API will return a slice of structures in response to a
// request for information. Sometimes you will have set criteria (e.g., the ID)
// that means you'll get back a one-length slice. This plural function works
// brilliantly for that situation too.
func flattenComplexArguments(apiObjects []*fis.ComplexArgument) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []any

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		l = append(l, flattenComplexArgument(apiObject))
	}

	return l
}

// TIP: Remember, as mentioned above, expanders take a Terraform data structure
// and return something that you can send to the AWS API. In other words,
// expanders translate from Terraform -> AWS.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/
func expandComplexArgument(tfMap map[string]any) *fis.ComplexArgument {
	if tfMap == nil {
		return nil
	}

	a := &fis.ComplexArgument{}

	if v, ok := tfMap["sub_field_one"].(string); ok && v != "" {
		a.SubFieldOne = aws.String(v)
	}

	if v, ok := tfMap["sub_field_two"].(string); ok && v != "" {
		a.SubFieldTwo = aws.String(v)
	}

	return a
}

// TIP: Even when you have a list with max length of 1, this plural function
// works brilliantly. However, if the AWS API takes a structure rather than a
// slice of structures, you will not need it.
func expandComplexArguments(tfList []any) []*fis.ComplexArgument {
	// TIP: The AWS API can be picky about whether you send a nil or zero-
	// length for an argument that should be cleared. For example, in some
	// cases, if you send a nil value, the AWS API interprets that as "make no
	// changes" when what you want to say is "remove everything." Sometimes
	// using a zero-length list will cause an error.
	//
	// As a result, here are two options. Usually, option 1, nil, will work as
	// expected, clearing the field. But, test going from something to nothing
	// to make sure it works. If not, try the second option.

	// TIP: Option 1: Returning nil for zero-length list
	if len(tfList) == 0 {
		return nil
	}

	var s []*fis.ComplexArgument

	// TIP: Option 2: Return zero-length list for zero-length list. If option 1 does
	// not work, after testing going from something to nothing (if that is
	// possible), uncomment out the next line and remove option 1.
	//
	// s := make([]*fis.ComplexArgument, 0)

	for _, r := range tfList {
		m, ok := r.(map[string]any)

		if !ok {
			continue
		}

		a := expandComplexArgument(m)

		if a == nil {
			continue
		}

		s = append(s, a)
	}

	return s
}
