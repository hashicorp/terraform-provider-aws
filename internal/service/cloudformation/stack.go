package cloudformation

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

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
			Create: schema.DefaultTimeout(StackCreatedDefaultTimeout),
			Update: schema.DefaultTimeout(StackUpdatedDefaultTimeout),
			Delete: schema.DefaultTimeout(StackDeletedDefaultTimeout),
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
	conn := meta.(*conns.AWSClient).CloudFormationConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	requestToken := resource.UniqueId()
	name := d.Get("name").(string)
	input := &cloudformation.CreateStackInput{
		ClientRequestToken: aws.String(requestToken),
		StackName:          aws.String(name),
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
	if v, ok := d.GetOk("capabilities"); ok {
		input.Capabilities = flex.ExpandStringSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("disable_rollback"); ok {
		input.DisableRollback = aws.Bool(v.(bool))
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
	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}
	if v, ok := d.GetOk("timeout_in_minutes"); ok {
		m := int64(v.(int))
		input.TimeoutInMinutes = aws.Int64(m)
	}
	if v, ok := d.GetOk("iam_role_arn"); ok {
		input.RoleARN = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating CloudFormation Stack: %s", input)
	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateStackWithContext(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, "ValidationError", "is invalid or cannot be assumed") {
				return true, err
			}

			return false, err
		},
	)

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
	conn := meta.(*conns.AWSClient).CloudFormationConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &cloudformation.DescribeStacksInput{
		StackName: aws.String(d.Id()),
	}
	resp, err := conn.DescribeStacksWithContext(ctx, input)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, "ValidationError") {
		create.LogNotFoundRemoveState(names.CloudFormation, create.ErrActionReading, ResNameStack, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.CloudFormation, create.ErrActionReading, ResNameStack, d.Id(), err)
	}

	stacks := resp.Stacks
	if !d.IsNewResource() && len(stacks) < 1 {
		create.LogNotFoundRemoveState(names.CloudFormation, create.ErrActionReading, ResNameStack, d.Id())
		d.SetId("")
		return diags
	}

	if d.IsNewResource() && len(stacks) < 1 {
		return create.DiagError(names.CloudFormation, create.ErrActionReading, ResNameStack, d.Id(), errors.New("not found after creation"))
	}

	stack := stacks[0]
	if !d.IsNewResource() && aws.StringValue(stack.StackStatus) == cloudformation.StackStatusDeleteComplete {
		log.Printf("[WARN] CloudFormation stack (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if d.IsNewResource() && aws.StringValue(stack.StackStatus) == cloudformation.StackStatusDeleteComplete {
		return create.DiagError(names.CloudFormation, create.ErrActionReading, ResNameStack, d.Id(), errors.New("status delete complete after creation"))
	}

	tInput := cloudformation.GetTemplateInput{
		StackName:     aws.String(d.Id()),
		TemplateStage: aws.String("Original"),
	}
	out, err := conn.GetTemplateWithContext(ctx, &tInput)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFormation Stack (%s): reading template: %s", d.Id(), err)
	}

	template, err := verify.NormalizeJSONOrYAMLString(*out.TemplateBody)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "template body contains an invalid JSON or YAML: %s", err)
	}
	d.Set("template_body", template)

	log.Printf("[DEBUG] Received CloudFormation stack: %s", stack)

	d.Set("name", stack.StackName)
	d.Set("iam_role_arn", stack.RoleARN)
	d.Set("timeout_in_minutes", stack.TimeoutInMinutes)

	if stack.DisableRollback != nil {
		d.Set("disable_rollback", stack.DisableRollback)

		// takes into account that disable_rollback conflicts with on_failure and
		// prevents forced new creation if disable_rollback is reset during refresh
		if d.Get("on_failure") != nil {
			d.Set("disable_rollback", false)
		}
	}
	if len(stack.NotificationARNs) > 0 {
		err = d.Set("notification_arns", flex.FlattenStringSet(stack.NotificationARNs))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading CloudFormation Stack (%s): %s", d.Id(), err)
		}
	}

	originalParams := d.Get("parameters").(map[string]interface{})
	err = d.Set("parameters", flattenParameters(stack.Parameters, originalParams))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFormation Stack (%s): %s", d.Id(), err)
	}

	tags := KeyValueTags(stack.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	err = d.Set("outputs", flattenOutputs(stack.Outputs))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFormation Stack (%s): %s", d.Id(), err)
	}

	if len(stack.Capabilities) > 0 {
		err = d.Set("capabilities", flex.FlattenStringSet(stack.Capabilities))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading CloudFormation Stack (%s): %s", d.Id(), err)
		}
	}

	return diags
}

func resourceStackUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	requestToken := resource.UniqueId()
	input := &cloudformation.UpdateStackInput{
		StackName:          aws.String(d.Id()),
		ClientRequestToken: aws.String(requestToken),
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

	// Capabilities must be present whether they are changed or not
	if v, ok := d.GetOk("capabilities"); ok {
		input.Capabilities = flex.ExpandStringSet(v.(*schema.Set))
	}

	if d.HasChange("notification_arns") {
		input.NotificationARNs = flex.ExpandStringSet(d.Get("notification_arns").(*schema.Set))
	}

	// Parameters must be present whether they are changed or not
	if v, ok := d.GetOk("parameters"); ok {
		input.Parameters = expandParameters(v.(map[string]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
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

	if d.HasChange("iam_role_arn") {
		input.RoleARN = aws.String(d.Get("iam_role_arn").(string))
	}

	log.Printf("[DEBUG] Updating CloudFormation Stack: %s", input)
	_, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.UpdateStackWithContext(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, "ValidationError", "is invalid or cannot be assumed") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil && !tfawserr.ErrMessageContains(err, "ValidationError", "No updates are to be performed") {
		return sdkdiag.AppendErrorf(diags, "updating CloudFormation Stack (%s): %s", d.Id(), err)
	}

	_, err = WaitStackUpdated(ctx, conn, d.Id(), requestToken, d.Timeout(schema.TimeoutUpdate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudFormation Stack (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceStackRead(ctx, d, meta)...)
}

func resourceStackDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationConn()

	requestToken := resource.UniqueId()
	input := &cloudformation.DeleteStackInput{
		StackName:          aws.String(d.Id()),
		ClientRequestToken: aws.String(requestToken),
	}
	_, err := conn.DeleteStackWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, "ValidationError") {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFormation Stack (%s): %s", d.Id(), err)
	}

	_, err = WaitStackDeleted(ctx, conn, d.Id(), requestToken, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFormation Stack (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}
