// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_event_permission", name="Permission")
func resourcePermission() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePermissionCreate,
		ReadWithoutTimeout:   resourcePermissionRead,
		UpdateWithoutTimeout: resourcePermissionUpdate,
		DeleteWithoutTimeout: resourcePermissionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAction: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "events:PutEvents",
				ValidateFunc: validatePermissionAction,
			},
			names.AttrCondition: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"aws:PrincipalOrgID"}, false),
						},
						names.AttrType: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"StringEquals"}, false),
						},
						names.AttrValue: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
					},
				},
			},
			"event_bus_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validBusName,
				Default:      DefaultEventBusName,
			},
			names.AttrPrincipal: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validatePermissionPrincipal,
			},
			"statement_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validatePermissionStatementID,
			},
		},
	}
}

func resourcePermissionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	eventBusName := d.Get("event_bus_name").(string)
	statementID := d.Get("statement_id").(string)
	id := permissionCreateResourceID(eventBusName, statementID)
	input := &eventbridge.PutPermissionInput{
		Action:       aws.String(d.Get(names.AttrAction).(string)),
		Condition:    expandCondition(d.Get(names.AttrCondition).([]interface{})),
		EventBusName: aws.String(eventBusName),
		Principal:    aws.String(d.Get(names.AttrPrincipal).(string)),
		StatementId:  aws.String(statementID),
	}

	_, err := conn.PutPermission(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Permission (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourcePermissionRead(ctx, d, meta)...)
}

// See also: https://docs.aws.amazon.com/eventbridge/latest/APIReference/API_DescribeEventBus.html
func resourcePermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	eventBusName, statementID, err := permissionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	outputRaw, err := tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findPermissionByTwoPartKey(ctx, conn, eventBusName, statementID)
	})

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Permission (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Permission (%s): %s", d.Id(), err)
	}

	policyStatement := outputRaw.(*permissionPolicyStatement)

	d.Set(names.AttrAction, policyStatement.Action)
	if err := d.Set(names.AttrCondition, flattenPermissionPolicyStatementCondition(policyStatement.Condition)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting condition: %s", err)
	}
	d.Set("event_bus_name", eventBusName)
	switch principal := policyStatement.Principal.(type) {
	case string:
		d.Set(names.AttrPrincipal, principal)
	case map[string]interface{}:
		if v, ok := principal["AWS"].(string); ok {
			if arn.IsARN(v) {
				principalARN, err := arn.Parse(v)
				if err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}

				d.Set(names.AttrPrincipal, principalARN.AccountID)
			} else {
				d.Set(names.AttrPrincipal, v)
			}
		}
	}
	d.Set("statement_id", policyStatement.Sid)

	return diags
}

func resourcePermissionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	eventBusName, statementID, err := permissionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &eventbridge.PutPermissionInput{
		Action:       aws.String(d.Get(names.AttrAction).(string)),
		Condition:    expandCondition(d.Get(names.AttrCondition).([]interface{})),
		EventBusName: aws.String(eventBusName),
		Principal:    aws.String(d.Get(names.AttrPrincipal).(string)),
		StatementId:  aws.String(statementID),
	}

	_, err = conn.PutPermission(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EventBridge Permission (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePermissionRead(ctx, d, meta)...)
}

func resourcePermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	eventBusName, statementID, err := permissionParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting EventBridge Permission: %s", d.Id())
	_, err = conn.RemovePermission(ctx, &eventbridge.RemovePermissionInput{
		EventBusName: aws.String(eventBusName),
		StatementId:  aws.String(statementID),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Permission (%s): %s", d.Id(), err)
	}
	return diags
}

