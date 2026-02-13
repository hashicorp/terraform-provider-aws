// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package redshift

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_hsm_configuration", name="HSM Configuration")
// @Tags(identifierAttribute="arn")
// @Testing(importIgnore="hsm_partition_password;hsm_server_public_certificate")
func resourceHSMConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHSMConfigurationCreate,
		ReadWithoutTimeout:   resourceHSMConfigurationRead,
		UpdateWithoutTimeout: resourceHSMConfigurationUpdate,
		DeleteWithoutTimeout: resourceHSMConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hsm_configuration_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hsm_ip_address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hsm_partition_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hsm_partition_password": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"hsm_server_public_certificate": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceHSMConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	hsmConfigurationID := d.Get("hsm_configuration_identifier").(string)
	input := redshift.CreateHsmConfigurationInput{
		Description:                aws.String(d.Get(names.AttrDescription).(string)),
		HsmConfigurationIdentifier: aws.String(hsmConfigurationID),
		HsmIpAddress:               aws.String(d.Get("hsm_ip_address").(string)),
		HsmPartitionName:           aws.String(d.Get("hsm_partition_name").(string)),
		HsmPartitionPassword:       aws.String(d.Get("hsm_partition_password").(string)),
		HsmServerPublicCertificate: aws.String(d.Get("hsm_server_public_certificate").(string)),
		Tags:                       getTagsIn(ctx),
	}

	output, err := conn.CreateHsmConfiguration(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift HSM Configuration (%s): %s", hsmConfigurationID, err)
	}

	d.SetId(aws.ToString(output.HsmConfiguration.HsmConfigurationIdentifier))

	return append(diags, resourceHSMConfigurationRead(ctx, d, meta)...)
}

func resourceHSMConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.RedshiftClient(ctx)

	hsmConfiguration, err := findHSMConfigurationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Redshift HSM Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift HSM Configuration (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, hsmConfigurationARN(ctx, c, d.Id()))
	d.Set(names.AttrDescription, hsmConfiguration.Description)
	d.Set("hsm_configuration_identifier", hsmConfiguration.HsmConfigurationIdentifier)
	d.Set("hsm_ip_address", hsmConfiguration.HsmIpAddress)
	d.Set("hsm_partition_name", hsmConfiguration.HsmPartitionName)
	d.Set("hsm_partition_password", d.Get("hsm_partition_password").(string))
	d.Set("hsm_server_public_certificate", d.Get("hsm_server_public_certificate").(string))

	setTagsOut(ctx, hsmConfiguration.Tags)

	return diags
}

func resourceHSMConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceHSMConfigurationRead(ctx, d, meta)...)
}

func resourceHSMConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	log.Printf("[DEBUG] Deleting Redshift HSM Configuration: %s", d.Id())
	input := redshift.DeleteHsmConfigurationInput{
		HsmConfigurationIdentifier: aws.String(d.Id()),
	}
	_, err := conn.DeleteHsmConfiguration(ctx, &input)

	if errs.IsA[*awstypes.HsmConfigurationNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift HSM Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func hsmConfigurationARN(ctx context.Context, c *conns.AWSClient, id string) string {
	return c.RegionalARN(ctx, names.Redshift, "hsmconfiguration:"+id)
}
