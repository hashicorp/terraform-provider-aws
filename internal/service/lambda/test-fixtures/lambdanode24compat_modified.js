/**
 * Copyright IBM Corp. 2014, 2026
 * SPDX-License-Identifier: MPL-2.0
 */

// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

var http = require('http')

exports.handler = async function(event, context) {
    return {
        body: "ok_modified",
        statusCode: 200
    }
}
