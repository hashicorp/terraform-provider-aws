package vpclattice

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_listener_rule", name="Listener Rule")
// @Tags(identifierAttribute="arn")
func ResourceListenerRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceListenerRuleCreate,
		ReadWithoutTimeout:   resourceListenerRuleRead,
		UpdateWithoutTimeout: resourceListenerRuleUpdate,
		DeleteWithoutTimeout: resourceListenerRuleDelete,

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
			"action": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fixed_response": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"status_code": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(100, 599),
									},
								},
							},
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
						},
						"forward": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"target_groups": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 2,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"target_group_identifier": {
													Type:     schema.TypeString,
													Required: true,
												},
												"weight": {
													Type:         schema.TypeInt,
													ValidateFunc: validation.IntBetween(0, 999),
													Default:      1,
													Optional:     true,
												},
											},
										},
									},
								},
							},
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
						},
					},
				},
			},
			"match": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_match": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"method": {
										Type:     schema.TypeString,
										Computed: true,
										Optional: true,
									},
									"headers_matches": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										MaxItems: 5,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"case_sensitive": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"match": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"contains": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"exact": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"prefix": {
																Type:     schema.TypeString,
																Optional: true,
															},
														},
													},
												},
												"name": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"path_match": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"case_sensitive": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"match": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"exact": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"prefix": {
																Type:     schema.TypeString,
																Optional: true,
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
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 128),
			},
			"priority": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: false,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customdiff.All(
			verify.SetTagsDiff,
		),
	}
}

const (
	ResNameListenerRule = "Listener Rule"
)

func resourceListenerRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient()

	in := &vpclattice.CreateRuleInput{
		Name: aws.String(d.Get("name").(string)),

		// ServiceIdentifier: vpclattice.CreateRuleInput.ServiceIdentifier()
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	// TIP: -- 3. Call the AWS create function
	out, err := conn.CreateListenerRule(ctx, in)
	if err != nil {
		// TIP: Since d.SetId() has not been called yet, you cannot use d.Id()
		// in error messages at this point.
		return create.DiagError(names.VpcLattice, create.ErrActionCreating, ResNameListenerRule, d.Get("name").(string), err)
	}

	if out == nil || out.ListenerRule == nil {
		return create.DiagError(names.VpcLattice, create.ErrActionCreating, ResNameListenerRule, d.Get("name").(string), errors.New("empty output"))
	}

	// TIP: -- 4. Set the minimum arguments and/or attributes for the Read function to
	// work.
	d.SetId(aws.ToString(out.ListenerRule.ListenerRuleID))

	// TIP: -- 5. Use a waiter to wait for create to complete
	if _, err := waitListenerRuleCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.VpcLattice, create.ErrActionWaitingForCreation, ResNameListenerRule, d.Id(), err)
	}

	// TIP: -- 6. Call the Read function in the Create return
	return resourceListenerRuleRead(ctx, d, meta)
}

func resourceListenerRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TIP: ==== RESOURCE READ ====
	// Generally, the Read function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Get the resource from AWS
	// 3. Set ID to empty where resource is not new and not found
	// 4. Set the arguments and attributes
	// 5. Set the tags
	// 6. Return nil

	// TIP: -- 1. Get a client connection to the relevant service
	conn := meta.(*conns.AWSClient).VpcLatticeClient()

	// TIP: -- 2. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findListenerRuleByID(ctx, conn, d.Id())

	// TIP: -- 3. Set ID to empty where resource is not new and not found
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VpcLattice ListenerRule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.VpcLattice, create.ErrActionReading, ResNameListenerRule, d.Id(), err)
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

	// TIP: Setting a complex type.
	// For more information, see:
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#data-handling-and-conversion
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#flatten-functions-for-blocks
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#root-typeset-of-resource-and-aws-list-of-structure
	if err := d.Set("complex_argument", flattenComplexArguments(out.ComplexArguments)); err != nil {
		return create.DiagError(names.VpcLattice, create.ErrActionSetting, ResNameListenerRule, d.Id(), err)
	}

	// TIP: Setting a JSON string to avoid errorneous diffs.
	p, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.ToString(out.Policy))
	if err != nil {
		return create.DiagError(names.VpcLattice, create.ErrActionSetting, ResNameListenerRule, d.Id(), err)
	}

	p, err = structure.NormalizeJsonString(p)
	if err != nil {
		return create.DiagError(names.VpcLattice, create.ErrActionSetting, ResNameListenerRule, d.Id(), err)
	}

	d.Set("policy", p)

	// TIP: -- 5. Set the tags
	//
	// TIP: Not all resources support tags and tags don't always make sense. If
	// your resource doesn't need tags, you can remove the tags lines here and
	// below. Many resources do include tags so this a reminder to include them
	// where possible.
	tags, err := ListTags(ctx, conn, d.Id())
	if err != nil {
		return create.DiagError(names.VpcLattice, create.ErrActionReading, ResNameListenerRule, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.VpcLattice, create.ErrActionSetting, ResNameListenerRule, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.VpcLattice, create.ErrActionSetting, ResNameListenerRule, d.Id(), err)
	}

	// TIP: -- 6. Return nil
	return nil
}

func resourceListenerRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	conn := meta.(*conns.AWSClient).VpcLatticeClient()

	// TIP: -- 2. Populate a modify input structure and check for changes
	//
	// When creating the input structure, only include mandatory fields. Other
	// fields are set as needed. You can use a flag, such as update below, to
	// determine if a certain portion of arguments have been changed and
	// whether to call the AWS update function.
	update := false

	in := &vpclattice.UpdateListenerRuleInput{
		Id: aws.String(d.Id()),
	}

	if d.HasChanges("an_argument") {
		in.AnArgument = aws.String(d.Get("an_argument").(string))
		update = true
	}

	if !update {
		// TIP: If update doesn't do anything at all, which is rare, you can
		// return nil. Otherwise, return a read call, as below.
		return nil
	}

	// TIP: -- 3. Call the AWS modify/update function
	log.Printf("[DEBUG] Updating VpcLattice ListenerRule (%s): %#v", d.Id(), in)
	out, err := conn.UpdateListenerRule(ctx, in)
	if err != nil {
		return create.DiagError(names.VpcLattice, create.ErrActionUpdating, ResNameListenerRule, d.Id(), err)
	}

	// TIP: -- 4. Use a waiter to wait for update to complete
	if _, err := waitListenerRuleUpdated(ctx, conn, aws.ToString(out.OperationId), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.DiagError(names.VpcLattice, create.ErrActionWaitingForUpdate, ResNameListenerRule, d.Id(), err)
	}

	// TIP: -- 5. Call the Read function in the Update return
	return resourceListenerRuleRead(ctx, d, meta)
}

func resourceListenerRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	// 5. Return nil

	// TIP: -- 1. Get a client connection to the relevant service
	conn := meta.(*conns.AWSClient).VpcLatticeClient()

	// TIP: -- 2. Populate a delete input structure
	log.Printf("[INFO] Deleting VpcLattice ListenerRule %s", d.Id())

	// TIP: -- 3. Call the AWS delete function
	_, err := conn.DeleteListenerRule(ctx, &vpclattice.DeleteListenerRuleInput{
		Id: aws.String(d.Id()),
	})

	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.VpcLattice, create.ErrActionDeleting, ResNameListenerRule, d.Id(), err)
	}

	// TIP: -- 4. Use a waiter to wait for delete to complete
	if _, err := waitListenerRuleDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.VpcLattice, create.ErrActionWaitingForDeletion, ResNameListenerRule, d.Id(), err)
	}

	// TIP: -- 5. Return nil
	return nil
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

func waitListenerRuleCreated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.ListenerRule, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusListenerRule(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.ListenerRule); ok {
		return out, err
	}

	return nil, err
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.

func waitListenerRuleUpdated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.ListenerRule, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusListenerRule(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.ListenerRule); ok {
		return out, err
	}

	return nil, err
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.

func waitListenerRuleDeleted(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.ListenerRule, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusListenerRule(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*vpclattice.ListenerRule); ok {
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

func statusListenerRule(ctx context.Context, conn *vpclattice.Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findListenerRuleByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.Status), nil
	}
}

func findListenerRuleByID(ctx context.Context, conn *vpclattice.Client, id string) (*vpclattice.ListenerRule, error) {
	in := &vpclattice.GetListenerRuleInput{
		Id: aws.String(id),
	}
	out, err := conn.GetListenerRule(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.ListenerRule == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.ListenerRule, nil
}
func flattenRuleAction(apiObject types.RuleAction) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v, ok := apiObject.(*types.RuleActionMemberFixedResponse); ok {
		tfMap["fixed_response"] = flattenRuleActionMemberFixedResponse(v)
	}
	if v, ok := apiObject.(*types.RuleActionMemberForward); ok {
		tfMap["forward"] = flattenForwardAction(v)
	}

	return tfMap
}

func flattenRuleActionMemberFixedResponse(apiObject *types.RuleActionMemberFixedResponse) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Value.StatusCode; v != nil {
		tfMap["status"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenForwardAction(apiObject *types.ForwardAction) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.TargetGroups; v != nil {
		tfMap["forward"] = flattenWeightedTargetGroups(v)
	}

	return tfMap
}

func flattenWeightedTargetGroups(apiObjects []types.WeightedTargetGroup) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenWeightedTargetGroup(&apiObject))
	}

	return tfList
}

