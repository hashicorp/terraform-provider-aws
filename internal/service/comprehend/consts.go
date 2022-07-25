package comprehend

import (
	"time"
)

const iamPropagationTimeout = 2 * time.Minute

// Avoid service throttling
const entityRegcognizerMinInterval = 1 * time.Second
