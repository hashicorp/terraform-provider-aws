// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datasync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datasync/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_datasync_location_fsx_windows_file_system", name="Location FSx for Windows File Server File System")
// @Tags(identifierAttribute="id")
func resourceLocationFSxWindowsFileSystem() *schema.Resource {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreationTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomain: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 253),
			},
			"fsx_filesystem_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrPassword: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(1, 104),
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
			names.AttrURI: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 104),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLocationFSxWindowsFileSystemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	input := &datasync.CreateLocationFsxWindowsInput{
		FsxFilesystemArn:  aws.String(d.Get("fsx_filesystem_arn").(string)),
		Password:          aws.String(d.Get(names.AttrPassword).(string)),
		SecurityGroupArns: flex.ExpandStringValueSet(d.Get("security_group_arns").(*schema.Set)),
		Tags:              getTagsIn(ctx),
		User:              aws.String(d.Get("user").(string)),
	}

	if v, ok := d.GetOk(names.AttrDomain); ok {
		input.Domain = aws.String(v.(string))
	}

	if v, ok := d.GetOk("subdirectory"); ok {
		input.Subdirectory = aws.String(v.(string))
	}

	output, err := conn.CreateLocationFsxWindows(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataSync Location FSx for Windows File Server File System: %s", err)
	}

	d.SetId(aws.ToString(output.LocationArn))

	return append(diags, resourceLocationFSxWindowsFileSystemRead(ctx, d, meta)...)
}

func resourceLocationFSxWindowsFileSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	output, err := findLocationFSxWindowsByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Location FSx for Windows File Server File System (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Location FSx for Windows File Server File System (%s): %s", d.Id(), err)
	}

	uri := aws.ToString(output.LocationUri)
	subdirectory, err := subdirectoryFromLocationURI(uri)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrARN, output.LocationArn)
	d.Set(names.AttrCreationTime, output.CreationTime.Format(time.RFC3339))
	d.Set(names.AttrDomain, output.Domain)
	d.Set("fsx_filesystem_arn", d.Get("fsx_filesystem_arn"))
	d.Set("security_group_arns", output.SecurityGroupArns)
	d.Set("subdirectory", subdirectory)
	d.Set(names.AttrURI, uri)
	d.Set("user", output.User)

	return diags
}

func resourceLocationFSxWindowsFileSystemUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceLocationFSxWindowsFileSystemRead(ctx, d, meta)...)
}

func resourceLocationFSxWindowsFileSystemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	log.Printf("[DEBUG] Deleting DataSync Location FSx for Windows File Server File System: %s", d.Id())
	_, err := conn.DeleteLocation(ctx, &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	})

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DataSync Location FSx for Windows File Server File System (%s): %s", d.Id(), err)
	}

	return diags
}

func findLocationFSxWindowsByARN(ctx context.Context, conn *datasync.Client, arn string) (*datasync.DescribeLocationFsxWindowsOutput, error) {
	input := &datasync.DescribeLocationFsxWindowsInput{
		LocationArn: aws.String(arn),
	}

	output, err := conn.DescribeLocationFsxWindows(ctx, input)

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
