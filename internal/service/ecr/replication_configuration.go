package ecr

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceReplicationConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceReplicationConfigurationPut,
		Read:   resourceReplicationConfigurationRead,
		Update: resourceReplicationConfigurationPut,
		Delete: resourceReplicationConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 10,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"destination": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 25,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"region": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: verify.ValidRegionName,
												},
												"registry_id": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: verify.ValidAccountID,
												},
											},
										},
									},
									"repository_filter": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										MaxItems: 100,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"filter": {
													Type:     schema.TypeString,
													Required: true,
												},
												"filter_type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(ecr.RepositoryFilterType_Values(), false),
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
		},
	}
}

func resourceReplicationConfigurationPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRConn

	input := ecr.PutReplicationConfigurationInput{
		ReplicationConfiguration: expandReplicationConfigurationReplicationConfiguration(d.Get("replication_configuration").([]interface{})),
	}

	_, err := conn.PutReplicationConfiguration(&input)
	if err != nil {
		return fmt.Errorf("error creating ECR Replication Configuration: %w", err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)

	return resourceReplicationConfigurationRead(d, meta)
}

func resourceReplicationConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRConn

	log.Printf("[DEBUG] Reading ECR Replication Configuration %s", d.Id())
	out, err := conn.DescribeRegistry(&ecr.DescribeRegistryInput{})
	if err != nil {
		return fmt.Errorf("error reading ECR Replication Configuration: %w", err)
	}

	d.Set("registry_id", out.RegistryId)

	if err := d.Set("replication_configuration", flattenReplicationConfigurationReplicationConfiguration(out.ReplicationConfiguration)); err != nil {
		return fmt.Errorf("error setting replication_configuration for ECR Replication Configuration: %w", err)
	}

	return nil
}

func resourceReplicationConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRConn

	input := ecr.PutReplicationConfigurationInput{
		ReplicationConfiguration: &ecr.ReplicationConfiguration{
			Rules: []*ecr.ReplicationRule{},
		},
	}

	_, err := conn.PutReplicationConfiguration(&input)
	if err != nil {
		return fmt.Errorf("error deleting ECR Replication Configuration: %w", err)
	}

	return nil
}

func expandReplicationConfigurationReplicationConfiguration(data []interface{}) *ecr.ReplicationConfiguration {
	if len(data) == 0 || data[0] == nil {
		return nil
	}

	ec := data[0].(map[string]interface{})
	config := &ecr.ReplicationConfiguration{
		Rules: expandReplicationConfigurationReplicationConfigurationRules(ec["rule"].([]interface{})),
	}
	return config
}

func flattenReplicationConfigurationReplicationConfiguration(ec *ecr.ReplicationConfiguration) []map[string]interface{} {
	if ec == nil {
		return nil
	}

	config := map[string]interface{}{
		"rule": flattenReplicationConfigurationReplicationConfigurationRules(ec.Rules),
	}

	return []map[string]interface{}{
		config,
	}
}

func expandReplicationConfigurationReplicationConfigurationRules(data []interface{}) []*ecr.ReplicationRule {
	if len(data) == 0 || data[0] == nil {
		return nil
	}

	var rules []*ecr.ReplicationRule

	for _, rule := range data {
		ec := rule.(map[string]interface{})
		config := &ecr.ReplicationRule{
			Destinations:      expandReplicationConfigurationReplicationConfigurationRulesDestinations(ec["destination"].([]interface{})),
			RepositoryFilters: expandReplicationConfigurationReplicationConfigurationRulesRepositoryFilters(ec["repository_filter"].([]interface{})),
		}

		rules = append(rules, config)

	}
	return rules
}

func flattenReplicationConfigurationReplicationConfigurationRules(ec []*ecr.ReplicationRule) []interface{} {
	if len(ec) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range ec {
		tfMap := map[string]interface{}{
			"destination":       flattenReplicationConfigurationReplicationConfigurationRulesDestinations(apiObject.Destinations),
			"repository_filter": flattenReplicationConfigurationReplicationConfigurationRulesRepositoryFilters(apiObject.RepositoryFilters),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandReplicationConfigurationReplicationConfigurationRulesDestinations(data []interface{}) []*ecr.ReplicationDestination {
	if len(data) == 0 || data[0] == nil {
		return nil
	}

	var dests []*ecr.ReplicationDestination

	for _, dest := range data {
		ec := dest.(map[string]interface{})
		config := &ecr.ReplicationDestination{
			Region:     aws.String(ec["region"].(string)),
			RegistryId: aws.String(ec["registry_id"].(string)),
		}

		dests = append(dests, config)
	}
	return dests
}

func flattenReplicationConfigurationReplicationConfigurationRulesDestinations(ec []*ecr.ReplicationDestination) []interface{} {
	if len(ec) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range ec {
		tfMap := map[string]interface{}{
			"region":      aws.StringValue(apiObject.Region),
			"registry_id": aws.StringValue(apiObject.RegistryId),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandReplicationConfigurationReplicationConfigurationRulesRepositoryFilters(data []interface{}) []*ecr.RepositoryFilter {
	if len(data) == 0 || data[0] == nil {
		return nil
	}

	var filters []*ecr.RepositoryFilter

	for _, filter := range data {
		ec := filter.(map[string]interface{})
		config := &ecr.RepositoryFilter{
			Filter:     aws.String(ec["filter"].(string)),
			FilterType: aws.String(ec["filter_type"].(string)),
		}

		filters = append(filters, config)
	}
	return filters
}

func flattenReplicationConfigurationReplicationConfigurationRulesRepositoryFilters(ec []*ecr.RepositoryFilter) []interface{} {
	if len(ec) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range ec {
		tfMap := map[string]interface{}{
			"filter":      aws.StringValue(apiObject.Filter),
			"filter_type": aws.StringValue(apiObject.FilterType),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
