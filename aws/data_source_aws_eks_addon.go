package aws

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEksAddon() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsEksAddonRead,
		Schema: map[string]*schema.Schema{
			"addon_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cluster_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"addon_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_account_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modified_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsEksAddonRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).eksconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	addonName := d.Get("addon_name").(string)
	clusterName := d.Get("cluster_name").(string)

	input := &eks.DescribeAddonInput{
		AddonName:   aws.String(addonName),
		ClusterName: aws.String(clusterName),
	}

	output, err := conn.DescribeAddonWithContext(ctx, input)
	if err != nil {
		return diag.Errorf("error reading EKS Addon (%s): %s", addonName, err)
	}

	addon := output.Addon
	if addon == nil {
		return diag.Errorf("EKS Addon (%s) not found", addonName)
	}

	d.SetId(addonName)
	d.Set("arn", addon.AddonArn)
	d.Set("addon_version", addon.AddonVersion)
	d.Set("service_account_role_arn", addon.ServiceAccountRoleArn)
	d.Set("status", addon.Status)
	d.Set("created_at", aws.TimeValue(addon.CreatedAt).String())
	d.Set("modified_at", aws.TimeValue(addon.ModifiedAt).String())

	if err := d.Set("tags", keyvaluetags.EksKeyValueTags(addon.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags attribute: %s", err)
	}

	return nil
}
