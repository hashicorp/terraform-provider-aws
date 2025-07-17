// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_imagebuilder_workflow", name="Workflow")
// @Tags(identifierAttribute="id")
func resourceWorkflow() *schema.Resource {
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
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.WorkflowType](),
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
	}
}

func resourceWorkflowCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &imagebuilder.CreateWorkflowInput{
		ClientToken:     aws.String(id.UniqueId()),
		Name:            aws.String(name),
		SemanticVersion: aws.String(d.Get(names.AttrVersion).(string)),
		Type:            awstypes.WorkflowType(d.Get(names.AttrType).(string)),
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

	output, err := conn.CreateWorkflow(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Workflow (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.WorkflowBuildVersionArn))

	return append(diags, resourceWorkflowRead(ctx, d, meta)...)
}

func resourceWorkflowRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	workflow, err := findWorkflowByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Image Builder Workflow (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Workflow (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, workflow.Arn)
	d.Set("change_description", workflow.ChangeDescription)
	d.Set("data", workflow.Data)
	d.Set("date_created", workflow.DateCreated)
	d.Set(names.AttrDescription, workflow.Description)
	d.Set(names.AttrName, workflow.Name)
	d.Set(names.AttrKMSKeyID, workflow.KmsKeyId)
	d.Set(names.AttrOwner, workflow.Owner)
	d.Set(names.AttrType, workflow.Type)
	d.Set(names.AttrVersion, workflow.Version)

	setTagsOut(ctx, workflow.Tags)

	return diags
}

func resourceWorkflowUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceWorkflowRead(ctx, d, meta)...)
}

func resourceWorkflowDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	log.Printf("[DEBUG] Deleting Image Builder Workflow: %s", d.Id())
	_, err := conn.DeleteWorkflow(ctx, &imagebuilder.DeleteWorkflowInput{
		WorkflowBuildVersionArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Image Builder Workflow (%s): %s", d.Id(), err)
	}

	return diags
}

func findWorkflowByARN(ctx context.Context, conn *imagebuilder.Client, arn string) (*awstypes.Workflow, error) {
	input := &imagebuilder.GetWorkflowInput{
		WorkflowBuildVersionArn: aws.String(arn),
	}

	output, err := conn.GetWorkflow(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Workflow == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Workflow, nil
}
