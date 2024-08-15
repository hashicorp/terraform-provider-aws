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
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.AzureAccessTierHot,
				ValidateDiagFunc: enum.Validate[awstypes.AzureAccessTier](),
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
			"authentication_type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AzureBlobAuthenticationType](),
			},
			"blob_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.AzureBlobTypeBlock,
				ValidateDiagFunc: enum.Validate[awstypes.AzureBlobType](),
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
			names.AttrURI: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLocationAzureBlobCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	input := &datasync.CreateLocationAzureBlobInput{
		AgentArns:          flex.ExpandStringValueSet(d.Get("agent_arns").(*schema.Set)),
		AuthenticationType: awstypes.AzureBlobAuthenticationType(d.Get("authentication_type").(string)),
		ContainerUrl:       aws.String(d.Get("container_url").(string)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("access_tier"); ok {
		input.AccessTier = awstypes.AzureAccessTier(v.(string))
	}

	if v, ok := d.GetOk("blob_type"); ok {
		input.BlobType = awstypes.AzureBlobType(v.(string))
	}

	if v, ok := d.GetOk("sas_configuration"); ok {
		input.SasConfiguration = expandAzureBlobSasConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("subdirectory"); ok {
		input.Subdirectory = aws.String(v.(string))
	}

	output, err := conn.CreateLocationAzureBlob(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataSync Location Microsoft Azure Blob Storage: %s", err)
	}

	d.SetId(aws.ToString(output.LocationArn))

	return append(diags, resourceLocationAzureBlobRead(ctx, d, meta)...)
}

func resourceLocationAzureBlobRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	output, err := findLocationAzureBlobByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Location Microsoft Azure Blob Storage (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Location Microsoft Azure Blob Storage (%s): %s", d.Id(), err)
	}

	uri := aws.ToString(output.LocationUri)
	accountHostName, err := globalIDFromLocationURI(aws.ToString(output.LocationUri))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	subdirectory, err := subdirectoryFromLocationURI(uri)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	containerName := subdirectory[:strings.IndexAny(subdirectory[1:], "/")+1]
	containerURL := fmt.Sprintf("https://%s%s", accountHostName, containerName)

	d.Set("access_tier", output.AccessTier)
	d.Set("agent_arns", output.AgentArns)
	d.Set(names.AttrARN, output.LocationArn)
	d.Set("authentication_type", output.AuthenticationType)
	d.Set("blob_type", output.BlobType)
	d.Set("container_url", containerURL)
	d.Set("sas_configuration", d.Get("sas_configuration"))
	d.Set("subdirectory", subdirectory[strings.IndexAny(subdirectory[1:], "/")+1:])
	d.Set(names.AttrURI, uri)

	return diags
}

func resourceLocationAzureBlobUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &datasync.UpdateLocationAzureBlobInput{
			LocationArn: aws.String(d.Id()),
		}

		if d.HasChange("access_tier") {
			input.AccessTier = awstypes.AzureAccessTier(d.Get("access_tier").(string))
		}

		if d.HasChange("agent_arns") {
			input.AgentArns = flex.ExpandStringValueSet(d.Get("agent_arns").(*schema.Set))
		}

		if d.HasChange("authentication_type") {
			input.AuthenticationType = awstypes.AzureBlobAuthenticationType(d.Get("authentication_type").(string))
		}

		if d.HasChange("blob_type") {
			input.BlobType = awstypes.AzureBlobType(d.Get("blob_type").(string))
		}

		if d.HasChange("sas_configuration") {
			input.SasConfiguration = expandAzureBlobSasConfiguration(d.Get("sas_configuration").([]interface{}))
		}

		if d.HasChange("subdirectory") {
			input.Subdirectory = aws.String(d.Get("subdirectory").(string))
		}

		_, err := conn.UpdateLocationAzureBlob(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DataSync Location Microsoft Azure Blob Storage (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceLocationAzureBlobRead(ctx, d, meta)...)
}

func resourceLocationAzureBlobDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	log.Printf("[DEBUG] Deleting DataSync LocationMicrosoft Azure Blob Storage: %s", d.Id())
	_, err := conn.DeleteLocation(ctx, &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	})

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DataSync Location Microsoft Azure Blob Storage (%s): %s", d.Id(), err)
	}

	return diags
}

func findLocationAzureBlobByARN(ctx context.Context, conn *datasync.Client, arn string) (*datasync.DescribeLocationAzureBlobOutput, error) {
	input := &datasync.DescribeLocationAzureBlobInput{
		LocationArn: aws.String(arn),
	}

	output, err := conn.DescribeLocationAzureBlob(ctx, input)

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

func expandAzureBlobSasConfiguration(l []interface{}) *awstypes.AzureBlobSasConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	apiObject := &awstypes.AzureBlobSasConfiguration{
		Token: aws.String(m["token"].(string)),
	}

	return apiObject
}
