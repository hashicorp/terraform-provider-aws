package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceWorkflow() *schema.Resource {
	return &schema.Resource{
		Create: resourceWorkflowCreate,
		Read:   resourceWorkflowRead,
		Update: resourceWorkflowUpdate,
		Delete: resourceWorkflowDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_run_properties": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"max_concurrent_runs": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceWorkflowCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	name := d.Get("name").(string)

	input := &glue.CreateWorkflowInput{
		Name: aws.String(name),
		Tags: tags.IgnoreAws().GlueTags(),
	}

	if kv, ok := d.GetOk("default_run_properties"); ok {
		input.DefaultRunProperties = flex.ExpandStringMap(kv.(map[string]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_concurrent_runs"); ok {
		input.MaxConcurrentRuns = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Creating Glue Workflow: %s", input)
	_, err := conn.CreateWorkflow(input)
	if err != nil {
		return fmt.Errorf("error creating Glue Trigger (%s): %w", name, err)
	}
	d.SetId(name)

	return resourceWorkflowRead(d, meta)
}

func resourceWorkflowRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &glue.GetWorkflowInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Glue Workflow: %#v", input)
	output, err := conn.GetWorkflow(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue Workflow (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Glue Workflow (%s): %w", d.Id(), err)
	}

	workflow := output.Workflow
	if workflow == nil {
		log.Printf("[WARN] Glue Workflow (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	workFlowArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "glue",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("workflow/%s", d.Id()),
	}.String()
	d.Set("arn", workFlowArn)

	if err := d.Set("default_run_properties", aws.StringValueMap(workflow.DefaultRunProperties)); err != nil {
		return fmt.Errorf("error setting default_run_properties: %w", err)
	}
	d.Set("description", workflow.Description)
	d.Set("max_concurrent_runs", workflow.MaxConcurrentRuns)
	d.Set("name", workflow.Name)

	tags, err := tftags.GlueListTags(conn, workFlowArn)

	if err != nil {
		return fmt.Errorf("error listing tags for Glue Workflow (%s): %w", workFlowArn, err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceWorkflowUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn

	if d.HasChanges("default_run_properties", "description", "max_concurrent_runs") {
		input := &glue.UpdateWorkflowInput{
			Name: aws.String(d.Get("name").(string)),
		}

		if kv, ok := d.GetOk("default_run_properties"); ok {
			input.DefaultRunProperties = flex.ExpandStringMap(kv.(map[string]interface{}))
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("max_concurrent_runs"); ok {
			input.MaxConcurrentRuns = aws.Int64(int64(v.(int)))
		}

		log.Printf("[DEBUG] Updating Glue Workflow: %#v", input)
		_, err := conn.UpdateWorkflow(input)
		if err != nil {
			return fmt.Errorf("error updating Glue Workflow (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := tftags.GlueUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceWorkflowRead(d, meta)
}

func resourceWorkflowDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn

	log.Printf("[DEBUG] Deleting Glue Workflow: %s", d.Id())
	err := deleteWorkflow(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error deleting Glue Workflow (%s): %w", d.Id(), err)
	}

	return nil
}

func deleteWorkflow(conn *glue.Glue, name string) error {
	input := &glue.DeleteWorkflowInput{
		Name: aws.String(name),
	}

	_, err := conn.DeleteWorkflow(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, glue.ErrCodeEntityNotFoundException, "") {
			return nil
		}
		return err
	}

	return nil
}
