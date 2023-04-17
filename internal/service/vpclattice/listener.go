package vpclattice

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_vpclattice_listener")
func ResourceListener() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceListenerCreate,
		ReadWithoutTimeout:   resourceListenerRead,
		UpdateWithoutTimeout: resourceListenerUpdate,
		DeleteWithoutTimeout: resourceListenerDelete,
		// Id returned by GetListener does not contain required service name, use a custom import function
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected SERVICE-ID/LISTENER-ID", d.Id())
				}
				d.Set("service_identifier", idParts[0])
				d.Set("listener_id", idParts[1])

				return []*schema.ResourceData{d}, nil
			},
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
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_action": {
				Type:     schema.TypeList,
				MaxItems: 1,
				MinItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fixed_response": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							// ExactlyOneOf: []string{
							// 	"default_action.0.fixed_response",
							// 	"default_action.0.forward",
							// },
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"status_code": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
						"forward": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							// ExactlyOneOf: []string{
							// 	"default_action.0.fixed_response",
							// 	"default_action.0.forward",
							// },
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"target_groups": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"target_group_identifier": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"weight": {
													Type:     schema.TypeInt,
													Default:  100,
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
			"last_updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"listener_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"port": {
				Type:         schema.TypeInt,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.IsPortNumber,
			},
			"protocol": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"HTTP", "HTTPS"}, true),
			},
			"service_arn": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				AtLeastOneOf: []string{"service_arn", "service_identifier"},
			},
			"service_identifier": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				AtLeastOneOf: []string{"service_arn", "service_identifier"},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameListener = "Listener"
)

func resourceListenerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient()

	in := &vpclattice.CreateListenerInput{
		Name:          aws.String(d.Get("name").(string)),
		DefaultAction: expandDefaultAction(d.Get("default_action").([]interface{})),
		Protocol:      types.ListenerProtocol(d.Get("protocol").(string)),
		Tags:          GetTagsIn(ctx),
	}

	if v, ok := d.GetOk("port"); ok && v != nil {
		in.Port = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("service_identifier"); ok {
		in.ServiceIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("service_arn"); ok {
		in.ServiceIdentifier = aws.String(v.(string))
	}

	if in.ServiceIdentifier == nil {
		return diag.FromErr(fmt.Errorf("must specify either service_arn or service_identifier"))
	}

	// TIP: Not all resources support tags and tags don't always make sense. If
	// your resource doesn't need tags, you can remove the tags lines here and
	// below. Many resources do include tags so this a reminder to include them
	// where possible.
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateListener(ctx, in)
	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionCreating, ResNameListener, d.Get("name").(string), err)
	}

	if out == nil || out.Arn == nil {
		return create.DiagError(names.VPCLattice, create.ErrActionCreating, ResNameListener, d.Get("name").(string), errors.New("empty output"))
	}

	// Id returned by GetListener does not contain required service name
	// Create a composite ID using service ID and listener ID
	d.Set("listener_id", out.Id)
	d.Set("service_identifier", out.ServiceId)

	parts := []string{
		d.Get("service_identifier").(string),
		d.Get("listener_id").(string),
	}

	d.SetId(strings.Join(parts, "/"))

	// TIP: -- 5. Use a waiter to wait for create to complete
	// if _, err := waitListenerCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
	// 	return create.DiagError(names.VPCLattice, create.ErrActionWaitingForCreation, ResNameListener, d.Id(), err)
	// }

	return resourceListenerRead(ctx, d, meta)
}

func resourceListenerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient()

	// GetListener requires the ID or Amazon Resource Name (ARN) of the service
	serviceId := d.Get("service_identifier").(string)
	listenerId := d.Get("listener_id").(string)

	out, err := findListenerByIdAndServiceId(ctx, conn, listenerId, serviceId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPCLattice Listener (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, ResNameListener, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("created_at", aws.ToTime(out.CreatedAt).String())
	d.Set("last_updated_at", aws.ToTime(out.LastUpdatedAt).String())
	d.Set("listener_id", out.Id)
	d.Set("name", out.Name)
	d.Set("protocol", out.Protocol)
	d.Set("port", out.Port)
	d.Set("service_arn", out.ServiceArn)
	d.Set("service_identifier", out.ServiceId)

	if err := d.Set("default_action", flattenListenerRuleActions(out.DefaultAction)); err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionSetting, ResNameListener, d.Id(), err)
	}

	// ListTags requires ARN, not ID
	tags, err := ListTags(ctx, conn, d.Get("arn").(string))
	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, ResNameListener, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionSetting, ResNameListener, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionSetting, ResNameListener, d.Id(), err)
	}

	return nil
}

func resourceListenerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient()

	serviceId := d.Get("service_identifier").(string)
	listenerId := d.Get("listener_id").(string)

	if d.HasChangesExcept("tags", "tags_all") {

		in := &vpclattice.UpdateListenerInput{
			ListenerIdentifier: aws.String(listenerId),
			ServiceIdentifier:  aws.String(serviceId),
		}

		// Cannot edit listener name, protocol, or port after creation
		if d.HasChanges("default_action") {
			in.DefaultAction = expandDefaultAction(d.Get("default_action").([]interface{}))
		}

		// TIP: -- 3. Call the AWS modify/update function
		log.Printf("[DEBUG] Updating VPC Lattice Listener (%s): %#v", d.Id(), in)
		_, err := conn.UpdateListener(ctx, in)
		if err != nil {
			return create.DiagError(names.VPCLattice, create.ErrActionUpdating, ResNameListener, d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating tags: %s", err)
		}
	}

	// TIP: -- 4. Use a waiter to wait for update to complete
	// if _, err := waitListenerUpdated(ctx, conn, aws.ToString(out.Id), d.Timeout(schema.TimeoutUpdate)); err != nil {
	// 	return create.DiagError(names.VPCLattice, create.ErrActionWaitingForUpdate, ResNameListener, d.Id(), err)
	// }

	// TIP: -- 5. Call the Read function in the Update return
	return resourceListenerRead(ctx, d, meta)
}

func resourceListenerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient()

	log.Printf("[INFO] Deleting VPC Lattice Listener %s", d.Id())

	serviceId := d.Get("service_identifier").(string)
	listenerId := d.Get("listener_id").(string)

	_, err := conn.DeleteListener(ctx, &vpclattice.DeleteListenerInput{
		ListenerIdentifier: aws.String(listenerId),
		ServiceIdentifier:  aws.String(serviceId),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.VPCLattice, create.ErrActionDeleting, ResNameListener, d.Id(), err)
	}

	// TIP: -- 4. Use a waiter to wait for delete to complete
	// if _, err := waitListenerDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
	// 	return create.DiagError(names.VPCLattice, create.ErrActionWaitingForDeletion, ResNameListener, d.Id(), err)
	// }

	return nil
}

// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., amp.WorkspaceStatusCodeActive).
// const (
// 	statusChangePending = "Pending"
// 	statusDeleting      = "Deleting"
// 	statusNormal        = "Normal"
// 	statusUpdated       = "Updated"
// )

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

// func waitListenerCreated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.Listener, error) {
// 	stateConf := &retry.StateChangeConf{
// 		Pending:                   []string{},
// 		Target:                    []string{statusNormal},
// 		Refresh:                   statusListener(ctx, conn, id),
// 		Timeout:                   timeout,
// 		NotFoundChecks:            20,
// 		ContinuousTargetOccurence: 2,
// 	}

// 	outputRaw, err := stateConf.WaitForStateContext(ctx)
// 	if out, ok := outputRaw.(*vpclattice.Listener); ok {
// 		return out, err
// 	}

// 	return nil, err
// }

// // TIP: It is easier to determine whether a resource is updated for some
// // resources than others. The best case is a status flag that tells you when
// // the update has been fully realized. Other times, you can check to see if a
// // key resource argument is updated to a new value or not.

// func waitListenerUpdated(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.Listener, error) {
// 	stateConf := &retry.StateChangeConf{
// 		Pending:                   []string{statusChangePending},
// 		Target:                    []string{statusUpdated},
// 		Refresh:                   statusListener(ctx, conn, id),
// 		Timeout:                   timeout,
// 		NotFoundChecks:            20,
// 		ContinuousTargetOccurence: 2,
// 	}

// 	outputRaw, err := stateConf.WaitForStateContext(ctx)
// 	if out, ok := outputRaw.(*vpclattice.Listener); ok {
// 		return out, err
// 	}

// 	return nil, err
// }

// // TIP: A deleted waiter is almost like a backwards created waiter. There may
// // be additional pending states, however.

// func waitListenerDeleted(ctx context.Context, conn *vpclattice.Client, id string, timeout time.Duration) (*vpclattice.Listener, error) {
// 	stateConf := &retry.StateChangeConf{
// 		Pending:                   []string{statusDeleting, statusNormal},
// 		Target:                    []string{},
// 		Refresh:                   statusListener(ctx, conn, id),
// 		Timeout:                   timeout,
// 	}

// 	outputRaw, err := stateConf.WaitForStateContext(ctx)
// 	if out, ok := outputRaw.(*vpclattice.Listener); ok {
// 		return out, err
// 	}

