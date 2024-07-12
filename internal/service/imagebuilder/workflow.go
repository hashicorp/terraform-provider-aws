// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_imagebuilder_workflow", name="Workflow")
// @Tags(identifierAttribute="id")
func ResourceWorkflow() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkflowCreate,
		ReadWithoutTimeout:   resourceWorkflowRead,
		UpdateWithoutTimeout: resourceWorkflowUpdate,
		DeleteWithoutTimeout: resourceWorkflowDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"change_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"data": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"data", names.AttrURI},
				ValidateFunc: validation.StringLenBetween(1, 16000),
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			names.AttrOwner: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(imagebuilder.WorkflowType_Values(), false),
			},
			names.AttrURI: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"data", names.AttrURI},
			},
			names.AttrVersion: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`), "valid semantic version must be provided"),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceWorkflowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn(ctx)

	input := &imagebuilder.CreateWorkflowInput{
		ClientToken:     aws.String(id.UniqueId()),
		Name:            aws.String(d.Get(names.AttrName).(string)),
		SemanticVersion: aws.String(d.Get(names.AttrVersion).(string)),
		Type:            aws.String(d.Get(names.AttrType).(string)),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk("change_description"); ok {
		input.ChangeDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data"); ok {
		input.Data = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrURI); ok {
		input.Uri = aws.String(v.(string))
	}

	output, err := conn.CreateWorkflowWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Workflow: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Workflow: empty response")
	}

	d.SetId(aws.StringValue(output.WorkflowBuildVersionArn))

	return append(diags, resourceWorkflowRead(ctx, d, meta)...)
}

func resourceWorkflowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn(ctx)

	input := &imagebuilder.GetWorkflowInput{
		WorkflowBuildVersionArn: aws.String(d.Id()),
	}

	output, err := conn.GetWorkflowWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Image Builder Workflow (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if output == nil || output.Workflow == nil {
		return sdkdiag.AppendErrorf(diags, "getting Image Builder Workflow (%s): empty response", d.Id())
	}

	workflow := output.Workflow

	d.Set(names.AttrARN, workflow.Arn)
	d.Set("change_description", workflow.ChangeDescription)
	d.Set("data", workflow.Data)
	d.Set("date_created", workflow.DateCreated)
	d.Set(names.AttrDescription, workflow.Description)
	d.Set(names.AttrName, workflow.Name)
	d.Set(names.AttrKMSKeyID, workflow.KmsKeyId)
	d.Set(names.AttrOwner, workflow.Owner)

	setTagsOut(ctx, workflow.Tags)

	d.Set(names.AttrType, workflow.Type)
	d.Set(names.AttrVersion, workflow.Version)

	return diags
}

func resourceWorkflowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceWorkflowRead(ctx, d, meta)...)
}

func resourceWorkflowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn(ctx)

	input := &imagebuilder.DeleteWorkflowInput{
		WorkflowBuildVersionArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteWorkflowWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Image Builder Workflow (%s): %s", d.Id(), err)
	}

	return diags
}
