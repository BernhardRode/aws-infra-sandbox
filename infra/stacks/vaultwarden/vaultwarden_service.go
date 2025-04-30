package vaultwarden

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecspatterns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsefs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type VaultwardenServiceProps struct {
	Cluster        awsecs.Cluster
	ImageRepository awsecr.Repository
	Version        string
	Filesystem     awsefs.FileSystem
	DomainName     *string
	DesiredCount   int
	Cpu            int
	MemoryMiB      int
	HostedZone     awsroute53.IHostedZone
}

// VaultwardenService creates an ECS service to run Vaultwarden containers
type VaultwardenService struct {
	Service awsecspatterns.ApplicationLoadBalancedFargateService
}

func NewVaultwardenService(scope constructs.Construct, id string, props *VaultwardenServiceProps) *VaultwardenService {
	construct := constructs.NewConstruct(scope, &id)
	
	// Create an execution role for the ECS task
	executionRole := awsiam.NewRole(construct, jsii.String("TaskExecRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ecs-tasks.amazonaws.com"), nil),
	})
	
	// Grant the execution role permission to pull from the ECR repository
	props.ImageRepository.GrantPull(executionRole)
	
	// Create a certificate if a domain name is provided
	var certificate awscertificatemanager.Certificate
	if props.DomainName != nil && props.HostedZone != nil {
		certificate = awscertificatemanager.NewCertificate(construct, jsii.String("VaultwardenCertificate"), &awscertificatemanager.CertificateProps{
			DomainName: props.DomainName,
			Validation: awscertificatemanager.CertificateValidation_FromDns(props.HostedZone),
		})
		
		// Output a warning about certificate validation
		awscdk.NewCfnOutput(construct, jsii.String("CertificateValidationWarning"), &awscdk.CfnOutputProps{
			Value: jsii.String("IMPORTANT: You need to validate the SSL certificate by adding DNS records. " +
				"The deployment will wait until validation is complete. Check the ACM console for details."),
		})
	}
	
	// Create a security group for specific egress rules
	securityGroup := awsec2.NewSecurityGroup(construct, jsii.String("VaultwardenSecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc: props.Cluster.Vpc(),
		AllowAllOutbound: jsii.Bool(false),
		Description: jsii.String("Security group for Vaultwarden service"),
	})
	
	// Add specific egress rules
	securityGroup.AddEgressRule(
		awsec2.Peer_AnyIpv4(),
		awsec2.Port_Tcp(jsii.Number(443)),
		jsii.String("Allow HTTPS outbound traffic"),
		nil,
	)
	
	// Use default values if not provided
	desiredCount := 1
	if props.DesiredCount > 0 {
		desiredCount = props.DesiredCount
	}
	
	cpu := 256
	if props.Cpu > 0 {
		cpu = props.Cpu
	}
	
	memoryMiB := 512
	if props.MemoryMiB > 0 {
		memoryMiB = props.MemoryMiB
	}
	
	// Create the Fargate service with an Application Load Balancer
	serviceProps := &awsecspatterns.ApplicationLoadBalancedFargateServiceProps{
		Cluster:        props.Cluster,
		DesiredCount:   jsii.Number(float64(desiredCount)),
		Cpu:            jsii.Number(float64(cpu)),
		MemoryLimitMiB: jsii.Number(float64(memoryMiB)),
		
		TaskImageOptions: &awsecspatterns.ApplicationLoadBalancedTaskImageOptions{
			Image:         awsecs.ContainerImage_FromEcrRepository(props.ImageRepository, jsii.String(props.Version)),
			ExecutionRole: executionRole,
			Environment:   generateVaultwardenEnvironmentVariables(),
		},
		
		TaskSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},
		
		PublicLoadBalancer: jsii.Bool(true),
		Certificate:        certificate,
		
		// Fix the minHealthyPercent warning
		MinHealthyPercent: jsii.Number(100),
		
		// Add the security group
		SecurityGroups: &[]awsec2.ISecurityGroup{securityGroup},
	}
	
	// Create the Fargate service
	service := awsecspatterns.NewApplicationLoadBalancedFargateService(construct, jsii.String("VaultwardenService"), serviceProps)
	
	// Add EFS volume to the task definition
	service.TaskDefinition().AddVolume(&awsecs.Volume{
		Name: jsii.String("efs"),
		EfsVolumeConfiguration: &awsecs.EfsVolumeConfiguration{
			FileSystemId:      props.Filesystem.FileSystemId(),
			TransitEncryption: jsii.String("ENABLED"),
		},
	})
	
	// Mount the EFS volume to the container
	service.TaskDefinition().DefaultContainer().AddMountPoints(&awsecs.MountPoint{
		SourceVolume:  jsii.String("efs"),
		ContainerPath: jsii.String("/data"),
		ReadOnly:      jsii.Bool(false),
	})
	
	// Allow network connectivity between the service and the EFS filesystem
	service.Service().Connections().AllowFrom(props.Filesystem, awsec2.Port_Tcp(jsii.Number(2049)), jsii.String("Allow EFS access from Vaultwarden"))
	service.Service().Connections().AllowTo(props.Filesystem, awsec2.Port_Tcp(jsii.Number(2049)), jsii.String("Allow Vaultwarden to access EFS"))
	
	// Create Route53 A record for the custom domain if domain name is provided
	if props.DomainName != nil && props.HostedZone != nil {
		awsroute53.NewARecord(construct, jsii.String("VaultwardenDnsRecord"), &awsroute53.ARecordProps{
			Zone:       props.HostedZone,
			RecordName: props.DomainName,
			Target:     awsroute53.RecordTarget_FromAlias(awsroute53targets.NewLoadBalancerTarget(service.LoadBalancer(), nil)),
		})
	}
	
	// Output the load balancer DNS name
	awscdk.NewCfnOutput(construct, jsii.String("LoadBalancerDnsName"), &awscdk.CfnOutputProps{
		Description: jsii.String("The DNS name of the load balancer for the Vaultwarden service"),
		Value:       service.LoadBalancer().LoadBalancerDnsName(),
	})
	
	return &VaultwardenService{
		Service: service,
	}
}

// generateVaultwardenEnvironmentVariables creates environment variables for Vaultwarden configuration
func generateVaultwardenEnvironmentVariables() *map[string]*string {
	// In a real implementation, you would read these from environment variables or parameters
	// For now, we'll just return some basic configuration
	return &map[string]*string{
		"WEBSOCKET_ENABLED": jsii.String("true"),
		"LOG_LEVEL":         jsii.String("info"),
	}
}
