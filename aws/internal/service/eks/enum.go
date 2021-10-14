package eks

const (
	IdentityProviderConfigTypeOIDC = "oidc"
)

const (
	ResourcesSecrets = "secrets"
)

func Resources_Values() []string {
	return []string{
		ResourcesSecrets,
	}
}
