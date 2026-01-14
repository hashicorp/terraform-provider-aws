/**
 * Copyright IBM Corp. 2014, 2026
 * SPDX-License-Identifier: MPL-2.0
 */

// Copyright 2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
const { ApiGatewayManagementApiClient, PostToConnectionCommand } = require('@aws-sdk/client-apigatewaymanagementapi');
const { DynamoDBClient } = require('@aws-sdk/client-dynamodb');
const { DynamoDBDocumentClient, ScanCommand, DeleteCommand } = require('@aws-sdk/lib-dynamodb');

const ddbClient = new DynamoDBClient({});
const ddb = DynamoDBDocumentClient.from(ddbClient);

const { TABLE_NAME } = process.env;

exports.handler = async event => {
	let connectionData;

	try {
		connectionData = await ddb.send(new ScanCommand({ TableName: TABLE_NAME, ProjectionExpression: 'connectionId' }));
	} catch (e) {
		return { statusCode: 500, body: e.stack };
	}

	const apigwManagementApi = new ApiGatewayManagementApiClient({
		endpoint: `https://${event.requestContext.domainName}/${event.requestContext.stage}`
	});

	const postData = JSON.parse(event.body).data;

	const postCalls = connectionData.Items.map(async ({ connectionId }) => {
		try {
			await apigwManagementApi.send(new PostToConnectionCommand({ ConnectionId: connectionId, Data: postData }));
		} catch (e) {
			if (e.statusCode === 410) {
				console.log(`Found stale connection, deleting ${connectionId}`);
				await ddb.send(new DeleteCommand({ TableName: TABLE_NAME, Key: { connectionId } }));
			} else {
				throw e;
			}
		}
	});

	try {
		await Promise.all(postCalls);
	} catch (e) {
		return { statusCode: 500, body: e.stack };
	}

	return { statusCode: 200, body: 'Data sent.' };
};
