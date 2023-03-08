val orgacctServices = mapOf(
    "accessanalyzer" to ServiceSpec("Access Analyzer"),
    "backup" to ServiceSpec("Backup", "TestAccBackupGlobalSettings_basic"),
    "cloudformation" to ServiceSpec(
        "CloudFormation",
        "TestAccCloudFormationStackSet_PermissionModel_serviceManaged|TestAccCloudFormationStackSetInstance_deploymentTargets"
    ),
    "cloudtrail" to ServiceSpec("CloudTrail"),
    "config" to ServiceSpec("Config" /*"TestAccConfig_serial|TestAccConfigConfigurationAggregator_"*/),
    "fms" to ServiceSpec("FMS"),
    "guardduty" to ServiceSpec("GuardDuty"),
    "macie2" to ServiceSpec("Macie2"),
    "organizations" to ServiceSpec("Organizations"),
    "securityhub" to ServiceSpec(
        "SecurityHub",
        "TestAccSecurityHub_serial/Account|TestAccSecurityHub_serial/OrganizationAdminAccount|TestAccSecurityHub_serial/OrganizationConfiguration"
    ),
)
