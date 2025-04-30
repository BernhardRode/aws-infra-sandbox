package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"

	"aws-infra-sandbox/lib"
	"aws-infra-sandbox/stacks/core"
	"aws-infra-sandbox/stacks/lambda"
	"aws-infra-sandbox/stacks/vaultwarden"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	// Get environment information from context
	environment := lib.GetEnvironmentFromContext(app)

	// Add environment tags to all resources
	awscdk.Tags_Of(app).Add(jsii.String("Environment"), jsii.String(environment.Name), nil)
	awscdk.Tags_Of(app).Add(jsii.String("Username"), jsii.String(environment.Username), nil)

	if environment.IsPR {
		awscdk.Tags_Of(app).Add(jsii.String("PR"), jsii.String("true"), nil)
		awscdk.Tags_Of(app).Add(jsii.String("PR"), jsii.String(environment.PRNumber), nil)
	}
	if environment.Version != "" {
		awscdk.Tags_Of(app).Add(jsii.String("Version"), jsii.String(environment.Version), nil)
	}

	// Create stacks with environment-specific names
	coreStackName := environment.GetStackName("CoreStack")
	lambdaStackName := environment.GetStackName("LambdaStack")
	vaultwardenStackName := environment.GetStackName("VaultwardenStack")

	// Configure domain for all stacks
	domainConfig := &lib.DomainConfig{
		RootDomain:   "ebbo.dev",
		HostedZoneId: "Z02287733RP9AY57D3IRQ",
	}

	// Create props for each stack with environment information
	coreProps := &core.CoreStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
		Environment:  environment,
		DomainConfig: domainConfig,
	}

	lambdaProps := &lambda.LambdaStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
		Environment:  environment,
		DomainConfig: domainConfig,
	}

	// Configure Vaultwarden stack
	vaultwardenConfig := &vaultwarden.VaultwardenConfig{
		// Base configuration
		BaseImageName: "vaultwarden/server",
		BaseVersion:   "latest",
		DomainName:    "", // Will be auto-generated as vault.<env>.ebbo.dev

		// VPC configuration
		VpcCidr: "20.0.0.0/24",
		MaxAzs:  2,

		// ECS configuration
		ClusterName:  environment.GetStackName("vaultwarden-cluster"),
		DesiredCount: 1,
		Cpu:          256, // 0.25 vCPU
		MemoryMiB:    512, // 512 MB RAM

		// EFS configuration
		FileSystemName:            environment.GetStackName("vaultwarden-fs"),
		EnableAutomaticBackups:    true,
		LifecyclePolicyDays:       14,
		OutOfInfrequentAccessHits: 1,
	}

	vaultwardenProps := &vaultwarden.VaultwardenStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
		Environment:  environment,
		Config:       vaultwardenConfig,
		DomainConfig: domainConfig,
	}

	// Create the stacks
	core.NewCoreStack(app, coreStackName, coreProps)
	lambdaStack := lambda.NewLambdaStack(app, lambdaStackName, lambdaProps)
	vaultwarden.NewVaultwardenStack(app, vaultwardenStackName, vaultwardenProps)

	// Add stack outputs for PR environments
	if environment.IsPR {
		awscdk.NewCfnOutput(lambdaStack, jsii.String("EnvironmentType"), &awscdk.CfnOutputProps{
			Value: jsii.String("PR"),
		})
		awscdk.NewCfnOutput(lambdaStack, jsii.String("PRNumber"), &awscdk.CfnOutputProps{
			Value: jsii.String(environment.PRNumber),
		})
	}

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// For development, we'll use environment-agnostic stacks
	// The actual account and region will be provided by the CDK CLI during deployment
	return nil
}
