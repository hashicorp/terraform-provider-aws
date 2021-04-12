package waiter

import (
	"time"
)

const (
	CreateTableTimeout                  = 2 * time.Minute
	UpdateTableTimeout                  = 20 * time.Minute
	UpdateTableContinuousBackupsTimeout = 20 * time.Minute
	DeleteTableTimeout                  = 5 * time.Minute
)
