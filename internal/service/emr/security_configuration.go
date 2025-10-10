// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emr/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_emr_security_configuration", name="Security Configuration")
func resourceSecurityConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecurityConfigurationCreate,
		ReadWithoutTimeout:   resourceSecurityConfigurationRead,
		DeleteWithoutTimeout: resourceSecurityConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrConfiguration: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsJSON,
			},
			names.AttrCreationDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validation.StringLenBetween(0, 10280),
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validation.StringLenBetween(0, 10280-id.UniqueIDSuffixLength),
			},
		},
	}
}

func resourceSecurityConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	name := create.NewNameGenerator(
		create.WithConfiguredName(d.Get(names.AttrName).(string)),
		create.WithConfiguredPrefix(d.Get(names.AttrNamePrefix).(string)),
		create.WithDefaultPrefix("tf-emr-sc-"),
	).Generate()
	input := &emr.CreateSecurityConfigurationInput{
		Name:                  aws.String(name),
		SecurityConfiguration: aws.String(d.Get(names.AttrConfiguration).(string)),
	}

	output, err := conn.CreateSecurityConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EMR Security Configuration (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Name))

	return append(diags, resourceSecurityConfigurationRead(ctx, d, meta)...)
}

func resourceSecurityConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	output, err := findSecurityConfigurationByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EMR Security Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EMR Security Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrConfiguration, output.SecurityConfiguration)
	d.Set(names.AttrCreationDate, aws.ToTime(output.CreationDateTime).Format(time.RFC3339))
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(output.Name)))

	return diags
}

func resourceSecurityConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	log.Printf("[INFO] Deleting EMR Security Configuration: %s", d.Id())
	_, err := conn.DeleteSecurityConfiguration(ctx, &emr.DeleteSecurityConfigurationInput{
		Name: aws.String(d.Id()),
	})

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EMR Security Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func findSecurityConfigurationByName(ctx context.Context, conn *emr.Client, name string) (*emr.DescribeSecurityConfigurationOutput, error) {
	input := &emr.DescribeSecurityConfigurationInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeSecurityConfiguration(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "does not exist") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, err
}
