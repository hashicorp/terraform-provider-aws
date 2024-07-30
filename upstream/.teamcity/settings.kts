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

version = "2024.03"

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

    if (DslContext.getParameter("build_pullrequest", "").toBoolean() || DslContext.getParameter("pullrequest_build", "").toBoolean()) {
        buildType(PullRequest)
    }

    if (DslContext.getParameter("build_sweeperonly", "").toBoolean()) {
        buildType(Sweeper)
    }

    buildType(Sanity)
    buildType(Performance)

    params {
        if (acctestParallelism != "") {
            text("ACCTEST_PARALLELISM", acctestParallelism, allowEmpty = false)
        }
        text("TEST_PATTERN", "TestAcc", display = ParameterDisplay.HIDDEN)
        text("TEST_EXCLUDE_PATTERN", "", display = ParameterDisplay.HIDDEN)
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

    subProject(Services)
}

object PullRequest : BuildType({
    name = "Pull Request"

    vcs {
        root(AbsoluteId(DslContext.getParameter("vcs_root_id")))

        cleanCheckout = true
    }

    failureConditions {
        val defaultPullRequestTimeoutHours: Long = 6
        executionTimeoutMin = Duration.ofHours(defaultPullRequestTimeoutHours).toMinutes().toInt()
    }

    val accTestRoleARN = DslContext.getParameter("aws_account.role_arn", "")
    steps {
        ConfigureGoEnv()
        script {
            name = "Run Tests"
            scriptContent = File("./scripts/pullrequest_tests/tests.sh").readText()
        }
    }

    features {
        feature {
            type = "JetBrains.SharedResources"
            param("locks-param", "${DslContext.getParameter("aws_account.lock_id")} readLock")
        }
        val alternateAccountLockId = DslContext.getParameter("aws_alt_account.lock_id", "")
        if (alternateAccountLockId != "") {
            feature {
                type = "JetBrains.SharedResources"
                param("locks-param", "$alternateAccountLockId readLock")
            }
        }

        val notifierConnectionID = DslContext.getParameter("notifier.id", "")
        val notifier: Notifier? = if (notifierConnectionID != "") {
            Notifier(notifierConnectionID, DslContext.getParameter("notifier.destination"))
        } else {
            null
        }

        if (notifier != null) {
            notifications {
                notifierSettings = slackNotifier {
                    connection = notifier.connectionID
                    sendTo = notifier.destination
                    messageFormat = verboseMessageFormat {
                        addBranch = true
                        addStatusText = true
                    }
                }
                branchFilter = "+:*"

                buildStarted = true
                buildFailedToStart = true
                buildFailed = true
                buildFinishedSuccessfully = true
                firstBuildErrorOccurs = true
                buildProbablyHanging = false
            }
        }
    }
})

