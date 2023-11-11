// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_eks_cluster")
func dataSourceCluster() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceClusterRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_authority": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled_cluster_log_types": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"oidc": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"issuer": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"kubernetes_network_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_family": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_ipv4_cidr": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_ipv6_cidr": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validClusterName,
			},
			"outpost_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"control_plane_instance_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"control_plane_placement": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"group_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"outpost_arns": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"platform_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster_security_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"endpoint_private_access": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"endpoint_public_access": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"public_access_cidrs": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	cluster, err := findClusterByName(ctx, conn, name)

	if err != nil {
		return diag.Errorf("reading EKS Cluster (%s): %s", name, err)
	}

	d.SetId(name)
	d.Set("arn", cluster.Arn)
	if err := d.Set("certificate_authority", flattenCertificate(cluster.CertificateAuthority)); err != nil {
		return diag.Errorf("setting certificate_authority: %s", err)
	}
	// cluster_id is only relevant for clusters on Outposts.
	if cluster.OutpostConfig != nil {
		d.Set("cluster_id", cluster.Id)
	}
	d.Set("created_at", aws.ToTime(cluster.CreatedAt).String())
	if err := d.Set("enabled_cluster_log_types", flattenLogging(cluster.Logging)); err != nil {
		return diag.Errorf("setting enabled_cluster_log_types: %s", err)
	}
	d.Set("endpoint", cluster.Endpoint)
	if err := d.Set("identity", flattenIdentity(cluster.Identity)); err != nil {
		return diag.Errorf("setting identity: %s", err)
	}
	if err := d.Set("kubernetes_network_config", flattenKubernetesNetworkConfigResponse(cluster.KubernetesNetworkConfig)); err != nil {
		return diag.Errorf("setting kubernetes_network_config: %s", err)
	}
	d.Set("name", cluster.Name)
	if err := d.Set("outpost_config", flattenOutpostConfigResponse(cluster.OutpostConfig)); err != nil {
		return diag.Errorf("setting outpost_config: %s", err)
	}
	d.Set("platform_version", cluster.PlatformVersion)
	d.Set("role_arn", cluster.RoleArn)
	d.Set("status", cluster.Status)
	d.Set("version", cluster.Version)
	if err := d.Set("vpc_config", flattenVPCConfigResponse(cluster.ResourcesVpcConfig)); err != nil {
		return diag.Errorf("setting vpc_config: %s", err)
	}

	if err := d.Set("tags", KeyValueTags(ctx, cluster.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