func findPermissionByTwoPartKey(ctx context.Context, conn *eventbridge.Client, eventBusName, statementID string) (*permissionPolicyStatement, error) {
	output, err := findEventBusPolicyByName(ctx, conn, eventBusName)

	if err != nil {
		return nil, err
	}

	var policyDoc permissionPolicyDoc
	if err := json.Unmarshal([]byte(aws.ToString(output)), &policyDoc); err != nil {
		return nil, err
	}

	for _, statement := range policyDoc.Statements {
		if statement.Sid == statementID {
			return &statement, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

const permissionResourceIDSeparator = "/"

func permissionCreateResourceID(eventBusName, statementID string) string {
	if eventBusName == "" || eventBusName == DefaultEventBusName {
		return statementID
	}

	parts := []string{eventBusName, statementID}
	id := strings.Join(parts, permissionResourceIDSeparator)

	return id
}

func permissionParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, permissionResourceIDSeparator)

	if len(parts) == 1 && parts[0] != "" {
		return DefaultEventBusName, parts[0], nil
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected EVENTBUSNAME%[2]sSTATEMENTID or STATEMENTID", id, permissionResourceIDSeparator)
}

// PermissionPolicyDoc represents the Policy attribute of DescribeEventBus
// See also: https://docs.aws.amazon.com/eventbridge/latest/APIReference/API_DescribeEventBus.html
type permissionPolicyDoc struct {
	Version    string
	ID         string                      `json:"Id,omitempty"`
	Statements []permissionPolicyStatement `json:"Statement"`
}

// PermissionPolicyStatement represents the Statement attribute of PermissionPolicyDoc
// See also: https://docs.aws.amazon.com/eventbridge/latest/APIReference/API_DescribeEventBus.html
type permissionPolicyStatement struct {
	Sid       string
	Effect    string
	Action    string
	Condition *permissionPolicyStatementCondition `json:"Condition,omitempty"`
	Principal interface{}                         // "*" or {"AWS": "arn:aws:iam::111111111111:root"}
	Resource  string
}

// PermissionPolicyStatementCondition represents the Condition attribute of PermissionPolicyStatement
// See also: https://docs.aws.amazon.com/eventbridge/latest/APIReference/API_DescribeEventBus.html
type permissionPolicyStatementCondition struct {
	Key   string
	Type  string
	Value string
}

func (c *permissionPolicyStatementCondition) UnmarshalJSON(b []byte) error {
	var out permissionPolicyStatementCondition

	// JSON representation: \"Condition\":{\"StringEquals\":{\"aws:PrincipalOrgID\":\"o-0123456789\"}}
	var data map[string]map[string]string
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	for typeKey, typeValue := range data {
		for conditionKey, conditionValue := range typeValue {
			out = permissionPolicyStatementCondition{
				Key:   conditionKey,
				Type:  typeKey,
				Value: conditionValue,
			}
		}
	}

	*c = out
	return nil
}

func expandCondition(l []interface{}) *types.Condition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	condition := &types.Condition{
		Key:   aws.String(m[names.AttrKey].(string)),
		Type:  aws.String(m[names.AttrType].(string)),
		Value: aws.String(m[names.AttrValue].(string)),
	}

	return condition
}

func flattenPermissionPolicyStatementCondition(c *permissionPolicyStatementCondition) []interface{} {
	if c == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrKey:   c.Key,
		names.AttrType:  c.Type,
		names.AttrValue: c.Value,
	}

	return []interface{}{m}
}

// https://docs.aws.amazon.com/eventbridge/latest/APIReference/API_PutPermission.html#API_PutPermission_RequestParameters
func validatePermissionAction(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if (len(value) < 1) || (len(value) > 64) {
		es = append(es, fmt.Errorf("%q must be between 1 and 64 characters", k))
	}

	if !regexache.MustCompile(`^events:[A-Za-z]+$`).MatchString(value) {
		es = append(es, fmt.Errorf("%q must be: events: followed by one or more alphabetic characters", k))
	}
	return
}

// https://docs.aws.amazon.com/eventbridge/latest/APIReference/API_PutPermission.html#API_PutPermission_RequestParameters
func validatePermissionPrincipal(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if !regexache.MustCompile(`^(\d{12}|\*)$`).MatchString(value) {
		es = append(es, fmt.Errorf("%q must be * or a 12 digit AWS account ID", k))
	}
	return
}

// https://docs.aws.amazon.com/eventbridge/latest/APIReference/API_PutPermission.html#API_PutPermission_RequestParameters
func validatePermissionStatementID(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if (len(value) < 1) || (len(value) > 64) {
		es = append(es, fmt.Errorf("%q must be between 1 and 64 characters", k))
	}

	if !regexache.MustCompile(`^[0-9A-Za-z_-]+$`).MatchString(value) {
		es = append(es, fmt.Errorf("%q must be one or more alphanumeric, hyphen, or underscore characters", k))
	}
	return
}
