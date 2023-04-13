package devicefarm

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_devicefarm_project", name="Project")
// @Tags(identifierAttribute="arn")
func ResourceProject() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProjectCreate,
		ReadWithoutTimeout:   resourceProjectRead,
		UpdateWithoutTimeout: resourceProjectUpdate,
		DeleteWithoutTimeout: resourceProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"default_job_timeout_minutes": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn()

	name := d.Get("name").(string)
	input := &devicefarm.CreateProjectInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("default_job_timeout_minutes"); ok {
		input.DefaultJobTimeoutMinutes = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Creating DeviceFarm Project: %s", name)
	out, err := conn.CreateProjectWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error creating DeviceFarm Project: %s", err)
	}

	arn := aws.StringValue(out.Project.Arn)
	log.Printf("[DEBUG] Successsfully Created DeviceFarm Project: %s", arn)
	d.SetId(arn)

	if tags := KeyValueTags(ctx, GetTagsIn(ctx)); len(tags) > 0 {
		if err := UpdateTags(ctx, conn, arn, nil, tags); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DeviceFarm Project (%s) tags: %s", arn, err)
		}
	}

	return append(diags, resourceProjectRead(ctx, d, meta)...)
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn()

	project, err := FindProjectByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DeviceFarm Project (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DeviceFarm Project (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(project.Arn)
	d.Set("name", project.Name)
	d.Set("arn", arn)
	d.Set("default_job_timeout_minutes", project.DefaultJobTimeoutMinutes)

	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn()

	if d.HasChangesExcept("tags", "tags_all") {
		input := &devicefarm.UpdateProjectInput{
			Arn: aws.String(d.Id()),
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("default_job_timeout_minutes") {
			input.DefaultJobTimeoutMinutes = aws.Int64(int64(d.Get("default_job_timeout_minutes").(int)))
		}

		log.Printf("[DEBUG] Updating DeviceFarm Project: %s", d.Id())
		_, err := conn.UpdateProjectWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "Error Updating DeviceFarm Project: %s", err)
		}
	}

	return append(diags, resourceProjectRead(ctx, d, meta)...)
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn()

	input := &devicefarm.DeleteProjectInput{
		Arn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DeviceFarm Project: %s", d.Id())
	_, err := conn.DeleteProjectWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, devicefarm.ErrCodeNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "Error deleting DeviceFarm Project: %s", err)
	}

	return diags
}
