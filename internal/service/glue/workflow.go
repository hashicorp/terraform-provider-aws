// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_workflow", name="Workflow")
// @Tags(identifierAttribute="arn")
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
				Elem:     &schema.Schema{Type: schema.TypeString},
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceWorkflowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

	name := d.Get("name").(string)
	input := &glue.CreateWorkflowInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
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
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

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

	return diags
}

func resourceWorkflowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

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

	return append(diags, resourceWorkflowRead(ctx, d, meta)...)
}

func resourceWorkflowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

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
