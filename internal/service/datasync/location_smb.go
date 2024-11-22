// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"context"
	"log"

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

// @SDKResource("aws_datasync_location_smb", name="Location SMB")
// @Tags(identifierAttribute="id")
func resourceLocationSMB() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLocationSMBCreate,
		ReadWithoutTimeout:   resourceLocationSMBRead,
		UpdateWithoutTimeout: resourceLocationSMBUpdate,
		DeleteWithoutTimeout: resourceLocationSMBDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"agent_arns": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomain: {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 253),
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
							Default:          awstypes.SmbVersionAutomatic,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.SmbVersion](),
						},
					},
				},
			},
			names.AttrPassword: {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(1, 104),
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
				/*// Ignore missing trailing slash
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "/" {
						return false
					}
					if strings.TrimSuffix(old, "/") == strings.TrimSuffix(new, "/") {
						return true
					}
					return false
				},
				*/
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrURI: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 104),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLocationSMBCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	input := &datasync.CreateLocationSmbInput{
		AgentArns:      flex.ExpandStringValueSet(d.Get("agent_arns").(*schema.Set)),
		MountOptions:   expandSMBMountOptions(d.Get("mount_options").([]interface{})),
		Password:       aws.String(d.Get(names.AttrPassword).(string)),
		ServerHostname: aws.String(d.Get("server_hostname").(string)),
		Subdirectory:   aws.String(d.Get("subdirectory").(string)),
		Tags:           getTagsIn(ctx),
		User:           aws.String(d.Get("user").(string)),
	}

	if v, ok := d.GetOk(names.AttrDomain); ok {
		input.Domain = aws.String(v.(string))
	}

	output, err := conn.CreateLocationSmb(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataSync Location SMB: %s", err)
	}

	d.SetId(aws.ToString(output.LocationArn))

	return append(diags, resourceLocationSMBRead(ctx, d, meta)...)
}

func resourceLocationSMBRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	output, err := findLocationSMBByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Location SMB (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Location SMB (%s): %s", d.Id(), err)
	}

	uri := aws.ToString(output.LocationUri)
	serverHostName, err := globalIDFromLocationURI(uri)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	subdirectory, err := subdirectoryFromLocationURI(aws.ToString(output.LocationUri))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("agent_arns", output.AgentArns)
	d.Set(names.AttrARN, output.LocationArn)
	d.Set(names.AttrDomain, output.Domain)
	if err := d.Set("mount_options", flattenSMBMountOptions(output.MountOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting mount_options: %s", err)
	}
	d.Set("server_hostname", serverHostName)
	d.Set("subdirectory", subdirectory)
	d.Set(names.AttrURI, uri)
	d.Set("user", output.User)

	return diags
}

func resourceLocationSMBUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &datasync.UpdateLocationSmbInput{
			LocationArn:  aws.String(d.Id()),
			AgentArns:    flex.ExpandStringValueSet(d.Get("agent_arns").(*schema.Set)),
			MountOptions: expandSMBMountOptions(d.Get("mount_options").([]interface{})),
			Password:     aws.String(d.Get(names.AttrPassword).(string)),
			Subdirectory: aws.String(d.Get("subdirectory").(string)),
			User:         aws.String(d.Get("user").(string)),
		}

		if v, ok := d.GetOk(names.AttrDomain); ok {
			input.Domain = aws.String(v.(string))
		}

		_, err := conn.UpdateLocationSmb(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DataSync Location SMB (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceLocationSMBRead(ctx, d, meta)...)
}

func resourceLocationSMBDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	log.Printf("[DEBUG] Deleting DataSync Location SMB: %s", d.Id())
	_, err := conn.DeleteLocation(ctx, &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	})

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DataSync Location SMB (%s): %s", d.Id(), err)
	}

	return diags
}

func findLocationSMBByARN(ctx context.Context, conn *datasync.Client, arn string) (*datasync.DescribeLocationSmbOutput, error) {
	input := &datasync.DescribeLocationSmbInput{
		LocationArn: aws.String(arn),
	}

	output, err := conn.DescribeLocationSmb(ctx, input)

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

func flattenSMBMountOptions(mountOptions *awstypes.SmbMountOptions) []interface{} {
	if mountOptions == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrVersion: string(mountOptions.Version),
	}

	return []interface{}{m}
}

func expandSMBMountOptions(l []interface{}) *awstypes.SmbMountOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	smbMountOptions := &awstypes.SmbMountOptions{
		Version: awstypes.SmbVersion(m[names.AttrVersion].(string)),
	}

	return smbMountOptions
}
