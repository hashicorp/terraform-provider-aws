package storagegateway

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFileSystemAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFileSystemAssociationCreate,
		ReadWithoutTimeout:   resourceFileSystemAssociationRead,
		UpdateWithoutTimeout: resourceFileSystemAssociationUpdate,
		DeleteWithoutTimeout: resourceFileSystemAssociationDelete,
		CustomizeDiff:        customdiff.Sequence(verify.SetTagsDiff),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
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
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^[ -~]+$`), ""),
					validation.StringLenBetween(1, 1024),
				),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexp.MustCompile(`^\w[\w\.\- ]*$`), ""),
					validation.StringLenBetween(1, 1024),
				),
			},
		},
	}
}

func resourceFileSystemAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	gatewayARN := d.Get("gateway_arn").(string)
	input := &storagegateway.AssociateFileSystemInput{
		ClientToken: aws.String(resource.UniqueId()),
		GatewayARN:  aws.String(gatewayARN),
		LocationARN: aws.String(d.Get("location_arn").(string)),
		Password:    aws.String(d.Get("password").(string)),
		Tags:        Tags(tags.IgnoreAWS()),
		UserName:    aws.String(d.Get("username").(string)),
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
	conn := meta.(*conns.AWSClient).StorageGatewayConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	filesystem, err := FindFileSystemAssociationByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Storage Gateway File System Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway File System Association (%s): %s", d.Id(), err)
	}

	d.Set("arn", filesystem.FileSystemAssociationARN)
	d.Set("audit_destination_arn", filesystem.AuditDestinationARN)
	d.Set("gateway_arn", filesystem.GatewayARN)
	d.Set("location_arn", filesystem.LocationARN)

	if err := d.Set("cache_attributes", flattenFileSystemAssociationCacheAttributes(filesystem.CacheAttributes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cache_attributes: %s", err)
	}

	tags := KeyValueTags(filesystem.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceFileSystemAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	if d.HasChangesExcept("tags_all") {
		input := &storagegateway.UpdateFileSystemAssociationInput{
			AuditDestinationARN:      aws.String(d.Get("audit_destination_arn").(string)),
			Password:                 aws.String(d.Get("password").(string)),
			UserName:                 aws.String(d.Get("username").(string)),
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
	conn := meta.(*conns.AWSClient).StorageGatewayConn()

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
