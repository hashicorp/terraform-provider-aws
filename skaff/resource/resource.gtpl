// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package {{ .ServicePackage }}

{{- if .IncludeComments }}
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
{{- end }}

import (
{{- if .IncludeComments }}
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/{{ .SDKPackage }}/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
{{- end }}
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/{{ .SDKPackage }}"
	"github.com/aws/aws-sdk-go-v2/service/{{ .SDKPackage }}/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
{{- if .IncludeTags }}
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
{{- end }}
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)
{{ if .IncludeComments }}
// TIP: ==== FILE STRUCTURE ====
// All resources should follow this basic outline. Improve this resource's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main resource function with schema
// 4. Create, read, update, delete functions (in that order)
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)
{{- end }}

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("{{ .ProviderResourceName }}", name="{{ .HumanResourceName }}")
{{- if .IncludeTags }}
// Tagging annotations are used for "transparent tagging".
// Change the "identifierAttribute" value to the name of the attribute used in ListTags and UpdateTags calls (e.g. "arn").
// @Tags(identifierAttribute="id")
{{- end }}
func Resource{{ .Resource }}() *schema.Resource {
	return &schema.Resource{
		{{- if .IncludeComments }}
		// TIP: ==== ASSIGN CRUD FUNCTIONS ====
		// These 4 functions handle CRUD responsibilities below.
		{{- end }}
		CreateWithoutTimeout: resource{{ .Resource }}Create,
		ReadWithoutTimeout:   resource{{ .Resource }}Read,
		UpdateWithoutTimeout: resource{{ .Resource }}Update,
		DeleteWithoutTimeout: resource{{ .Resource }}Delete,
		{{ if .IncludeComments }}
		// TIP: ==== TERRAFORM IMPORTING ====
		// If Read can get all the information it needs from the Identifier
		// (i.e., d.Id()), you can use the Passthrough importer. Otherwise,
		// you'll need a custom import function.
		//
		// See more:
		// https://hashicorp.github.io/terraform-provider-aws/add-import-support/
		// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#implicit-state-passthrough
		// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#virtual-attributes
		{{- end }}
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		{{ if .IncludeComments }}
		// TIP: ==== CONFIGURABLE TIMEOUTS ====
		// Users can configure timeout lengths but you need to use the times they
		// provide. Access the timeout they configure (or the defaults) using,
		// e.g., d.Timeout(schema.TimeoutCreate) (see below). The times here are
		// the defaults if they don't configure timeouts.
		{{- end }}
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		{{ if .IncludeComments }}
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
		{{- end }}
		Schema: map[string]*schema.Schema{
			"arn": { {{- if .IncludeComments }} // TIP: Many, but not all, resources have an `arn` attribute.{{- end }}
				Type:     schema.TypeString,
				Computed: true,
			},
			"replace_with_arguments": { {{- if .IncludeComments }} // TIP: Add all your arguments and attributes.{{- end }}
				Type:     schema.TypeString,
				Optional: true,
			},
			"complex_argument": { {{- if .IncludeComments }} // TIP: See setting, getting, flattening, expanding examples below for this complex argument.{{- end }}
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sub_field_one": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
						"sub_field_two": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			{{- if .IncludeTags }}
			names.AttrTags:    tftags.TagsSchema(), {{- if .IncludeComments }} // TIP: Many, but not all, resources have `tags` and `tags_all` attributes.{{- end }}
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			{{- end }}
		},
		{{- if .IncludeTags }}
		CustomizeDiff: verify.SetTagsDiff,
		{{- end }}
	}
}

const (
	ResName{{ .Resource }} = "{{ .HumanResourceName }}"
)

func resource{{ .Resource }}Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	{{- if .IncludeComments }}
	// TIP: ==== RESOURCE CREATE ====
	// Generally, the Create function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Populate a create input structure
	// 3. Call the AWS create/put function
	// 4. Using the output from the create function, set the minimum arguments
	//    and attributes for the Read function to work. At a minimum, set the
	//    resource ID. E.g., d.SetId(<Identifier, such as AWS ID or ARN>)
	// 5. Use a waiter to wait for create to complete
	// 6. Call the Read function in the Create return
	{{- end }}
	{{- if .IncludeComments }}

	// TIP: -- 1. Get a client connection to the relevant service
	{{- end }}
	conn := meta.(*conns.AWSClient).{{ .Service }}Client(ctx)
	{{ if .IncludeComments }}
	// TIP: -- 2. Populate a create input structure
	{{- end }}
	in := &{{ .SDKPackage }}.Create{{ .Resource }}Input{
		{{- if .IncludeComments }}
		// TIP: Mandatory or fields that will always be present can be set when
		// you create the Input structure. (Replace these with real fields.)
		{{- end }}
		{{ .Resource }}Name: aws.String(d.Get("name").(string)),
		{{ .Resource }}Type: aws.String(d.Get("type").(string)),
		{{ if .IncludeComments }}
		// TIP: Not all resources support tags and tags don't always make sense. If
		// your resource doesn't need tags, you can remove the tags lines here and
		// below. Many resources do include tags so this a reminder to include them
		// where possible.
		{{- end }}
		{{- if .IncludeTags }}
		Tags: getTagsIn(ctx),
		{{- end }}
	}

	if v, ok := d.GetOk("max_size"); ok {
		{{- if .IncludeComments }}
		// TIP: Optional fields should be set based on whether or not they are
		// used.
		{{- end }}
		in.MaxSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("complex_argument"); ok && len(v.([]interface{})) > 0 {
		{{- if .IncludeComments }}
		// TIP: Use an expander to assign a complex argument.
		{{- end }}
		in.ComplexArguments = expandComplexArguments(v.([]interface{}))
	}

	{{ if .IncludeComments }}
	// TIP: -- 3. Call the AWS create function
	{{- end }}
	out, err := conn.Create{{ .Resource }}(ctx, in)
	if err != nil {
		{{- if .IncludeComments }}
		// TIP: Since d.SetId() has not been called yet, you cannot use d.Id()
		// in error messages at this point.
		{{- end }}
		return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionCreating, ResName{{ .Resource }}, d.Get("name").(string), err)
	}

	if out == nil || out.{{ .Resource }} == nil {
		return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionCreating, ResName{{ .Resource }}, d.Get("name").(string), errors.New("empty output"))
	}
	{{ if .IncludeComments }}
	// TIP: -- 4. Set the minimum arguments and/or attributes for the Read function to
	// work.
	{{- end }}
	d.SetId(aws.ToString(out.{{ .Resource }}.{{ .Resource }}ID))
	{{ if .IncludeComments }}
	// TIP: -- 5. Use a waiter to wait for create to complete
	{{- end }}
	if _, err := wait{{ .Resource }}Created(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionWaitingForCreation, ResName{{ .Resource }}, d.Id(), err)
	}
	{{ if .IncludeComments }}
	// TIP: -- 6. Call the Read function in the Create return
	{{- end }}
	return append(diags, resource{{ .Resource }}Read(ctx, d, meta)...)
}

