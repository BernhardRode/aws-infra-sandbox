package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	
	"aws-infra-sandbox/lib"
)

type AwsInfraSandboxStackProps struct {
	awscdk.StackProps
}

func CoreInfraSandboxStack(scope constructs.Construct, id string, props *AwsInfraSandboxStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)
	
	// Get environment information
	app := awscdk.App_Of(scope)
	environment := lib.GetEnvironmentFromContext(app)
	
	// Add core infrastructure resources here
	// These resources will be shared across all environments
	
	return stack
}

func NewAwsInfraSandboxStack(scope constructs.Construct, id string, props *AwsInfraSandboxStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Get environment information
	app := awscdk.App_Of(scope)
	environment := lib.GetEnvironmentFromContext(app)

	// iterate over all folders in functions and create lambdas
	// read the folders from functions folder, from the filesystem and operating system
	folders, err := readFolders("../functions")
	if err != nil {
		fmt.Println("Error reading folders:", err)
		return nil
	}

	for _, folder := range folders {
		// Create a LamdaFn Name from the folder name which will be like foo-bar, just parse the string
		folderName := toPascalCase(folder)
		
		// Create environment-specific resource names
		lambdaName := environment.GetResourceName(folderName + "LambdaFn")
		apiName := environment.GetResourceName(folder + "Endpoint")
		
		// Create the Lambda function
		lambdaFn := awslambda.NewFunction(stack, jsii.String(lambdaName), &awslambda.FunctionProps{
			Code:         awslambda.Code_FromAsset(jsii.String("../build/dist/"+folder+".zip"), &awss3assets.AssetOptions{}),
			Timeout:      awscdk.Duration_Seconds(jsii.Number(300)),
			Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
			Architecture: awslambda.Architecture_ARM_64(),
			Handler:      jsii.String("bootstrap"), // Must be "bootstrap" for provided.al2023
			FunctionName: jsii.String(lambdaName),
		})

		// Create API Gateway with Lambda integration
		api := awsapigateway.NewLambdaRestApi(stack, jsii.String(apiName), &awsapigateway.LambdaRestApiProps{
			Handler: lambdaFn,
			RestApiName: jsii.String(apiName),
		})
		
		// Add API URL as stack output
		awscdk.NewCfnOutput(stack, jsii.String(folder+"ApiUrl"), &awscdk.CfnOutputProps{
			Value: api.Url(),
		})
	}

	return stack
}

func toPascalCase(input string) string {
	// Split the string by hyphen
	parts := strings.Split(input, "-")

	// Capitalize first letter of each part
	for i := 0; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			// Convert part to lowercase first to handle any mixed case
			parts[i] = strings.ToLower(parts[i])
			// Then capitalize first letter
			r := []rune(parts[i])
			r[0] = unicode.ToUpper(r[0])
			parts[i] = string(r)
		}
	}

	// Join all parts together
	return strings.Join(parts, "")
}

func readFolders(path string) ([]string, error) {
	var folders []string

	// Read the directory
	entries, err := os.ReadDir(path)
	if err != nil {
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

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)
	
	// Get environment information from context
	environment := lib.GetEnvironmentFromContext(app)
	
	// Add environment tags to all resources
	awscdk.Tags_Of(app).Add(jsii.String("Environment"), jsii.String(environment.Name), nil)
	if environment.IsPreview {
		awscdk.Tags_Of(app).Add(jsii.String("Preview"), jsii.String("true"), nil)
		awscdk.Tags_Of(app).Add(jsii.String("PR"), jsii.String(environment.PRNumber), nil)
	}
	if environment.Version != "" {
		awscdk.Tags_Of(app).Add(jsii.String("Version"), jsii.String(environment.Version), nil)
	}
	
	// Create stacks with environment-specific names
	coreStackName := environment.GetStackName("CoreInfraSandboxStack")
	appStackName := environment.GetStackName("AwsInfraSandboxStack")
	
	CoreInfraSandboxStack(app, coreStackName, &AwsInfraSandboxStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	stack := NewAwsInfraSandboxStack(app, appStackName, &AwsInfraSandboxStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})
	
	// Add stack outputs for PR environments
	if environment.IsPreview {
		awscdk.NewCfnOutput(stack, jsii.String("EnvironmentType"), &awscdk.CfnOutputProps{
			Value: jsii.String("Preview"),
		})
		awscdk.NewCfnOutput(stack, jsii.String("PRNumber"), &awscdk.CfnOutputProps{
			Value: jsii.String(environment.PRNumber),
		})
	}

	app.Synth(nil)
}

// Keep your existing env() function as is

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
