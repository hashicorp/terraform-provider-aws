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

// @SDKResource("aws_datasync_location_fsx_ontap_file_system", name="Location FSx Ontap File System")
// @Tags(identifierAttribute="id")
func ResourceLocationFSxOntapFileSystem() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLocationFSxOntapFileSystemCreate,
		ReadWithoutTimeout:   resourceLocationFSxOntapFileSystemRead,
		UpdateWithoutTimeout: resourceLocationFSxOntapFileSystemUpdate,
		DeleteWithoutTimeout: resourceLocationFSxOntapFileSystemDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "#")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected DataSyncLocationArn#FsxSVMArn", d.Id())
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
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"nfs": {
							Type:         schema.TypeList,
							Optional:     true,
							ForceNew:     true,
							MaxItems:     1,
							ExactlyOneOf: []string{"protocol.0.nfs", "protocol.0.smb"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mount_options": {
										Type:     schema.TypeList,
										Required: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"version": {
													Type:         schema.TypeString,
													Default:      datasync.NfsVersionNfs3,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringInSlice(FSxOntapNfsVersion_Values(), false),
												},
											},
										},
									},
								},
							},
						},
						"smb": {
							Type:         schema.TypeList,
							Optional:     true,
							ForceNew:     true,
							MaxItems:     1,
							ExactlyOneOf: []string{"protocol.0.nfs", "protocol.0.smb"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"domain": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 253),
									},
									"mount_options": {
										Type:     schema.TypeList,
										Required: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"version": {
													Type:         schema.TypeString,
													Default:      datasync.SmbVersionAutomatic,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringInSlice(FSxOntapSmbVersion_Values(), false),
												},
											},
										},
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
								},
							},
						},
					},
				},
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
			"storage_virtual_machine_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
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

const (
	ResNameLocationFSxOntapFileSystem = "Location FSx Ontap File System"
)

func resourceLocationFSxOntapFileSystemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	input := &datasync.CreateLocationFsxOntapInput{
		Protocol:                 expandProtocol(d.Get("protocol").([]interface{})),
		SecurityGroupArns:        flex.ExpandStringSet(d.Get("security_group_arns").(*schema.Set)),
		StorageVirtualMachineArn: aws.String(d.Get("storage_virtual_machine_arn").(string)),
		Tags:                     getTagsIn(ctx),
	}

	if v, ok := d.GetOk("subdirectory"); ok {
		input.Subdirectory = aws.String(v.(string))
	}

	output, err := conn.CreateLocationFsxOntapWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataSync Location Fsx Ontap File System: %s", err)
	}

	d.SetId(aws.StringValue(output.LocationArn))

	return append(diags, resourceLocationFSxOntapFileSystemRead(ctx, d, meta)...)
}

func resourceLocationFSxOntapFileSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	input := &datasync.DescribeLocationFsxOntapInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading DataSync Location Fsx Ontap: %#v", input)
	output, err := conn.DescribeLocationFsxOntapWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
		log.Printf("[WARN] DataSync Location Fsx Ontap %q not found - removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Location Fsx Ontap (%s): %s", d.Id(), err)
	}

	subdirectory, err := subdirectoryFromLocationURI(aws.StringValue(output.LocationUri))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Location Fsx Ontap (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.LocationArn)
	d.Set("fsx_filesystem_arn", output.FsxFilesystemArn)
	d.Set("storage_virtual_machine_arn", output.StorageVirtualMachineArn)
	d.Set("subdirectory", subdirectory)
	d.Set("uri", output.LocationUri)

	if err := d.Set("creation_time", output.CreationTime.Format(time.RFC3339)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting creation_time: %s", err)
	}

	if err := d.Set("protocol", flattenProtocol(output.Protocol)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting protocol: %s", err)
	}

	if err := d.Set("security_group_arns", flex.FlattenStringSet(output.SecurityGroupArns)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting security_group_arns: %s", err)
	}

	return diags
}

func resourceLocationFSxOntapFileSystemUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceLocationFSxOntapFileSystemRead(ctx, d, meta)...)
}

func resourceLocationFSxOntapFileSystemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	input := &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataSync Location Fsx Ontap File System: %#v", input)
	_, err := conn.DeleteLocationWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DataSync Location Fsx Ontap (%s): %s", d.Id(), err)
	}

	return diags
}

func FSxOntapNfsVersion_Values() []string {
	return []string{
		datasync.NfsVersionNfs3,
	}
}

func FSxOntapSmbVersion_Values() []string {
	return []string{
		datasync.SmbVersionAutomatic,
		datasync.SmbVersionSmb2,
		datasync.SmbVersionSmb3,
		datasync.SmbVersionSmb20,
	}
}
