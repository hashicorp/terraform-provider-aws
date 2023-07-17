// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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

// @SDKResource("aws_emr_studio", name="Studio")
// @Tags(identifierAttribute="id")
func ResourceStudio() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStudioCreate,
		ReadWithoutTimeout:   resourceStudioRead,
		UpdateWithoutTimeout: resourceStudioUpdate,
		DeleteWithoutTimeout: resourceStudioDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_mode": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(emr.AuthMode_Values(), false),
			},
			"default_s3_location": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"engine_security_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"idp_auth_url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"idp_relay_state_parameter_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"service_role": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				MaxItems: 5,
				MinItems: 1,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"user_role": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"workspace_security_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceStudioCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRConn(ctx)

	input := &emr.CreateStudioInput{
		AuthMode:                 aws.String(d.Get("auth_mode").(string)),
		DefaultS3Location:        aws.String(d.Get("default_s3_location").(string)),
		EngineSecurityGroupId:    aws.String(d.Get("engine_security_group_id").(string)),
		Name:                     aws.String(d.Get("name").(string)),
		ServiceRole:              aws.String(d.Get("service_role").(string)),
		SubnetIds:                flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		Tags:                     getTagsIn(ctx),
		VpcId:                    aws.String(d.Get("vpc_id").(string)),
		WorkspaceSecurityGroupId: aws.String(d.Get("workspace_security_group_id").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("idp_auth_url"); ok {
		input.IdpAuthUrl = aws.String(v.(string))
	}

	if v, ok := d.GetOk("idp_relay_state_parameter_name"); ok {
		input.IdpRelayStateParameterName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("user_role"); ok {
		input.UserRole = aws.String(v.(string))
	}

	var result *emr.CreateStudioOutput
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error
		result, err = conn.CreateStudioWithContext(ctx, input)
		if tfawserr.ErrMessageContains(err, emr.ErrCodeInvalidRequestException, "entity does not have permissions to assume role") {
			return retry.RetryableError(err)
		}
		if tfawserr.ErrMessageContains(err, emr.ErrCodeInvalidRequestException, "Service role does not have permission to access") {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		result, err = conn.CreateStudioWithContext(ctx, input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EMR Studio: %s", err)
	}

	d.SetId(aws.StringValue(result.StudioId))

	return append(diags, resourceStudioRead(ctx, d, meta)...)
}

func resourceStudioUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &emr.UpdateStudioInput{
			StudioId: aws.String(d.Id()),
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("default_s3_location") {
			input.DefaultS3Location = aws.String(d.Get("default_s3_location").(string))
		}

		if d.HasChange("subnet_ids") {
			input.SubnetIds = flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set))
		}

		_, err := conn.UpdateStudioWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EMR Studio: %s", err)
		}
	}

	return append(diags, resourceStudioRead(ctx, d, meta)...)
}

func resourceStudioRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRConn(ctx)

	studio, err := FindStudioByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EMR Studio (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EMR Studio (%s): %s", d.Id(), err)
	}

	d.Set("arn", studio.StudioArn)
	d.Set("auth_mode", studio.AuthMode)
	d.Set("default_s3_location", studio.DefaultS3Location)
	d.Set("description", studio.Description)
	d.Set("engine_security_group_id", studio.EngineSecurityGroupId)
	d.Set("idp_auth_url", studio.IdpAuthUrl)
	d.Set("idp_relay_state_parameter_name", studio.IdpRelayStateParameterName)
	d.Set("name", studio.Name)
	d.Set("service_role", studio.ServiceRole)
	d.Set("url", studio.Url)
	d.Set("user_role", studio.UserRole)
	d.Set("vpc_id", studio.VpcId)
	d.Set("workspace_security_group_id", studio.WorkspaceSecurityGroupId)
	d.Set("subnet_ids", flex.FlattenStringSet(studio.SubnetIds))

	setTagsOut(ctx, studio.Tags)

	return diags
}

func resourceStudioDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRConn(ctx)

	request := &emr.DeleteStudioInput{
		StudioId: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting EMR Studio: %s", d.Id())
	_, err := conn.DeleteStudioWithContext(ctx, request)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, emr.ErrCodeInternalServerException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting EMR Studio (%s): %s", d.Id(), err)
	}

	return diags
}
