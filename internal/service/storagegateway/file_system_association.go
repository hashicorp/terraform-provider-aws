// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_storagegateway_file_system_association", name="File System Association")
// @Tags(identifierAttribute="arn")
func resourceFileSystemAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFileSystemAssociationCreate,
		ReadWithoutTimeout:   resourceFileSystemAssociationRead,
		UpdateWithoutTimeout: resourceFileSystemAssociationUpdate,
		DeleteWithoutTimeout: resourceFileSystemAssociationDelete,

		CustomizeDiff: customdiff.Sequence(verify.SetTagsDiff),

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"audit_destination_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
				Default:      "",
			},
			"cache_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cache_stale_timeout_in_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
							ValidateFunc: validation.Any(
								validation.IntInSlice([]int{0}),
								validation.IntBetween(300, 2592000),
							),
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			"gateway_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"location_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrPassword: {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^[ -~]+$`), ""),
					validation.StringLenBetween(1, 1024),
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrUsername: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^\w[\w\.\- ]*$`), ""),
					validation.StringLenBetween(1, 1024),
				),
			},
		},
	}
}

func resourceFileSystemAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

	gatewayARN := d.Get("gateway_arn").(string)
	input := &storagegateway.AssociateFileSystemInput{
		ClientToken: aws.String(id.UniqueId()),
		GatewayARN:  aws.String(gatewayARN),
		LocationARN: aws.String(d.Get("location_arn").(string)),
		Password:    aws.String(d.Get(names.AttrPassword).(string)),
		Tags:        getTagsIn(ctx),
		UserName:    aws.String(d.Get(names.AttrUsername).(string)),
	}

	if v, ok := d.GetOk("audit_destination_arn"); ok {
		input.AuditDestinationARN = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cache_attributes"); ok {
		input.CacheAttributes = expandFileSystemAssociationCacheAttributes(v.([]interface{}))
	}

	output, err := conn.AssociateFileSystemWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Storage Gateway (%s) File System Association: %s", gatewayARN, err)
	}

	d.SetId(aws.StringValue(output.FileSystemAssociationARN))

	if _, err = waitFileSystemAssociationAvailable(ctx, conn, d.Id(), fileSystemAssociationCreateTimeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway File System Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceFileSystemAssociationRead(ctx, d, meta)...)
}

func resourceFileSystemAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

	filesystem, err := FindFileSystemAssociationByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Storage Gateway File System Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway File System Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, filesystem.FileSystemAssociationARN)
	d.Set("audit_destination_arn", filesystem.AuditDestinationARN)
	d.Set("gateway_arn", filesystem.GatewayARN)
	d.Set("location_arn", filesystem.LocationARN)

	if err := d.Set("cache_attributes", flattenFileSystemAssociationCacheAttributes(filesystem.CacheAttributes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cache_attributes: %s", err)
	}

	setTagsOut(ctx, filesystem.Tags)

	return diags
}

func resourceFileSystemAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

	if d.HasChangesExcept(names.AttrTagsAll) {
		input := &storagegateway.UpdateFileSystemAssociationInput{
			AuditDestinationARN:      aws.String(d.Get("audit_destination_arn").(string)),
			Password:                 aws.String(d.Get(names.AttrPassword).(string)),
			UserName:                 aws.String(d.Get(names.AttrUsername).(string)),
			FileSystemAssociationARN: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("cache_attributes"); ok {
			input.CacheAttributes = expandFileSystemAssociationCacheAttributes(v.([]interface{}))
		}

		_, err := conn.UpdateFileSystemAssociationWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Storage Gateway File System Association (%s): %s", d.Id(), err)
		}

		if _, err = waitFileSystemAssociationAvailable(ctx, conn, d.Id(), fileSystemAssociationUpdateTimeout); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway File System Association (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFileSystemAssociationRead(ctx, d, meta)...)
}

func resourceFileSystemAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

	input := &storagegateway.DisassociateFileSystemInput{
		FileSystemAssociationARN: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Storage Gateway File System Association: %s", input)
	_, err := conn.DisassociateFileSystemWithContext(ctx, input)

	if operationErrorCode(err) == operationErrCodeFileSystemAssociationNotFound {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Storage Gateway File System Association (%s): %s", d.Id(), err)
	}

	if _, err = waitFileSystemAssociationDeleted(ctx, conn, d.Id(), fileSystemAssociationDeleteTimeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway File System Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandFileSystemAssociationCacheAttributes(l []interface{}) *storagegateway.CacheAttributes {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	ca := &storagegateway.CacheAttributes{
		CacheStaleTimeoutInSeconds: aws.Int64(int64(m["cache_stale_timeout_in_seconds"].(int))),
	}

	return ca
}

func flattenFileSystemAssociationCacheAttributes(ca *storagegateway.CacheAttributes) []interface{} {
	if ca == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cache_stale_timeout_in_seconds": aws.Int64Value(ca.CacheStaleTimeoutInSeconds),
	}

	return []interface{}{m}
}
