package eks

const (
	IdentityProviderConfigTypeOidc = "oidc"
)

const (
	ResourcesSecrets = "secrets"
)

func Resources_Values() []string {
	return []string{
		ResourcesSecrets,
	}
}
