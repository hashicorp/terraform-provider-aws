// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
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
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_listener", name="Listener")
// @Tags(identifierAttribute="arn")
func resourceListener() *schema.Resource {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDefaultAction: {
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
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrStatusCode: {
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
												names.AttrWeight: {
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
			names.AttrName: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			names.AttrPort: {
				Type:         schema.TypeInt,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsPortNumber,
			},
			names.AttrProtocol: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.ListenerProtocol](),
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	in := &vpclattice.CreateListenerInput{
		Name:          aws.String(d.Get(names.AttrName).(string)),
		DefaultAction: expandDefaultAction(d.Get(names.AttrDefaultAction).([]interface{})),
		Protocol:      types.ListenerProtocol(d.Get(names.AttrProtocol).(string)),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrPort); ok && v != nil {
		in.Port = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("service_identifier"); ok {
		in.ServiceIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("service_arn"); ok {
		in.ServiceIdentifier = aws.String(v.(string))
	}

	if in.ServiceIdentifier == nil {
		return sdkdiag.AppendErrorf(diags, "must specify either service_arn or service_identifier")
	}

	out, err := conn.CreateListener(ctx, in)

	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionCreating, ResNameListener, d.Get(names.AttrName).(string), err)
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

	return append(diags, resourceListenerRead(ctx, d, meta)...)
}

func resourceListenerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	// GetListener requires the ID or Amazon Resource Name (ARN) of the service
	serviceId := d.Get("service_identifier").(string)
	listenerId := d.Get("listener_id").(string)

	out, err := findListenerByTwoPartKey(ctx, conn, listenerId, serviceId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPCLattice Listener (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionReading, ResNameListener, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrCreatedAt, aws.ToTime(out.CreatedAt).String())
	d.Set("last_updated_at", aws.ToTime(out.LastUpdatedAt).String())
	d.Set("listener_id", out.Id)
	d.Set(names.AttrName, out.Name)
	d.Set(names.AttrProtocol, out.Protocol)
	d.Set(names.AttrPort, out.Port)
	d.Set("service_arn", out.ServiceArn)
	d.Set("service_identifier", out.ServiceId)

	if err := d.Set(names.AttrDefaultAction, flattenListenerRuleActions(out.DefaultAction)); err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionSetting, ResNameListener, d.Id(), err)
	}

	return diags
}

func resourceListenerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	serviceId := d.Get("service_identifier").(string)
	listenerId := d.Get("listener_id").(string)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		in := &vpclattice.UpdateListenerInput{
			ListenerIdentifier: aws.String(listenerId),
			ServiceIdentifier:  aws.String(serviceId),
		}

		// Cannot edit listener name, protocol, or port after creation
		if d.HasChanges(names.AttrDefaultAction) {
			in.DefaultAction = expandDefaultAction(d.Get(names.AttrDefaultAction).([]interface{}))
		}

		log.Printf("[DEBUG] Updating VPC Lattice Listener (%s): %#v", d.Id(), in)
		_, err := conn.UpdateListener(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionUpdating, ResNameListener, d.Id(), err)
		}
	}

	return append(diags, resourceListenerRead(ctx, d, meta)...)
}

func resourceListenerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	serviceId := d.Get("service_identifier").(string)
	listenerId := d.Get("listener_id").(string)

	log.Printf("[INFO] Deleting VPCLattice Listener %s", d.Id())
	_, err := conn.DeleteListener(ctx, &vpclattice.DeleteListenerInput{
		ListenerIdentifier: aws.String(listenerId),
		ServiceIdentifier:  aws.String(serviceId),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.VPCLattice, create.ErrActionDeleting, ResNameListener, d.Id(), err)
	}

	return diags
}

func findListenerByTwoPartKey(ctx context.Context, conn *vpclattice.Client, listenerID, serviceID string) (*vpclattice.GetListenerOutput, error) {
	in := &vpclattice.GetListenerInput{
		ListenerIdentifier: aws.String(listenerID),
		ServiceIdentifier:  aws.String(serviceID),
	}
	out, err := conn.GetListener(ctx, in)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

// Flatten function for listener rule actions
func flattenListenerRuleActions(config types.RuleAction) []interface{} {
	m := map[string]interface{}{}

	if config == nil {
		return []interface{}{}
	}

	switch v := config.(type) {
	case *types.RuleActionMemberFixedResponse:
		m["fixed_response"] = flattenFixedResponseAction(&v.Value)
	case *types.RuleActionMemberForward:
		m["forward"] = flattenComplexDefaultActionForward(&v.Value)
	}

	return []interface{}{m}
}

// Flatten function for fixed_response action
func flattenFixedResponseAction(response *types.FixedResponseAction) []interface{} {
	tfMap := map[string]interface{}{}

	if v := response.StatusCode; v != nil {
		tfMap[names.AttrStatusCode] = aws.ToInt32(v)
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
			names.AttrWeight:          aws.ToInt32(targetGroup.Weight),
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
			Weight:                aws.Int32(int32(tfMap[names.AttrWeight].(int))),
		}

		targetGroups = append(targetGroups, *targetGroup)
	}

	return targetGroups
}

// Expand function for fixed_response action
func expandDefaultActionFixedResponseStatus(l []interface{}) *types.FixedResponseAction {
	lRaw := l[0].(map[string]interface{})

	fixedResponseAction := &types.FixedResponseAction{}

	if v, ok := lRaw[names.AttrStatusCode].(int); ok {
		fixedResponseAction.StatusCode = aws.Int32(int32(v))
	}

	return fixedResponseAction
}
