package lambda

import (
	"fmt"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"aws-infra-sandbox/lib"
)

type LambdaStackProps struct {
	awscdk.StackProps
	Environment lib.Environment
	DomainConfig *lib.DomainConfig
}

func NewLambdaStack(scope constructs.Construct, id string, props *LambdaStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Use domain config from props or default
	domainConfig := props.DomainConfig
	if domainConfig == nil {
		domainConfig = lib.DefaultDomainConfig()
	}

	// Reference the hosted zone from CoreStack
	hostedZone := awsroute53.HostedZone_FromHostedZoneAttributes(stack, jsii.String("EbboDevZone"), &awsroute53.HostedZoneAttributes{
		HostedZoneId: jsii.String(domainConfig.HostedZoneId),
		ZoneName:     jsii.String(domainConfig.RootDomain),
	})

	// Create a wildcard certificate for the environment domain
	environmentDomain := domainConfig.GetEnvironmentDomain(props.Environment)
	wildcardDomain := fmt.Sprintf("*.%s", environmentDomain)
	certificate := awscertificatemanager.NewCertificate(stack, jsii.String("ApiCertificate"), &awscertificatemanager.CertificateProps{
		DomainName: jsii.String(wildcardDomain),
		Validation: awscertificatemanager.CertificateValidation_FromDns(hostedZone),
	})

	// Create a single API Gateway for all Lambda functions
	apiName := fmt.Sprintf("%s-api", props.Environment.GetEnvPrefix())
	mainApi := awsapigateway.NewRestApi(stack, jsii.String("MainApi"), &awsapigateway.RestApiProps{
		RestApiName: jsii.String(apiName),
		// Enable CORS
		DefaultCorsPreflightOptions: &awsapigateway.CorsOptions{
			AllowOrigins: awsapigateway.Cors_ALL_ORIGINS(),
			AllowMethods: awsapigateway.Cors_ALL_METHODS(),
			AllowHeaders: jsii.Strings("Content-Type", "X-Amz-Date", "Authorization", "X-Api-Key"),
		},
		// Configure binary media types
		BinaryMediaTypes: jsii.Strings("*/*"),
		// Configure deployment options
		DeployOptions: &awsapigateway.StageOptions{
			StageName:      jsii.String("prod"),
			LoggingLevel:   awsapigateway.MethodLoggingLevel_INFO,
			MetricsEnabled: jsii.Bool(true),
		},
	})

	// Create a Lambda function for the root path
	rootLambdaFn := awslambda.NewFunction(stack, jsii.String("RootLambda"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_NODEJS_18_X(),
		Handler: jsii.String("index.handler"),
		Code: awslambda.Code_FromInline(jsii.String(`
			exports.handler = async function(event) {
				return {
					statusCode: 200,
					headers: {
						"Content-Type": "application/json",
						"Access-Control-Allow-Origin": "*",
						"Access-Control-Allow-Methods": "GET,OPTIONS",
						"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key"
					},
					body: JSON.stringify({ message: "Hello, from ebbo.dev" })
				};
			}
		`)),
	})

	// Add the root Lambda integration to the root path
	rootIntegration := awsapigateway.NewLambdaIntegration(rootLambdaFn, &awsapigateway.LambdaIntegrationOptions{
		Proxy: jsii.Bool(true),
	})

	// Add GET method to the root path
	mainApi.Root().AddMethod(jsii.String("GET"), rootIntegration, nil)

	// Create custom domain name for the API
	apiDomainName := domainConfig.GetAppDomain("api", props.Environment)
	apiDomain := awsapigateway.NewDomainName(stack, jsii.String("api-serverDomain"), &awsapigateway.DomainNameProps{
		DomainName:   jsii.String(apiDomainName),
		Certificate:  certificate,
		EndpointType: awsapigateway.EndpointType_REGIONAL,
	})

	// Map the API to the custom domain
	awsapigateway.NewBasePathMapping(stack, jsii.String("ApiPathMapping"), &awsapigateway.BasePathMappingProps{
		DomainName: apiDomain,
		RestApi:    mainApi,
	})

	// Create Route53 A record for the custom domain
	awsroute53.NewARecord(stack, jsii.String("api-dnsRecord"), &awsroute53.ARecordProps{
		Zone:       hostedZone,
		RecordName: jsii.String(fmt.Sprintf("api.%s", props.Environment.GetEnvPrefix())),
		Target:     awsroute53.RecordTarget_FromAlias(awsroute53targets.NewApiGatewayDomain(apiDomain)),
	})

	// Add custom domain URL as stack output
	awscdk.NewCfnOutput(stack, jsii.String("ApiCustomDomainUrl"), &awscdk.CfnOutputProps{
		Value: jsii.String(fmt.Sprintf("https://%s", apiDomainName)),
	})

	// iterate over all folders in functions and create lambdas
	// read the folders from functions folder, from the filesystem and operating system
	folders, err := readFolders("./functions")
	if err != nil {
		fmt.Println("Error reading folders:", err)
		return nil
	}

	for _, folder := range folders {
		lambdaName := props.Environment.GetStackName(folder) + folder

		// Create the Lambda function
		lambdaFn := awslambda.NewFunction(stack, jsii.String(lambdaName), &awslambda.FunctionProps{
			Code:         awslambda.Code_FromAsset(jsii.String("build/dist/"+folder+".zip"), &awss3assets.AssetOptions{}),
			Timeout:      awscdk.Duration_Seconds(jsii.Number(300)),
			Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
			Architecture: awslambda.Architecture_ARM_64(),
			Handler:      jsii.String("bootstrap"), // Must be "bootstrap" for provided.al2023
		})

		lambdaFn.Role().AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(
			jsii.String("AWSLambda_ReadOnlyAccess"),
		))

		// Create a resource for this Lambda in the main API Gateway
		resource := mainApi.Root().AddResource(jsii.String(folder), nil)

		// Add a proxy resource to handle all paths under this resource
		proxyResource := resource.AddResource(jsii.String("{proxy+}"), nil)

		// Add Lambda integration with proxy configuration
		integration := awsapigateway.NewLambdaIntegration(lambdaFn, &awsapigateway.LambdaIntegrationOptions{
			// Enable proxy integration to pass all request data to Lambda
			Proxy: jsii.Bool(true),
		})

		// Add methods to the resources
		resource.AddMethod(jsii.String("ANY"), integration, nil)

		// For paths under the function (e.g., /gin-server/{proxy+})
		proxyResource.AddMethod(jsii.String("ANY"), integration, nil)

		// Add Lambda URL as stack output
		awscdk.NewCfnOutput(stack, jsii.String(folder+"LambdaEndpoint"), &awscdk.CfnOutputProps{
			Value: jsii.String(fmt.Sprintf("https://%s/%s", apiDomainName, folder)),
		})
	}

	return stack
}

// readFolders reads all folders in the specified path
func readFolders(path string) ([]string, error) {
	var folders []string

	// Read the directory
	entries, err := os.ReadDir(path)
	if err != nil {
		// Return empty slice instead of error if directory doesn't exist
		if os.IsNotExist(err) {
			return folders, nil
		}
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	// Iterate through entries and add folder names
	for _, entry := range entries {
		if entry.IsDir() {
			folders = append(folders, entry.Name())
		}
	}

	return folders, nil
}