func flattenWeightedTargetGroup(apiObject *types.WeightedTargetGroup) map[string]interface{} {

	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.TargetGroupIdentifier; v != nil {
		tfMap["target_group_identifier"] = aws.ToString(v)
	}

	if v := apiObject.Weight; v != nil {
		tfMap["weight"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenHttpMatch(apiObject *types.HttpMatch) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.HeaderMatches; v != nil {
		tfMap["headers_matches"] = flattenHeaderMatches(v)
	}

	if v := apiObject.PathMatch; v != nil {
		tfMap["path_match"] = []interface{}{flattenPathMatch(v)}
	}

	return tfMap
}

func flattenHeaderMatches(apiObjects []types.HeaderMatch) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenHeaderMatch(&apiObject))
	}

	return tfList
}

func flattenHeaderMatch(apiObject *types.HeaderMatch) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CaseSensitive; v != nil {
		tfMap["case_sensitive"] = aws.ToBool(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.ToString(v)
	}

	if v := apiObject.Match; v != nil {
		tfMap["match"] = []interface{}{flattenHeaderMatchType(v.(*types.HeaderMatchType))}
	}

	return tfMap
}

func flattenHeaderMatchType(apiObject types.HeaderMatchType) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v, ok := apiObject.(*types.HeaderMatchTypeMemberExact); ok {
		tfMap["exact"] = []interface{}{flattenHeaderMatchTypeMemberExact(v)}
	}

	if v, ok := apiObject.(*types.HeaderMatchTypeMemberPrefix); ok {
		tfMap["prefix"] = []interface{}{flattenHeaderMatchTypeMemberPrefix(v)}
	}

	if v, ok := apiObject.(*types.HeaderMatchTypeMemberContains); ok {
		tfMap["prefix"] = []interface{}{flattenHeaderMatchTypeMemberContains(v)}
	}

	return tfMap
}

func flattenHeaderMatchTypeMemberContains(apiObject *types.HeaderMatchTypeMemberContains) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"contains": apiObject.Value,
	}

	return tfMap
}

func flattenHeaderMatchTypeMemberExact(apiObject *types.HeaderMatchTypeMemberExact) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"exact": apiObject.Value,
	}

	return tfMap
}

func flattenHeaderMatchTypeMemberPrefix(apiObject *types.HeaderMatchTypeMemberPrefix) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"prefix": apiObject.Value,
	}

	return tfMap
}

func flattenPathMatch(apiObject *types.PathMatch) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CaseSensitive; v != nil {
		tfMap["case_sensitive"] = aws.ToBool(v)
	}

	if v := apiObject.Match; v != nil {
		tfMap["match"] = []interface{}{flattenPathMatchType(v)}
	}

	return tfMap
}

func flattenPathMatchType(apiObject types.PathMatchType) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v, ok := apiObject.(*types.PathMatchTypeMemberExact); ok {
		tfMap["exact"] = flattenPathMatchTypeMemberExact(v)
	}

	if v, ok := apiObject.(*types.PathMatchTypeMemberPrefix); ok {
		tfMap["prefix"] = flattenPathMatchTypeMemberPrefix(v)
	}

	return tfMap
}

func flattenPathMatchTypeMemberExact(apiObject *types.PathMatchTypeMemberExact) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"exact": apiObject.Value,
	}

	return tfMap
}

func flattenPathMatchTypeMemberPrefix(apiObject *types.PathMatchTypeMemberPrefix) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"prefix": apiObject.Value,
	}

	return tfMap
}