func resource{{ .Resource }}Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	{{- if .IncludeComments }}
	// TIP: ==== RESOURCE READ ====
	// Generally, the Read function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Get the resource from AWS
	// 3. Set ID to empty where resource is not new and not found
	// 4. Set the arguments and attributes
	// 5. Set the tags
	// 6. Return diags
	{{- end }}
	{{- if .IncludeComments }}

	// TIP: -- 1. Get a client connection to the relevant service
	{{- end }}
	conn := meta.(*conns.AWSClient).{{ .Service }}Client(ctx)
	{{ if .IncludeComments }}
	// TIP: -- 2. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	{{- end }}
	out, err := find{{ .Resource }}ByID(ctx, conn, d.Id())
	{{ if .IncludeComments }}
	// TIP: -- 3. Set ID to empty where resource is not new and not found
	{{- end }}
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] {{ .Service }} {{ .Resource }} (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionReading, ResName{{ .Resource }}, d.Id(), err)
	}
	{{ if .IncludeComments }}
	// TIP: -- 4. Set the arguments and attributes
	//
	// For simple data types (i.e., schema.TypeString, schema.TypeBool,
	// schema.TypeInt, and schema.TypeFloat), a simple Set call (e.g.,
	// d.Set("arn", out.Arn) is sufficient. No error or nil checking is
	// necessary.
	//
	// However, there are some situations where more handling is needed.
	// a. Complex data types (e.g., schema.TypeList, schema.TypeSet)
	// b. Where errorneous diffs occur. For example, a schema.TypeString may be
	//    a JSON. AWS may return the JSON in a slightly different order but it
	//    is equivalent to what is already set. In that case, you may check if
	//    it is equivalent before setting the different JSON.
	{{- end }}
	d.Set("arn", out.Arn)
	d.Set("name", out.Name)
	{{ if .IncludeComments }}
	// TIP: Setting a complex type.
	// For more information, see:
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#data-handling-and-conversion
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#flatten-functions-for-blocks
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#root-typeset-of-resource-and-aws-list-of-structure
	{{- end }}
	if err := d.Set("complex_argument", flattenComplexArguments(out.ComplexArguments)); err != nil {
		return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionSetting, ResName{{ .Resource }}, d.Id(), err)
	}
	{{ if .IncludeComments }}
	// TIP: Setting a JSON string to avoid errorneous diffs.
	{{- end }}
	p, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.ToString(out.Policy))
	if err != nil {
		return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionSetting, ResName{{ .Resource }}, d.Id(), err)
	}

	p, err = structure.NormalizeJsonString(p)
	if err != nil {
		return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionSetting, ResName{{ .Resource }}, d.Id(), err)
	}

	d.Set("policy", p)

	{{ if .IncludeComments }}
	// TIP: -- 6. Return diags
	{{- end }}
	return diags
}

