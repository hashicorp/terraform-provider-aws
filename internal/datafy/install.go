package datafy

import "fmt"

func InstallScript(token, version string) string {
	if version == "latest" {
		version = ""
	}
	return fmt.Sprintf(`curl -sSfL https://agent.datafy.io/install?version=%s | TOKEN="%s" sh`, version, token)
}

func InstallScriptWithSecretManager(tokenSecretName string, version string) string {
	if version == "latest" {
		version = ""
	}

	return fmt.Sprintf(`
DATAFY_TOKEN=$(aws secretsmanager get-secret-value --secret-id "%s" --query 'SecretString' --output text)
if [ $? -ne 0 ]; then
  echo "Failed to retrieve the secret value"
else
  curl -sSfL https://agent.datafy.io/install?version=%s | TOKEN="${DATAFY_TOKEN}" sh
fi
`, tokenSecretName, version)
}
