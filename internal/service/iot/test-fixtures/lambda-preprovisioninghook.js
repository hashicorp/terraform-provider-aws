/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

//https://docs.aws.amazon.com/iot/latest/developerguide/pre-provisioning-hook.html#pre-provisioning-example
exports.handler = function (event, context, callback) {
    console.log(JSON.stringify(event, null, 2));
    var reply = {
        allowProvisioning: true,
        parameterOverrides: {
            DeviceLocation: 'Seattle'
        }
    };
    callback(null, reply);
}
