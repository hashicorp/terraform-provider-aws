/**
 * Copyright (c) HashiCorp, Inc.
 * SPDX-License-Identifier: MPL-2.0
 */

const AWS = require('aws-sdk')
const ssmClient = new AWS.SSM();

exports.handler = async (event) => {
    let tf_key = "tf";
    if (tf_key in event) {
        if (event[tf_key].action == "delete" && process.env.TEST_DATA != "") {
            await ssmClient.putParameter({
                Name: process.env.TEST_DATA,
                Value: JSON.stringify(event),
                Type: "String"
            }).promise();
        }
    }
    return event;
}
