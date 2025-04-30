package vaultwarden

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsefs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"aws-infra-sandbox/lib"
)

type VaultwardenStackProps struct {
	awscdk.StackProps
	Environment  lib.Environment
	Config       *VaultwardenConfig
	DomainConfig *lib.DomainConfig
}

// VaultwardenStack encapsulates all resources needed to run Vaultwarden on AWS
func NewVaultwardenStack(scope constructs.Construct, id string, props *VaultwardenStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Add environment tags
	awscdk.Tags_Of(stack).Add(jsii.String("x:stack"), jsii.String("vaultwarden"), nil)

	// Use configuration from props or environment variables
	config := props.Config
	if config == nil {
		config = DefaultVaultwardenConfig()
	}

	// Use domain config from props or default
	domainConfig := props.DomainConfig
	if domainConfig == nil {
		domainConfig = lib.DefaultDomainConfig()
	}

	// Override with environment variables if set
	baseVersion := os.Getenv("VAULTWARDEN_BASE_VERSION")
	if baseVersion != "" {
		config.BaseVersion = baseVersion
	}

	// If domain name is not set in config but is set in env var, use the env var
	envDomainName := os.Getenv("VAULTWARDEN_DOMAIN_NAME")
	if envDomainName != "" {
		config.DomainName = envDomainName
	}

	// If domain name is not set at all, use the app domain from domain config
	if config.DomainName == "" {
		config.DomainName = domainConfig.GetAppDomain("vault", props.Environment)
	}

	// Reference the hosted zone from CoreStack
	hostedZone := awsroute53.HostedZone_FromHostedZoneAttributes(stack, jsii.String("EbboDevZone"), &awsroute53.HostedZoneAttributes{
		HostedZoneId: jsii.String(domainConfig.HostedZoneId),
		ZoneName:     jsii.String(domainConfig.RootDomain),
	})

	// Create the image repository
	imageRepository := NewImageRepository(stack, "ImageRepository", &ImageRepositoryProps{
		ImageName: config.BaseImageName,
		Version:   config.BaseVersion,
	}, props.Environment)

	// Create the network infrastructure
	network := NewNetwork(stack, "Network", &NetworkProps{
		VpcCidr: config.VpcCidr,
		MaxAzs:  config.MaxAzs,
	})

	// Create the ECS cluster
	cluster := awsecs.NewCluster(stack, jsii.String("VaultwardenCluster"), &awsecs.ClusterProps{
		ClusterName: jsii.String(config.ClusterName),
		Vpc:         network.Vpc,
	})

	// Allow the cluster to access AWS services via VPC endpoints
	network.EcrEndpoint.Connections().AllowDefaultPortFrom(cluster.Connections(), jsii.String("Allow ECR API access"))
	network.EcrRepositoryEndpoint.Connections().AllowDefaultPortFrom(cluster.Connections(), jsii.String("Allow ECR Repository access"))
	network.CloudwatchEndpoint.Connections().AllowDefaultPortFrom(cluster.Connections(), jsii.String("Allow CloudWatch access"))

	// Create an EFS filesystem for persistent storage
	var lifecyclePolicy awsefs.LifecyclePolicy
	if config.LifecyclePolicyDays == 7 {
		lifecyclePolicy = awsefs.LifecyclePolicy_AFTER_7_DAYS
	} else if config.LifecyclePolicyDays == 14 {
		lifecyclePolicy = awsefs.LifecyclePolicy_AFTER_14_DAYS
	} else if config.LifecyclePolicyDays == 30 {
		lifecyclePolicy = awsefs.LifecyclePolicy_AFTER_30_DAYS
	} else if config.LifecyclePolicyDays == 60 {
		lifecyclePolicy = awsefs.LifecyclePolicy_AFTER_60_DAYS
	} else if config.LifecyclePolicyDays == 90 {
		lifecyclePolicy = awsefs.LifecyclePolicy_AFTER_90_DAYS
	} else {
		lifecyclePolicy = awsefs.LifecyclePolicy_AFTER_14_DAYS // Default
	}

	var outOfInfrequentAccessPolicy awsefs.OutOfInfrequentAccessPolicy
	if config.OutOfInfrequentAccessHits == 1 {
		outOfInfrequentAccessPolicy = awsefs.OutOfInfrequentAccessPolicy_AFTER_1_ACCESS
	} else {
		outOfInfrequentAccessPolicy = awsefs.OutOfInfrequentAccessPolicy_AFTER_1_ACCESS // Default
	}

	filesystem := awsefs.NewFileSystem(stack, jsii.String("VaultwardenFS"), &awsefs.FileSystemProps{
		FileSystemName: jsii.String(config.FileSystemName),

		Vpc: network.Vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},

		Encrypted:              jsii.Bool(true),
		PerformanceMode:        awsefs.PerformanceMode_GENERAL_PURPOSE,
		EnableAutomaticBackups: jsii.Bool(config.EnableAutomaticBackups),

		LifecyclePolicy:             lifecyclePolicy,
		OutOfInfrequentAccessPolicy: outOfInfrequentAccessPolicy,
	})

	// Create the Vaultwarden service with domain name
	NewVaultwardenService(stack, "VaultwardenService", &VaultwardenServiceProps{
		Cluster:         cluster,
		ImageRepository: imageRepository.Repository,
		Version:         config.BaseVersion,
		Filesystem:      filesystem,
		DomainName:      jsii.String(config.DomainName),
		DesiredCount:    config.DesiredCount,
		Cpu:             config.Cpu,
		MemoryMiB:       config.MemoryMiB,
		HostedZone:      hostedZone,
	})

	// Output the domain name
	awscdk.NewCfnOutput(stack, jsii.String("VaultwardenDomainName"), &awscdk.CfnOutputProps{
		Description: jsii.String("The domain name for the Vaultwarden service"),
		Value:       jsii.String(config.DomainName),
	})

	return stack
}
