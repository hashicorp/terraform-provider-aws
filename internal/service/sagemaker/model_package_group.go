package sagemaker

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceModelPackageGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceModelPackageGroupCreate,
		ReadWithoutTimeout:   resourceModelPackageGroupRead,
		UpdateWithoutTimeout: resourceModelPackageGroupUpdate,
		DeleteWithoutTimeout: resourceModelPackageGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"model_package_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}$`),
						"Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"model_package_group_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceModelPackageGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("model_package_group_name").(string)
	input := &sagemaker.CreateModelPackageGroupInput{
		ModelPackageGroupName: aws.String(name),
	}

	if v, ok := d.GetOk("model_package_group_description"); ok {
		input.ModelPackageGroupDescription = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	_, err := conn.CreateModelPackageGroupWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Model Package Group %s: %s", name, err)
	}

	d.SetId(name)

	if _, err := WaitModelPackageGroupCompleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Model Package Group (%s) to be created: %s", d.Id(), err)
	}

	return append(diags, resourceModelPackageGroupRead(ctx, d, meta)...)
}

func resourceModelPackageGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	mpg, err := FindModelPackageGroupByName(ctx, conn, d.Id())
	if err != nil {
		if tfawserr.ErrMessageContains(err, "ValidationException", "does not exist") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker Model Package Group (%s); removing from state", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Model Package Group (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(mpg.ModelPackageGroupArn)
	d.Set("model_package_group_name", mpg.ModelPackageGroupName)
	d.Set("arn", arn)
	d.Set("model_package_group_description", mpg.ModelPackageGroupDescription)

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for SageMaker Model Package Group (%s): %s", d.Id(), err)
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

func resourceModelPackageGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker Model Package Group (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceModelPackageGroupRead(ctx, d, meta)...)
}

func resourceModelPackageGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn()

	input := &sagemaker.DeleteModelPackageGroupInput{
		ModelPackageGroupName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteModelPackageGroupWithContext(ctx, input); err != nil {
		if tfawserr.ErrMessageContains(err, "ValidationException", "does not exist") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Model Package Group (%s): %s", d.Id(), err)
	}

	if _, err := WaitModelPackageGroupDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrMessageContains(err, "ValidationException", "does not exist") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Model Package Group (%s) to delete: %s", d.Id(), err)
	}

	return diags
}
