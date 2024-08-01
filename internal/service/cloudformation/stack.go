// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudformation_stack", name="Stack")
// @Tags
func resourceStack() *schema.Resource {
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
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.Capability](),
				},
			},
			"disable_rollback": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			names.AttrIAMRoleARN: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"notification_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"on_failure": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.OnFailure](),
			},
			"outputs": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrParameters: {
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

		CustomizeDiff: customdiff.All(
			verify.SetTagsDiff,
			customdiff.ComputedIf("outputs", stackHasActualChanges),
		),
	}
}

func resourceStackCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	requestToken := id.UniqueId()
	name := d.Get(names.AttrName).(string)
	input := &cloudformation.CreateStackInput{
		ClientRequestToken: aws.String(requestToken),
		StackName:          aws.String(name),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("capabilities"); ok {
		input.Capabilities = flex.ExpandStringyValueSet[awstypes.Capability](v.(*schema.Set))
	}
	if v, ok := d.GetOk("disable_rollback"); ok {
		input.DisableRollback = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk(names.AttrIAMRoleARN); ok {
		input.RoleARN = aws.String(v.(string))
	}
	if v, ok := d.GetOk("notification_arns"); ok {
		input.NotificationARNs = flex.ExpandStringValueSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("on_failure"); ok {
		input.OnFailure = awstypes.OnFailure(v.(string))
	}
	if v, ok := d.GetOk(names.AttrParameters); ok {
		input.Parameters = expandParameters(v.(map[string]interface{}))
	}
	if v, ok := d.GetOk("policy_body"); ok {
		policy, err := structure.NormalizeJsonString(v)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		input.StackPolicyBody = aws.String(policy)
	}
	if v, ok := d.GetOk("policy_url"); ok {
		input.StackPolicyURL = aws.String(v.(string))
	}
	if v, ok := d.GetOk("template_body"); ok {
		template, err := verify.NormalizeJSONOrYAMLString(v)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		input.TemplateBody = aws.String(template)
	}
	if v, ok := d.GetOk("template_url"); ok {
		input.TemplateURL = aws.String(v.(string))
	}
	if v, ok := d.GetOk("timeout_in_minutes"); ok {
		input.TimeoutInMinutes = aws.Int32(int32(v.(int)))
	}

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateStack(ctx, input)
	}, errCodeValidationError, "is invalid or cannot be assumed")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFormation Stack (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*cloudformation.CreateStackOutput).StackId))

	if _, err := waitStackCreated(ctx, conn, d.Id(), requestToken, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudFormation Stack (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceStackRead(ctx, d, meta)...)
}

func resourceStackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	stack, err := findStackByName(ctx, conn, d.Id())

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
		TemplateStage: awstypes.TemplateStageOriginal,
	}

	output, err := conn.GetTemplate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFormation Stack (%s) template: %s", d.Id(), err)
	}

	template, err := verify.NormalizeJSONOrYAMLString(aws.ToString(output.TemplateBody))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	d.Set("template_body", template)

	if len(stack.Capabilities) > 0 {
		d.Set("capabilities", stack.Capabilities)
	}
	if stack.DisableRollback != nil {
		d.Set("disable_rollback", stack.DisableRollback)

		// takes into account that disable_rollback conflicts with on_failure and
		// prevents forced new creation if disable_rollback is reset during refresh
		if d.Get("on_failure") != nil {
			d.Set("disable_rollback", false)
		}
	}
	d.Set(names.AttrIAMRoleARN, stack.RoleARN)
	d.Set(names.AttrName, stack.StackName)
	if len(stack.NotificationARNs) > 0 {
		d.Set("notification_arns", stack.NotificationARNs)
	}
	if err := d.Set("outputs", flattenOutputs(stack.Outputs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting outputs: %s", err)
	}
	if err := d.Set(names.AttrParameters, flattenParameters(stack.Parameters, d.Get(names.AttrParameters).(map[string]interface{}))); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameters: %s", err)
	}
	d.Set("timeout_in_minutes", stack.TimeoutInMinutes)

	setTagsOut(ctx, stack.Tags)

	return diags
}

func resourceStackUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	requestToken := id.UniqueId()
	input := &cloudformation.UpdateStackInput{
		ClientRequestToken: aws.String(requestToken),
		StackName:          aws.String(d.Id()),
		Tags:               []awstypes.Tag{},
	}

	// Capabilities must be present whether they are changed or not
	if v, ok := d.GetOk("capabilities"); ok {
		input.Capabilities = flex.ExpandStringyValueSet[awstypes.Capability](v.(*schema.Set))
	}
	if d.HasChange(names.AttrIAMRoleARN) {
		input.RoleARN = aws.String(d.Get(names.AttrIAMRoleARN).(string))
	}
	if d.HasChange("notification_arns") {
		input.NotificationARNs = flex.ExpandStringValueSet(d.Get("notification_arns").(*schema.Set))
	}
	// Parameters must be present whether they are changed or not
	if v, ok := d.GetOk(names.AttrParameters); ok {
		input.Parameters = expandParameters(v.(map[string]interface{}))
	}
	if d.HasChange("policy_body") {
		policy, err := structure.NormalizeJsonString(d.Get("policy_body"))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
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
			return sdkdiag.AppendFromErr(diags, err)
		}
		input.TemplateBody = aws.String(template)
	}

	if tags := getTagsIn(ctx); len(tags) > 0 {
		input.Tags = tags
	}

	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.UpdateStack(ctx, input)
	}, errCodeValidationError, "is invalid or cannot be assumed")

	if tfawserr.ErrMessageContains(err, errCodeValidationError, "No updates are to be performed") {
		return append(diags, resourceStackRead(ctx, d, meta)...)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudFormation Stack (%s): %s", d.Id(), err)
	}

	if _, err := waitStackUpdated(ctx, conn, d.Id(), requestToken, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudFormation Stack (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceStackRead(ctx, d, meta)...)
}

func resourceStackDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	log.Printf("[INFO] Deleting CloudFormation Stack: %s", d.Id())
	requestToken := id.UniqueId()
	_, err := conn.DeleteStack(ctx, &cloudformation.DeleteStackInput{
		ClientRequestToken: aws.String(requestToken),
		StackName:          aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeValidationError) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFormation Stack (%s): %s", d.Id(), err)
	}

	if _, err := waitStackDeleted(ctx, conn, d.Id(), requestToken, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudFormation Stack (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findStackByName(ctx context.Context, conn *cloudformation.Client, name string) (*awstypes.Stack, error) {
	input := &cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	}

	output, err := findStack(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.StackStatus; status == awstypes.StackStatusDeleteComplete {
		return nil, &retry.NotFoundError{
			LastRequest: input,
			Message:     string(status),
		}
	}

	return output, nil
}

func findStack(ctx context.Context, conn *cloudformation.Client, input *cloudformation.DescribeStacksInput) (*awstypes.Stack, error) {
	output, err := findStacks(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findStacks(ctx context.Context, conn *cloudformation.Client, input *cloudformation.DescribeStacksInput) ([]awstypes.Stack, error) {
	var output []awstypes.Stack

	pages := cloudformation.NewDescribeStacksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrMessageContains(err, errCodeValidationError, "does not exist") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Stacks...)
	}

	return output, nil
}

func statusStack(ctx context.Context, conn *cloudformation.Client, name string) retry.StateRefreshFunc {
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

		return output, string(output.StackStatus), nil
	}
}

func waitStackCreated(ctx context.Context, conn *cloudformation.Client, name, requestToken string, timeout time.Duration) (*awstypes.Stack, error) {
	const (
		minTimeout = 1 * time.Second
	)
	stateConf := retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.StackStatusCreateInProgress, awstypes.StackStatusDeleteInProgress, awstypes.StackStatusRollbackInProgress),
		Target:     enum.Slice(awstypes.StackStatusCreateComplete, awstypes.StackStatusCreateFailed, awstypes.StackStatusDeleteComplete, awstypes.StackStatusDeleteFailed, awstypes.StackStatusRollbackComplete, awstypes.StackStatusRollbackFailed),
		Timeout:    timeout,
		MinTimeout: minTimeout,
		Delay:      10 * time.Second,
		Refresh:    statusStack(ctx, conn, name),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, err
	}

	output, ok := outputRaw.(*awstypes.Stack)
	if !ok {
		return nil, err
	}

	var reasonErr error

	switch output.StackStatus {
	// This will be the case if either disable_rollback is false or on_failure is ROLLBACK
	case awstypes.StackStatusRollbackComplete, awstypes.StackStatusRollbackFailed:
		if events := getStackRollbackEvents(ctx, conn, name, requestToken); len(events) > 0 {
			reasonErr = stackEventsError(events)
		} else {
			reasonErr = errors.New(aws.ToString(output.StackStatusReason))
		}

	// This will be the case if on_failure is DELETE
	case awstypes.StackStatusDeleteComplete, awstypes.StackStatusDeleteFailed:
		if events := getStackDeletionEvents(ctx, conn, name, requestToken); len(events) > 0 {
			reasonErr = stackEventsError(events)
		} else {
			reasonErr = errors.New(aws.ToString(output.StackStatusReason))
		}

	// This will be the case if either disable_rollback is true or on_failure is DO_NOTHING
	case awstypes.StackStatusCreateFailed:
		if events := getStackFailureEvents(ctx, conn, name, requestToken); len(events) > 0 {
			reasonErr = stackEventsError(events)
		} else {
			reasonErr = errors.New(aws.ToString(output.StackStatusReason))
		}
	}

	if reasonErr != nil {
		err = fmt.Errorf("stack status (%s): %w", output.StackStatus, reasonErr)
	}

	return output, err
}

func waitStackUpdated(ctx context.Context, conn *cloudformation.Client, name, requestToken string, timeout time.Duration) (*awstypes.Stack, error) {
	const (
		minTimeout = 5 * time.Second
	)
	stateConf := retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.StackStatusUpdateCompleteCleanupInProgress, awstypes.StackStatusUpdateInProgress, awstypes.StackStatusUpdateRollbackInProgress, awstypes.StackStatusUpdateRollbackCompleteCleanupInProgress),
		Target:     enum.Slice(awstypes.StackStatusCreateComplete, awstypes.StackStatusUpdateComplete, awstypes.StackStatusUpdateRollbackComplete, awstypes.StackStatusUpdateRollbackFailed),
		Timeout:    timeout,
		MinTimeout: minTimeout,
		Delay:      10 * time.Second,
		Refresh:    statusStack(ctx, conn, name),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, err
	}

	output, ok := outputRaw.(*awstypes.Stack)
	if !ok {
		return nil, err
	}

	var reasonErr error

	switch output.StackStatus {
	case awstypes.StackStatusUpdateRollbackComplete, awstypes.StackStatusUpdateRollbackFailed:
		if events := getStackRollbackEvents(ctx, conn, name, requestToken); len(events) > 0 {
			reasonErr = stackEventsError(events)
		} else {
			reasonErr = errors.New(aws.ToString(output.StackStatusReason))
		}
	}

	if reasonErr != nil {
		err = fmt.Errorf("stack status (%s): %w", output.StackStatus, reasonErr)
	}

	return output, err
}

