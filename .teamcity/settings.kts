import jetbrains.buildServer.configs.kotlin.* // ktlint-disable no-wildcard-imports
import jetbrains.buildServer.configs.kotlin.buildFeatures.golang
import jetbrains.buildServer.configs.kotlin.buildFeatures.notifications
import jetbrains.buildServer.configs.kotlin.buildSteps.script
import jetbrains.buildServer.configs.kotlin.failureConditions.failOnText
import jetbrains.buildServer.configs.kotlin.failureConditions.BuildFailureOnText
import jetbrains.buildServer.configs.kotlin.triggers.schedule
import java.io.File
import java.time.Duration
import java.time.LocalTime
import java.time.ZoneId
import java.time.format.DateTimeFormatter

version = "2023.05"

val defaultRegion = DslContext.getParameter("default_region")
val alternateRegion = DslContext.getParameter("alternate_region", "")
val acmCertificateRootDomain = DslContext.getParameter("acm_certificate_root_domain", "")
val sweeperRegions = DslContext.getParameter("sweeper_regions")
val awsAccountID = DslContext.getParameter("aws_account.account_id")
val acctestParallelism = DslContext.getParameter("acctest_parallelism", "")
val tfAccAssumeRoleArn = DslContext.getParameter("tf_acc_assume_role_arn", "")
val awsAlternateAccountID = DslContext.getParameter("aws_alt_account.account_id", "")
val tfLog = DslContext.getParameter("tf_log", "")

// Legacy User credentials
val legacyAWSAccessKeyID = DslContext.getParameter("aws_account.legacy_access_key_id", "")
val legacyAWSSecretAccessKey = DslContext.getParameter("aws_account.legacy_secret_access_key", "")

// Legacy Alternate User credentials
val legacyAWSAlternateAccessKeyID = DslContext.getParameter("aws_alt_account.legacy_access_key_id", "")
val legacyAWSAlternateSecretAccessKey = DslContext.getParameter("aws_alt_account.legacy_secret_access_key", "")

// Assume Role credentials
val accTestRoleARN = DslContext.getParameter("aws_account.role_arn", "")
val awsAccessKeyID = if (accTestRoleARN != "") { DslContext.getParameter("aws_account.access_key_id") } else { "" }
val awsSecretAccessKey = if (accTestRoleARN != "") { DslContext.getParameter("aws_account.secret_access_key") } else { "" }

// Alternate Assume Role credentials
val alternateAccTestRoleARN = DslContext.getParameter("aws_alt_account.role_arn", "")
val alternateAWSAccessKeyID = if (alternateAccTestRoleARN != "") { DslContext.getParameter("aws_alt_account.access_key_id") } else { "" }
val alternateAWSSecretAccessKey = if (alternateAccTestRoleARN != "") { DslContext.getParameter("aws_alt_account.secret_access_key") } else { "" }

