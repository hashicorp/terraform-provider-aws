package cloudformation

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceStackSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceStackSetCreate,
		Read:   resourceStackSetRead,
		Update: resourceStackSetUpdate,
		Delete: resourceStackSetDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(StackSetUpdatedDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"administration_role_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"auto_deployment"},
				ValidateFunc:  verify.ValidARN,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_deployment": {
				Type:     schema.TypeList,
				MinItems: 1,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				ConflictsWith: []string{
					"administration_role_arn",
					"execution_role_name",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"retain_stacks_on_account_removal": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"call_as": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(cloudformation.CallAs_Values(), false),
				Default:      cloudformation.CallAsSelf,
			},
			"capabilities": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(cloudformation.Capability_Values(), false),
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"execution_role_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"auto_deployment"},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]+$`), "must contain only alphanumeric and hyphen characters"),
				),
			},
			"operation_preferences": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"failure_tolerance_count": {
							Type:          schema.TypeInt,
							Optional:      true,
							ValidateFunc:  validation.IntAtLeast(0),
							ConflictsWith: []string{"operation_preferences.0.failure_tolerance_percentage"},
						},
						"failure_tolerance_percentage": {
							Type:          schema.TypeInt,
							Optional:      true,
							ValidateFunc:  validation.IntBetween(0, 100),
							ConflictsWith: []string{"operation_preferences.0.failure_tolerance_count"},
						},
						"max_concurrent_count": {
							Type:          schema.TypeInt,
							Optional:      true,
							ValidateFunc:  validation.IntAtLeast(1),
							ConflictsWith: []string{"operation_preferences.0.max_concurrent_percentage"},
						},
						"max_concurrent_percentage": {
							Type:          schema.TypeInt,
							Optional:      true,
							ValidateFunc:  validation.IntBetween(1, 100),
							ConflictsWith: []string{"operation_preferences.0.max_concurrent_count"},
						},
						"region_concurrency_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(cloudformation.RegionConcurrencyType_Values(), false),
						},
						"region_order": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]{1,128}$`), ""),
							},
						},
					},
				},
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"permission_model": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(cloudformation.PermissionModels_Values(), false),
				Default:      cloudformation.PermissionModelsSelfManaged,
			},
			"stack_set_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"template_body": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ConflictsWith:    []string{"template_url"},
				DiffSuppressFunc: verify.SuppressEquivalentJSONOrYAMLDiffs,
				ValidateFunc:     verify.ValidStringIsJSONOrYAML,
			},
			"template_url": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"template_body"},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceStackSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFormationConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &cloudformation.CreateStackSetInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		StackSetName:       aws.String(name),
	}

	if v, ok := d.GetOk("administration_role_arn"); ok {
		input.AdministrationRoleARN = aws.String(v.(string))
	}

	if v, ok := d.GetOk("auto_deployment"); ok {
		input.AutoDeployment = expandAutoDeployment(v.([]interface{}))
	}

	if v, ok := d.GetOk("capabilities"); ok {
		input.Capabilities = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("execution_role_name"); ok {
		input.ExecutionRoleName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameters"); ok {
		input.Parameters = expandParameters(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("permission_model"); ok {
		input.PermissionModel = aws.String(v.(string))
	}

	if v, ok := d.GetOk("call_as"); ok {
		input.CallAs = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("template_body"); ok {
		input.TemplateBody = aws.String(v.(string))
	}

	if v, ok := d.GetOk("template_url"); ok {
		input.TemplateURL = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating CloudFormation StackSet: %s", input)
	_, err := conn.CreateStackSet(input)

	if err != nil {
		return fmt.Errorf("error creating CloudFormation StackSet (%s): %w", name, err)
	}

	d.SetId(name)

	return resourceStackSetRead(d, meta)
}

func resourceStackSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFormationConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	callAs := d.Get("call_as").(string)

	stackSet, err := FindStackSetByName(conn, d.Id(), callAs)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFormation StackSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CloudFormation StackSet (%s): %w", d.Id(), err)
	}

	d.Set("administration_role_arn", stackSet.AdministrationRoleARN)
	d.Set("arn", stackSet.StackSetARN)

	if err := d.Set("auto_deployment", flattenStackSetAutoDeploymentResponse(stackSet.AutoDeployment)); err != nil {
		return fmt.Errorf("error setting auto_deployment: %s", err)
	}

	if err := d.Set("capabilities", aws.StringValueSlice(stackSet.Capabilities)); err != nil {
		return fmt.Errorf("error setting capabilities: %s", err)
	}

	d.Set("description", stackSet.Description)
	d.Set("execution_role_name", stackSet.ExecutionRoleName)
	d.Set("name", stackSet.StackSetName)
	d.Set("permission_model", stackSet.PermissionModel)

	if err := d.Set("parameters", flattenAllParameters(stackSet.Parameters)); err != nil {
		return fmt.Errorf("error setting parameters: %s", err)
	}

	d.Set("stack_set_id", stackSet.StackSetId)

	tags := KeyValueTags(stackSet.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("template_body", stackSet.TemplateBody)

	return nil
}

func resourceStackSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFormationConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &cloudformation.UpdateStackSetInput{
		OperationId:  aws.String(resource.UniqueId()),
		StackSetName: aws.String(d.Id()),
		Tags:         []*cloudformation.Tag{},
		TemplateBody: aws.String(d.Get("template_body").(string)),
	}

	if v, ok := d.GetOk("administration_role_arn"); ok {
		input.AdministrationRoleARN = aws.String(v.(string))
	}

	if v, ok := d.GetOk("capabilities"); ok {
		input.Capabilities = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("execution_role_name"); ok {
		input.ExecutionRoleName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("operation_preferences"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.OperationPreferences = expandOperationPreferences(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("parameters"); ok {
		input.Parameters = expandParameters(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("permission_model"); ok {
		input.PermissionModel = aws.String(v.(string))
	}

	callAs := d.Get("call_as").(string)
	if v, ok := d.GetOk("call_as"); ok {
		input.CallAs = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("template_url"); ok {
		// ValidationError: Exactly one of TemplateBody or TemplateUrl must be specified
		// TemplateBody is always present when TemplateUrl is used so remove TemplateBody if TemplateUrl is set
		input.TemplateBody = nil
		input.TemplateURL = aws.String(v.(string))
	}

	// When `auto_deployment` is set, ignore `administration_role_arn` and
	// `execution_role_name` fields since it's using the SERVICE_MANAGED
	// permission model
	if v, ok := d.GetOk("auto_deployment"); ok {
		input.AdministrationRoleARN = nil
		input.ExecutionRoleName = nil
		input.AutoDeployment = expandAutoDeployment(v.([]interface{}))
	}

	log.Printf("[DEBUG] Updating CloudFormation StackSet: %s", input)
	output, err := conn.UpdateStackSet(input)

	if err != nil {
		return fmt.Errorf("error updating CloudFormation StackSet (%s): %w", d.Id(), err)
	}

	if _, err := WaitStackSetOperationSucceeded(conn, d.Id(), aws.StringValue(output.OperationId), callAs, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return fmt.Errorf("error waiting for CloudFormation StackSet (%s) update: %w", d.Id(), err)
	}

	return resourceStackSetRead(d, meta)
}

func resourceStackSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFormationConn

	input := &cloudformation.DeleteStackSetInput{
		StackSetName: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("call_as"); ok {
		input.CallAs = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Deleting CloudFormation StackSet: %s", d.Id())
	_, err := conn.DeleteStackSet(input)

	if tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeStackSetNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CloudFormation StackSet (%s): %w", d.Id(), err)
	}

	return nil
}

func expandAutoDeployment(l []interface{}) *cloudformation.AutoDeployment {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	autoDeployment := &cloudformation.AutoDeployment{
		Enabled:                      aws.Bool(m["enabled"].(bool)),
		RetainStacksOnAccountRemoval: aws.Bool(m["retain_stacks_on_account_removal"].(bool)),
	}

	return autoDeployment
}

func flattenStackSetAutoDeploymentResponse(autoDeployment *cloudformation.AutoDeployment) []map[string]interface{} {
	if autoDeployment == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"enabled":                          aws.BoolValue(autoDeployment.Enabled),
		"retain_stacks_on_account_removal": aws.BoolValue(autoDeployment.RetainStacksOnAccountRemoval),
	}

	return []map[string]interface{}{m}
}
