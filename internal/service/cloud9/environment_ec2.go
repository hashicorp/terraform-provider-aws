// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloud9

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloud9_environment_ec2", name="Environment EC2")
// @Tags(identifierAttribute="arn")
func ResourceEnvironmentEC2() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEnvironmentEC2Create,
		ReadWithoutTimeout:   resourceEnvironmentEC2Read,
		UpdateWithoutTimeout: resourceEnvironmentEC2Update,
		DeleteWithoutTimeout: resourceEnvironmentEC2Delete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"automatic_stop_time_minutes": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtMost(20160),
			},
			"connection_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      cloud9.ConnectionTypeConnectSsh,
				ValidateFunc: validation.StringInSlice(cloud9.ConnectionType_Values(), false),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 200),
			},
			"image_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"amazonlinux-1-x86_64",
					"amazonlinux-2-x86_64",
					"ubuntu-18.04-x86_64",
					"resolve:ssm:/aws/service/cloud9/amis/amazonlinux-1-x86_64",
					"resolve:ssm:/aws/service/cloud9/amis/amazonlinux-2-x86_64",
					"resolve:ssm:/aws/service/cloud9/amis/ubuntu-18.04-x86_64",
				}, false),
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 60),
			},
			"owner_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEnvironmentEC2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Cloud9Conn(ctx)

	name := d.Get("name").(string)
	input := &cloud9.CreateEnvironmentEC2Input{
		ClientRequestToken: aws.String(id.UniqueId()),
		ConnectionType:     aws.String(d.Get("connection_type").(string)),
		InstanceType:       aws.String(d.Get("instance_type").(string)),
		Name:               aws.String(name),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("automatic_stop_time_minutes"); ok {
		input.AutomaticStopTimeMinutes = aws.Int64(int64(v.(int)))
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("image_id"); ok {
		input.ImageId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("owner_arn"); ok {
		input.OwnerArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("subnet_id"); ok {
		input.SubnetId = aws.String(v.(string))
	}

	log.Printf("[INFO] Creating Cloud9 EC2 Environment: %s", input)
	var output *cloud9.CreateEnvironmentEC2Output
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error
		output, err = conn.CreateEnvironmentEC2WithContext(ctx, input)

		if err != nil {
			// NotFoundException: User arn:aws:iam::*******:user/****** does not exist.
			if tfawserr.ErrMessageContains(err, cloud9.ErrCodeNotFoundException, "User") {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateEnvironmentEC2WithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Cloud9 EC2 Environment (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.EnvironmentId))

	_, err = waitEnvironmentReady(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Cloud9 EC2 Environment (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceEnvironmentEC2Read(ctx, d, meta)...)
}

func resourceEnvironmentEC2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Cloud9Conn(ctx)

	env, err := FindEnvironmentByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Cloud9 EC2 Environment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Cloud9 EC2 Environment (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(env.Arn)
	d.Set("arn", arn)
	d.Set("connection_type", env.ConnectionType)
	d.Set("description", env.Description)
	d.Set("name", env.Name)
	d.Set("owner_arn", env.OwnerArn)
	d.Set("type", env.Type)

	return diags
}

func resourceEnvironmentEC2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Cloud9Conn(ctx)

	if d.HasChangesExcept("tags_all", "tags") {
		input := cloud9.UpdateEnvironmentInput{
			Description:   aws.String(d.Get("description").(string)),
			EnvironmentId: aws.String(d.Id()),
			Name:          aws.String(d.Get("name").(string)),
		}

		log.Printf("[INFO] Updating Cloud9 EC2 Environment: %s", input)
		_, err := conn.UpdateEnvironmentWithContext(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Cloud9 EC2 Environment (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceEnvironmentEC2Read(ctx, d, meta)...)
}

func resourceEnvironmentEC2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Cloud9Conn(ctx)

	log.Printf("[INFO] Deleting Cloud9 EC2 Environment: %s", d.Id())
	_, err := conn.DeleteEnvironmentWithContext(ctx, &cloud9.DeleteEnvironmentInput{
		EnvironmentId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloud9.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Cloud9 EC2 Environment (%s): %s", d.Id(), err)
	}

	_, err = waitEnvironmentDeleted(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Cloud9 EC2 Environment (%s) delete: %s", d.Id(), err)
	}

	return diags
}
