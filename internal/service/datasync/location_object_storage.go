// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"context"
	"fmt"
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

// @SDKResource("aws_datasync_location_object_storage", name="Location Object Storage")
// @Tags(identifierAttribute="id")
func resourceLocationObjectStorage() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLocationObjectStorageCreate,
		ReadWithoutTimeout:   resourceLocationObjectStorageRead,
		UpdateWithoutTimeout: resourceLocationObjectStorageUpdate,
		DeleteWithoutTimeout: resourceLocationObjectStorageDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccessKey: {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrBucketName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 63),
			},
			names.AttrSecretKey: {
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
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.ObjectStorageServerProtocolHttps,
				ValidateDiagFunc: enum.Validate[awstypes.ObjectStorageServerProtocol](),
			},
			"subdirectory": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 4096),
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

func resourceLocationObjectStorageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	input := &datasync.CreateLocationObjectStorageInput{
		AgentArns:      flex.ExpandStringValueSet(d.Get("agent_arns").(*schema.Set)),
		BucketName:     aws.String(d.Get(names.AttrBucketName).(string)),
		ServerHostname: aws.String(d.Get("server_hostname").(string)),
		Subdirectory:   aws.String(d.Get("subdirectory").(string)),
		Tags:           getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrAccessKey); ok {
		input.AccessKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrSecretKey); ok {
		input.SecretKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("server_certificate"); ok {
		input.ServerCertificate = []byte(v.(string))
	}

	if v, ok := d.GetOk("server_port"); ok {
		input.ServerPort = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("server_protocol"); ok {
		input.ServerProtocol = awstypes.ObjectStorageServerProtocol(v.(string))
	}

	output, err := conn.CreateLocationObjectStorage(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataSync Location Object Storage: %s", err)
	}

	d.SetId(aws.ToString(output.LocationArn))

	return append(diags, resourceLocationObjectStorageRead(ctx, d, meta)...)
}

func resourceLocationObjectStorageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	output, err := findLocationObjectStorageByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Location Object Storage (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Location Object Storage (%s): %s", d.Id(), err)
	}

	uri := aws.ToString(output.LocationUri)
	hostname, bucketName, subdirectory, err := decodeObjectStorageURI(uri)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrAccessKey, output.AccessKey)
	d.Set("agent_arns", output.AgentArns)
	d.Set(names.AttrARN, output.LocationArn)
	d.Set(names.AttrBucketName, bucketName)
	d.Set("server_certificate", string(output.ServerCertificate))
	d.Set("server_hostname", hostname)
	d.Set("server_port", output.ServerPort)
	d.Set("server_protocol", output.ServerProtocol)
	d.Set("subdirectory", subdirectory)
	d.Set(names.AttrURI, uri)

	return diags
}

func resourceLocationObjectStorageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &datasync.UpdateLocationObjectStorageInput{
			LocationArn: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrAccessKey) {
			input.AccessKey = aws.String(d.Get(names.AttrAccessKey).(string))
		}

		if d.HasChange("agent_arns") {
			input.AgentArns = flex.ExpandStringValueSet(d.Get("agent_arns").(*schema.Set))

			// Access key must be specified when updating agent ARNs
			input.AccessKey = aws.String("")
			if v, ok := d.GetOk(names.AttrAccessKey); ok {
				input.AccessKey = aws.String(v.(string))
			}

			// Secret key must be specified when updating agent ARNs
			input.SecretKey = aws.String("")
			if v, ok := d.GetOk(names.AttrSecretKey); ok {
				input.SecretKey = aws.String(v.(string))
			}
		}

		if d.HasChange(names.AttrSecretKey) {
			input.SecretKey = aws.String(d.Get(names.AttrSecretKey).(string))
		}

		if d.HasChange("server_certificate") {
			input.ServerCertificate = []byte(d.Get("server_certificate").(string))
		}

		if d.HasChange("server_port") {
			input.ServerPort = aws.Int32(int32(d.Get("server_port").(int)))
		}

		if d.HasChange("server_protocol") {
			input.ServerProtocol = awstypes.ObjectStorageServerProtocol(d.Get("server_protocol").(string))
		}

		if d.HasChange("subdirectory") {
			input.Subdirectory = aws.String(d.Get("subdirectory").(string))
		}

		_, err := conn.UpdateLocationObjectStorage(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DataSync Location Object Storage (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceLocationObjectStorageRead(ctx, d, meta)...)
}

func resourceLocationObjectStorageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	log.Printf("[DEBUG] Deleting DataSync Location Object Storage: %s", d.Id())
	_, err := conn.DeleteLocation(ctx, &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	})

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DataSync Location Object Storage (%s): %s", d.Id(), err)
	}

	return diags
}

func findLocationObjectStorageByARN(ctx context.Context, conn *datasync.Client, arn string) (*datasync.DescribeLocationObjectStorageOutput, error) {
	input := &datasync.DescribeLocationObjectStorageInput{
		LocationArn: aws.String(arn),
	}

	output, err := conn.DescribeLocationObjectStorage(ctx, input)

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
