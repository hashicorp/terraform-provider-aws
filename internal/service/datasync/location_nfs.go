// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_datasync_location_nfs", name="Location NFS")
// @Tags(identifierAttribute="id")
func ResourceLocationNFS() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLocationNFSCreate,
		ReadWithoutTimeout:   resourceLocationNFSRead,
		UpdateWithoutTimeout: resourceLocationNFSUpdate,
		DeleteWithoutTimeout: resourceLocationNFSDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"on_prem_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"agent_arns": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
			},
			"mount_options": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:         schema.TypeString,
							Default:      datasync.NfsVersionAutomatic,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(datasync.NfsVersion_Values(), false),
						},
					},
				},
			},
			"server_hostname": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"subdirectory": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 4096),
				// Ignore missing trailing slash
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "/" {
						return false
					}
					if strings.TrimSuffix(old, "/") == strings.TrimSuffix(new, "/") {
						return true
					}
					return false
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLocationNFSCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	input := &datasync.CreateLocationNfsInput{
		OnPremConfig:   expandOnPremConfig(d.Get("on_prem_config").([]interface{})),
		ServerHostname: aws.String(d.Get("server_hostname").(string)),
		Subdirectory:   aws.String(d.Get("subdirectory").(string)),
		Tags:           getTagsIn(ctx),
	}

	if v, ok := d.GetOk("mount_options"); ok {
		input.MountOptions = expandNFSMountOptions(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating DataSync Location NFS: %s", input)
	output, err := conn.CreateLocationNfsWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataSync Location NFS: %s", err)
	}

	d.SetId(aws.StringValue(output.LocationArn))

	return append(diags, resourceLocationNFSRead(ctx, d, meta)...)
}

func resourceLocationNFSRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	input := &datasync.DescribeLocationNfsInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading DataSync Location NFS: %s", input)
	output, err := conn.DescribeLocationNfsWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
		log.Printf("[WARN] DataSync Location NFS %q not found - removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Location NFS (%s): %s", d.Id(), err)
	}

	subdirectory, err := subdirectoryFromLocationURI(aws.StringValue(output.LocationUri))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Location NFS (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.LocationArn)

	if err := d.Set("on_prem_config", flattenOnPremConfig(output.OnPremConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting on_prem_config: %s", err)
	}

	if err := d.Set("mount_options", flattenNFSMountOptions(output.MountOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting mount_options: %s", err)
	}

	d.Set("subdirectory", subdirectory)
	d.Set("uri", output.LocationUri)

	return diags
}

func resourceLocationNFSUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	if d.HasChangesExcept("tags_all", "tags") {
		input := &datasync.UpdateLocationNfsInput{
			LocationArn:  aws.String(d.Id()),
			OnPremConfig: expandOnPremConfig(d.Get("on_prem_config").([]interface{})),
			Subdirectory: aws.String(d.Get("subdirectory").(string)),
		}

		if v, ok := d.GetOk("mount_options"); ok {
			input.MountOptions = expandNFSMountOptions(v.([]interface{}))
		}

		_, err := conn.UpdateLocationNfsWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DataSync Location NFS (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceLocationNFSRead(ctx, d, meta)...)
}

func resourceLocationNFSDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	input := &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataSync Location NFS: %s", input)
	_, err := conn.DeleteLocationWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DataSync Location NFS (%s): %s", d.Id(), err)
	}

	return diags
}

func expandNFSMountOptions(l []interface{}) *datasync.NfsMountOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	nfsMountOptions := &datasync.NfsMountOptions{
		Version: aws.String(m["version"].(string)),
	}

	return nfsMountOptions
}

func flattenNFSMountOptions(mountOptions *datasync.NfsMountOptions) []interface{} {
	if mountOptions == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"version": aws.StringValue(mountOptions.Version),
	}

	return []interface{}{m}
}

func flattenOnPremConfig(onPremConfig *datasync.OnPremConfig) []interface{} {
	if onPremConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"agent_arns": flex.FlattenStringSet(onPremConfig.AgentArns),
	}

	return []interface{}{m}
}

func expandOnPremConfig(l []interface{}) *datasync.OnPremConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	onPremConfig := &datasync.OnPremConfig{
		AgentArns: flex.ExpandStringSet(m["agent_arns"].(*schema.Set)),
	}

	return onPremConfig
}
