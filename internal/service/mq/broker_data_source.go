// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mq

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mq"
	"github.com/aws/aws-sdk-go-v2/service/mq/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_mq_broker", name="Broker")
func dataSourceBroker() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBrokerRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_strategy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAutoMinorVersionUpgrade: {
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
			names.AttrConfiguration: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
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
						names.AttrKMSKeyID: {
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
			names.AttrEngineVersion: {
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
						names.AttrEndpoints: {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrIPAddress: {
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
			names.AttrPubliclyAccessible: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrSecurityGroups: {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			names.AttrStorageType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
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
						names.AttrUsername: {
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
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MQClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &mq.ListBrokersInput{}
	broker, err := findBroker(ctx, conn, input, func(b *types.BrokerSummary) bool {
		if v, ok := d.GetOk("broker_id"); ok && v.(string) != aws.ToString(b.BrokerId) {
			return false
		}

		if v, ok := d.GetOk("broker_name"); ok && v.(string) != aws.ToString(b.BrokerName) {
			return false
		}

		return true
	})

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("MQ Broker", err))
	}

	brokerID := aws.ToString(broker.BrokerId)
	output, err := findBrokerByID(ctx, conn, brokerID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MQ Broker (%s): %s", brokerID, err)
	}

	d.SetId(brokerID)
	d.Set(names.AttrARN, output.BrokerArn)
	d.Set("authentication_strategy", output.AuthenticationStrategy)
	d.Set(names.AttrAutoMinorVersionUpgrade, output.AutoMinorVersionUpgrade)
	d.Set("broker_id", brokerID)
	d.Set("broker_name", output.BrokerName)
	d.Set("deployment_mode", output.DeploymentMode)
	d.Set("engine_type", output.EngineType)
	d.Set(names.AttrEngineVersion, output.EngineVersion)
	d.Set("host_instance_type", output.HostInstanceType)
	d.Set("instances", flattenBrokerInstances(output.BrokerInstances))
	d.Set(names.AttrPubliclyAccessible, output.PubliclyAccessible)
	d.Set(names.AttrSecurityGroups, output.SecurityGroups)
	d.Set(names.AttrStorageType, output.StorageType)
	d.Set(names.AttrSubnetIDs, output.SubnetIds)

	if err := d.Set(names.AttrConfiguration, flattenConfiguration(output.Configurations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting configuration: %s", err)
	}

	if err := d.Set("encryption_options", flattenEncryptionOptions(output.EncryptionOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting encryption_options: %s", err)
	}

	var password string
	if v, ok := d.GetOk("ldap_server_metadata.0.service_account_password"); ok {
		password = v.(string)
	}

	if err := d.Set("ldap_server_metadata", flattenLDAPServerMetadata(output.LdapServerMetadata, password)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ldap_server_metadata: %s", err)
	}

	if err := d.Set("logs", flattenLogs(output.Logs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting logs: %s", err)
	}

	if err := d.Set("maintenance_window_start_time", flattenWeeklyStartTime(output.MaintenanceWindowStartTime)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting maintenance_window_start_time: %s", err)
	}

	rawUsers, err := expandUsersForBroker(ctx, conn, brokerID, output.Users)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MQ Broker (%s) users: %s", brokerID, err)
	}

	if err := d.Set("user", flattenUsers(rawUsers, d.Get("user").(*schema.Set).List())); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting user: %s", err)
	}

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}

func findBroker(ctx context.Context, conn *mq.Client, input *mq.ListBrokersInput, filter tfslices.Predicate[*types.BrokerSummary]) (*types.BrokerSummary, error) {
	output, err := findBrokers(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findBrokers(ctx context.Context, conn *mq.Client, input *mq.ListBrokersInput, filter tfslices.Predicate[*types.BrokerSummary]) ([]types.BrokerSummary, error) {
	var output []types.BrokerSummary

	pages := mq.NewListBrokersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.BrokerSummaries {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
