// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appstream_image_builder", name="Image Builder")
// @Tags(identifierAttribute="arn")
func ResourceImageBuilder() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceImageBuilderCreate,
		ReadWithoutTimeout:   resourceImageBuilderRead,
		UpdateWithoutTimeout: resourceImageBuilderUpdate,
		DeleteWithoutTimeout: resourceImageBuilderDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"access_endpoint": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 4,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appstream.AccessEndpointType_Values(), false),
						},
						"vpce_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"appstream_agent_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"display_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			"domain_join_info": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"directory_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"organizational_unit_distinguished_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"enable_default_internet_access": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"iam_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"image_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"image_arn", "image_name"},
				ValidateFunc: verify.ValidARN,
			},
			"image_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"image_name", "image_arn"},
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceImageBuilderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppStreamConn(ctx)

	name := d.Get("name").(string)
	input := &appstream.CreateImageBuilderInput{
		InstanceType: aws.String(d.Get("instance_type").(string)),
		Name:         aws.String(name),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("access_endpoint"); ok && v.(*schema.Set).Len() > 0 {
		input.AccessEndpoints = expandAccessEndpoints(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("appstream_agent_version"); ok {
		input.AppstreamAgentVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("display_name"); ok {
		input.DisplayName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("domain_join_info"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DomainJoinInfo = expandDomainJoinInfo(v.([]interface{}))
	}

	if v, ok := d.GetOk("enable_default_internet_access"); ok {
		input.EnableDefaultInternetAccess = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("iam_role_arn"); ok {
		input.IamRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_arn"); ok {
		input.ImageArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_name"); ok {
		input.ImageName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.VpcConfig = expandImageBuilderVPCConfig(v.([]interface{}))
	}

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, iamPropagationTimeout, func() (interface{}, error) {
		return conn.CreateImageBuilderWithContext(ctx, input)
	}, appstream.ErrCodeInvalidRoleException, "encountered an error because your IAM role")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppStream ImageBuilder (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*appstream.CreateImageBuilderOutput).ImageBuilder.Name))

	if _, err = waitImageBuilderStateRunning(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for AppStream ImageBuilder (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceImageBuilderRead(ctx, d, meta)...)
}

func resourceImageBuilderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppStreamConn(ctx)

	imageBuilder, err := FindImageBuilderByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppStream ImageBuilder (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppStream ImageBuilder (%s): %s", d.Id(), err)
	}

	if err = d.Set("access_endpoint", flattenAccessEndpoints(imageBuilder.AccessEndpoints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting access_endpoint: %s", err)
	}
	d.Set("appstream_agent_version", imageBuilder.AppstreamAgentVersion)
	arn := aws.StringValue(imageBuilder.Arn)
	d.Set("arn", arn)
	d.Set("created_time", aws.TimeValue(imageBuilder.CreatedTime).Format(time.RFC3339))
	d.Set("description", imageBuilder.Description)
	d.Set("display_name", imageBuilder.DisplayName)
	if imageBuilder.DomainJoinInfo != nil {
		if err = d.Set("domain_join_info", []interface{}{flattenDomainInfo(imageBuilder.DomainJoinInfo)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting domain_join_info: %s", err)
		}
	} else {
		d.Set("domain_join_info", nil)
	}
	d.Set("enable_default_internet_access", imageBuilder.EnableDefaultInternetAccess)
	d.Set("iam_role_arn", imageBuilder.IamRoleArn)
	d.Set("image_arn", imageBuilder.ImageArn)
	d.Set("instance_type", imageBuilder.InstanceType)
	d.Set("name", imageBuilder.Name)
	d.Set("state", imageBuilder.State)
	if imageBuilder.VpcConfig != nil {
		if err = d.Set("vpc_config", []interface{}{flattenVPCConfig(imageBuilder.VpcConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting vpc_config: %s", err)
		}
	} else {
		d.Set("vpc_config", nil)
	}

	return diags
}

func resourceImageBuilderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceImageBuilderRead(ctx, d, meta)
}

func resourceImageBuilderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppStreamConn(ctx)

	log.Printf("[DEBUG] Deleting AppStream ImageBuilder: %s", d.Id())
	_, err := conn.DeleteImageBuilderWithContext(ctx, &appstream.DeleteImageBuilderInput{
		Name: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppStream ImageBuilder (%s): %s", d.Id(), err)
	}

	if _, err = waitImageBuilderStateDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for AppStream ImageBuilder (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandImageBuilderVPCConfig(tfList []interface{}) *appstream.VpcConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	apiObject := &appstream.VpcConfig{}

	if v, ok := tfMap["security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringSet(v)
	}
	if v, ok := tfMap["subnet_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringSet(v)
	}

	return apiObject
}
