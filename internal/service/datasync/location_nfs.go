// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datasync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datasync/types"
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

// @SDKResource("aws_datasync_location_nfs", name="Location NFS")
// @Tags(identifierAttribute="id")
func resourceLocationNFS() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLocationNFSCreate,
		ReadWithoutTimeout:   resourceLocationNFSRead,
		UpdateWithoutTimeout: resourceLocationNFSUpdate,
		DeleteWithoutTimeout: resourceLocationNFSDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mount_options": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrVersion: {
							Type:             schema.TypeString,
							Default:          awstypes.NfsVersionAutomatic,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.NfsVersion](),
						},
					},
				},
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
			names.AttrURI: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLocationNFSCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	input := &datasync.CreateLocationNfsInput{
		OnPremConfig:   expandOnPremConfig(d.Get("on_prem_config").([]interface{})),
		ServerHostname: aws.String(d.Get("server_hostname").(string)),
		Subdirectory:   aws.String(d.Get("subdirectory").(string)),
		Tags:           getTagsIn(ctx),
	}

	if v, ok := d.GetOk("mount_options"); ok {
		input.MountOptions = expandNFSMountOptions(v.([]interface{}))
	}

	output, err := conn.CreateLocationNfs(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataSync Location NFS: %s", err)
	}

	d.SetId(aws.ToString(output.LocationArn))

	return append(diags, resourceLocationNFSRead(ctx, d, meta)...)
}

func resourceLocationNFSRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	output, err := findLocationNFSByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Location NFS (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Location NFS (%s): %s", d.Id(), err)
	}

	uri := aws.ToString(output.LocationUri)
	serverHostName, err := globalIDFromLocationURI(uri)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	subdirectory, err := subdirectoryFromLocationURI(uri)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrARN, output.LocationArn)
	if err := d.Set("mount_options", flattenNFSMountOptions(output.MountOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting mount_options: %s", err)
	}
	if err := d.Set("on_prem_config", flattenOnPremConfig(output.OnPremConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting on_prem_config: %s", err)
	}
	d.Set("server_hostname", serverHostName)
	d.Set("subdirectory", subdirectory)
	d.Set(names.AttrURI, uri)

	return diags
}

func resourceLocationNFSUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &datasync.UpdateLocationNfsInput{
			LocationArn:  aws.String(d.Id()),
			OnPremConfig: expandOnPremConfig(d.Get("on_prem_config").([]interface{})),
			Subdirectory: aws.String(d.Get("subdirectory").(string)),
		}

		if v, ok := d.GetOk("mount_options"); ok {
			input.MountOptions = expandNFSMountOptions(v.([]interface{}))
		}

		_, err := conn.UpdateLocationNfs(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DataSync Location NFS (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceLocationNFSRead(ctx, d, meta)...)
}

func resourceLocationNFSDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	log.Printf("[DEBUG] Deleting DataSync Location NFS: %s", d.Id())
	_, err := conn.DeleteLocation(ctx, &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	})

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DataSync Location NFS (%s): %s", d.Id(), err)
	}

	return diags
}

func findLocationNFSByARN(ctx context.Context, conn *datasync.Client, arn string) (*datasync.DescribeLocationNfsOutput, error) {
	input := &datasync.DescribeLocationNfsInput{
		LocationArn: aws.String(arn),
	}

	output, err := conn.DescribeLocationNfs(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "not found") {
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

	return output, nil
}

func expandNFSMountOptions(l []interface{}) *awstypes.NfsMountOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	nfsMountOptions := &awstypes.NfsMountOptions{
		Version: awstypes.NfsVersion(m[names.AttrVersion].(string)),
	}

	return nfsMountOptions
}

func flattenNFSMountOptions(mountOptions *awstypes.NfsMountOptions) []interface{} {
	if mountOptions == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrVersion: string(mountOptions.Version),
	}

	return []interface{}{m}
}

func flattenOnPremConfig(onPremConfig *awstypes.OnPremConfig) []interface{} {
	if onPremConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"agent_arns": onPremConfig.AgentArns,
	}

	return []interface{}{m}
}

func expandOnPremConfig(l []interface{}) *awstypes.OnPremConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	onPremConfig := &awstypes.OnPremConfig{
		AgentArns: flex.ExpandStringValueSet(m["agent_arns"].(*schema.Set)),
	}

	return onPremConfig
}
