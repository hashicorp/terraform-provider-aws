// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"context"
	"fmt"
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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_datasync_location_object_storage", name="Location Object Storage")
// @Tags(identifierAttribute="id")
func ResourceLocationObjectStorage() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLocationObjectStorageCreate,
		ReadWithoutTimeout:   resourceLocationObjectStorageRead,
		UpdateWithoutTimeout: resourceLocationObjectStorageUpdate,
		DeleteWithoutTimeout: resourceLocationObjectStorageDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"access_key": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(8, 200),
			},
			"agent_arns": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"bucket_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 63),
			},
			"secret_key": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(8, 200),
			},
			"server_certificate": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"server_hostname": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"server_port": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      443,
				ValidateFunc: validation.IsPortNumber,
			},
			"server_protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      datasync.ObjectStorageServerProtocolHttps,
				ValidateFunc: validation.StringInSlice(datasync.ObjectStorageServerProtocol_Values(), false),
			},
			"subdirectory": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 4096),
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

func resourceLocationObjectStorageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	input := &datasync.CreateLocationObjectStorageInput{
		AgentArns:      flex.ExpandStringSet(d.Get("agent_arns").(*schema.Set)),
		Subdirectory:   aws.String(d.Get("subdirectory").(string)),
		BucketName:     aws.String(d.Get("bucket_name").(string)),
		ServerHostname: aws.String(d.Get("server_hostname").(string)),
		Tags:           getTagsIn(ctx),
	}

	if v, ok := d.GetOk("access_key"); ok {
		input.AccessKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("server_protocol"); ok {
		input.ServerProtocol = aws.String(v.(string))
	}

	if v, ok := d.GetOk("server_port"); ok {
		input.ServerPort = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("secret_key"); ok {
		input.SecretKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("server_certificate"); ok {
		input.ServerCertificate = []byte(v.(string))
	}

	output, err := conn.CreateLocationObjectStorageWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataSync Location Object Storage: %s", err)
	}

	d.SetId(aws.StringValue(output.LocationArn))

	return append(diags, resourceLocationObjectStorageRead(ctx, d, meta)...)
}

func resourceLocationObjectStorageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	output, err := FindLocationObjectStorageByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Location Object Storage (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Location Object Storage (%s): %s", d.Id(), err)
	}

	uri := aws.StringValue(output.LocationUri)
	hostname, bucketName, subdirectory, err := decodeObjectStorageURI(uri)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("access_key", output.AccessKey)
	d.Set("agent_arns", aws.StringValueSlice(output.AgentArns))
	d.Set("arn", output.LocationArn)
	d.Set("bucket_name", bucketName)
	d.Set("server_certificate", string(output.ServerCertificate))
	d.Set("server_hostname", hostname)
	d.Set("server_port", output.ServerPort)
	d.Set("server_protocol", output.ServerProtocol)
	d.Set("subdirectory", subdirectory)
	d.Set("uri", uri)

	return diags
}

func resourceLocationObjectStorageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &datasync.UpdateLocationObjectStorageInput{
			LocationArn: aws.String(d.Id()),
		}

		if d.HasChange("access_key") {
			input.AccessKey = aws.String(d.Get("access_key").(string))
		}

		if d.HasChange("secret_key") {
			input.SecretKey = aws.String(d.Get("secret_key").(string))
		}

		if d.HasChange("server_certificate") {
			input.ServerCertificate = []byte(d.Get("server_certificate").(string))
		}

		if d.HasChange("server_port") {
			input.ServerPort = aws.Int64(int64(d.Get("server_port").(int)))
		}

		if d.HasChange("server_protocol") {
			input.ServerProtocol = aws.String(d.Get("server_protocol").(string))
		}

		if d.HasChange("subdirectory") {
			input.Subdirectory = aws.String(d.Get("subdirectory").(string))
		}

		_, err := conn.UpdateLocationObjectStorageWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DataSync Location Object Storage (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceLocationObjectStorageRead(ctx, d, meta)...)
}

func resourceLocationObjectStorageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	input := &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataSync Location Object Storage: %s", d.Id())
	_, err := conn.DeleteLocationWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DataSync Location Object Storage (%s): %s", d.Id(), err)
	}

	return diags
}

func decodeObjectStorageURI(uri string) (string, string, string, error) {
	prefix := "object-storage://"
	if !strings.HasPrefix(uri, prefix) {
		return "", "", "", fmt.Errorf("incorrect uri format needs to start with %s", prefix)
	}
	trimmedUri := strings.TrimPrefix(uri, prefix)
	uriParts := strings.Split(trimmedUri, "/")

	if len(uri) < 2 {
		return "", "", "", fmt.Errorf("incorrect uri format needs to start with %sSERVER-NAME/BUCKET-NAME/SUBDIRECTORY", prefix)
	}

	return uriParts[0], uriParts[1], "/" + strings.Join(uriParts[2:], "/"), nil
}
