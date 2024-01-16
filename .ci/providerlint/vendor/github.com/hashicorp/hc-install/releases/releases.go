// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package releases

import (
	"io/ioutil"
	"log"
	"time"
)

var (
	defaultInstallTimeout = 30 * time.Second
	defaultListTimeout    = 10 * time.Second
	discardLogger         = log.New(ioutil.Discard, "", 0)
)