object FullBuild : BuildType({
    name = "Service Tests"

    type = BuildTypeSettings.Type.COMPOSITE

    vcs {
        showDependenciesChanges = true
    }

    val notifierConnectionID = DslContext.getParameter("notifier.id", "")
    val notifier: Notifier? = if (notifierConnectionID != "") {
        Notifier(notifierConnectionID, DslContext.getParameter("notifier.destination"))
    } else {
        null
    }

    dependencies {
        snapshot(SetUp) {
            reuseBuilds = ReuseBuilds.NO
            onDependencyFailure = FailureAction.ADD_PROBLEM
            onDependencyCancel = FailureAction.IGNORE
        }

        val testType = DslContext.getParameter("test_type", "")
        val serviceList = if (testType == "orgacct") orgacctServices else services
        serviceList.forEach { (serviceName, displayName) ->
            snapshot(Service(serviceName, displayName).buildType(notifier)) {
                reuseBuilds = ReuseBuilds.NO
                onDependencyFailure = FailureAction.ADD_PROBLEM
                onDependencyCancel = FailureAction.IGNORE
            }
        }

        snapshot(CleanUp) {
            reuseBuilds = ReuseBuilds.NO
            onDependencyFailure = FailureAction.IGNORE
            onDependencyCancel = FailureAction.IGNORE
        }
    }

    val runNightly = DslContext.getParameter("run_nightly_build", "")
    if (runNightly.toBoolean()) {
        val triggerTimeRaw = DslContext.getParameter("trigger_time")
        val formatter = DateTimeFormatter.ofPattern("HH':'mm' 'VV")
        val triggerTime = formatter.parse(triggerTimeRaw)
        val triggerDay = if (DslContext.getParameter("trigger_day", "") != "") {
            DslContext.getParameter("trigger_day", "")
        } else {
            "Sun-Thu"
        }

        val enableTestTriggersGlobally = DslContext.getParameter("enable_test_triggers_globally", "true").equals("true", ignoreCase = true)
        if (enableTestTriggersGlobally) {
            triggers {
                schedule {
                    schedulingPolicy = cron {
                        dayOfWeek = triggerDay
                        val triggerHM = LocalTime.from(triggerTime)
                        hours = triggerHM.getHour().toString()
                        minutes = triggerHM.getMinute().toString()
                        timezone = ZoneId.from(triggerTime).toString()
                    }
                    branchFilter = "" // For a Composite build, the branch filter must be empty
                    triggerBuild = always()
                    withPendingChangesOnly = false
                    enableQueueOptimization = false
                    enforceCleanCheckoutForDependencies = true
                }
            }
        }
    }

    features {
        feature {
            type = "JetBrains.SharedResources"
            param("locks-param", "${DslContext.getParameter("aws_account.lock_id")} writeLock")
        }
        val alternateAccountLockId = DslContext.getParameter("aws_alt_account.lock_id", "")
        if (alternateAccountLockId != "") {
            feature {
                type = "JetBrains.SharedResources"
                param("locks-param", "$alternateAccountLockId readLock")
            }
        }

        if (notifier != null) {
            notifications {
                notifierSettings = slackNotifier {
                    connection = notifier.connectionID
                    sendTo = notifier.destination
                    messageFormat = simpleMessageFormat()
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

object SetUp : BuildType({
    name = "1. Set Up"

    vcs {
        root(AbsoluteId(DslContext.getParameter("vcs_root_id")))

        cleanCheckout = true
    }

    steps {
        ConfigureGoEnv()
        script {
            name = "Run provider unit tests"
            scriptContent = File("./scripts/provider_tests/unit_tests.sh").readText()
        }
        script {
            name = "Run provider acceptance tests"
            scriptContent = File("./scripts/provider_tests/acceptance_tests.sh").readText()
        }
        script {
            name = "Pre-Sweeper"
            executionMode = BuildStep.ExecutionMode.RUN_ON_FAILURE
            scriptContent = File("./scripts/sweeper.sh").readText()
        }
    }

    features {
        golang {
            testFormat = "json"
        }

        val notifierConnectionID = DslContext.getParameter("notifier.id", "")
        val notifier: Notifier? = if (notifierConnectionID != "") {
            Notifier(notifierConnectionID, DslContext.getParameter("notifier.destination"))
        } else {
            null
        }

        if (notifier != null) {
            notifications {
                notifierSettings = slackNotifier {
                    connection = notifier.connectionID
                    sendTo = notifier.destination
                    messageFormat = verboseMessageFormat {
                        addBranch = true
                        addStatusText = true
                    }
                }
                buildStarted = false // With the number of tests, this would be too noisy
                buildFailedToStart = true
                buildFailed = true
                buildFinishedSuccessfully = false // With the number of tests, this would be too noisy
                firstSuccessAfterFailure = true
                buildProbablyHanging = false
                // Ideally we'd have this enabled, but we have too many failures and this would get very noisy
                // firstBuildErrorOccurs = true
            }
        }
    }
})

object Services : Project({
    id = DslContext.createId("Services")

    name = "Services"

    val notifierConnectionID = DslContext.getParameter("notifier.id", "")
    val notifier: Notifier? = if (notifierConnectionID != "") {
        Notifier(notifierConnectionID, DslContext.getParameter("notifier.destination"))
    } else {
        null
    }

    val buildChain = sequential {
        buildType(SetUp)

        val testType = DslContext.getParameter("test_type", "")
        val serviceList = if (testType == "orgacct") orgacctServices else services
        parallel(options = { onDependencyFailure = FailureAction.IGNORE }) {
            serviceList.forEach { (serviceName, displayName) ->
                buildType(Service(serviceName, displayName).buildType(notifier))
            }
        }

        buildType(CleanUp, options = {
            reuseBuilds = ReuseBuilds.NO
            onDependencyFailure = FailureAction.IGNORE
            onDependencyCancel = FailureAction.IGNORE
        })
    }
    buildChain.buildTypes().forEach { buildType(it) }
})

object CleanUp : BuildType({
    name = "3. Clean Up"

    vcs {
        root(AbsoluteId(DslContext.getParameter("vcs_root_id")))

        cleanCheckout = true
    }

    steps {
        ConfigureGoEnv()
        script {
            name = "Post-Sweeper"
            enabled = false
            scriptContent = File("./scripts/sweeper.sh").readText()
        }
    }
})

object Sweeper : BuildType({
    name = "Sweeper"

    vcs {
        root(AbsoluteId(DslContext.getParameter("vcs_root_id")))

        cleanCheckout = true
    }

    steps {
        ConfigureGoEnv()
        script {
            name = "Sweeper"
            scriptContent = File("./scripts/sweeper.sh").readText()
        }
    }

    val triggerTimeRaw = DslContext.getParameter("sweeper_trigger_time", "")
    if (triggerTimeRaw != "") {
        val formatter = DateTimeFormatter.ofPattern("HH':'mm' 'VV")
        val triggerTime = formatter.parse(triggerTimeRaw)
        val enableTestTriggersGlobally = DslContext.getParameter("enable_test_triggers_globally", "true").equals("true", ignoreCase = true)
        if (enableTestTriggersGlobally) {
            triggers {
                schedule {
                    schedulingPolicy = daily {
                        val triggerHM = LocalTime.from(triggerTime)
                        hour = triggerHM.getHour()
                        minute = triggerHM.getMinute()
                        timezone = ZoneId.from(triggerTime).toString()
                    }
                    branchFilter = "+:refs/heads/main"
                    triggerBuild = always()
                    withPendingChangesOnly = false
                    enableQueueOptimization = true
                    enforceCleanCheckoutForDependencies = true
                }
            }
        }
    }

    failureConditions {
        failOnText {
            conditionType = BuildFailureOnText.ConditionType.REGEXP
            pattern = """Sweeper Tests for region \(([-a-z0-9]+)\) ran unsuccessfully"""
            failureMessage = """Sweeper failure for region "${'$'}1""""
            reverse = false
            reportOnlyFirstMatch = false
        }
    }

    features {
        feature {
            type = "JetBrains.SharedResources"
            param("locks-param", "${DslContext.getParameter("aws_account.lock_id")} writeLock")
        }

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
                        addBranch = branchRef != "refs/heads/main"
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

object Sanity : BuildType({
    name = "Sanity"

    vcs {
        root(AbsoluteId(DslContext.getParameter("vcs_root_id")))

        cleanCheckout = true
    }

    steps {
        ConfigureGoEnv()
        script {
            name = "IAM"
            scriptContent = File("./scripts/sanity.sh").readText()
        }
        script {
            name = "Logs"
            scriptContent = File("./scripts/sanity.sh").readText()
        }
        script {
            name = "EC2"
            scriptContent = File("./scripts/sanity.sh").readText()
        }
        script {
            name = "ECS"
            scriptContent = File("./scripts/sanity.sh").readText()
        }
        script {
            name = "ELBv2"
            scriptContent = File("./scripts/sanity.sh").readText()
        }
        script {
            name = "KMS"
            scriptContent = File("./scripts/sanity.sh").readText()
        }
        script {
            name = "IAM"
            scriptContent = File("./scripts/sanity.sh").readText()
        }
        script {
            name = "Lambda"
            scriptContent = File("./scripts/sanity.sh").readText()
        }
        script {
            name = "Meta"
            scriptContent = File("./scripts/sanity.sh").readText()
        }
        script {
            name = "Route53"
            scriptContent = File("./scripts/sanity.sh").readText()
        }
        script {
            name = "S3"
            scriptContent = File("./scripts/sanity.sh").readText()
        }
        script {
            name = "Secrets Manager"
            scriptContent = File("./scripts/sanity.sh").readText()
        }
        script {
            name = "STS"
            scriptContent = File("./scripts/sanity.sh").readText()
        }  
        script {
            name = "Report Success"
            scriptContent = File("./scripts/sanity.sh").readText()
        }    
    }

    val triggerTimeRaw = DslContext.getParameter("sanity_trigger_time", "")
    if (triggerTimeRaw != "") {
        val formatter = DateTimeFormatter.ofPattern("HH':'mm' 'VV")
        val triggerTime = formatter.parse(triggerTimeRaw)
        val enableTestTriggersGlobally = DslContext.getParameter("enable_test_triggers_globally", "true").equals("true", ignoreCase = true)
        if (enableTestTriggersGlobally) {
            triggers {
                schedule {
                    schedulingPolicy = daily {
                        val triggerHM = LocalTime.from(triggerTime)
                        hour = triggerHM.getHour()
                        minute = triggerHM.getMinute()
                        timezone = ZoneId.from(triggerTime).toString()
                    }
                    branchFilter = "+:refs/heads/main"
                    triggerBuild = always()
                    withPendingChangesOnly = false
                    enableQueueOptimization = true
                    enforceCleanCheckoutForDependencies = true
                }
            }
        }
    }

    features {
        feature {
            type = "JetBrains.SharedResources"
            param("locks-param", "${DslContext.getParameter("aws_account.lock_id")} writeLock")
        }
        feature {
            type = "JetBrains.SharedResources"
            param("locks-param", "${DslContext.getParameter("aws_account.vpc_lock_id")} readLock")
        }
    }
})

object Performance : BuildType({
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

    val triggerTimeRaw = DslContext.getParameter("performance_trigger_time", "")
    if (triggerTimeRaw != "") {
        val formatter = DateTimeFormatter.ofPattern("HH':'mm' 'VV")
        val triggerTime = formatter.parse(triggerTimeRaw)
        val enableTestTriggersGlobally = DslContext.getParameter("enable_test_triggers_globally", "true").equals("true", ignoreCase = true)
        if (enableTestTriggersGlobally) {
            triggers {
                schedule {
                    schedulingPolicy = daily {
                        val triggerHM = LocalTime.from(triggerTime)
                        hour = triggerHM.getHour()
                        minute = triggerHM.getMinute()
                        timezone = ZoneId.from(triggerTime).toString()
                    }
                    branchFilter = "+:refs/heads/main"
                    triggerBuild = always()
                    withPendingChangesOnly = false
                    enableQueueOptimization = true
                    enforceCleanCheckoutForDependencies = true
                }
            }
        }
    }

    features {
        feature {
            type = "JetBrains.SharedResources"
            param("locks-param", "${DslContext.getParameter("aws_account.lock_id")} writeLock")
        }
        feature {
            type = "JetBrains.SharedResources"
            param("locks-param", "${DslContext.getParameter("aws_account.vpc_lock_id")} readLock")
        }
    }
})
