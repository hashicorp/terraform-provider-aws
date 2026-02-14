/**
 * Copyright IBM Corp. 2014, 2026
 * SPDX-License-Identifier: MPL-2.0
 */

exports.handler = async (event) => {
    if (process.env.TEST_DATA) {
        event.key3 = process.env.TEST_DATA;
    }
    return event;
}
