package cognitoidp

import "time"

const (
	ResIdentityProvider  = "Identity Provider"
	ResResourceServer    = "Resource Server"
	ResRiskConfiguration = "Risk Configuration"
	ResUserGroup         = "User Group"
	ResUserPoolClient    = "User Pool Client"
	ResUserPoolDomain    = "User Pool Domain"
	ResUserPool          = "User Pool"
	ResUser              = "User"
)

const (
	propagationTimeout = 2 * time.Minute
)
