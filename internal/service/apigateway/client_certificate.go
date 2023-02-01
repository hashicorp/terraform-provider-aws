package apigateway

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceClientCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClientCertificateCreate,
		ReadWithoutTimeout:   resourceClientCertificateRead,
		UpdateWithoutTimeout: resourceClientCertificateUpdate,
		DeleteWithoutTimeout: resourceClientCertificateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expiration_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pem_encoded_certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClientCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := apigateway.GenerateClientCertificateInput{}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}
	log.Printf("[DEBUG] Generating API Gateway Client Certificate: %s", input)
	out, err := conn.GenerateClientCertificateWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Failed to generate client certificate: %s", err)
	}

	d.SetId(aws.StringValue(out.ClientCertificateId))

	return append(diags, resourceClientCertificateRead(ctx, d, meta)...)
}

func resourceClientCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := apigateway.GetClientCertificateInput{
		ClientCertificateId: aws.String(d.Id()),
	}
	out, err := conn.GetClientCertificateWithContext(ctx, &input)
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
			log.Printf("[WARN] API Gateway Client Certificate (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Client Certificate (%s): %s", d.Id(), err)
	}

	tags := KeyValueTags(out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/clientcertificates/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	d.Set("description", out.Description)
	d.Set("created_date", out.CreatedDate.String())
	d.Set("expiration_date", out.ExpirationDate.String())
	d.Set("pem_encoded_certificate", out.PemEncodedCertificate)

	return diags
}

func resourceClientCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	operations := make([]*apigateway.PatchOperation, 0)
	if d.HasChange("description") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/description"),
			Value: aws.String(d.Get("description").(string)),
		})
	}

	input := apigateway.UpdateClientCertificateInput{
		ClientCertificateId: aws.String(d.Id()),
		PatchOperations:     operations,
	}

	log.Printf("[DEBUG] Updating API Gateway Client Certificate: %s", input)
	_, err := conn.UpdateClientCertificateWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Updating API Gateway Client Certificate failed: %s", err)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	return append(diags, resourceClientCertificateRead(ctx, d, meta)...)
}

func resourceClientCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()
	log.Printf("[DEBUG] Deleting API Gateway Client Certificate: %s", d.Id())
	input := apigateway.DeleteClientCertificateInput{
		ClientCertificateId: aws.String(d.Id()),
	}
	_, err := conn.DeleteClientCertificateWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Deleting API Gateway Client Certificate failed: %s", err)
	}

	return diags
}
