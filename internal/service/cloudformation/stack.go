// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudformation_stack", name="Stack")
// @Tags
func ResourceStack() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStackCreate,
		ReadWithoutTimeout:   resourceStackRead,
		UpdateWithoutTimeout: resourceStackUpdate,
		DeleteWithoutTimeout: resourceStackDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"capabilities": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(cloudformation.Capability_Values(), false),
				},
				Set: schema.HashString,
			},
			"disable_rollback": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"iam_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"notification_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"on_failure": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(cloudformation.OnFailure_Values(), false),
			},
			"outputs": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"policy_body": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsJSON,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"policy_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"template_body": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidStringIsJSONOrYAML,
				StateFunc: func(v interface{}) string {
					template, _ := verify.NormalizeJSONOrYAMLString(v)
					return template
				},
			},
			"template_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"timeout_in_minutes": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceStackCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationConn(ctx)

	requestToken := id.UniqueId()
	name := d.Get("name").(string)
	input := &cloudformation.CreateStackInput{
		ClientRequestToken: aws.String(requestToken),
		StackName:          aws.String(name),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("capabilities"); ok {
		input.Capabilities = flex.ExpandStringSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("disable_rollback"); ok {
		input.DisableRollback = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("iam_role_arn"); ok {
		input.RoleARN = aws.String(v.(string))
	}
	if v, ok := d.GetOk("notification_arns"); ok {
		input.NotificationARNs = flex.ExpandStringSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("on_failure"); ok {
		input.OnFailure = aws.String(v.(string))
	}
	if v, ok := d.GetOk("parameters"); ok {
		input.Parameters = expandParameters(v.(map[string]interface{}))
	}
	if v, ok := d.GetOk("policy_body"); ok {
		policy, err := structure.NormalizeJsonString(v)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "policy body contains an invalid JSON: %s", err)
		}
		input.StackPolicyBody = aws.String(policy)
	}
	if v, ok := d.GetOk("policy_url"); ok {
		input.StackPolicyURL = aws.String(v.(string))
	}
	if v, ok := d.GetOk("template_body"); ok {
		template, err := verify.NormalizeJSONOrYAMLString(v)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "template body contains an invalid JSON or YAML: %s", err)
		}
		input.TemplateBody = aws.String(template)
	}
	if v, ok := d.GetOk("template_url"); ok {
		input.TemplateURL = aws.String(v.(string))
	}
	if v, ok := d.GetOk("timeout_in_minutes"); ok {
		input.TimeoutInMinutes = aws.Int64(int64(v.(int)))
	}

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateStackWithContext(ctx, input)
	}, errCodeValidationError, "is invalid or cannot be assumed")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFormation Stack (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*cloudformation.CreateStackOutput).StackId))

	if _, err := WaitStackCreated(ctx, conn, d.Id(), requestToken, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudFormation Stack (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceStackRead(ctx, d, meta)...)
}

func resourceStackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationConn(ctx)

	stack, err := FindStackByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFormation Stack %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFormation Stack (%s): %s", d.Id(), err)
	}

	input := &cloudformation.GetTemplateInput{
		StackName:     aws.String(d.Id()),
		TemplateStage: aws.String(cloudformation.TemplateStageOriginal),
	}

	output, err := conn.GetTemplateWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFormation Stack (%s) template: %s", d.Id(), err)
	}

	template, err := verify.NormalizeJSONOrYAMLString(aws.StringValue(output.TemplateBody))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	d.Set("template_body", template)

	if len(stack.Capabilities) > 0 {
		d.Set("capabilities", aws.StringValueSlice(stack.Capabilities))
	}
	if stack.DisableRollback != nil {
		d.Set("disable_rollback", stack.DisableRollback)

		// takes into account that disable_rollback conflicts with on_failure and
		// prevents forced new creation if disable_rollback is reset during refresh
		if d.Get("on_failure") != nil {
			d.Set("disable_rollback", false)
		}
	}
	d.Set("iam_role_arn", stack.RoleARN)
	d.Set("name", stack.StackName)
	if len(stack.NotificationARNs) > 0 {
		d.Set("notification_arns", aws.StringValueSlice(stack.NotificationARNs))
	}
	if err := d.Set("outputs", flattenOutputs(stack.Outputs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting outputs: %s", err)
	}
	if err := d.Set("parameters", flattenParameters(stack.Parameters, d.Get("parameters").(map[string]interface{}))); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameters: %s", err)
	}
	d.Set("timeout_in_minutes", stack.TimeoutInMinutes)

	setTagsOut(ctx, stack.Tags)

	return diags
}

func resourceStackUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationConn(ctx)

	requestToken := id.UniqueId()
	input := &cloudformation.UpdateStackInput{
		ClientRequestToken: aws.String(requestToken),
		StackName:          aws.String(d.Id()),
		Tags:               []*cloudformation.Tag{},
	}

	// Capabilities must be present whether they are changed or not
	if v, ok := d.GetOk("capabilities"); ok {
		input.Capabilities = flex.ExpandStringSet(v.(*schema.Set))
	}
	if d.HasChange("iam_role_arn") {
		input.RoleARN = aws.String(d.Get("iam_role_arn").(string))
	}
	if d.HasChange("notification_arns") {
		input.NotificationARNs = flex.ExpandStringSet(d.Get("notification_arns").(*schema.Set))
	}
	// Parameters must be present whether they are changed or not
	if v, ok := d.GetOk("parameters"); ok {
		input.Parameters = expandParameters(v.(map[string]interface{}))
	}

	if d.HasChange("policy_body") {
		policy, err := structure.NormalizeJsonString(d.Get("policy_body"))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "policy body contains an invalid JSON: %s", err)
		}
		input.StackPolicyBody = aws.String(policy)
	}
	if d.HasChange("policy_url") {
		input.StackPolicyURL = aws.String(d.Get("policy_url").(string))
	}

	// Either TemplateBody, TemplateURL or UsePreviousTemplate are required
	if v, ok := d.GetOk("template_url"); ok {
		input.TemplateURL = aws.String(v.(string))
	}
	if v, ok := d.GetOk("template_body"); ok && input.TemplateURL == nil {
		template, err := verify.NormalizeJSONOrYAMLString(v)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "template body contains an invalid JSON or YAML: %s", err)
		}
		input.TemplateBody = aws.String(template)
	}

	if tags := getTagsIn(ctx); len(tags) > 0 {
		input.Tags = tags
	}

	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.UpdateStackWithContext(ctx, input)
	}, errCodeValidationError, "is invalid or cannot be assumed")

	if tfawserr.ErrMessageContains(err, errCodeValidationError, "No updates are to be performed") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudFormation Stack (%s): %s", d.Id(), err)
	}

	if _, err := WaitStackUpdated(ctx, conn, d.Id(), requestToken, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudFormation Stack (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceStackRead(ctx, d, meta)...)
}

func resourceStackDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationConn(ctx)

	log.Printf("[INFO] Deleting CloudFormation Stack: %s", d.Id())
	requestToken := id.UniqueId()
	_, err := conn.DeleteStackWithContext(ctx, &cloudformation.DeleteStackInput{
		ClientRequestToken: aws.String(requestToken),
		StackName:          aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeValidationError) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFormation Stack (%s): %s", d.Id(), err)
	}

	if _, err := WaitStackDeleted(ctx, conn, d.Id(), requestToken, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudFormation Stack (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindStackByName(ctx context.Context, conn *cloudformation.CloudFormation, name string) (*cloudformation.Stack, error) {
	input := &cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	}

	output, err := findStack(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.StackStatus); status == cloudformation.StackStatusDeleteComplete {
		return nil, &retry.NotFoundError{
			LastRequest: input,
			Message:     status,
		}
	}

	return output, nil
}

func findStack(ctx context.Context, conn *cloudformation.CloudFormation, input *cloudformation.DescribeStacksInput) (*cloudformation.Stack, error) {
	output, err := findStacks(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findStacks(ctx context.Context, conn *cloudformation.CloudFormation, input *cloudformation.DescribeStacksInput) ([]*cloudformation.Stack, error) {
	var output []*cloudformation.Stack

	err := conn.DescribeStacksPagesWithContext(ctx, input, func(page *cloudformation.DescribeStacksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Stacks {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrMessageContains(err, errCodeValidationError, "does not exist") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusStack(ctx context.Context, conn *cloudformation.CloudFormation, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindStackByName as it maps useful status codes to NotFoundError.
		output, err := findStack(ctx, conn, &cloudformation.DescribeStacksInput{
			StackName: aws.String(name),
		})

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.StackStatus), nil
	}
}

func WaitStackCreated(ctx context.Context, conn *cloudformation.CloudFormation, name, requestToken string, timeout time.Duration) (*cloudformation.Stack, error) {
	const (
		minTimeout = 1 * time.Second
	)
	stateConf := retry.StateChangeConf{
		Pending: []string{
			cloudformation.StackStatusCreateInProgress,
			cloudformation.StackStatusDeleteInProgress,
			cloudformation.StackStatusRollbackInProgress,
		},
		Target: []string{
			cloudformation.StackStatusCreateComplete,
			cloudformation.StackStatusCreateFailed,
			cloudformation.StackStatusDeleteComplete,
			cloudformation.StackStatusDeleteFailed,
			cloudformation.StackStatusRollbackComplete,
			cloudformation.StackStatusRollbackFailed,
		},
		Timeout:    timeout,
		MinTimeout: minTimeout,
		Delay:      10 * time.Second,
		Refresh:    statusStack(ctx, conn, name),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, err
	}

	output, ok := outputRaw.(*cloudformation.Stack)
	if !ok {
		return nil, err
	}

	switch lastStatus := aws.StringValue(output.StackStatus); lastStatus {
	// This will be the case if either disable_rollback is false or on_failure is ROLLBACK
	case cloudformation.StackStatusRollbackComplete, cloudformation.StackStatusRollbackFailed:
		if reasons := getRollbackReasons(ctx, conn, name, requestToken); len(reasons) > 0 {
			return output, fmt.Errorf("failed to create CloudFormation stack, rollback requested (%s): %q", lastStatus, reasons)
		} else {
			return output, fmt.Errorf("failed to create CloudFormation stack (%s): %s", lastStatus, aws.StringValue(output.StackStatusReason))
		}

	// This will be the case if on_failure is DELETE
	case cloudformation.StackStatusDeleteComplete, cloudformation.StackStatusDeleteFailed:
		if reasons := getDeletionReasons(ctx, conn, name, requestToken); len(reasons) > 0 {
			return output, fmt.Errorf("failed to create CloudFormation stack, delete requested (%s): %q", lastStatus, reasons)
		} else {
			return output, fmt.Errorf("failed to create CloudFormation stack (%s): %s", lastStatus, aws.StringValue(output.StackStatusReason))
		}

	// This will be the case if either disable_rollback is true or on_failure is DO_NOTHING
	case cloudformation.StackStatusCreateFailed:
		if reasons := getFailureReasons(ctx, conn, name, requestToken); len(reasons) > 0 {
			return output, fmt.Errorf("failed to create CloudFormation stack (%s): %q", lastStatus, reasons)
		} else {
			return output, fmt.Errorf("failed to create CloudFormation stack (%s): %s", lastStatus, aws.StringValue(output.StackStatusReason))
		}
	}

	return output, err
}

func WaitStackUpdated(ctx context.Context, conn *cloudformation.CloudFormation, name, requestToken string, timeout time.Duration) (*cloudformation.Stack, error) {
	const (
		minTimeout = 5 * time.Second
	)
	stateConf := retry.StateChangeConf{
		Pending: []string{
			cloudformation.StackStatusUpdateCompleteCleanupInProgress,
			cloudformation.StackStatusUpdateInProgress,
			cloudformation.StackStatusUpdateRollbackInProgress,
			cloudformation.StackStatusUpdateRollbackCompleteCleanupInProgress,
		},
		Target: []string{
			cloudformation.StackStatusCreateComplete,
			cloudformation.StackStatusUpdateComplete,
			cloudformation.StackStatusUpdateRollbackComplete,
			cloudformation.StackStatusUpdateRollbackFailed,
		},
		Timeout:    timeout,
		MinTimeout: minTimeout,
		Delay:      10 * time.Second,
		Refresh:    statusStack(ctx, conn, name),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, err
	}

	output, ok := outputRaw.(*cloudformation.Stack)
	if !ok {
		return nil, err
	}

	switch lastStatus := aws.StringValue(output.StackStatus); lastStatus {
	case cloudformation.StackStatusUpdateRollbackComplete, cloudformation.StackStatusUpdateRollbackFailed:
		if reasons := getRollbackReasons(ctx, conn, name, requestToken); len(reasons) > 0 {
			return output, fmt.Errorf("failed to update CloudFormation stack (%s): %q", lastStatus, reasons)
		} else {
			return output, fmt.Errorf("failed to update CloudFormation stack (%s): %s", lastStatus, aws.StringValue(output.StackStatusReason))
		}
	}

	return output, err
}

func WaitStackDeleted(ctx context.Context, conn *cloudformation.CloudFormation, name, requestToken string, timeout time.Duration) (*cloudformation.Stack, error) {
	const (
		minTimeout = 5 * time.Second
	)
	stateConf := retry.StateChangeConf{
		Pending: []string{
			cloudformation.StackStatusDeleteInProgress,
			cloudformation.StackStatusRollbackInProgress,
		},
		Target: []string{
			cloudformation.StackStatusDeleteComplete,
			cloudformation.StackStatusDeleteFailed,
		},
		Timeout:    timeout,
		MinTimeout: minTimeout,
		Delay:      10 * time.Second,
		Refresh:    statusStack(ctx, conn, name),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, err
	}

	output, ok := outputRaw.(*cloudformation.Stack)
	if !ok {
		return nil, err
	}

	switch lastStatus := aws.StringValue(output.StackStatus); lastStatus {
	case cloudformation.StackStatusDeleteFailed:
		if reasons := getFailureReasons(ctx, conn, name, requestToken); len(reasons) > 0 {
			return output, fmt.Errorf("failed to delete CloudFormation stack (%s): %q", lastStatus, reasons)
		} else {
			return output, fmt.Errorf("failed to delete CloudFormation stack (%s): %s", lastStatus, aws.StringValue(output.StackStatusReason))
		}
	}

	return output, err
}

func findStackEventsForOperation(ctx context.Context, conn *cloudformation.CloudFormation, name, requestToken string, filter slices.Predicate[*cloudformation.StackEvent]) ([]*cloudformation.StackEvent, error) {
	input := &cloudformation.DescribeStackEventsInput{
		StackName: aws.String(name),
	}
	var output []*cloudformation.StackEvent
	tokenSeen := false

	err := conn.DescribeStackEventsPagesWithContext(ctx, input, func(page *cloudformation.DescribeStackEventsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.StackEvents {
			if v == nil {
				continue
			}

			if currentToken := aws.StringValue(v.ClientRequestToken); !tokenSeen {
				if currentToken != requestToken {
					continue
				}
				tokenSeen = true
			} else {
				if currentToken != requestToken {
					return false
				}
			}

			if filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	return output, err
}

func getDeletionReasons(ctx context.Context, conn *cloudformation.CloudFormation, name, requestToken string) []string {
	events, err := findStackEventsForOperation(ctx, conn, name, requestToken, func(event *cloudformation.StackEvent) bool {
		return isFailedEvent(event) || isStackDeletionEvent(event)
	})

	if err != nil {
		return nil
	}

	return slices.ApplyToAll(events, reasonFromEvent)
}

func getRollbackReasons(ctx context.Context, conn *cloudformation.CloudFormation, name, requestToken string) []string {
	events, err := findStackEventsForOperation(ctx, conn, name, requestToken, func(event *cloudformation.StackEvent) bool {
		return isFailedEvent(event) || isRollbackEvent(event)
	})

	if err != nil {
		return nil
	}

	return slices.ApplyToAll(events, reasonFromEvent)
}

func getFailureReasons(ctx context.Context, conn *cloudformation.CloudFormation, name, requestToken string) []string {
	events, err := findStackEventsForOperation(ctx, conn, name, requestToken, func(event *cloudformation.StackEvent) bool {
		return isFailedEvent(event)
	})

	if err != nil {
		return nil
	}

	return slices.ApplyToAll(events, reasonFromEvent)
}

func isFailedEvent(event *cloudformation.StackEvent) bool {
	return strings.HasSuffix(aws.StringValue(event.ResourceStatus), "_FAILED") && event.ResourceStatusReason != nil
}

func isRollbackEvent(event *cloudformation.StackEvent) bool {
	return strings.HasPrefix(aws.StringValue(event.ResourceStatus), "ROLLBACK_") && event.ResourceStatusReason != nil
}

func isStackDeletionEvent(event *cloudformation.StackEvent) bool {
	return aws.StringValue(event.ResourceStatus) == cloudformation.ResourceStatusDeleteInProgress &&
		aws.StringValue(event.ResourceType) == "AWS::CloudFormation::Stack" &&
		event.ResourceStatusReason != nil
}

func reasonFromEvent(event *cloudformation.StackEvent) string {
	return aws.StringValue(event.ResourceStatusReason)
}
