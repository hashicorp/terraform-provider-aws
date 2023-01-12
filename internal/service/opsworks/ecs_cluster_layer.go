package opsworks

import (
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceECSClusterLayer() *schema.Resource {
	layerType := &opsworksLayerType{
		TypeName:         opsworks.LayerTypeEcsCluster,
		DefaultLayerName: "Ecs Cluster",

		Attributes: map[string]*opsworksLayerTypeAttribute{
			"ecs_cluster_arn": {
				AttrName:     opsworks.LayerAttributesKeysEcsClusterArn,
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}

	return layerType.resourceSchema()
}
