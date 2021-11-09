import jetbrains.buildServer.configs.kotlin.v2019_2.AbsoluteId
import jetbrains.buildServer.configs.kotlin.v2019_2.BuildType
import jetbrains.buildServer.configs.kotlin.v2019_2.DslContext
import jetbrains.buildServer.configs.kotlin.v2019_2.ParameterDisplay
import jetbrains.buildServer.configs.kotlin.v2019_2.buildSteps.script
import java.io.File

data class ServiceSpec(val readableName: String, val patternOverride: String = "")

class Service(name: String, displayName: ServiceSpec) {
    val packageName = name
    val displayName = displayName

    fun buildType(): BuildType {
        return BuildType {
            id = DslContext.createId("ServiceTest_$packageName")

            name = "2. ${displayName.readableName} - Tests"

            vcs {
                root(AbsoluteId(DslContext.getParameter("vcs_root_id")))
                cleanCheckout = true
            }

            if (displayName.patternOverride != "") {
                params {
                    text("TEST_PATTERN", displayName.patternOverride, display = ParameterDisplay.HIDDEN)
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
        }
    }
}
