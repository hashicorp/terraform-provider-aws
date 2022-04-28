package emrcontainers

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceVirtualCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVirtualClusterRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"container_provider": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						// According to https://docs.aws.amazon.com/emr-on-eks/latest/APIReference/API_ContainerProvider.html
						// The info and the eks_info are optional but the API raises ValidationException without the fields
						"info": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"eks_info": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"namespace": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceVirtualClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EMRContainersConn

	var id string
	if cid, ok := d.GetOk("id"); ok {
		id = cid.(string)
	}

	vc, err := FindVirtualClusterByID(ctx, conn, id)

	if err != nil {
		return diag.Errorf("reading EMR Containers Virtual Cluster (%s): %s", id, err)
	}

	d.SetId(aws.StringValue(vc.Id))
	d.Set("arn", vc.Arn)
	if err := d.Set("container_provider", flattenEMRContainersContainerProvider(vc.ContainerProvider)); err != nil {
		return diag.Errorf("setting container_provider: %s", err)
	}
	d.Set("created_at", aws.TimeValue(vc.CreatedAt).String())
	d.Set("name", vc.Name)
	d.Set("state", vc.State)

	return nil
}
