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
)

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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
	conn := meta.(*conns.AWSClient).DataSyncConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &datasync.CreateLocationObjectStorageInput{
		AgentArns:      flex.ExpandStringSet(d.Get("agent_arns").(*schema.Set)),
		Subdirectory:   aws.String(d.Get("subdirectory").(string)),
		BucketName:     aws.String(d.Get("bucket_name").(string)),
		ServerHostname: aws.String(d.Get("server_hostname").(string)),
		Tags:           Tags(tags.IgnoreAWS()),
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

	if v, ok := d.GetOk("server_certficate"); ok {
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
	conn := meta.(*conns.AWSClient).DataSyncConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindLocationObjectStorageByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Location Object Storage (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Location Object Storage (%s): %s", d.Id(), err)
	}

	subdirectory, err := SubdirectoryFromLocationURI(aws.StringValue(output.LocationUri))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing DataSync Location Object Storage (%s) location URI: %s", d.Id(), err)
	}

	d.Set("agent_arns", flex.FlattenStringSet(output.AgentArns))
	d.Set("arn", output.LocationArn)
	d.Set("server_protocol", output.ServerProtocol)
	d.Set("subdirectory", subdirectory)
	d.Set("access_key", output.AccessKey)
	d.Set("server_port", output.ServerPort)
	d.Set("server_certificate", string(output.ServerCertificate))

	uri := aws.StringValue(output.LocationUri)

	hostname, bucketName, err := decodeObjectStorageURI(uri)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing DataSync Location Object Storage (%s) object-storage URI: %s", d.Id(), err)
	}

	d.Set("server_hostname", hostname)
	d.Set("bucket_name", bucketName)

	d.Set("uri", uri)

	tags, err := ListTags(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for DataSync Location Object Storage (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceLocationObjectStorageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn()

	if d.HasChangesExcept("tags_all", "tags") {
		input := &datasync.UpdateLocationObjectStorageInput{
			LocationArn: aws.String(d.Id()),
		}

		if d.HasChange("server_protocol") {
			input.ServerProtocol = aws.String(d.Get("server_protocol").(string))
		}

		if d.HasChange("server_port") {
			input.ServerPort = aws.Int64(int64(d.Get("server_port").(int)))
		}

		if d.HasChange("access_key") {
			input.AccessKey = aws.String(d.Get("access_key").(string))
		}

		if d.HasChange("secret_key") {
			input.SecretKey = aws.String(d.Get("secret_key").(string))
		}

		if d.HasChange("subdirectory") {
			input.Subdirectory = aws.String(d.Get("subdirectory").(string))
		}

		if d.HasChange("server_certficate") {
			input.ServerCertificate = []byte(d.Get("server_certficate").(string))
		}

		_, err := conn.UpdateLocationObjectStorageWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DataSync Location Object Storage (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DataSync Location Object Storage (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceLocationObjectStorageRead(ctx, d, meta)...)
}

func resourceLocationObjectStorageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn()

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

func decodeObjectStorageURI(uri string) (string, string, error) {
	prefix := "object-storage://"
	if !strings.HasPrefix(uri, prefix) {
		return "", "", fmt.Errorf("incorrect uri format needs to start with %s", prefix)
	}
	trimmedUri := strings.TrimPrefix(uri, prefix)
	uriParts := strings.Split(trimmedUri, "/")

	if len(uri) < 2 {
		return "", "", fmt.Errorf("incorrect uri format needs to start with %sSERVER-NAME/BUCKET-NAME/SUBDIRECTORY", prefix)
	}

	return uriParts[0], uriParts[1], nil
}
