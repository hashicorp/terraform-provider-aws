package appmesh

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceVirtualService() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVirtualServiceRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"mesh_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"mesh_owner": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"resource_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"spec": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provider": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"virtual_node": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"virtual_node_name": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},

									"virtual_router": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"virtual_router_name": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},

			"tags": tftags.TagsSchema(),
		},
	}
}

func dataSourceVirtualServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &appmesh.DescribeVirtualServiceInput{
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		VirtualServiceName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("mesh_owner"); ok {
		req.MeshOwner = aws.String(v.(string))
	}

	resp, err := conn.DescribeVirtualServiceWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Virtual Service: %s", err)
	}

	arn := aws.StringValue(resp.VirtualService.Metadata.Arn)

	d.SetId(aws.StringValue(resp.VirtualService.VirtualServiceName))

	d.Set("name", resp.VirtualService.VirtualServiceName)
	d.Set("mesh_name", resp.VirtualService.MeshName)
	d.Set("mesh_owner", resp.VirtualService.Metadata.MeshOwner)
	d.Set("arn", arn)
	d.Set("created_date", resp.VirtualService.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", resp.VirtualService.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("resource_owner", resp.VirtualService.Metadata.ResourceOwner)

	err = d.Set("spec", flattenVirtualServiceSpec(resp.VirtualService.Spec))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for App Mesh Virtual Service (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
