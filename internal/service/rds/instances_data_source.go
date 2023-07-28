// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/generate/namevaluesfilters"
	"github.com/hashicorp/terraform-provider-aws/names"
	"log"
)

// @SDKDataSource("aws_db_instances")
func DataSourceInstances() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstancesRead,

		Schema: map[string]*schema.Schema{
			"filter": namevaluesfilters.Schema(),
			"tag": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"value": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},
			// Computed values.
			"instance_arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"instance_identifiers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

const (
	DSNameInstances = "Instances Data Source"
)

func dataSourceInstancesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	input := &rds.DescribeDBInstancesInput{}
	var instanceARNS []string
	var instanceIdentifiers []string

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = namevaluesfilters.New(v.(*schema.Set)).RDSFilters()

		err := conn.DescribeDBInstancesPagesWithContext(ctx, input, func(page *rds.DescribeDBInstancesOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, dbInstance := range page.DBInstances {
				if dbInstance == nil {
					continue
				}

				instanceARNS = append(instanceARNS, aws.StringValue(dbInstance.DBInstanceArn))
				instanceIdentifiers = append(instanceIdentifiers, aws.StringValue(dbInstance.DBInstanceIdentifier))
			}

			return !lastPage
		})
		if err != nil {
			return create.DiagError(names.RDS, create.ErrActionReading, DSNameInstances, "", err)
		}
	}

	if v, ok := d.GetOk("tag"); ok {
		// Build map of tags to check, based on user request.
		tags := v.(*schema.Set).List()
		tagsToCheck := make(map[string]string)
		for _, tag := range tags {
			tagMap := tag.(map[string]interface{})
			key := tagMap["key"].(string)
			value := tagMap["value"].(string)
			tagsToCheck[key] = value
		}

		err := conn.DescribeDBInstancesPagesWithContext(ctx, input, func(page *rds.DescribeDBInstancesOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, dbInstance := range page.DBInstances {
				log.Printf("[DEBUG] DBInstanceIdentifier: %v", aws.StringValue(dbInstance.DBInstanceIdentifier))
				if tagMatchKeyAndValue(tagsToCheck, dbInstance.TagList) {
					instanceIdentifiers = append(instanceIdentifiers, aws.StringValue(dbInstance.DBInstanceIdentifier))
					instanceARNS = append(instanceARNS, aws.StringValue(dbInstance.DBInstanceArn))
				}
			}
			return !lastPage
		})

		if err != nil {
			return create.DiagError(names.RDS, create.ErrActionReading, DSNameInstances, "", err)
		}

		log.Printf("[DEBUG] instanceARNS: %+v\n", instanceARNS)
		log.Printf("[DEBUG] instanceIdentifiers: %+v\n", instanceIdentifiers)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("instance_arns", instanceARNS)
	d.Set("instance_identifiers", instanceIdentifiers)

	return nil
}

func tagMatchKeyAndValue(tagsToCheck map[string]string, rdsInstanceTags []*rds.Tag) bool {
	log.Printf("[DEBUG] Check if the following tags are present in instance tags: %v", tagsToCheck)
	for key, desiredValue := range tagsToCheck {
		isAMatch := false
		for _, tag := range rdsInstanceTags {
			if aws.StringValue(tag.Key) == key {
				if aws.StringValue(tag.Value) == desiredValue {
					isAMatch = true
					log.Printf("[DEBUG] Matching key (%v) and value (%v)", aws.StringValue(tag.Key), aws.StringValue(tag.Value))
					break
				} else {
					log.Printf("[DEBUG] Matching key (%v) but not value (%v)", aws.StringValue(tag.Key), aws.StringValue(tag.Value))
				}
			}
		}
		if !isAMatch {
			return false
		}
	}
	return true
}
