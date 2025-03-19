// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emr/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_emr_studio", name="Studio")
// @Tags(identifierAttribute="id")
func resourceStudio() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStudioCreate,
		ReadWithoutTimeout:   resourceStudioRead,
		UpdateWithoutTimeout: resourceStudioUpdate,
		DeleteWithoutTimeout: resourceStudioDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_mode": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AuthMode](),
			},
			"default_s3_location": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"encryption_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
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
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			names.AttrServiceRole: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrSubnetIDs: {
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
			names.AttrURL: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCID: {
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

func resourceStudioCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &emr.CreateStudioInput{
		AuthMode:                 awstypes.AuthMode(d.Get("auth_mode").(string)),
		DefaultS3Location:        aws.String(d.Get("default_s3_location").(string)),
		EngineSecurityGroupId:    aws.String(d.Get("engine_security_group_id").(string)),
		Name:                     aws.String(name),
		ServiceRole:              aws.String(d.Get(names.AttrServiceRole).(string)),
		SubnetIds:                flex.ExpandStringValueSet(d.Get(names.AttrSubnetIDs).(*schema.Set)),
		Tags:                     getTagsIn(ctx),
		VpcId:                    aws.String(d.Get(names.AttrVPCID).(string)),
		WorkspaceSecurityGroupId: aws.String(d.Get("workspace_security_group_id").(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("encryption_key_arn"); ok {
		input.EncryptionKeyArn = aws.String(v.(string))
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

	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (any, error) {
			return conn.CreateStudio(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "entity does not have permissions to assume role") ||
				errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "Service role does not have permission to access") ||
				errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "'ServiceRole' does not have permission to access") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EMR Studio (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*emr.CreateStudioOutput).StudioId))

	return append(diags, resourceStudioRead(ctx, d, meta)...)
}

func resourceStudioRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	studio, err := findStudioByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EMR Studio (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EMR Studio (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, studio.StudioArn)
	d.Set("auth_mode", studio.AuthMode)
	d.Set("default_s3_location", studio.DefaultS3Location)
	d.Set(names.AttrDescription, studio.Description)
	d.Set("encryption_key_arn", studio.EncryptionKeyArn)
	d.Set("engine_security_group_id", studio.EngineSecurityGroupId)
	d.Set("idp_auth_url", studio.IdpAuthUrl)
	d.Set("idp_relay_state_parameter_name", studio.IdpRelayStateParameterName)
	d.Set(names.AttrName, studio.Name)
	d.Set(names.AttrServiceRole, studio.ServiceRole)
	d.Set(names.AttrSubnetIDs, studio.SubnetIds)
	d.Set(names.AttrURL, studio.Url)
	d.Set("user_role", studio.UserRole)
	d.Set(names.AttrVPCID, studio.VpcId)
	d.Set("workspace_security_group_id", studio.WorkspaceSecurityGroupId)

	setTagsOut(ctx, studio.Tags)

	return diags
}

func resourceStudioUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &emr.UpdateStudioInput{
			StudioId: aws.String(d.Id()),
		}

		if d.HasChange("default_s3_location") {
			input.DefaultS3Location = aws.String(d.Get("default_s3_location").(string))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("encryption_key_arn") {
			input.EncryptionKeyArn = aws.String(d.Get("encryption_key_arn").(string))
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange(names.AttrSubnetIDs) {
			input.SubnetIds = flex.ExpandStringValueSet(d.Get(names.AttrSubnetIDs).(*schema.Set))
		}

		_, err := conn.UpdateStudio(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EMR Studio: %s", err)
		}
	}

	return append(diags, resourceStudioRead(ctx, d, meta)...)
}

func resourceStudioDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	log.Printf("[INFO] Deleting EMR Studio: %s", d.Id())
	_, err := conn.DeleteStudio(ctx, &emr.DeleteStudioInput{
		StudioId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.InternalServerException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EMR Studio (%s): %s", d.Id(), err)
	}

	return diags
}

func findStudioByID(ctx context.Context, conn *emr.Client, id string) (*awstypes.Studio, error) {
	input := &emr.DescribeStudioInput{
		StudioId: aws.String(id),
	}

	output, err := conn.DescribeStudio(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "Studio does not exist") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Studio == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Studio, nil
}
