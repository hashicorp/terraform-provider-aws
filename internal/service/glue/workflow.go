package glue

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceWorkflow() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkflowCreate,
		ReadWithoutTimeout:   resourceWorkflowRead,
		UpdateWithoutTimeout: resourceWorkflowUpdate,
		DeleteWithoutTimeout: resourceWorkflowDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceWorkflowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	name := d.Get("name").(string)

	input := &glue.CreateWorkflowInput{
		Name: aws.String(name),
		Tags: Tags(tags.IgnoreAWS()),
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
	_, err := conn.CreateWorkflowWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Trigger (%s): %s", name, err)
	}
	d.SetId(name)

	return append(diags, resourceWorkflowRead(ctx, d, meta)...)
}

func resourceWorkflowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &glue.GetWorkflowInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Glue Workflow: %#v", input)
	output, err := conn.GetWorkflowWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			log.Printf("[WARN] Glue Workflow (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Glue Workflow (%s): %s", d.Id(), err)
	}

	workflow := output.Workflow
	if workflow == nil {
		log.Printf("[WARN] Glue Workflow (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
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
		return sdkdiag.AppendErrorf(diags, "setting default_run_properties: %s", err)
	}
	d.Set("description", workflow.Description)
	d.Set("max_concurrent_runs", workflow.MaxConcurrentRuns)
	d.Set("name", workflow.Name)

	tags, err := ListTags(ctx, conn, workFlowArn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Glue Workflow (%s): %s", workFlowArn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceWorkflowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()

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
		_, err := conn.UpdateWorkflowWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Workflow (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	return append(diags, resourceWorkflowRead(ctx, d, meta)...)
}

func resourceWorkflowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()

	log.Printf("[DEBUG] Deleting Glue Workflow: %s", d.Id())
	err := DeleteWorkflow(ctx, conn, d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Workflow (%s): %s", d.Id(), err)
	}

	return diags
}

func DeleteWorkflow(ctx context.Context, conn *glue.Glue, name string) error {
	input := &glue.DeleteWorkflowInput{
		Name: aws.String(name),
	}

	_, err := conn.DeleteWorkflowWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return nil
		}
		return err
	}

	return nil
}
