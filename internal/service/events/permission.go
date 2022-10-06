package events

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourcePermission() *schema.Resource {
	return &schema.Resource{
		Create: resourcePermissionCreate,
		Read:   resourcePermissionRead,
		Update: resourcePermissionUpdate,
		Delete: resourcePermissionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"action": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "events:PutEvents",
				ValidateFunc: validatePermissionAction,
			},
			"condition": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"aws:PrincipalOrgID"}, false),
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"StringEquals"}, false),
						},
						"value": {
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
				ValidateFunc: validBusNameOrARN,
				Default:      DefaultEventBusName,
			},
			"principal": {
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

func resourcePermissionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn

	eventBusName := d.Get("event_bus_name").(string)
	statementID := d.Get("statement_id").(string)

	input := eventbridge.PutPermissionInput{
		Action:       aws.String(d.Get("action").(string)),
		Condition:    expandCondition(d.Get("condition").([]interface{})),
		EventBusName: aws.String(eventBusName),
		Principal:    aws.String(d.Get("principal").(string)),
		StatementId:  aws.String(statementID),
	}

	log.Printf("[DEBUG] Creating EventBridge permission: %s", input)
	_, err := conn.PutPermission(&input)
	if err != nil {
		return fmt.Errorf("Creating EventBridge permission failed: %w", err)
	}

	id := PermissionCreateResourceID(eventBusName, statementID)
	d.SetId(id)

	return resourcePermissionRead(d, meta)
}

// See also: https://docs.aws.amazon.com/eventbridge/latest/APIReference/API_DescribeEventBus.html
func resourcePermissionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn

	eventBusName, statementID, err := PermissionParseResourceID(d.Id())
	if err != nil {
		return fmt.Errorf("error reading EventBridge permission (%s): %w", d.Id(), err)
	}
	input := eventbridge.DescribeEventBusInput{
		Name: aws.String(eventBusName),
	}
	var output *eventbridge.DescribeEventBusOutput
	var policyStatement *PermissionPolicyStatement

	// Especially with concurrent PutPermission calls there can be a slight delay
	err = resource.Retry(propagationTimeout, func() *resource.RetryError {
		log.Printf("[DEBUG] Reading EventBridge bus: %s", input)
		output, err = conn.DescribeEventBus(&input)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("reading EventBridge permission (%s) failed: %w", d.Id(), err))
		}

		policyStatement, err = getPolicyStatement(output, statementID)
		if err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.DescribeEventBus(&input)
		if output != nil {
			policyStatement, err = getPolicyStatement(output, statementID)
		}
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge permission (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading EventBridge permission (%s): %w", d.Id(), err)
	}

	d.Set("action", policyStatement.Action)
	busName := aws.StringValue(output.Name)
	if busName == "" {
		busName = DefaultEventBusName
	}
	d.Set("event_bus_name", busName)

	if err := d.Set("condition", flattenPermissionPolicyStatementCondition(policyStatement.Condition)); err != nil {
		return fmt.Errorf("error setting condition: %w", err)
	}

	switch principal := policyStatement.Principal.(type) {
	case string:
		d.Set("principal", principal)
	case map[string]interface{}:
		if v, ok := principal["AWS"].(string); ok {
			if arn.IsARN(v) {
				principalARN, err := arn.Parse(v)

				if err != nil {
					return fmt.Errorf("error parsing EventBridge Permission (%s) principal as ARN (%s): %w", d.Id(), v, err)
				}

				d.Set("principal", principalARN.AccountID)
			} else {
				d.Set("principal", v)
			}
		}
	}

	d.Set("statement_id", policyStatement.Sid)

	return nil
}

func getPolicyStatement(output *eventbridge.DescribeEventBusOutput, statementID string) (*PermissionPolicyStatement, error) {
	var policyDoc PermissionPolicyDoc

	if output == nil || output.Policy == nil {
		return nil, &resource.NotFoundError{
			Message:      fmt.Sprintf("EventBridge permission %q not found", statementID),
			LastResponse: output,
		}
	}

	err := json.Unmarshal([]byte(*output.Policy), &policyDoc)
	if err != nil {
		return nil, fmt.Errorf("error reading EventBridge permission (%s): %w", statementID, err)
	}

	return FindPermissionPolicyStatementByID(&policyDoc, statementID)
}

func resourcePermissionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn

	eventBusName, statementID, err := PermissionParseResourceID(d.Id())
	if err != nil {
		return fmt.Errorf("error updating EventBridge permission (%s): %w", d.Id(), err)
	}
	input := eventbridge.PutPermissionInput{
		Action:       aws.String(d.Get("action").(string)),
		Condition:    expandCondition(d.Get("condition").([]interface{})),
		EventBusName: aws.String(eventBusName),
		Principal:    aws.String(d.Get("principal").(string)),
		StatementId:  aws.String(statementID),
	}

	_, err = conn.PutPermission(&input)
	if err != nil {
		return fmt.Errorf("error updating EventBridge permission (%s): %w", d.Id(), err)
	}

	return resourcePermissionRead(d, meta)
}

func resourcePermissionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn

	eventBusName, statementID, err := PermissionParseResourceID(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting EventBridge permission (%s): %w", d.Id(), err)
	}
	input := eventbridge.RemovePermissionInput{
		EventBusName: aws.String(eventBusName),
		StatementId:  aws.String(statementID),
	}

	log.Printf("[DEBUG] Delete EventBridge permission: %s", input)
	_, err = conn.RemovePermission(&input)
	if tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting EventBridge permission (%s): %w", d.Id(), err)
	}
	return nil
}

// https://docs.aws.amazon.com/eventbridge/latest/APIReference/API_PutPermission.html#API_PutPermission_RequestParameters
func validatePermissionAction(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if (len(value) < 1) || (len(value) > 64) {
		es = append(es, fmt.Errorf("%q must be between 1 and 64 characters", k))
	}

	if !regexp.MustCompile(`^events:[a-zA-Z]+$`).MatchString(value) {
		es = append(es, fmt.Errorf("%q must be: events: followed by one or more alphabetic characters", k))
	}
	return
}

// https://docs.aws.amazon.com/eventbridge/latest/APIReference/API_PutPermission.html#API_PutPermission_RequestParameters
func validatePermissionPrincipal(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if !regexp.MustCompile(`^(\d{12}|\*)$`).MatchString(value) {
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

	if !regexp.MustCompile(`^[a-zA-Z0-9-_]+$`).MatchString(value) {
		es = append(es, fmt.Errorf("%q must be one or more alphanumeric, hyphen, or underscore characters", k))
	}
	return
}

// PermissionPolicyDoc represents the Policy attribute of DescribeEventBus
// See also: https://docs.aws.amazon.com/eventbridge/latest/APIReference/API_DescribeEventBus.html
type PermissionPolicyDoc struct {
	Version    string
	ID         string                      `json:"Id,omitempty"`
	Statements []PermissionPolicyStatement `json:"Statement"`
}

// String returns the string representation
func (d PermissionPolicyDoc) String() string {
	return awsutil.Prettify(d)
}

// GoString returns the string representation
func (d PermissionPolicyDoc) GoString() string {
	return d.String()
}

// PermissionPolicyStatement represents the Statement attribute of PermissionPolicyDoc
// See also: https://docs.aws.amazon.com/eventbridge/latest/APIReference/API_DescribeEventBus.html
type PermissionPolicyStatement struct {
	Sid       string
	Effect    string
	Action    string
	Condition *PermissionPolicyStatementCondition `json:"Condition,omitempty"`
	Principal interface{}                         // "*" or {"AWS": "arn:aws:iam::111111111111:root"}
	Resource  string
}

// String returns the string representation
func (s PermissionPolicyStatement) String() string {
	return awsutil.Prettify(s)
}

// GoString returns the string representation
func (s PermissionPolicyStatement) GoString() string {
	return s.String()
}

// PermissionPolicyStatementCondition represents the Condition attribute of PermissionPolicyStatement
// See also: https://docs.aws.amazon.com/eventbridge/latest/APIReference/API_DescribeEventBus.html
type PermissionPolicyStatementCondition struct {
	Key   string
	Type  string
	Value string
}

// String returns the string representation
func (c PermissionPolicyStatementCondition) String() string {
	return awsutil.Prettify(c)
}

// GoString returns the string representation
func (c PermissionPolicyStatementCondition) GoString() string {
	return c.String()
}

func (c *PermissionPolicyStatementCondition) UnmarshalJSON(b []byte) error {
	var out PermissionPolicyStatementCondition

	// JSON representation: \"Condition\":{\"StringEquals\":{\"aws:PrincipalOrgID\":\"o-0123456789\"}}
	var data map[string]map[string]string
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	for typeKey, typeValue := range data {
		for conditionKey, conditionValue := range typeValue {
			out = PermissionPolicyStatementCondition{
				Key:   conditionKey,
				Type:  typeKey,
				Value: conditionValue,
			}
		}
	}

	*c = out
	return nil
}

func FindPermissionPolicyStatementByID(policy *PermissionPolicyDoc, id string) (
	*PermissionPolicyStatement, error) {

	log.Printf("[DEBUG] Finding statement (%s) in EventBridge permission policy: %s", id, policy)
	for _, statement := range policy.Statements {
		if statement.Sid == id {
			return &statement, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastRequest:  id,
		LastResponse: policy,
		Message:      fmt.Sprintf("Failed to find statement (%s) in EventBridge permission policy: %s", id, policy),
	}
}

func expandCondition(l []interface{}) *eventbridge.Condition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	condition := &eventbridge.Condition{
		Key:   aws.String(m["key"].(string)),
		Type:  aws.String(m["type"].(string)),
		Value: aws.String(m["value"].(string)),
	}

	return condition
}

func flattenPermissionPolicyStatementCondition(c *PermissionPolicyStatementCondition) []interface{} {
	if c == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"key":   c.Key,
		"type":  c.Type,
		"value": c.Value,
	}

	return []interface{}{m}
}
