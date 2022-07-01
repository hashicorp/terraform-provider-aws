package cloudformation

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceStack() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceStackRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"template_body": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capabilities": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disable_rollback": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"notification_arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"parameters": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"outputs": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"timeout_in_minutes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"iam_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceStackRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFormationConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	input := &cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	}

	log.Printf("[DEBUG] Reading CloudFormation Stack: %s", input)
	out, err := conn.DescribeStacks(input)
	if err != nil {
		return fmt.Errorf("Failed describing CloudFormation stack (%s): %w", name, err)
	}
	if l := len(out.Stacks); l != 1 {
		return fmt.Errorf("Expected 1 CloudFormation stack (%s), found %d", name, l)
	}
	stack := out.Stacks[0]
	d.SetId(aws.StringValue(stack.StackId))

	d.Set("description", stack.Description)
	d.Set("disable_rollback", stack.DisableRollback)
	d.Set("timeout_in_minutes", stack.TimeoutInMinutes)
	d.Set("iam_role_arn", stack.RoleARN)

	if len(stack.NotificationARNs) > 0 {
		d.Set("notification_arns", flex.FlattenStringSet(stack.NotificationARNs))
	}

	d.Set("parameters", flattenAllParameters(stack.Parameters))
	if err := d.Set("tags", KeyValueTags(stack.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}
	d.Set("outputs", flattenOutputs(stack.Outputs))

	if len(stack.Capabilities) > 0 {
		d.Set("capabilities", flex.FlattenStringSet(stack.Capabilities))
	}

	tInput := cloudformation.GetTemplateInput{
		StackName: aws.String(name),
	}
	tOut, err := conn.GetTemplate(&tInput)
	if err != nil {
		return err
	}

	template, err := verify.NormalizeJSONOrYAMLString(*tOut.TemplateBody)
	if err != nil {
		return fmt.Errorf("template body contains an invalid JSON or YAML: %w", err)
	}
	d.Set("template_body", template)

	return nil
}
