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
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpclattice_listener", name="Listener")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func resourceListener() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceListenerCreate,
		ReadWithoutTimeout:   resourceListenerRead,
		UpdateWithoutTimeout: resourceListenerUpdate,
		DeleteWithoutTimeout: resourceListenerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
	}
}

func resourceListenerCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := vpclattice.CreateListenerInput{
		ClientToken:   aws.String(sdkid.UniqueId()),
		Name:          aws.String(name),
		DefaultAction: expandDefaultAction(d.Get(names.AttrDefaultAction).([]any)),
		Protocol:      types.ListenerProtocol(d.Get(names.AttrProtocol).(string)),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrPort); ok && v != nil {
		input.Port = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("service_identifier"); ok {
		input.ServiceIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("service_arn"); ok {
		input.ServiceIdentifier = aws.String(v.(string))
	}

	output, err := conn.CreateListener(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating VPCLattice Listener (%s): %s", name, err)
	}

	d.SetId(listenerCreateResourceID(aws.ToString(output.ServiceId), aws.ToString(output.Id)))

	return append(diags, resourceListenerRead(ctx, d, meta)...)
}

func resourceListenerRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	serviceID, listenerID, err := listenerParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findListenerByTwoPartKey(ctx, conn, serviceID, listenerID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPCLattice Listener (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPCLattice Listener (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.Arn)
	d.Set(names.AttrCreatedAt, aws.ToTime(output.CreatedAt).String())
	if err := d.Set(names.AttrDefaultAction, flattenListenerRuleActions(output.DefaultAction)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_action: %s", err)
	}
	d.Set("last_updated_at", aws.ToTime(output.LastUpdatedAt).String())
	d.Set("listener_id", output.Id)
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrProtocol, output.Protocol)
	d.Set(names.AttrPort, output.Port)
	d.Set("service_arn", output.ServiceArn)
	d.Set("service_identifier", output.ServiceId)

	return diags
}

func resourceListenerUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	serviceID, listenerID, err := listenerParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := vpclattice.UpdateListenerInput{
			ListenerIdentifier: aws.String(listenerID),
			ServiceIdentifier:  aws.String(serviceID),
		}

		if d.HasChanges(names.AttrDefaultAction) {
			input.DefaultAction = expandDefaultAction(d.Get(names.AttrDefaultAction).([]any))
		}

		_, err := conn.UpdateListener(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating VPCLattice Listener (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceListenerRead(ctx, d, meta)...)
}

func resourceListenerDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	serviceID, listenerID, err := listenerParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting VPCLattice Listener: %s", d.Id())
	input := vpclattice.DeleteListenerInput{
		ListenerIdentifier: aws.String(listenerID),
		ServiceIdentifier:  aws.String(serviceID),
	}
	_, err = conn.DeleteListener(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting VPCLattice Listener (%s): %s", d.Id(), err)
	}

	return diags
}

const listenerResourceIDSeparator = "/"

func listenerCreateResourceID(serviceID, listenerID string) string {
	parts := []string{serviceID, listenerID}
	id := strings.Join(parts, listenerResourceIDSeparator)

	return id
}

func listenerParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, listenerResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected SERVICE-ID%[2]sLISTENER-ID", id, listenerResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findListenerByTwoPartKey(ctx context.Context, conn *vpclattice.Client, serviceID, listenerID string) (*vpclattice.GetListenerOutput, error) {
	input := vpclattice.GetListenerInput{
		ListenerIdentifier: aws.String(listenerID),
		ServiceIdentifier:  aws.String(serviceID),
	}
	output, err := findListener(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if output.Id == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findListener(ctx context.Context, conn *vpclattice.Client, input *vpclattice.GetListenerInput) (*vpclattice.GetListenerOutput, error) {
	output, err := conn.GetListener(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func flattenListenerRuleActions(apiObject types.RuleAction) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	switch v := apiObject.(type) {
	case *types.RuleActionMemberFixedResponse:
		tfMap["fixed_response"] = flattenFixedResponseAction(&v.Value)
	case *types.RuleActionMemberForward:
		tfMap["forward"] = flattenComplexDefaultActionForward(&v.Value)
	}

	return []any{tfMap}
}

func flattenFixedResponseAction(apiObject *types.FixedResponseAction) []any {
	tfMap := map[string]any{}

	if v := apiObject.StatusCode; v != nil {
		tfMap[names.AttrStatusCode] = aws.ToInt32(v)
	}

	return []any{tfMap}
}

func flattenComplexDefaultActionForward(apiObject *types.ForwardAction) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"target_groups": flattenDefaultActionForwardTargetGroups(apiObject.TargetGroups),
	}

	return []any{tfMap}
}

func flattenDefaultActionForwardTargetGroups(apiObjects []types.WeightedTargetGroup) []any {
	if len(apiObjects) == 0 {
		return []any{}
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"target_group_identifier": aws.ToString(apiObject.TargetGroupIdentifier),
			names.AttrWeight:          aws.ToInt32(apiObject.Weight),
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandDefaultAction(tfList []any) types.RuleAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMapRaw := tfList[0].(map[string]any)

	if v, ok := tfMapRaw["forward"].([]any); ok && len(v) > 0 {
		return &types.RuleActionMemberForward{
			Value: expandDefaultActionForwardAction(v),
		}
	} else if v, ok := tfMapRaw["fixed_response"].([]any); ok && len(v) > 0 {
		return &types.RuleActionMemberFixedResponse{
			Value: expandDefaultActionFixedResponseStatus(v),
		}
	}

	return nil
}

func expandDefaultActionForwardAction(tfList []any) types.ForwardAction {
	lRaw := tfList[0].(map[string]any)

	apiObject := types.ForwardAction{}

	if v, ok := lRaw["target_groups"].([]any); ok && len(v) > 0 {
		apiObject.TargetGroups = expandForwardTargetGroupList(v)
	}

	return apiObject
}

func expandForwardTargetGroupList(tfList []any) []types.WeightedTargetGroup {
	var apiObjects []types.WeightedTargetGroup

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := types.WeightedTargetGroup{
			TargetGroupIdentifier: aws.String((tfMap["target_group_identifier"].(string))),
			Weight:                aws.Int32(int32(tfMap[names.AttrWeight].(int))),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandDefaultActionFixedResponseStatus(tfList []any) types.FixedResponseAction {
	tfMapRaw := tfList[0].(map[string]any)

	apiObject := types.FixedResponseAction{}

	if v, ok := tfMapRaw[names.AttrStatusCode].(int); ok {
		apiObject.StatusCode = aws.Int32(int32(v))
	}

	return apiObject
}
