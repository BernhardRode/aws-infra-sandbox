package vaultwarden

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type NetworkProps struct {
	VpcCidr string
	MaxAzs  int
}

// Network creates a dedicated VPC network for use with the Vaultwarden application cluster
type Network struct {
	Vpc                 awsec2.Vpc
	EcrEndpoint         awsec2.InterfaceVpcEndpoint
	EcrRepositoryEndpoint awsec2.InterfaceVpcEndpoint
	CloudwatchEndpoint  awsec2.InterfaceVpcEndpoint
	S3Endpoint          awsec2.GatewayVpcEndpoint
}

func NewNetwork(scope constructs.Construct, id string, props *NetworkProps) *Network {
	construct := constructs.NewConstruct(scope, &id)
	
	// Use default values if not provided
	vpcCidr := "20.0.0.0/24"
	if props != nil && props.VpcCidr != "" {
		vpcCidr = props.VpcCidr
	}
	
	maxAzs := 2
	if props != nil && props.MaxAzs > 0 {
		maxAzs = props.MaxAzs
	}
	
	// Create a VPC for the Vaultwarden application
	vpc := awsec2.NewVpc(construct, jsii.String("VaultwardenVpc"), &awsec2.VpcProps{
		// Use ipAddresses instead of cidr (which is deprecated)
		IpAddresses: awsec2.IpAddresses_Cidr(jsii.String(vpcCidr)),
		MaxAzs:  jsii.Number(float64(maxAzs)),
		
		// Configure subnet groups for isolated workloads and public ingress
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("isolated"),
				SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
				CidrMask:   jsii.Number(26),
			},
			{
				Name:       jsii.String("ingress"),
				SubnetType: awsec2.SubnetType_PUBLIC,
				CidrMask:   jsii.Number(26),
			},
		},
	})
	
	// Create VPC endpoints for private connectivity to AWS services
	// ECR API endpoint for authentication and authorization
	ecrEndpoint := vpc.AddInterfaceEndpoint(jsii.String("EcrEndpoint"), &awsec2.InterfaceVpcEndpointOptions{
		Service:           awsec2.InterfaceVpcEndpointAwsService_ECR(),
		PrivateDnsEnabled: jsii.Bool(true),
		Subnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},
	})
	
	// ECR Repository endpoint for image pulls
	ecrRepositoryEndpoint := vpc.AddInterfaceEndpoint(jsii.String("EcrRepositoryEndpoint"), &awsec2.InterfaceVpcEndpointOptions{
		Service:           awsec2.InterfaceVpcEndpointAwsService_ECR_DOCKER(),
		PrivateDnsEnabled: jsii.Bool(true),
		Subnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},
	})
	
	// S3 Gateway endpoint for ECR storage access
	s3Endpoint := vpc.AddGatewayEndpoint(jsii.String("S3Endpoint"), &awsec2.GatewayVpcEndpointOptions{
		Service: awsec2.GatewayVpcEndpointAwsService_S3(),
		Subnets: &[]*awsec2.SubnetSelection{
			{
				SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
			},
		},
	})
	
	// CloudWatch Logs endpoint for container logging
	cloudwatchEndpoint := vpc.AddInterfaceEndpoint(jsii.String("CloudwatchEndpoint"), &awsec2.InterfaceVpcEndpointOptions{
		Service:           awsec2.InterfaceVpcEndpointAwsService_CLOUDWATCH_LOGS(),
		PrivateDnsEnabled: jsii.Bool(true),
		Subnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},
	})
	
	return &Network{
		Vpc:                 vpc,
		EcrEndpoint:         ecrEndpoint,
		EcrRepositoryEndpoint: ecrRepositoryEndpoint,
		CloudwatchEndpoint:  cloudwatchEndpoint,
		S3Endpoint:          s3Endpoint,
	}
}
