var http = require('http')

exports.handler = async function(event, context) {
    return {
        body: "ok_modified",
        statusCode: 200
    }
}
