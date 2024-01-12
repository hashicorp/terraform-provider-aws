package servicecatalogappregistry

import (
	"context"
	"log"

	servicecatalogappregistry_sdkv2 "github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_servicecatalog_application", name="Application")
// @Tags(identifierAttribute="arn")
func ResourceApplication() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout:   resourceApplicationRead,
		CreateWithoutTimeout: resourceApplicationCreate,
		UpdateWithoutTimeout: resourceApplicationUpdate,
		DeleteWithoutTimeout: resourceApplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogAppRegistryClient(ctx)

	input := &servicecatalogappregistry_sdkv2.CreateApplicationInput{
		Name:        aws.String(d.Get("name").(string)),
		Description: aws.String(d.Get("description").(string)),
		Tags:        getTagsIn(ctx),
	}

	resp, err := conn.CreateApplication(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog App Registry Application: %s", err)
	}

	d.SetId(aws.StringValue(resp.Application.Id))

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogAppRegistryClient(ctx)

	input := &servicecatalogappregistry_sdkv2.GetApplicationInput{
		Application: aws.String(d.Id()),
	}

	resp, err := conn.GetApplication(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Service Catalog App Registry Application: %s", err)
	}

	d.Set("arn", aws.StringValue(resp.Arn))
	d.Set("name", aws.StringValue(resp.Name))
	d.Set("description", aws.StringValue(resp.Description))

	setTagsOut(ctx, resp.Tags)

	return diags
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogAppRegistryClient(ctx)

	input := &servicecatalogappregistry_sdkv2.UpdateApplicationInput{
		Application: aws.String(d.Id()),
		// Disable updating the name of the application since this field is deprecated
		//Name:        aws.String(d.Get("name").(string)),
		Description: aws.String(d.Get("description").(string)),
	}

	_, err := conn.UpdateApplication(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Service Catalog App Registry Application: %s", err)
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogAppRegistryClient(ctx)

	input := &servicecatalogappregistry_sdkv2.DeleteApplicationInput{
		Application: aws.String(d.Id()),
	}

	_, err := conn.DeleteApplication(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Service Catalog App Registry Application: %s", err)
	}

	return diags
}

func findApplicationByID(ctx context.Context, conn *servicecatalogappregistry_sdkv2.Client, name string) (*servicecatalogappregistry_sdkv2.GetApplicationOutput, error) {
	input := &servicecatalogappregistry_sdkv2.GetApplicationInput{
		Application: aws.String(name),
	}

	output, err := conn.GetApplication(ctx, input)

	log.Printf("[DEBUG] Service Catalog App Registry Application: %+v", output)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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