func resource{{ .Resource }}Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	{{- if .IncludeComments }}
	// TIP: ==== RESOURCE UPDATE ====
	// Not all resources have Update functions. There are a few reasons:
	// a. The AWS API does not support changing a resource
	// b. All arguments have ForceNew: true, set
	// c. The AWS API uses a create call to modify an existing resource
	//
	// In the cases of a. and b., the main resource function will not have a
	// UpdateWithoutTimeout defined. In the case of c., Update and Create are
	// the same.
	//
	// The rest of the time, there should be an Update function and it should
	// do the following things. Make sure there is a good reason if you don't
	// do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Populate a modify input structure and check for changes
	// 3. Call the AWS modify/update function
	// 4. Use a waiter to wait for update to complete
	// 5. Call the Read function in the Update return
	{{- end }}
	{{- if .IncludeComments }}

	// TIP: -- 1. Get a client connection to the relevant service
	{{- end }}
	conn := meta.(*conns.AWSClient).{{ .Service }}Client(ctx)
	{{ if .IncludeComments }}
	// TIP: -- 2. Populate a modify input structure and check for changes
	//
	// When creating the input structure, only include mandatory fields. Other
	// fields are set as needed. You can use a flag, such as update below, to
	// determine if a certain portion of arguments have been changed and
	// whether to call the AWS update function.
	{{- end }}
	update := false

	in := &{{ .ServiceLower }}.Update{{ .Resource }}Input{
		Id: aws.String(d.Id()),
	}

	if d.HasChanges("an_argument") {
		in.AnArgument = aws.String(d.Get("an_argument").(string))
		update = true
	}

	if !update {
		{{- if .IncludeComments }}
		// TIP: If update doesn't do anything at all, which is rare, you can
		// return diags. Otherwise, return a read call, as below.
		{{- end }}
		return diags
	}
	{{ if .IncludeComments }}
	// TIP: -- 3. Call the AWS modify/update function
	{{- end }}
	log.Printf("[DEBUG] Updating {{ .Service }} {{ .Resource }} (%s): %#v", d.Id(), in)
	out, err := conn.Update{{ .Resource }}(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionUpdating, ResName{{ .Resource }}, d.Id(), err)
	}
	{{ if .IncludeComments }}
	// TIP: -- 4. Use a waiter to wait for update to complete
	{{- end }}
	if _, err := wait{{ .Resource }}Updated(ctx, conn, aws.ToString(out.OperationId), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionWaitingForUpdate, ResName{{ .Resource }}, d.Id(), err)
	}
	{{ if .IncludeComments }}
	// TIP: -- 5. Call the Read function in the Update return
	{{- end }}
	return append(diags, resource{{ .Resource }}Read(ctx, d, meta)...)
}

