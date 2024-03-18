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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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

// @SDKResource("aws_datasync_location_azure_blob", name="Location Microsoft Azure Blob Storage")
// @Tags(identifierAttribute="id")
func resourceLocationAzureBlob() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLocationAzureBlobCreate,
		ReadWithoutTimeout:   resourceLocationAzureBlobRead,
		UpdateWithoutTimeout: resourceLocationAzureBlobUpdate,
		DeleteWithoutTimeout: resourceLocationAzureBlobDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"access_tier": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      datasync.AzureAccessTierHot,
				ValidateFunc: validation.StringInSlice(datasync.AzureAccessTier_Values(), false),
			},
			"agent_arns": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(datasync.AzureBlobAuthenticationType_Values(), false),
			},
			"blob_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      datasync.AzureBlobTypeBlock,
				ValidateFunc: validation.StringInSlice(datasync.AzureBlobType_Values(), false),
			},
			"container_url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"sas_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"token": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"subdirectory": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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

func resourceLocationAzureBlobCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	input := &datasync.CreateLocationAzureBlobInput{
		AgentArns:          flex.ExpandStringSet(d.Get("agent_arns").(*schema.Set)),
		AuthenticationType: aws.String(d.Get("authentication_type").(string)),
		ContainerUrl:       aws.String(d.Get("container_url").(string)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("access_tier"); ok {
		input.AccessTier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("blob_type"); ok {
		input.BlobType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sas_configuration"); ok {
		input.SasConfiguration = expandAzureBlobSasConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("subdirectory"); ok {
		input.Subdirectory = aws.String(v.(string))
	}

	output, err := conn.CreateLocationAzureBlobWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataSync Location Microsoft Azure Blob Storage: %s", err)
	}

	d.SetId(aws.StringValue(output.LocationArn))

	return append(diags, resourceLocationAzureBlobRead(ctx, d, meta)...)
}

func resourceLocationAzureBlobRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	output, err := findLocationAzureBlobByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Location Microsoft Azure Blob Storage (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Location Microsoft Azure Blob Storage (%s): %s", d.Id(), err)
	}

	uri := aws.StringValue(output.LocationUri)
	accountHostName, err := globalIDFromLocationURI(aws.StringValue(output.LocationUri))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	subdirectory, err := subdirectoryFromLocationURI(uri)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	containerName := subdirectory[:strings.IndexAny(subdirectory[1:], "/")+1]
	containerUrl := fmt.Sprintf("https://%s%s", accountHostName, containerName)
	d.Set("container_url", containerUrl)

	d.Set("access_tier", output.AccessTier)
	d.Set("agent_arns", aws.StringValueSlice(output.AgentArns))
	d.Set("arn", output.LocationArn)
	d.Set("authentication_type", output.AuthenticationType)
	d.Set("blob_type", output.BlobType)
	d.Set("container_url", d.Get("container_url"))
	d.Set("sas_configuration", d.Get("sas_configuration"))
	d.Set("subdirectory", subdirectory[strings.IndexAny(subdirectory[1:], "/")+1:])
	d.Set("uri", uri)

	return diags
}

func resourceLocationAzureBlobUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &datasync.UpdateLocationAzureBlobInput{
			LocationArn: aws.String(d.Id()),
		}

		if d.HasChange("access_tier") {
			input.AccessTier = aws.String(d.Get("access_tier").(string))
		}

		if d.HasChange("agent_arns") {
			input.AgentArns = flex.ExpandStringSet(d.Get("agent_arns").(*schema.Set))
		}

		if d.HasChange("authentication_type") {
			input.AuthenticationType = aws.String(d.Get("authentication_type").(string))
		}

		if d.HasChange("blob_type") {
			input.BlobType = aws.String(d.Get("blob_type").(string))
		}

		if d.HasChange("sas_configuration") {
			input.SasConfiguration = expandAzureBlobSasConfiguration(d.Get("sas_configuration").([]interface{}))
		}

		if d.HasChange("subdirectory") {
			input.Subdirectory = aws.String(d.Get("subdirectory").(string))
		}

		_, err := conn.UpdateLocationAzureBlobWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DataSync Location Microsoft Azure Blob Storage (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceLocationAzureBlobRead(ctx, d, meta)...)
}

func resourceLocationAzureBlobDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	log.Printf("[DEBUG] Deleting DataSync LocationMicrosoft Azure Blob Storage: %s", d.Id())
	_, err := conn.DeleteLocationWithContext(ctx, &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DataSync Location Microsoft Azure Blob Storage (%s): %s", d.Id(), err)
	}

	return diags
}

func findLocationAzureBlobByARN(ctx context.Context, conn *datasync.DataSync, arn string) (*datasync.DescribeLocationAzureBlobOutput, error) {
	input := &datasync.DescribeLocationAzureBlobInput{
		LocationArn: aws.String(arn),
	}

	output, err := conn.DescribeLocationAzureBlobWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
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

func expandAzureBlobSasConfiguration(l []interface{}) *datasync.AzureBlobSasConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	apiObject := &datasync.AzureBlobSasConfiguration{
		Token: aws.String(m["token"].(string)),
	}

	return apiObject
}
