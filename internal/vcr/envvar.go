// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vcr

import (
	"fmt"
	"os"

	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

const (
	envVarVCRMode = "VCR_MODE"
	envVarVCRPath = "VCR_PATH"

	vcrModeRecordOnly = "RECORD_ONLY"
	vcrModeReplayOnly = "REPLAY_ONLY"
)

// IsEnabled indicates whether VCR testing is enabled
//
// Returns true if both the VCR_MODE and VCR_PATH environment variables are set
// to non-empty values.
func IsEnabled() bool {
	return os.Getenv(envVarVCRMode) != "" && os.Getenv(envVarVCRPath) != ""
}

// Mode returns the VCR recording mode inferred from the VCR_MODE environment variable
func Mode() (recorder.Mode, error) {
	switch v := os.Getenv(envVarVCRMode); v {
	case vcrModeRecordOnly:
		return recorder.ModeRecordOnly, nil
	case vcrModeReplayOnly:
		return recorder.ModeReplayOnly, nil
	default:
		return recorder.ModePassthrough, fmt.Errorf("unsupported value for %s: %s", envVarVCRMode, v)
	}
}

// Path returns the directory in which VCR recordings should be stored
func Path() string {
	return os.Getenv(envVarVCRPath)
}
