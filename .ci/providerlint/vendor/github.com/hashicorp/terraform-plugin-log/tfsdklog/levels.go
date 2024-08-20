package tfsdklog

import (
	"sync"

	"github.com/hashicorp/go-hclog"
)

var (
	// rootLevel stores the effective level of the root SDK logger during
	// NewRootSDKLogger where the value is deterministically chosen based on
	// environment variables, etc. This call generally happens with each new
	// provider RPC request. If the environment variable values changed during
	// runtime between calls, then inflight provider requests checking this
	// value would receive the most up-to-date value which would potentially
	// differ with the actual in-context logger level. This tradeoff would only
	// effect the inflight requests and should not be an overall performance
	// concern in the case of this level causing more context checks until the
	// request is over.
	rootLevel hclog.Level = hclog.NoLevel

	// rootLevelMutex is a read-write mutex that protects rootLevel from
	// triggering the data race detector.
	rootLevelMutex = sync.RWMutex{}

	// subsystemLevels stores the effective level of all subsystem SDK loggers
	// during NewSubsystem where the value is deterministically chosen based on
	// environment variables, etc. This call generally happens with each new
	// provider RPC request. If the environment variable values changed during
	// runtime between calls, then inflight provider requests checking this
	// value would receive the most up-to-date value which would potentially
	// differ with the actual in-context logger level. This tradeoff would only
	// effect the inflight requests and should not be an overall performance
	// concern in the case of this level causing more context checks until the
	// request is over.
	subsystemLevels map[string]hclog.Level = make(map[string]hclog.Level)

	// subsystemLevelsMutex is a read-write mutex that protects the
	// subsystemLevels map from concurrent read and write panics.
	subsystemLevelsMutex = sync.RWMutex{}
)

// subsystemWouldLog returns true if the subsystem SDK logger would emit a log
// at the given level. This is performed outside the context-based logger for
// performance.
func subsystemWouldLog(subsystem string, level hclog.Level) bool {
	subsystemLevelsMutex.RLock()

	setLevel, ok := subsystemLevels[subsystem]

	subsystemLevelsMutex.RUnlock()

	if !ok {
		return false
	}

	return wouldLog(setLevel, level)
}

// rootWouldLog returns true if the root SDK logger would emit a log at the
// given level. This is performed outside the context-based logger for
// performance.
func rootWouldLog(level hclog.Level) bool {
	rootLevelMutex.RLock()

	setLevel := rootLevel

	rootLevelMutex.RUnlock()

	return wouldLog(setLevel, level)
}

// wouldLog returns true if the set level would emit a log at the given
// level. This is performed outside the context-based logger for performance.
func wouldLog(setLevel, checkLevel hclog.Level) bool {
	if checkLevel == hclog.Off {
		return false
	}

	return checkLevel >= setLevel
}
