package lib

import (
	"fmt"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

// Environment represents the deployment environment
type Environment struct {
	Name      string
	PRNumber  string
	Version   string
	IsPreview bool
}

// GetEnvironmentFromContext extracts environment information from CDK context
func GetEnvironmentFromContext(app awscdk.App) Environment {
	envName := app.Node().TryGetContext(jsii.String("environment"))
	prNumber := app.Node().TryGetContext(jsii.String("prNumber"))
	version := app.Node().TryGetContext(jsii.String("version"))

	env := Environment{
		Name:      "dev",
		IsPreview: false,
	}

	if envName != nil {
		env.Name = *envName.(*string)
	}

	if prNumber != nil {
		env.PRNumber = *prNumber.(*string)
		env.IsPreview = true
	}

	if version != nil {
		env.Version = *version.(*string)
	}

	return env
}

// GetStackName generates a stack name based on environment
func (e *Environment) GetStackName(baseName string) string {
	if e.IsPreview && e.PRNumber != "" {
		return fmt.Sprintf("%s-pr-%s", baseName, e.PRNumber)
	}
	
	if e.Name != "production" {
		return fmt.Sprintf("%s-%s", baseName, e.Name)
	}
	
	return baseName
}

// GetResourceName generates a resource name with environment suffix
func (e *Environment) GetResourceName(baseName string) string {
	if e.IsPreview && e.PRNumber != "" {
		return fmt.Sprintf("%s-pr%s", baseName, e.PRNumber)
	}
	
	if e.Name != "production" {
		return fmt.Sprintf("%s-%s", baseName, e.Name)
	}
	
	return baseName
}

// GetTags returns common tags for resources
func (e *Environment) GetTags() map[string]*string {
	tags := map[string]*string{
		"Environment": jsii.String(e.Name),
		"ManagedBy":   jsii.String("CDK"),
	}
	
	if e.IsPreview {
		tags["Preview"] = jsii.String("true")
		tags["PR"] = jsii.String(e.PRNumber)
	}
	
	if e.Version != "" {
		tags["Version"] = jsii.String(e.Version)
	}
	
	return tags
}
