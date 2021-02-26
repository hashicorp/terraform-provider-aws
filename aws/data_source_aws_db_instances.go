package aws

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsDbInstances() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsDbInstancesRead,

		Schema: map[string]*schema.Schema{
			"db_instance_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAwsDbInstancesRead(d *schema.ResourceData, meta interface{}) error {
	//conn := meta.(*AWSClient).rdsconn
	//ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
	//
	//opts := &rds.DescribeDBInstancesInput{
	//	DBInstanceIdentifier: aws.String(d.Get("db_instance_identifier").(string)),
	//}
	//
	//log.Printf("[DEBUG] Reading DB Instance: %s", opts)
	//
	//resp, err := conn.DescribeDBInstances(opts)
	//if err != nil {
	//	return err
	//}
	//
	//if len(resp.DBInstances) < 1 {
	//	return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	//}
	//if len(resp.DBInstances) > 1 {
	//	return fmt.Errorf("Your query returned more than one result. Please try a more specific search criteria.")
	//}
	//
	//dbInstance := *resp.DBInstances[0]
	//
	//d.SetId(d.Get("db_instance_identifier").(string))
	//
	//d.Set("allocated_storage", dbInstance.AllocatedStorage)
	//d.Set("auto_minor_version_upgrade", dbInstance.AutoMinorVersionUpgrade)
	//d.Set("availability_zone", dbInstance.AvailabilityZone)
	//d.Set("backup_retention_period", dbInstance.BackupRetentionPeriod)
	//d.Set("db_cluster_identifier", dbInstance.DBClusterIdentifier)
	//d.Set("db_instance_arn", dbInstance.DBInstanceArn)
	//d.Set("db_instance_class", dbInstance.DBInstanceClass)
	//d.Set("db_name", dbInstance.DBName)
	//d.Set("resource_id", dbInstance.DbiResourceId)
	//
	//var parameterGroups []string
	//for _, v := range dbInstance.DBParameterGroups {
	//	parameterGroups = append(parameterGroups, *v.DBParameterGroupName)
	//}
	//if err := d.Set("db_parameter_groups", parameterGroups); err != nil {
	//	return fmt.Errorf("Error setting db_parameter_groups attribute: %#v, error: %w", parameterGroups, err)
	//}
	//
	//var dbSecurityGroups []string
	//for _, v := range dbInstance.DBSecurityGroups {
	//	dbSecurityGroups = append(dbSecurityGroups, *v.DBSecurityGroupName)
	//}
	//if err := d.Set("db_security_groups", dbSecurityGroups); err != nil {
	//	return fmt.Errorf("Error setting db_security_groups attribute: %#v, error: %w", dbSecurityGroups, err)
	//}
	//
	//if dbInstance.DBSubnetGroup != nil {
	//	d.Set("db_subnet_group", dbInstance.DBSubnetGroup.DBSubnetGroupName)
	//} else {
	//	d.Set("db_subnet_group", "")
	//}
	//
	//d.Set("db_instance_port", dbInstance.DbInstancePort)
	//d.Set("engine", dbInstance.Engine)
	//d.Set("engine_version", dbInstance.EngineVersion)
	//d.Set("iops", dbInstance.Iops)
	//d.Set("kms_key_id", dbInstance.KmsKeyId)
	//d.Set("license_model", dbInstance.LicenseModel)
	//d.Set("master_username", dbInstance.MasterUsername)
	//d.Set("monitoring_interval", dbInstance.MonitoringInterval)
	//d.Set("monitoring_role_arn", dbInstance.MonitoringRoleArn)
	//d.Set("multi_az", dbInstance.MultiAZ)
	//d.Set("address", dbInstance.Endpoint.Address)
	//d.Set("port", dbInstance.Endpoint.Port)
	//d.Set("hosted_zone_id", dbInstance.Endpoint.HostedZoneId)
	//d.Set("endpoint", fmt.Sprintf("%s:%d", *dbInstance.Endpoint.Address, *dbInstance.Endpoint.Port))
	//
	//if err := d.Set("enabled_cloudwatch_logs_exports", aws.StringValueSlice(dbInstance.EnabledCloudwatchLogsExports)); err != nil {
	//	return fmt.Errorf("error setting enabled_cloudwatch_logs_exports: %w", err)
	//}
	//
	//var optionGroups []string
	//for _, v := range dbInstance.OptionGroupMemberships {
	//	optionGroups = append(optionGroups, *v.OptionGroupName)
	//}
	//if err := d.Set("option_group_memberships", optionGroups); err != nil {
	//	return fmt.Errorf("Error setting option_group_memberships attribute: %#v, error: %w", optionGroups, err)
	//}
	//
	//d.Set("preferred_backup_window", dbInstance.PreferredBackupWindow)
	//d.Set("preferred_maintenance_window", dbInstance.PreferredMaintenanceWindow)
	//d.Set("publicly_accessible", dbInstance.PubliclyAccessible)
	//d.Set("storage_encrypted", dbInstance.StorageEncrypted)
	//d.Set("storage_type", dbInstance.StorageType)
	//d.Set("timezone", dbInstance.Timezone)
	//d.Set("replicate_source_db", dbInstance.ReadReplicaSourceDBInstanceIdentifier)
	//d.Set("ca_cert_identifier", dbInstance.CACertificateIdentifier)
	//
	//var vpcSecurityGroups []string
	//for _, v := range dbInstance.VpcSecurityGroups {
	//	vpcSecurityGroups = append(vpcSecurityGroups, *v.VpcSecurityGroupId)
	//}
	//if err := d.Set("vpc_security_groups", vpcSecurityGroups); err != nil {
	//	return fmt.Errorf("Error setting vpc_security_groups attribute: %#v, error: %w", vpcSecurityGroups, err)
	//}
	//
	//tags, err := keyvaluetags.RdsListTags(conn, d.Get("db_instance_arn").(string))
	//
	//if err != nil {
	//	return fmt.Errorf("error listing tags for RDS DB Instance (%s): %w", d.Get("db_instance_arn").(string), err)
	//}
	//
	//if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
	//	return fmt.Errorf("error setting tags: %w", err)
	//}

	return nil
}