func resource{{ .Resource }}Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	{{- if .IncludeComments }}
	// TIP: ==== RESOURCE DELETE ====
	// Most resources have Delete functions. There are rare situations
	// where you might not need a delete:
	// a. The AWS API does not provide a way to delete the resource
	// b. The point of your resource is to perform an action (e.g., reboot a
	//    server) and deleting serves no purpose.
	//
	// The Delete function should do the following things. Make sure there
	// is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Populate a delete input structure
	// 3. Call the AWS delete function
	// 4. Use a waiter to wait for delete to complete
	// 5. Return diags
	{{- end }}
	{{- if .IncludeComments }}

	// TIP: -- 1. Get a client connection to the relevant service
	{{- end }}
	conn := meta.(*conns.AWSClient).{{ .Service }}Client(ctx)
	{{ if .IncludeComments }}
	// TIP: -- 2. Populate a delete input structure
	{{- end }}
	log.Printf("[INFO] Deleting {{ .Service }} {{ .Resource }} %s", d.Id())
	{{ if .IncludeComments }}
	// TIP: -- 3. Call the AWS delete function
	{{- end }}
	_, err := conn.Delete{{ .Resource }}(ctx, &{{ .ServiceLower }}.Delete{{ .Resource }}Input{
		Id: aws.String(d.Id()),
	})
	{{ if .IncludeComments }}
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	{{- end }}
	if errs.IsA[*types.ResourceNotFoundException](err){
		return diags
	}
	if err != nil {
		return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionDeleting, ResName{{ .Resource }}, d.Id(), err)
	}
	{{ if .IncludeComments }}
	// TIP: -- 4. Use a waiter to wait for delete to complete
	{{- end }}
	if _, err := wait{{ .Resource }}Deleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.{{ .Service }}, create.ErrActionWaitingForDeletion, ResName{{ .Resource }}, d.Id(), err)
	}
	{{ if .IncludeComments }}
	// TIP: -- 5. Return diags
	{{- end }}
	return diags
}
{{ if .IncludeComments }}
// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., amp.WorkspaceStatusCodeActive).
{{- end }}
const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)
{{ if .IncludeComments }}
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
{{- end }}
func wait{{ .Resource }}Created(ctx context.Context, conn *{{ .ServiceLower }}.Client, id string, timeout time.Duration) (*{{ .ServiceLower }}.{{ .Resource }}, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   status{{ .Resource }}(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*{{ .ServiceLower }}.{{ .Resource }}); ok {
		return out, err
	}

	return nil, err
}
{{ if .IncludeComments }}
// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
{{- end }}
func wait{{ .Resource }}Updated(ctx context.Context, conn *{{ .ServiceLower }}.Client, id string, timeout time.Duration) (*{{ .ServiceLower }}.{{ .Resource }}, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   status{{ .Resource }}(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*{{ .ServiceLower }}.{{ .Resource }}); ok {
		return out, err
	}

	return nil, err
}
{{ if .IncludeComments }}
// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
{{- end }}
func wait{{ .Resource }}Deleted(ctx context.Context, conn *{{ .ServiceLower }}.Client, id string, timeout time.Duration) (*{{ .ServiceLower }}.{{ .Resource }}, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusDeleting, statusNormal},
		Target:                    []string{},
		Refresh:                   status{{ .Resource }}(ctx, conn, id),
		Timeout:                   timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*{{ .ServiceLower }}.{{ .Resource }}); ok {
		return out, err
	}

	return nil, err
}
{{ if .IncludeComments }}
// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.
{{- end }}
func status{{ .Resource }}(ctx context.Context, conn *{{ .ServiceLower }}.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := find{{ .Resource }}ByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.Status), nil
	}
}
{{ if .IncludeComments }}
// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
{{- end }}
func find{{ .Resource }}ByID(ctx context.Context, conn *{{ .ServiceLower }}.Client, id string) (*{{ .ServiceLower }}.{{ .Resource }}, error) {
	in := &{{ .ServiceLower }}.Get{{ .Resource }}Input{
		Id: aws.String(id),
	}

	out, err := conn.Get{{ .Resource }}(ctx, in)
	if errs.IsA[*types.ResourceNotFoundException](err){
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.{{ .Resource }} == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.{{ .Resource }}, nil
}
{{ if .IncludeComments }}
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
{{- end }}
func flattenComplexArgument(apiObject *{{ .ServiceLower }}.ComplexArgument) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.SubFieldOne; v != nil {
		m["sub_field_one"] = aws.ToString(v)
	}

	if v := apiObject.SubFieldTwo; v != nil {
		m["sub_field_two"] = aws.ToString(v)
	}

	return m
}
{{ if .IncludeComments }}
// TIP: Often the AWS API will return a slice of structures in response to a
// request for information. Sometimes you will have set criteria (e.g., the ID)
// that means you'll get back a one-length slice. This plural function works
// brilliantly for that situation too.
{{- end }}
func flattenComplexArguments(apiObjects []*{{ .ServiceLower }}.ComplexArgument) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		l = append(l, flattenComplexArgument(apiObject))
	}

	return l
}
{{ if .IncludeComments }}
// TIP: Remember, as mentioned above, expanders take a Terraform data structure
// and return something that you can send to the AWS API. In other words,
// expanders translate from Terraform -> AWS.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/
{{- end }}
func expandComplexArgument(tfMap map[string]interface{}) *{{ .ServiceLower }}.ComplexArgument {
	if tfMap == nil {
		return nil
	}

	a := &{{ .ServiceLower }}.ComplexArgument{}

	if v, ok := tfMap["sub_field_one"].(string); ok && v != "" {
		a.SubFieldOne = aws.String(v)
	}

	if v, ok := tfMap["sub_field_two"].(string); ok && v != "" {
		a.SubFieldTwo = aws.String(v)
	}

	return a
}
{{ if .IncludeComments }}
// TIP: Even when you have a list with max length of 1, this plural function
// works brilliantly. However, if the AWS API takes a structure rather than a
// slice of structures, you will not need it.
{{- end }}
func expandComplexArguments(tfList []interface{}) []*{{ .ServiceLower }}.ComplexArgument {
	{{- if .IncludeComments }}
	// TIP: The AWS API can be picky about whether you send a nil or zero-
	// length for an argument that should be cleared. For example, in some
	// cases, if you send a nil value, the AWS API interprets that as "make no
	// changes" when what you want to say is "remove everything." Sometimes
	// using a zero-length list will cause an error.
	//
	// As a result, here are two options. Usually, option 1, nil, will work as
	// expected, clearing the field. But, test going from something to nothing
	// to make sure it works. If not, try the second option.
	{{- end }}
	{{ if .IncludeComments }}
	// TIP: Option 1: Returning nil for zero-length list
    if len(tfList) == 0 {
        return nil
    }
	{{- end }}

    var s []*{{ .ServiceLower }}.ComplexArgument
	{{ if .IncludeComments }}
	// TIP: Option 2: Return zero-length list for zero-length list. If option 1 does
	// not work, after testing going from something to nothing (if that is
	// possible), uncomment out the next line and remove option 1.
	//
	{{- end }}
	// s := make([]*{{ .ServiceLower }}.ComplexArgument, 0)

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

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
