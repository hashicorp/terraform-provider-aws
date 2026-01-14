/**
 * Copyright IBM Corp. 2014, 2026
 * SPDX-License-Identifier: MPL-2.0
 */

// Copyright 2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
const { DynamoDBClient } = require('@aws-sdk/client-dynamodb');
const { DynamoDBDocumentClient, PutCommand } = require('@aws-sdk/lib-dynamodb');

const client = new DynamoDBClient({});
const ddb = DynamoDBDocumentClient.from(client);

exports.handler = async event => {
	const putParams = {
		TableName: process.env.TABLE_NAME,
		Item: {
			connectionId: event.requestContext.connectionId
		}
	};

	try {
		await ddb.send(new PutCommand(putParams));
	} catch (err) {
		return { statusCode: 500, body: 'Failed to connect: ' + JSON.stringify(err) };
	}

	return { statusCode: 200, body: 'Connected.' };
};