func expandRuleAction(tfMap map[string]interface{}) types.RuleAction {
	var apiObject types.RuleAction

	if v, ok := tfMap["fixed_response_action"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject = expandFixedResponseAction(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["forward_action"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject = expandForwardAction(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandFixedResponseAction(tfMap map[string]interface{}) *types.RuleActionMemberFixedResponse {
	apiObject := &types.RuleActionMemberFixedResponse{}

	if v, ok := tfMap["status"].(int); ok && v != 0 {
		apiObject.Value.StatusCode = aws.Int32(int32(v))
	}

	return apiObject
}

func expandForwardAction(tfMap map[string]interface{}) *types.RuleActionMemberForward {
	apiObject := &types.RuleActionMemberForward{}

	if v, ok := tfMap["target_groups"].([]interface{}); ok && len(v) > 0 && v != nil {
		apiObject.Value.TargetGroups = expandWeightedTargetGroups(v)
	}

	return apiObject
}

func expandWeightedTargetGroups(tfList []interface{}) []types.WeightedTargetGroup {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.WeightedTargetGroup

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandWeightedTargetGroup(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandWeightedTargetGroup(tfMap map[string]interface{}) types.WeightedTargetGroup {
	apiObject := types.WeightedTargetGroup{}

	if v, ok := tfMap["target_group_identifier"].(string); ok && v != "" {
		apiObject.TargetGroupIdentifier = aws.String(v)
	}

	if v, ok := tfMap["weight"].(int); ok && v != 0 {
		apiObject.Weight = aws.Int32(int32(v))
	}

	return apiObject
}

func expandRuleMatch(tfMap map[string]interface{}) types.RuleMatch {
	apiObject := &types.RuleMatchMemberHttpMatch{}

	if v, ok := tfMap["match"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Value = expandHttpMatch(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandHttpMatch(tfMap map[string]interface{}) types.HttpMatch {
	apiObject := types.HttpMatch{}

	if v, ok := tfMap["header_matches"].([]interface{}); ok && len(v) > 0 && v != nil {
		apiObject.HeaderMatches = expandHeaderMatches(v)
	}

	if v, ok := tfMap["method"].(string); ok {
		apiObject.Method = aws.String(v)
	}

	if v, ok := tfMap["matcher"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.PathMatch = expandPathMatch(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandHeaderMatches(tfList []interface{}) []types.HeaderMatch {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.HeaderMatch

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandHeaderMatch(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandHeaderMatch(tfMap map[string]interface{}) types.HeaderMatch {
	apiObject := types.HeaderMatch{}

	if v, ok := tfMap["case_sensitive"].(bool); ok {
		apiObject.CaseSensitive = aws.Bool(v)
	}

	if v, ok := tfMap["name"].(string); ok {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["match"].([]interface{}); ok && len(v) > 0 {
		matchObj := v[0].(map[string]interface{})
		if matchV, ok := matchObj["exact"].(string); ok && matchV != "" {
			apiObject.Match = expandHeaderMatchTypeMemberExact(matchObj)
		}
		if matchV, ok := matchObj["prefix"].(string); ok && matchV != "" {
			apiObject.Match = expandHeaderMatchTypeMemberPrefix(matchObj)
		}
		if matchV, ok := matchObj["contains"].(string); ok && matchV != "" {
			apiObject.Match = expandHeaderMatchTypeMemberContains(matchObj)
		}
	}

	return apiObject
}

func expandHeaderMatchTypeMemberContains(tfMap map[string]interface{}) types.HeaderMatchType {
	apiObject := &types.HeaderMatchTypeMemberContains{}

	if v, ok := tfMap["contains"].(string); ok && v != "" {
		apiObject.Value = v
	}
	return apiObject
}

func expandHeaderMatchTypeMemberPrefix(tfMap map[string]interface{}) types.HeaderMatchType {
	apiObject := &types.HeaderMatchTypeMemberPrefix{}

	if v, ok := tfMap["prefix"].(string); ok && v != "" {
		apiObject.Value = v
	}
	return apiObject
}

func expandHeaderMatchTypeMemberExact(tfMap map[string]interface{}) types.HeaderMatchType {
	apiObject := &types.HeaderMatchTypeMemberExact{}

	if v, ok := tfMap["exact"].(string); ok && v != "" {
		apiObject.Value = v
	}
	return apiObject
}

func expandPathMatch(tfMap map[string]interface{}) *types.PathMatch {
	apiObject := &types.PathMatch{}

	if v, ok := tfMap["case_sensitive"].(bool); ok {
		apiObject.CaseSensitive = aws.Bool(v)
	}

	if v, ok := tfMap["match"].([]interface{}); ok && len(v) > 0 {
		matchObj := v[0].(map[string]interface{})

		if matchV, ok := matchObj["exact"].(string); ok && matchV != "" {
			apiObject.Match = expandPathMatchTypeMemberExact(matchObj)
		}

		if matchV, ok := matchObj["prefix"].(string); ok && matchV != "" {
			apiObject.Match = expandPathMatchTypeMemberPrefix(matchObj)
		}
	}

	return apiObject
}

func expandPathMatchTypeMemberExact(tfMap map[string]interface{}) types.PathMatchType {
	apiObject := &types.PathMatchTypeMemberExact{}

	if v, ok := tfMap["exact"].(string); ok && v != "" {
		apiObject.Value = v
	}
	return apiObject
}

func expandPathMatchTypeMemberPrefix(tfMap map[string]interface{}) types.PathMatchType {
	apiObject := &types.PathMatchTypeMemberPrefix{}

	if v, ok := tfMap["prefix"].(string); ok && v != "" {
		apiObject.Value = v
	}
	return apiObject
}
