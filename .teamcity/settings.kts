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
val accTestRoleARN = DslContext.getParameter("aws_account.role_arn", "")
val awsAlternateAccountID = DslContext.getParameter("aws_alt_account.account_id", "")
val tfLog = DslContext.getParameter("tf_log", "")

val aaki = DslContext.getParameter("AWS_ACCESS_KEY_ID", "")
val asak = DslContext.getParameter("AWS_SECRET_ACCESS_KEY", "")

project {
    if (DslContext.getParameter("build_full", "true").toBoolean()) {
        buildType(FullBuild)
    }

    params {
        if (acctestParallelism != "") {
            text("ACCTEST_PARALLELISM", acctestParallelism, allowEmpty = false)
        }
        text("TEST_PATTERN", "TestAcc", display = ParameterDisplay.HIDDEN)
        text("env.AWS_ACCOUNT_ID", awsAccountID, display = ParameterDisplay.HIDDEN, allowEmpty = false)
        text("env.AWS_DEFAULT_REGION", defaultRegion, allowEmpty = false)
        text("env.TF_LOG", tfLog)

        if (alternateRegion != "") {
            text("env.AWS_ALTERNATE_REGION", alternateRegion)
        }

        val brancRef = DslContext.getParameter("branch_name", "")
        if (brancRef != "") {
            text("BRANCH_NAME", brancRef, display = ParameterDisplay.HIDDEN)
        }

        // Assume Role credentials
        password("AWS_ACCESS_KEY_ID", aaki, display = ParameterDisplay.HIDDEN)
        password("AWS_SECRET_ACCESS_KEY", asak, display = ParameterDisplay.HIDDEN)
        text("ACCTEST_ROLE_ARN", accTestRoleARN, display = ParameterDisplay.HIDDEN)

        // Define this parameter even when not set to allow individual builds to set the value
        text("env.TF_ACC_TERRAFORM_VERSION", DslContext.getParameter("terraform_version", ""))
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
            name = "Performance"
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
