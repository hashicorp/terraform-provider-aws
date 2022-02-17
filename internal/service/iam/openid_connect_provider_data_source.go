package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceOpenIDConnectProvider() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceOpenIDConnectProviderRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"arn", "url"},
				ValidateFunc: verify.ValidARN,
			},
			"url": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"arn", "url"},
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_id_list": {
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Type:     schema.TypeList,
				Computed: true,
			},
			"thumbprint_list": {
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Type:     schema.TypeList,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceOpenIDConnectProviderRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &iam.GetOpenIDConnectProviderInput{}

	if v, ok := d.GetOk("arn"); ok {
		input.OpenIDConnectProviderArn = aws.String(v.(string))
	} else if v, ok := d.GetOk("url"); ok {
		url := v.(string)

		oidcpEntry, err := dataSourceGetOpenIDConnectProviderByURL(context.Background(), conn, url)
		if err != nil {
			return fmt.Errorf("error finding IAM OIDC Provider by url (%s): %w", url, err)
		}

		if oidcpEntry == nil {
			return fmt.Errorf("error finding IAM OIDC Provider by url (%s): not found", url)
		}
		input.OpenIDConnectProviderArn = oidcpEntry.Arn
	}

	resp, err := conn.GetOpenIDConnectProvider(input)

	if err != nil {
		return fmt.Errorf("error reading IAM OIDC Provider: %w", err)
	}

	d.Set("arn", input.OpenIDConnectProviderArn)
	d.Set("url", resp.Url)
	d.Set("created_at", aws.TimeValue(resp.CreateDate).String())
	d.Set("client_id_list", flex.FlattenStringList(resp.ClientIDList))
	d.Set("thumbprint_list", flex.FlattenStringList(resp.ThumbprintList))

	tags := KeyValueTags(resp.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func dataSourceGetOpenIDConnectProviderByURL(ctx context.Context, conn *iam.IAM, url string) (*iam.OpenIDConnectProviderListEntry, error) {
	var result *iam.OpenIDConnectProviderListEntry

	input := &iam.ListOpenIDConnectProvidersInput{}

	output, err := conn.ListOpenIDConnectProvidersWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	for _, oidcp := range output.OpenIDConnectProviderList {
		if oidcp == nil {
			continue
		}
		_, arnUrl, err := urlFromOpenIDConnectProviderArn(aws.StringValue(oidcp.Arn))
		if err != nil {
			return nil, err
		}

		oidcpUrl := fmt.Sprintf("https://%s", arnUrl)
		if oidcpUrl == url {
			return oidcp, nil
		}
	}

	return result, nil
}
