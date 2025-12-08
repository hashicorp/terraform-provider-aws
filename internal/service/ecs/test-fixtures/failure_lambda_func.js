/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */


exports.handler = async (event, context) => {
    console.log('Event received:', JSON.stringify(event));
    return { hookStatus: 'FAILED' };
};