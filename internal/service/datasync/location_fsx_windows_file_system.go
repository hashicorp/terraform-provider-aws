// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

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

// @SDKResource("aws_datasync_location_fsx_windows_file_system", name="Location FSx Windows File System")
// @Tags(identifierAttribute="id")
func ResourceLocationFSxWindowsFileSystem() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLocationFSxWindowsFileSystemCreate,
		ReadWithoutTimeout:   resourceLocationFSxWindowsFileSystemRead,
		UpdateWithoutTimeout: resourceLocationFSxWindowsFileSystemUpdate,
		DeleteWithoutTimeout: resourceLocationFSxWindowsFileSystemDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "#")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected DataSyncLocationArn#FsxArn", d.Id())
				}

				DSArn := idParts[0]
				FSxArn := idParts[1]

				d.Set("fsx_filesystem_arn", FSxArn)
				d.SetId(DSArn)

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"fsx_filesystem_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"password": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(1, 104),
			},
			"user": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 104),
			},
			"domain": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 253),
			},
			"security_group_arns": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 5,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"subdirectory": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 4096),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLocationFSxWindowsFileSystemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	fsxArn := d.Get("fsx_filesystem_arn").(string)
	input := &datasync.CreateLocationFsxWindowsInput{
		FsxFilesystemArn:  aws.String(fsxArn),
		User:              aws.String(d.Get("user").(string)),
		Password:          aws.String(d.Get("password").(string)),
		SecurityGroupArns: flex.ExpandStringSet(d.Get("security_group_arns").(*schema.Set)),
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("subdirectory"); ok {
		input.Subdirectory = aws.String(v.(string))
	}

	if v, ok := d.GetOk("domain"); ok {
		input.Domain = aws.String(v.(string))
	}

	output, err := conn.CreateLocationFsxWindowsWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataSync Location Fsx Windows File System: %s", err)
	}

	d.SetId(aws.StringValue(output.LocationArn))

	return append(diags, resourceLocationFSxWindowsFileSystemRead(ctx, d, meta)...)
}

func resourceLocationFSxWindowsFileSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	input := &datasync.DescribeLocationFsxWindowsInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading DataSync Location Fsx Windows: %#v", input)
	output, err := conn.DescribeLocationFsxWindowsWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
		log.Printf("[WARN] DataSync Location Fsx Windows %q not found - removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Location Fsx Windows (%s): %s", d.Id(), err)
	}

	subdirectory, err := subdirectoryFromLocationURI(aws.StringValue(output.LocationUri))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Location Fsx Windows (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.LocationArn)
	d.Set("subdirectory", subdirectory)
	d.Set("uri", output.LocationUri)
	d.Set("user", output.User)
	d.Set("domain", output.Domain)

	if err := d.Set("security_group_arns", flex.FlattenStringSet(output.SecurityGroupArns)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting security_group_arns: %s", err)
	}

	if err := d.Set("creation_time", output.CreationTime.Format(time.RFC3339)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting creation_time: %s", err)
	}

	return diags
}

func resourceLocationFSxWindowsFileSystemUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceLocationFSxWindowsFileSystemRead(ctx, d, meta)...)
}

func resourceLocationFSxWindowsFileSystemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	input := &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataSync Location Fsx Windows File System: %#v", input)
	_, err := conn.DeleteLocationWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DataSync Location Fsx Windows (%s): %s", d.Id(), err)
	}

	return diags
}
