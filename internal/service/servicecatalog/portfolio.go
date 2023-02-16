package servicecatalog

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePortfolio() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePortfolioCreate,
		ReadWithoutTimeout:   resourcePortfolioRead,
		UpdateWithoutTimeout: resourcePortfolioUpdate,
		DeleteWithoutTimeout: resourcePortfolioDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(PortfolioCreateTimeout),
			Read:   schema.DefaultTimeout(PortfolioReadTimeout),
			Update: schema.DefaultTimeout(PortfolioUpdateTimeout),
			Delete: schema.DefaultTimeout(PortfolioDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 2000),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"provider_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 50),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}
func resourcePortfolioCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &servicecatalog.CreatePortfolioInput{
		AcceptLanguage:   aws.String(AcceptLanguageEnglish),
		DisplayName:      aws.String(name),
		IdempotencyToken: aws.String(resource.UniqueId()),
		Tags:             Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("provider_name"); ok {
		input.ProviderName = aws.String(v.(string))
	}

	output, err := conn.CreatePortfolioWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Portfolio (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.PortfolioDetail.Id))

	return append(diags, resourcePortfolioRead(ctx, d, meta)...)
}

func resourcePortfolioRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindPortfolioByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Catalog Portfolio (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Service Catalog Portfolio (%s): %s", d.Id(), err)
	}

	portfolioDetail := output.PortfolioDetail
	d.Set("arn", portfolioDetail.ARN)
	d.Set("created_time", portfolioDetail.CreatedTime.Format(time.RFC3339))
	d.Set("description", portfolioDetail.Description)
	d.Set("name", portfolioDetail.DisplayName)
	d.Set("provider_name", portfolioDetail.ProviderName)

	tags := KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourcePortfolioUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn()

	input := &servicecatalog.UpdatePortfolioInput{
		AcceptLanguage: aws.String(AcceptLanguageEnglish),
		Id:             aws.String(d.Id()),
	}

	if d.HasChange("accept_language") {
		input.AcceptLanguage = aws.String(d.Get("accept_language").(string))
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("name") {
		input.DisplayName = aws.String(d.Get("name").(string))
	}

	if d.HasChange("provider_name") {
		input.ProviderName = aws.String(d.Get("provider_name").(string))
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		input.AddTags = Tags(tftags.New(n).IgnoreAWS())
		input.RemoveTags = aws.StringSlice(tftags.New(o).IgnoreAWS().Keys())
	}

	_, err := conn.UpdatePortfolioWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Service Catalog Portfolio (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePortfolioRead(ctx, d, meta)...)
}

func resourcePortfolioDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn()

	log.Printf("[DEBUG] Deleting Service Catalog Portfolio: %s", d.Id())
	_, err := conn.DeletePortfolioWithContext(ctx, &servicecatalog.DeletePortfolioInput{
		Id: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Service Catalog Portfolio (%s): %s", d.Id(), err)
	}

	return diags
}

func FindPortfolioByID(ctx context.Context, conn *servicecatalog.ServiceCatalog, id string) (*servicecatalog.DescribePortfolioOutput, error) {
	input := &servicecatalog.DescribePortfolioInput{
		AcceptLanguage: aws.String(AcceptLanguageEnglish),
		Id:             aws.String(id),
	}

	output, err := conn.DescribePortfolioWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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
