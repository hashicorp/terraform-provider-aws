// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mq

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types/nullable"
)

// @SDKDataSource("aws_mq_broker")
func DataSourceBroker() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBrokerRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_strategy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_minor_version_upgrade": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"broker_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"broker_name"},
			},
			"broker_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"broker_id"},
			},
			"configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"revision": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"deployment_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_key_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"use_aws_owned_key": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"engine_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"host_instance_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"console_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"endpoints": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"ldap_server_metadata": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hosts": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"role_base": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role_search_matching": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role_search_subtree": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"service_account_password": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_account_username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_base": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_role_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_search_matching": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_search_subtree": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"logs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"audit": {
							Type:     nullable.TypeNullableBool,
							Computed: true,
						},
						"general": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"maintenance_window_start_time": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"day_of_week": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"time_of_day": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"time_zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"publicly_accessible": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"storage_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"user": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"console_access": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"groups": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
						"replication_user": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"username": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceBrokerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MQConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &mq.ListBrokersInput{}
	var brokers []*mq.BrokerSummary

	err := conn.ListBrokersPagesWithContext(ctx, input, func(page *mq.ListBrokersResponse, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, broker := range page.BrokerSummaries {
			if broker == nil {
				continue
			}

			if v, ok := d.GetOk("broker_id"); ok && v.(string) != aws.StringValue(broker.BrokerId) {
				continue
			}

			if v, ok := d.GetOk("broker_name"); ok && v.(string) != aws.StringValue(broker.BrokerName) {
				continue
			}

			brokers = append(brokers, broker)
		}

		return !lastPage
	})

	if err != nil {
		return diag.Errorf("reading MQ Brokers: %s", err)
	}

	if n := len(brokers); n == 0 {
		return diag.Errorf("no MQ Brokers matched")
	} else if n > 1 {
		return diag.Errorf("%d MQ Brokers matched; use additional constraints to reduce matches to a single Broker", n)
	}

	brokerID := aws.StringValue(brokers[0].BrokerId)
	output, err := FindBrokerByID(ctx, conn, brokerID)

	if err != nil {
		return diag.Errorf("reading MQ Broker (%s): %s", brokerID, err)
	}

	d.SetId(brokerID)
	d.Set("arn", output.BrokerArn)
	d.Set("authentication_strategy", output.AuthenticationStrategy)
	d.Set("auto_minor_version_upgrade", output.AutoMinorVersionUpgrade)
	d.Set("broker_id", brokerID)
	d.Set("broker_name", output.BrokerName)
	d.Set("deployment_mode", output.DeploymentMode)
	d.Set("engine_type", output.EngineType)
	d.Set("engine_version", output.EngineVersion)
	d.Set("host_instance_type", output.HostInstanceType)
	d.Set("instances", flattenBrokerInstances(output.BrokerInstances))
	d.Set("publicly_accessible", output.PubliclyAccessible)
	d.Set("security_groups", aws.StringValueSlice(output.SecurityGroups))
	d.Set("storage_type", output.StorageType)
	d.Set("subnet_ids", aws.StringValueSlice(output.SubnetIds))

	if err := d.Set("configuration", flattenConfiguration(output.Configurations)); err != nil {
		return diag.Errorf("setting configuration: %s", err)
	}

	if err := d.Set("encryption_options", flattenEncryptionOptions(output.EncryptionOptions)); err != nil {
		return diag.Errorf("setting encryption_options: %s", err)
	}

	var password string
	if v, ok := d.GetOk("ldap_server_metadata.0.service_account_password"); ok {
		password = v.(string)
	}

	if err := d.Set("ldap_server_metadata", flattenLDAPServerMetadata(output.LdapServerMetadata, password)); err != nil {
		return diag.Errorf("setting ldap_server_metadata: %s", err)
	}

	if err := d.Set("logs", flattenLogs(output.Logs)); err != nil {
		return diag.Errorf("setting logs: %s", err)
	}

	if err := d.Set("maintenance_window_start_time", flattenWeeklyStartTime(output.MaintenanceWindowStartTime)); err != nil {
		return diag.Errorf("setting maintenance_window_start_time: %s", err)
	}

	rawUsers, err := expandUsersForBroker(ctx, conn, brokerID, output.Users)

	if err != nil {
		return diag.Errorf("reading MQ Broker (%s) users: %s", brokerID, err)
	}

	if err := d.Set("user", flattenUsers(rawUsers, d.Get("user").(*schema.Set).List())); err != nil {
		return diag.Errorf("setting user: %s", err)
	}

	if err := d.Set("tags", KeyValueTags(ctx, output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