// 	return nil, err
// }

// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.

// func statusListener(ctx context.Context, conn *vpclattice.Client, id string) retry.StateRefreshFunc {
// 	return func() (interface{}, string, error) {
// 		out, err := findListenerByID(ctx, conn, id)
// 		if tfresource.NotFound(err) {
// 			return nil, "", nil
// 		}

// 		if err != nil {
// 			return nil, "", err
// 		}

// 		return out, aws.ToString(out.), nil
// 	}
// }

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.

func findListenerByIdAndServiceId(ctx context.Context, conn *vpclattice.Client, id string, serviceId string) (*vpclattice.GetListenerOutput, error) {
	in := &vpclattice.GetListenerInput{
		ListenerIdentifier: aws.String(id),
		ServiceIdentifier:  aws.String(serviceId),
	}
	out, err := conn.GetListener(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

// flattenListenerRuleActions
func flattenListenerRuleActions(config types.RuleAction) []interface{} {
	// Create an empty map with key type `string` and value of `interface{}`
	m := map[string]interface{}{}

	if config == nil {
		return []interface{}{}
	}

	switch v := config.(type) {
	case *types.RuleActionMemberFixedResponse:
		m["fixed_response"] = flattenRuleActionMemberFixedResponse(&v.Value)
	case *types.RuleActionMemberForward:
		m["forward"] = flattenComplexDefaultActionForward(&v.Value)
	}

	return []interface{}{m}
}

// Flatten function for fixed_response action
func flattenRuleActionMemberFixedResponse(response *types.FixedResponseAction) []interface{} {
	tfMap := map[string]interface{}{}

	if v := response.StatusCode; v != nil {
		tfMap["status_code"] = aws.ToInt32(v)
	}

	return []interface{}{tfMap}
}

// Flatten function for forward action
func flattenComplexDefaultActionForward(forwardAction *types.ForwardAction) []interface{} {
	if forwardAction == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"target_groups": flattenDefaultActionForwardTargetGroups(forwardAction.TargetGroups),
	}

	return []interface{}{m}
}

// Flatten function for target_groups
func flattenDefaultActionForwardTargetGroups(groups []types.WeightedTargetGroup) []interface{} {
	if len(groups) == 0 {
		return []interface{}{}
	}

	var targetGroups []interface{}

	for _, targetGroup := range groups {
		m := map[string]interface{}{
			"target_group_identifier": aws.ToString(targetGroup.TargetGroupIdentifier),
			"weight":                  aws.ToInt32(targetGroup.Weight),
		}
		targetGroups = append(targetGroups, m)
	}

	return targetGroups
}

// Expand function for default_action
func expandDefaultAction(l []interface{}) types.RuleAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	lRaw := l[0].(map[string]interface{})

	if v, ok := lRaw["forward"].([]interface{}); ok && len(v) > 0 {
		return &types.RuleActionMemberForward{
			Value: *expandDefaultActionForwardAction(v),
		}
	} else if v, ok := lRaw["fixed_response"].([]interface{}); ok && len(v) > 0 {
		return &types.RuleActionMemberFixedResponse{
			Value: *expandDefaultActionFixedResponseStatus(v),
		}
	} else {
		return nil
	}
}

// Expand function for forward action
func expandDefaultActionForwardAction(l []interface{}) *types.ForwardAction {
	lRaw := l[0].(map[string]interface{})

	forwardAction := &types.ForwardAction{}

	if v, ok := lRaw["target_groups"].([]interface{}); ok && len(v) > 0 {
		forwardAction.TargetGroups = expandForwardTargetGroupList(v)
	}

	return forwardAction
}

// Expand function for target_groups
func expandForwardTargetGroupList(tfList []interface{}) []types.WeightedTargetGroup {
	var targetGroups []types.WeightedTargetGroup

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		targetGroup := &types.WeightedTargetGroup{
			TargetGroupIdentifier: aws.String((tfMap["target_group_identifier"].(string))),
			Weight:                aws.Int32(int32(tfMap["weight"].(int))),
		}

		targetGroups = append(targetGroups, *targetGroup)

	}

	return targetGroups
}

// Expand function for fixed_response action
func expandDefaultActionFixedResponseStatus(l []interface{}) *types.FixedResponseAction {
	lRaw := l[0].(map[string]interface{})

	fixedResponseAction := &types.FixedResponseAction{}

	if v, ok := lRaw["status_code"].(int); ok {
		fixedResponseAction.StatusCode = aws.Int32(int32(v))
	}

	return fixedResponseAction
}
