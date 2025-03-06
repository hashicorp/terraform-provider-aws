val orgacctServices = mapOf(
    "accessanalyzer" to ServiceSpec("IAM Access Analyzer"),
    "backup" to ServiceSpec("Backup", "TestAccBackupGlobalSettings_basic"),
    "cloudformation" to ServiceSpec("CloudFormation"),
    "cloudtrail" to ServiceSpec("CloudTrail", parallelismOverride = 5),
    "config" to ServiceSpec("Config" /*"TestAccConfig_serial|TestAccConfigConfigurationAggregator_"*/),
    "detective" to ServiceSpec("Detective"),
    "fms" to ServiceSpec("FMS (Firewall Manager)", regionOverride = "us-east-1"),
    "guardduty" to ServiceSpec("GuardDuty"),
    "inspector" to ServiceSpec("Inspector Classic"),
    "inspector2" to ServiceSpec("Inspector"),
    "licensemanager" to ServiceSpec("License Manager"),
    "macie2" to ServiceSpec("Macie"),
    "organizations" to ServiceSpec("Organizations"),
    "securityhub" to ServiceSpec(
        "Security Hub",
        "TestAccSecurityHub_serial/Account|TestAccSecurityHub_serial/OrganizationAdminAccount|TestAccSecurityHub_serial/OrganizationConfiguration"
    ),
)