func waitStackDeleted(ctx context.Context, conn *cloudformation.Client, name, requestToken string, timeout time.Duration) (*awstypes.Stack, error) {
	const (
		minTimeout = 5 * time.Second
	)
	stateConf := retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.StackStatusDeleteInProgress, awstypes.StackStatusRollbackInProgress),
		Target:     enum.Slice(awstypes.StackStatusDeleteComplete, awstypes.StackStatusDeleteFailed),
		Timeout:    timeout,
		MinTimeout: minTimeout,
		Delay:      10 * time.Second,
		Refresh:    statusStack(ctx, conn, name),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, err
	}

	output, ok := outputRaw.(*awstypes.Stack)
	if !ok {
		return nil, err
	}

	var reasonErr error

	switch output.StackStatus {
	case awstypes.StackStatusDeleteFailed:
		if events := getStackFailureEvents(ctx, conn, name, requestToken); len(events) > 0 {
			reasonErr = stackEventsError(events)
		} else {
			reasonErr = errors.New(aws.ToString(output.StackStatusReason))
		}
	}

	if reasonErr != nil {
		err = fmt.Errorf("stack status (%s): %w", output.StackStatus, reasonErr)
	}

	return output, err
}

func findStackEventsForOperation(ctx context.Context, conn *cloudformation.Client, name, requestToken string, filter tfslices.Predicate[*awstypes.StackEvent]) ([]awstypes.StackEvent, error) {
	input := &cloudformation.DescribeStackEventsInput{
		StackName: aws.String(name),
	}
	var output []awstypes.StackEvent

	pages := cloudformation.NewDescribeStackEventsPaginator(conn, input, func(o *cloudformation.DescribeStackEventsPaginatorOptions) {
		o.StopOnDuplicateToken = true
	})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.StackEvents {
			if aws.ToString(v.ClientRequestToken) != requestToken {
				continue
			}

			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func getStackDeletionEvents(ctx context.Context, conn *cloudformation.Client, name, requestToken string) []awstypes.StackEvent {
	events, err := findStackEventsForOperation(ctx, conn, name, requestToken, tfslices.PredicateOr(isFailedEvent, isStackDeletionEvent))

	if err != nil {
		return nil
	}

	return events
}

func getStackRollbackEvents(ctx context.Context, conn *cloudformation.Client, name, requestToken string) []awstypes.StackEvent {
	events, err := findStackEventsForOperation(ctx, conn, name, requestToken, tfslices.PredicateOr(isFailedEvent, isRollbackEvent))

	if err != nil {
		return nil
	}

	return events
}

func getStackFailureEvents(ctx context.Context, conn *cloudformation.Client, name, requestToken string) []awstypes.StackEvent {
	events, err := findStackEventsForOperation(ctx, conn, name, requestToken, isFailedEvent)

	if err != nil {
		return nil
	}

	return events
}

func isFailedEvent(event *awstypes.StackEvent) bool {
	return strings.HasSuffix(string(event.ResourceStatus), "_FAILED") && event.ResourceStatusReason != nil
}

func isRollbackEvent(event *awstypes.StackEvent) bool {
	return strings.HasPrefix(string(event.ResourceStatus), "ROLLBACK_") && event.ResourceStatusReason != nil
}

func isStackDeletionEvent(event *awstypes.StackEvent) bool {
	return event.ResourceStatus == awstypes.ResourceStatusDeleteInProgress &&
		aws.ToString(event.ResourceType) == "AWS::CloudFormation::Stack" &&
		event.ResourceStatusReason != nil
}

func stackEventsError(events []awstypes.StackEvent) error {
	return errors.Join(tfslices.ApplyToAll(events, func(event awstypes.StackEvent) error { return errors.New(aws.ToString(event.ResourceStatusReason)) })...)
}

func stackHasActualChanges(ctx context.Context, d *schema.ResourceDiff, meta any) bool {
	if d.Id() == "" {
		return false
	}

	for k, attr := range resourceStack().Schema {
		if attr.ForceNew {
			continue
		}
		if attr.Computed && !attr.Optional {
			continue
		}

		if d.HasChange(k) {
			if attr.StateFunc == nil {
				return true
			}
			o, n := d.GetChange(k)
			on := attr.StateFunc(o)
			nn := attr.StateFunc(n)
			if on != nn {
				return true
			}
		}
	}
	return false
}
