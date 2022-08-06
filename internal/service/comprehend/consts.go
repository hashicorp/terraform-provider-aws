package comprehend

import (
	"time"
)

const iamPropagationTimeout = 2 * time.Minute

// Avoid service throttling
const entityRegcognizerDelay = 1 * time.Minute
const entityRegcognizerPollInterval = 1 * time.Minute
