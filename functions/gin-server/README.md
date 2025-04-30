# Gin Server Lambda Function

This Lambda function implements a simple web server using the Gin framework.

## Path Handling

The function automatically strips the `/gin-server` prefix from incoming requests. This allows the function to be deployed at any path in API Gateway without requiring changes to the function code.

## Available Routes

- `/` - Returns a welcome message
- `/ping` - Returns a "pong" response

## Development

To add new routes, simply add them to the Gin router in the `init()` function.
