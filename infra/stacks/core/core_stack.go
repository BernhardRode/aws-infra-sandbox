package core

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"aws-infra-sandbox/lib"
)

type CoreStackProps struct {
	awscdk.StackProps
	Environment lib.Environment
	DomainConfig *lib.DomainConfig
}

func NewCoreStack(scope constructs.Construct, id string, props *CoreStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Add environment tags to all resources
	awscdk.NewCfnOutput(stack, jsii.String("Username"), &awscdk.CfnOutputProps{
		Value: jsii.String(props.Environment.Username),
	})

	// Use domain config from props or default
	domainConfig := props.DomainConfig
	if domainConfig == nil {
		domainConfig = lib.DefaultDomainConfig()
	}

	// Create a hosted zone for ebbo.dev if it doesn't exist
	// We're using the existing hosted zone with ID from the domain config
	hostedZone := awsroute53.HostedZone_FromHostedZoneAttributes(stack, jsii.String("EbboDevZone"), &awsroute53.HostedZoneAttributes{
		HostedZoneId: jsii.String(domainConfig.HostedZoneId),
		ZoneName:     jsii.String(domainConfig.RootDomain),
	})

	// Output the hosted zone ID
	awscdk.NewCfnOutput(stack, jsii.String("HostedZoneId"), &awscdk.CfnOutputProps{
		Value: hostedZone.HostedZoneId(),
	})

	// Output the environment domain
	environmentDomain := domainConfig.GetEnvironmentDomain(props.Environment)
	awscdk.NewCfnOutput(stack, jsii.String("EnvironmentDomain"), &awscdk.CfnOutputProps{
		Value: jsii.String(environmentDomain),
	})

	return stack
}