project {
    if (DslContext.getParameter("build_full", "true").toBoolean()) {
        buildType(FullBuild)
    }

    params {
        if (acctestParallelism != "") {
            text("ACCTEST_PARALLELISM", acctestParallelism, allowEmpty = false)
        }
        text("TEST_PATTERN", "TestAcc", display = ParameterDisplay.HIDDEN)
        text("SWEEPER_REGIONS", sweeperRegions, display = ParameterDisplay.HIDDEN, allowEmpty = false)
        text("env.AWS_ACCOUNT_ID", awsAccountID, display = ParameterDisplay.HIDDEN, allowEmpty = false)
        text("env.AWS_DEFAULT_REGION", defaultRegion, allowEmpty = false)
        text("env.TF_LOG", tfLog)

        if (alternateRegion != "") {
            text("env.AWS_ALTERNATE_REGION", alternateRegion)
        }

        if (acmCertificateRootDomain != "") {
            text("env.ACM_CERTIFICATE_ROOT_DOMAIN", acmCertificateRootDomain, display = ParameterDisplay.HIDDEN)
        }

        val securityGroupRulesPerGroup = DslContext.getParameter("security_group_rules_per_group", "")
        if (securityGroupRulesPerGroup != "") {
            text("env.EC2_SECURITY_GROUP_RULES_PER_GROUP_LIMIT", securityGroupRulesPerGroup)
        }

        val brancRef = DslContext.getParameter("branch_name", "")
        if (brancRef != "") {
            text("BRANCH_NAME", brancRef, display = ParameterDisplay.HIDDEN)
        }

        if (tfAccAssumeRoleArn != "") {
            text("env.TF_ACC_ASSUME_ROLE_ARN", tfAccAssumeRoleArn)
        }

        // Legacy User credentials
        if (legacyAWSAccessKeyID != "") {
            password("env.AWS_ACCESS_KEY_ID", legacyAWSAccessKeyID, display = ParameterDisplay.HIDDEN)
        }
        if (legacyAWSSecretAccessKey != "") {
            password("env.AWS_SECRET_ACCESS_KEY", legacyAWSSecretAccessKey, display = ParameterDisplay.HIDDEN)
        }

        // Legacy Alternate User credentials
        if (awsAlternateAccountID != "" || legacyAWSAlternateAccessKeyID != "" || legacyAWSAlternateSecretAccessKey != "") {
            text("env.AWS_ALTERNATE_ACCOUNT_ID", awsAlternateAccountID, display = ParameterDisplay.HIDDEN)
            password("env.AWS_ALTERNATE_ACCESS_KEY_ID", legacyAWSAlternateAccessKeyID, display = ParameterDisplay.HIDDEN)
            password("env.AWS_ALTERNATE_SECRET_ACCESS_KEY", legacyAWSAlternateSecretAccessKey, display = ParameterDisplay.HIDDEN)
        }

        // Assume Role credentials
        password("AWS_ACCESS_KEY_ID", awsAccessKeyID, display = ParameterDisplay.HIDDEN)
        password("AWS_SECRET_ACCESS_KEY", awsSecretAccessKey, display = ParameterDisplay.HIDDEN)
        text("ACCTEST_ROLE_ARN", accTestRoleARN, display = ParameterDisplay.HIDDEN)

        // Alternate Assume Role credentials
        password("AWS_ALTERNATE_ACCESS_KEY_ID", alternateAWSAccessKeyID, display = ParameterDisplay.HIDDEN)
        password("AWS_ALTERNATE_SECRET_ACCESS_KEY", alternateAWSSecretAccessKey, display = ParameterDisplay.HIDDEN)
        text("ACCTEST_ALTERNATE_ROLE_ARN", alternateAccTestRoleARN, display = ParameterDisplay.HIDDEN)

        // Define this parameter even when not set to allow individual builds to set the value
        text("env.TF_ACC_TERRAFORM_VERSION", DslContext.getParameter("terraform_version", ""))

        // These overrides exist because of the inherited dependency in the existing project structure and can
        // be removed when this is moved outside of it
        val isOnPrem = DslContext.getParameter("is_on_prem", "true").equals("true", ignoreCase = true)
        if (isOnPrem) {
            // These should be overridden in the base AWS project
            param("env.GOPATH", "")
            param("env.GO111MODULE", "") // No longer needed as of Go 1.16
            param("env.GO_VERSION", "") // We're using `goenv` and `.go-version`
        }
    }

    // subProject(Services)
}

object FullBuild : BuildType({
    name = "Performance"

    vcs {
        root(AbsoluteId(DslContext.getParameter("vcs_root_id")))

        cleanCheckout = true
    }

    steps {
        script {
            name = "Configure Go"
            scriptContent = File("./scripts/configure_goenv.sh").readText()
        }
        script {
            name = "VPC Main"
            scriptContent = File("./scripts/performance.sh").readText()
        }
        script {
            name = "SSM Main"
            scriptContent = File("./scripts/performance.sh").readText()
        }
        script {
            name = "VPC Latest Version"
            scriptContent = File("./scripts/performance.sh").readText()
        }
        script {
            name = "SSM Latest Version"
            scriptContent = File("./scripts/performance.sh").readText()
        }
        script {
            name = "Analysis"
            scriptContent = File("./scripts/performance.sh").readText()
        }

    }

    features {
        val notifierConnectionID = DslContext.getParameter("notifier.id", "")
        val notifier: Notifier? = if (notifierConnectionID != "") {
            Notifier(notifierConnectionID, DslContext.getParameter("notifier.destination"))
        } else {
            null
        }

        if (notifier != null) {
            val branchRef = DslContext.getParameter("branch_name", "")
            notifications {
                notifierSettings = slackNotifier {
                    connection = notifier.connectionID
                    sendTo = notifier.destination
                    messageFormat = verboseMessageFormat {
                        addBranch = branchRef != "refs/heads/f-teamcity-memcpu-prof"
                        addStatusText = true
                    }
                }
                buildStarted = true
                buildFailedToStart = true
                buildFailed = true
                buildFinishedSuccessfully = true
                firstBuildErrorOccurs = true
            }
        }
    }
})
