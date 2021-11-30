import jetbrains.buildServer.configs.kotlin.v2019_2.AbsoluteId
import jetbrains.buildServer.configs.kotlin.v2019_2.BuildType
import jetbrains.buildServer.configs.kotlin.v2019_2.DslContext
import jetbrains.buildServer.configs.kotlin.v2019_2.ParameterDisplay
import jetbrains.buildServer.configs.kotlin.v2019_2.buildSteps.script
import java.io.File

data class ServiceSpec(
    val readableName: String,
    val patternOverride: String? = null,
    val vpcLock: Boolean = false,
    val parallelismOverride: Int? = null,
)

class Service(name: String, spec: ServiceSpec) {
    val packageName = name
    val spec = spec

    fun buildType(): BuildType {
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

            if (spec.vpcLock) {
                features {
                    feature {
                        type = "JetBrains.SharedResources"
                        param("locks-param", "${DslContext.getParameter("aws_account.vpc_lock_id")} readLock")
                    }
                }
            }
        }
    }
}
