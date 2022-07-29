package comprehend

import (
	"time"
)

const iamPropagationTimeout = 2 * time.Minute

// Avoid service throttling
const entityRegcognizerDelay = 10 * time.Minute
const entityRegcognizerPollInterval = 60 * time.Second
