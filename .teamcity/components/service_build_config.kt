import jetbrains.buildServer.configs.kotlin.AbsoluteId
import jetbrains.buildServer.configs.kotlin.BuildSteps
import jetbrains.buildServer.configs.kotlin.BuildType
import jetbrains.buildServer.configs.kotlin.DslContext
import jetbrains.buildServer.configs.kotlin.ParameterDisplay
import jetbrains.buildServer.configs.kotlin.buildFeatures.notifications
import jetbrains.buildServer.configs.kotlin.buildSteps.ScriptBuildStep
import jetbrains.buildServer.configs.kotlin.buildSteps.script
import jetbrains.buildServer.configs.kotlin.failureConditions.failOnText
import jetbrains.buildServer.configs.kotlin.failureConditions.BuildFailureOnText
import java.io.File

data class ServiceSpec(
    val readableName: String,
    val patternOverride: String? = null,
    val vpcLock: Boolean = false,
    val parallelismOverride: Int? = null,
    val regionOverride: String? = null,
    val splitPackageRealPackage: String? = null,
    val excludePattern: String? = null,
)

data class Notifier(
    val connectionID: String,
    val destination: String,
)

class Service(name: String, spec: ServiceSpec) {
    private var packageName = name
    val spec = spec

    fun buildType(notifier: Notifier?): BuildType {
        return BuildType {
            id = DslContext.createId("ServiceTest_$packageName")

            name = "2. ${spec.readableName} - Tests"

            vcs {
                root(AbsoluteId(DslContext.getParameter("vcs_root_id")))
                cleanCheckout = true
            }

            if (spec.patternOverride != null) {
                params {
                    text("TEST_PATTERN", spec.patternOverride, display = ParameterDisplay.HIDDEN)
                }
            }
            if (spec.parallelismOverride != null) {
                params {
                    text("ACCTEST_PARALLELISM", spec.parallelismOverride.toString(), display = ParameterDisplay.HIDDEN)
                }
            }
            if (spec.regionOverride != null) {
                params {
                    text("env.AWS_DEFAULT_REGION", spec.regionOverride, display = ParameterDisplay.HIDDEN)
                }
            }
            if (spec.excludePattern != null) {
                params {
                    text("TEST_EXCLUDE_PATTERN", spec.excludePattern, display = ParameterDisplay.HIDDEN)
                }
            }
            if (spec.splitPackageRealPackage != null) {
                packageName = spec.splitPackageRealPackage
            }

            val serviceDir = "./internal/service/$packageName"
            steps {
                ConfigureGoEnv()
                script {
                    name = "Compile Test Binary"
                    workingDir = serviceDir
                    scriptContent = File("./scripts/service_tests/compile_test_binary.sh").readText()
                }
                script {
                    name = "Run Unit Tests"
                    workingDir = serviceDir
                    scriptContent = File("./scripts/service_tests/unit_tests.sh").readText()
                }
                script {
                    name = "Run Acceptance Tests"
                    workingDir = serviceDir
                    scriptContent = File("./scripts/service_tests/acceptance_tests.sh").readText()
                }
            }

            failureConditions {
                failOnText {
                    conditionType = BuildFailureOnText.ConditionType.REGEXP
                    pattern = """(?i)build canceled"""
                    failureMessage = "build canceled when agent unregistered"
                    reverse = false
                    stopBuildOnFailure = true
                    reportOnlyFirstMatch = false
                }
            }

            features {
                if (spec.vpcLock) {
                    feature {
                        type = "JetBrains.SharedResources"
                        param("locks-param", "${DslContext.getParameter("aws_account.vpc_lock_id")} readLock")
                    }
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
                        buildFailed = false // With the current number of faling tests, this would be too noisy
                        buildFinishedSuccessfully = false // With the number of tests, this would be too noisy
                        firstSuccessAfterFailure = true
                        buildProbablyHanging = false
                        // Ideally we'd have this enabled, but we have too many failures and this would get very noisy
                        // firstBuildErrorOccurs = true
                    }
                }
            }
        }
    }
}

fun BuildSteps.ConfigureGoEnv() {
    step(ScriptBuildStep {
        name = "Configure GOENV"
        scriptContent = File("./scripts/configure_goenv.sh").readText()
    })
}
