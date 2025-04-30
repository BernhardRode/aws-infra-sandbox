package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

// init the Gin Server
func init() {
	// stdout and stderr are sent to AWS CloudWatch Logs
	log.Printf("Gin cold start")
	r := gin.Default()
	
	// Add middleware to log the request path
	r.Use(func(c *gin.Context) {
		log.Printf("Request path: %s", c.Request.URL.Path)
		c.Next()
	})
	
	// Define routes with the /gin-server prefix
	r.GET("/gin-server/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	
	r.GET("/gin-server", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello from Gin and ebbo.dev!",
		})
	})
	
	r.GET("/gin-server/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello from Gin and ebbo.dev!",
		})
	})
	
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		log.Printf("No route found for: %s", path)
		c.JSON(404, gin.H{
			"message": "Route not found: " + path,
		})
	})

	ginLambda = ginadapter.New(r)
}

// Handler will deal with Gin working with Lambda
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Log the incoming request
	log.Printf("API Gateway Request: %+v", req)
	
	// Process the request without modifying the path
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
