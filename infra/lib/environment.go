package lib

import (
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

// Environment represents the deployment environment
type Environment struct {
	Name      string
	PRNumber  string
	Version   string
	Username  string
	IsPreview bool
}

// getCurrentUsername retrieves the current username from the environment variables
func getCurrentUsername() string {
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	if len(username) == 0 {
		username = "default"
	}
	return username
}

// kebabCase converts a string to kebab-case
func kebabCase(s string) string {
	// Split the string by uppercase letters
	words := regexp.MustCompile("[A-Z][^A-Z]*").FindAllString(s, -1)
	// Join the words with a hyphen
	kebab := strings.Join(words, "-")
	return strings.ToLower(kebab)
}

func (e *Environment) GetStackName(suffix string) string {
	// recursive print e
	println("Environment:", e)
	println("Environment Name:", e.Name)
	println("Environment PRNumber:", e.PRNumber)
	println("Environment Version:", e.Version)
	println("Environment Username:", e.Username)
	println("Environment IsPreview:", e.IsPreview)
	println("Environment Suffix:", suffix)

	if e.Name == "preview" {
		return e.Name + "-" + e.PRNumber + "-" + kebabCase(suffix)
	}
	if e.Name == "development" {
		// If username is empty, use the current username
		username := e.Username
		if username == "" {
			username = getCurrentUsername()
		}
		return e.Name + "-" + username + "-" + kebabCase(suffix)
	}

	return e.Name + "-" + kebabCase(suffix)
}

// GetEnvironmentFromContext extracts environment information from CDK context
func GetEnvironmentFromContext(app awscdk.App) Environment {
	env := Environment{
		Name:      "unknown",
		IsPreview: false,
	}

	// Add nil check for app
	if app == nil {
		return env
	}

	envName := app.Node().TryGetContext(jsii.String("environment"))
	prNumber := app.Node().TryGetContext(jsii.String("prNumber"))
	version := app.Node().TryGetContext(jsii.String("version"))
	username := app.Node().TryGetContext(jsii.String("username"))
	sha := app.Node().TryGetContext(jsii.String("sha"))

	println("Environment Name:", envName)
	println("Environment PRNumber:", prNumber)
	println("Environment Version:", version)
	println("Environment Username:", username)
	println("Environment IsPreview:", prNumber)
	println("Environment sha:", sha)

	env = Environment{}

	if envName != nil {
		if nameStr, ok := envName.(string); ok {
			env.Name = nameStr
		}
	}

	if prNumber != nil {
		if prStr, ok := prNumber.(string); ok {
			env.PRNumber = prStr
			env.IsPreview = true
		}
	}

	if version != nil {
		if versionStr, ok := version.(string); ok {
			env.Version = versionStr
		}
	}

	if sha != nil {
		if shaStr, ok := sha.(string); ok {
			env.Version = shaStr
		}
	}

	// Set username from context or use current username as fallback
	if username != nil {
		if usernameStr, ok := username.(string); ok && usernameStr != "" {
			env.Username = usernameStr
		} else {
			env.Username = getCurrentUsername()
		}
	} else {
		env.Username = getCurrentUsername()
	}

	return env
}
