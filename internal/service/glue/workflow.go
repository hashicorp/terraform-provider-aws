// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_run_properties": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"max_concurrent_runs": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			names.AttrName: {
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
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &glue.CreateWorkflowInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	if kv, ok := d.GetOk("default_run_properties"); ok {
		input.DefaultRunProperties = flex.ExpandStringValueMap(kv.(map[string]interface{}))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_concurrent_runs"); ok {
		input.MaxConcurrentRuns = aws.Int32(int32(v.(int)))
	}

	log.Printf("[DEBUG] Creating Glue Workflow: %+v", input)
	_, err := conn.CreateWorkflow(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Trigger (%s): %s", name, err)
	}
	d.SetId(name)

	return append(diags, resourceWorkflowRead(ctx, d, meta)...)
}

func resourceWorkflowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	input := &glue.GetWorkflowInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Glue Workflow: %#v", input)
	output, err := conn.GetWorkflow(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
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
	d.Set(names.AttrARN, workFlowArn)

	if err := d.Set("default_run_properties", workflow.DefaultRunProperties); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_run_properties: %s", err)
	}
	d.Set(names.AttrDescription, workflow.Description)
	d.Set("max_concurrent_runs", workflow.MaxConcurrentRuns)
	d.Set(names.AttrName, workflow.Name)

	return diags
}

func resourceWorkflowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	if d.HasChanges("default_run_properties", names.AttrDescription, "max_concurrent_runs") {
		input := &glue.UpdateWorkflowInput{
			Name: aws.String(d.Get(names.AttrName).(string)),
		}

		if kv, ok := d.GetOk("default_run_properties"); ok {
			input.DefaultRunProperties = flex.ExpandStringValueMap(kv.(map[string]interface{}))
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("max_concurrent_runs"); ok {
			input.MaxConcurrentRuns = aws.Int32(int32(v.(int)))
		}

		log.Printf("[DEBUG] Updating Glue Workflow: %#v", input)
		_, err := conn.UpdateWorkflow(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Workflow (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceWorkflowRead(ctx, d, meta)...)
}

func resourceWorkflowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	log.Printf("[DEBUG] Deleting Glue Workflow: %s", d.Id())
	err := DeleteWorkflow(ctx, conn, d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Workflow (%s): %s", d.Id(), err)
	}

	return diags
}

func DeleteWorkflow(ctx context.Context, conn *glue.Client, name string) error {
	input := &glue.DeleteWorkflowInput{
		Name: aws.String(name),
	}

	_, err := conn.DeleteWorkflow(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return nil
		}
		return err
	}

	return nil
}
