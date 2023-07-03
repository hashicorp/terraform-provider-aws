package codecatalyst


import (

	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codecatalyst"
	"github.com/aws/aws-sdk-go-v2/service/codecatalyst/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
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
// @SDKResource("aws_codecatalyst_devenvironment", name="Devenvironment")
// Tagging annotations are used for "transparent tagging".
// Change the "identifierAttribute" value to the name of the attribute used in ListTags and UpdateTags calls (e.g. "arn").
// @Tags(identifierAttribute="id")
func ResourceDevenvironment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDevenvironmentCreate,
		ReadWithoutTimeout:   resourceDevenvironmentRead,
		UpdateWithoutTimeout: resourceDevenvironmentUpdate,
		DeleteWithoutTimeout: resourceDevenvironmentDelete,
		

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		
		Schema: map[string]*schema.Schema{
			"arn": { 
				Type:     schema.TypeString,
				Computed: true,
			},

			"alias": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ides": { 
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Optional: true,
						},
						"runtime": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"space_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"inactivity_timeout_minutes":{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"instance_type":{
				Type:     schema.TypeString,
				Required: true,
			},
			"persistent_storage": { 
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"repositories": { 
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"branch_name": {
							Type:         schema.TypeString,
							Optional: true,
						},
						"repository_name": {
							Type:         schema.TypeString,
							Required: true,
						},
					},
				},
			},


			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameDevenvironment = "Devenvironment"
)

func resourceDevenvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodeCatalystClient(ctx)
	
	storage := expandPersistentStorage(d.Get("persistant_storage").([]interface{})[0].(map[string]interface{}))
	instanceType := types.InstanceType(d.Get("instance_type").(string))

	in := &codecatalyst.CreateDevEnvironmentInput{
		ProjectName: aws.String(d.Get("project_name").(string)),
		SpaceName:  aws.String(d.Get("space_name").(string)),
		PersistentStorage: storage,
		InstanceType: instanceType,

	}

	if v, ok := d.GetOk("inactivity_timeout_minutes"); ok {
		in.InactivityTimeoutMinutes = int32(v.(int))
	}

	if v, ok := d.GetOk("alias"); ok {
		in.Alias = aws.String(v.(string))
	}

	out, err := conn.CreateDevEnvironment(ctx, in)

	if err != nil {

		return append(diags, create.DiagError(names.CodeCatalyst, create.ErrActionCreating, ResNameDevenvironment, d.Get("name").(string), err)...)
	}

	if out == nil || out == nil {
		return append(diags, create.DiagError(names.CodeCatalyst, create.ErrActionCreating, ResNameDevenvironment, d.Get("name").(string), errors.New("empty output"))...)
	}
	

	d.SetId(aws.ToString(out.Id))
	
	if _, err := waitDevenvironmentCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return append(diags, create.DiagError(names.CodeCatalyst, create.ErrActionWaitingForCreation, ResNameDevenvironment, d.Id(), err)...)
	}
	
	return append(diags, resourceDevenvironmentRead(ctx, d, meta)...)
}

func resourceDevenvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CodeCatalystClient(ctx)
	
	// TIP: -- 2. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findDevenvironmentByID(ctx, conn, d.Id())
	
	// TIP: -- 3. Set ID to empty where resource is not new and not found
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Codecatalyst Devenvironment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return append(diags, create.DiagError(names.CodeCatalyst, create.ErrActionReading, ResNameDevenvironment, d.Id(), err)...)
	}
	
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
	d.Set("arn", out.Arn)
	d.Set("name", out.Name)
	
	if err := d.Set("complex_argument", flattenComplexArguments(out.ComplexArguments)); err != nil {
		return append(diags, create.DiagError(names.CodeCatalyst, create.ErrActionSetting, ResNameDevenvironment, d.Id(), err)...)
	}
	
	// TIP: Setting a JSON string to avoid errorneous diffs.
	p, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.ToString(out.Policy))
	if err != nil {
		return append(diags, create.DiagError(names.CodeCatalyst, create.ErrActionSetting, ResNameDevenvironment, d.Id(), err)...)
	}

	p, err = structure.NormalizeJsonString(p)
	if err != nil {
		return append(diags, create.DiagError(names.CodeCatalyst, create.ErrActionSetting, ResNameDevenvironment, d.Id(), err)...)
	}

	d.Set("policy", p)

	
	// TIP: -- 6. Return diags
	return diags
}

func resourceDevenvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
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

	// TIP: -- 1. Get a client connection to the relevant service
	conn := meta.(*conns.AWSClient).CodecatalystClient(ctx)
	
	// TIP: -- 2. Populate a modify input structure and check for changes
	//
	// When creating the input structure, only include mandatory fields. Other
	// fields are set as needed. You can use a flag, such as update below, to
	// determine if a certain portion of arguments have been changed and
	// whether to call the AWS update function.
	update := false

	in := &codecatalyst.UpdateDevenvironmentInput{
		Id: aws.String(d.Id()),
	}

	if d.HasChanges("an_argument") {
		in.AnArgument = aws.String(d.Get("an_argument").(string))
		update = true
	}

	if !update {
		// TIP: If update doesn't do anything at all, which is rare, you can
		// return diags. Otherwise, return a read call, as below.
		return diags
	}
	
	// TIP: -- 3. Call the AWS modify/update function
	log.Printf("[DEBUG] Updating Codecatalyst Devenvironment (%s): %#v", d.Id(), in)
	out, err := conn.UpdateDevenvironment(ctx, in)
	if err != nil {
		return append(diags, create.DiagError(names.Codecatalyst, create.ErrActionUpdating, ResNameDevenvironment, d.Id(), err)...)
	}
	
	// TIP: -- 4. Use a waiter to wait for update to complete
	if _, err := waitDevenvironmentUpdated(ctx, conn, aws.ToString(out.OperationId), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return append(diags, create.DiagError(names.Codecatalyst, create.ErrActionWaitingForUpdate, ResNameDevenvironment, d.Id(), err)...)
	}
	
	// TIP: -- 5. Call the Read function in the Update return
	return append(diags, resourceDevenvironmentRead(ctx, d, meta)...)
}

func resourceDevenvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
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

	// TIP: -- 1. Get a client connection to the relevant service
	conn := meta.(*conns.AWSClient).CodecatalystClient(ctx)
	
	// TIP: -- 2. Populate a delete input structure
	log.Printf("[INFO] Deleting Codecatalyst Devenvironment %s", d.Id())
	
	// TIP: -- 3. Call the AWS delete function
	_, err := conn.DeleteDevenvironment(ctx, &codecatalyst.DeleteDevenvironmentInput{
		Id: aws.String(d.Id()),
	})
	
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if errs.IsA[*types.ResourceNotFoundException](err){
		return diags
	}
	if err != nil {
		return append(diags, create.DiagError(names.Codecatalyst, create.ErrActionDeleting, ResNameDevenvironment, d.Id(), err)...)
	}
	
	// TIP: -- 4. Use a waiter to wait for delete to complete
	if _, err := waitDevenvironmentDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return append(diags, create.DiagError(names.Codecatalyst, create.ErrActionWaitingForDeletion, ResNameDevenvironment, d.Id(), err)...)
	}
	
	// TIP: -- 5. Return diags
	return diags
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

func waitDevenvironmentCreated(ctx context.Context, conn *codecatalyst.Client, id string, timeout time.Duration) (*codecatalyst.Devenvironment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusDevenvironment(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*codecatalyst.Devenvironment); ok {
		return out, err
	}

	return nil, err
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.

func waitDevenvironmentUpdated(ctx context.Context, conn *codecatalyst.Client, id string, timeout time.Duration) (*codecatalyst.Devenvironment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusDevenvironment(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*codecatalyst.Devenvironment); ok {
		return out, err
	}

	return nil, err
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.

func waitDevenvironmentDeleted(ctx context.Context, conn *codecatalyst.Client, id string, timeout time.Duration) (*codecatalyst.Devenvironment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusDeleting, statusNormal},
		Target:                    []string{},
		Refresh:                   statusDevenvironment(ctx, conn, id),
		Timeout:                   timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*codecatalyst.Devenvironment); ok {
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

func statusDevenvironment(ctx context.Context, conn *codecatalyst.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findDevenvironmentByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.Status), nil
	}
}

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.

func findDevenvironmentByID(ctx context.Context, conn *codecatalyst.Client, id string) (*codecatalyst.Devenvironment, error) {
	in := &codecatalyst.GetDevenvironmentInput{
		Id: aws.String(id),
	}
	out, err := conn.GetDevenvironment(ctx, in)
	if errs.IsA[*types.ResourceNotFoundException](err){
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.Devenvironment == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Devenvironment, nil
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
func flattenComplexArgument(apiObject *codecatalyst.ComplexArgument) map[string]interface{} {
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

// TIP: Often the AWS API will return a slice of structures in response to a
// request for information. Sometimes you will have set criteria (e.g., the ID)
// that means you'll get back a one-length slice. This plural function works
// brilliantly for that situation too.
func flattenComplexArguments(apiObjects []*codecatalyst.ComplexArgument) []interface{} {
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



func expandPersistentStorage(tfMap map[string]interface{}) *types.PersistentStorageConfiguration {
	apiObject := &types.PersistentStorageConfiguration{}

	if v, ok := tfMap["persistant_storage"].(int); ok && v != 0 {
		apiObject.SizeInGiB= aws.Int32(int32(v))
	}

	return apiObject
}
