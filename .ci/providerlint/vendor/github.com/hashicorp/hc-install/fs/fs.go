// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fs

import (
	"io/ioutil"
	"log"
	"time"
)

var (
	defaultTimeout = 10 * time.Second
	discardLogger  = log.New(ioutil.Discard, "", 0)
)

type fileCheckFunc func(path string) error
