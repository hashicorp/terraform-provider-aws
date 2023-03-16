val orgacctServices = mapOf(
    "accessanalyzer" to ServiceSpec("IAM Access Analyzer"),
    "backup" to ServiceSpec("Backup", "TestAccBackupGlobalSettings_basic"),
    "cloudformation" to ServiceSpec(
        "CloudFormation",
        "TestAccCloudFormationStackSet_PermissionModel_serviceManaged|TestAccCloudFormationStackSetInstance_deploymentTargets"
    ),
    "cloudtrail" to ServiceSpec("CloudTrail"),
    "config" to ServiceSpec("Config" /*"TestAccConfig_serial|TestAccConfigConfigurationAggregator_"*/),
    "fms" to ServiceSpec("FMS (Firewall Manager)"),
    "guardduty" to ServiceSpec("GuardDuty"),
    "licensemanager" to ServiceSpec("License Manager"),
    "macie2" to ServiceSpec("Macie"),
    "organizations" to ServiceSpec("Organizations"),
    "securityhub" to ServiceSpec(
        "Security Hub",
        "TestAccSecurityHub_serial/Account|TestAccSecurityHub_serial/OrganizationAdminAccount|TestAccSecurityHub_serial/OrganizationConfiguration"
    ),
)
