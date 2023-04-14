package cognitoidp

import (
	"time"
)

const (
	ResNameIdentityProvider  = "Identity Provider"
	ResNameResourceServer    = "Resource Server"
	ResNameRiskConfiguration = "Risk Configuration"
	ResNameUserGroup         = "User Group"
	ResNameUserPoolClient    = "User Pool Client"
	ResNameUserPoolDomain    = "User Pool Domain"
	ResNameUserPool          = "User Pool"
	ResNameUser              = "User"
)

const (
	propagationTimeout = 2 * time.Minute
)
