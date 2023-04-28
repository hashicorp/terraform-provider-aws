import java.io.File
import jetbrains.buildServer.configs.kotlin.v2019_2.AbsoluteId
import jetbrains.buildServer.configs.kotlin.v2019_2.BuildType
import jetbrains.buildServer.configs.kotlin.v2019_2.DslContext
import jetbrains.buildServer.configs.kotlin.v2019_2.ParameterDisplay
import jetbrains.buildServer.configs.kotlin.v2019_2.buildFeatures.notifications
import jetbrains.buildServer.configs.kotlin.v2019_2.buildSteps.script

data class ServiceSpec(
    val readableName: String,
    val patternOverride: String? = null,
    val vpcLock: Boolean = false,
    val parallelismOverride: Int? = null,
    val regionOverride: String? = null,
)

data class Notifier(
    val connectionID: String,
    val destination: String,
)

class Service(name: String, spec: ServiceSpec) {
    val packageName = name
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

            val serviceDir = "./internal/service/$packageName"
            steps {
                script {
                    name = "Setup GOENV"
                    scriptContent = File("./scripts/setup_goenv.sh").readText()
                }
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
                        buildStarted = true
                        buildFailedToStart = true
                        buildFailed = true
                        buildFinishedSuccessfully = true
                        // Ideally we'd have this enabled, but we have too many failures and this would get very noisy
                        // firstBuildErrorOccurs = true
                    }
                }
            }
        }
    }
}
