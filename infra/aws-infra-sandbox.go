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

type StackProps struct {
	awscdk.StackProps
	Environment lib.Environment
}

func CoreStack(scope constructs.Construct, id string, props *StackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Add environment tags to all resources
	awscdk.NewCfnOutput(stack, jsii.String("Username"), &awscdk.CfnOutputProps{
		Value: jsii.String(props.Environment.Username),
	})

	// Safely retrieve and output the domain name servers
	// nameServers := zone.HostedZoneNameServers()
	// if nameServers != nil {
	// 	for i, ns := range *nameServers {
	// 		awscdk.NewCfnOutput(stack, jsii.String(fmt.Sprintf("ns-%d", i+1)), &awscdk.CfnOutputProps{
	// 			Value: ns,
	// 		})
	// 	}
	// } else {
	// 	fmt.Println("Warning: No name servers available for the hosted zone")
	// }

	return stack
}

func LambdaStack(scope constructs.Construct, id string, props *StackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// iterate over all folders in functions and create lambdas
	// read the folders from functions folder, from the filesystem and operating system
	folders, err := readFolders("./functions")
	if err != nil {
		fmt.Println("Error reading folders:", err)
		return nil
	}

	for _, folder := range folders {
		// Create a LamdaFn Name from the folder name which will be like foo-bar, just parse the string
		folderName := toPascalCase(folder)
		lambdaName := props.Environment.GetStackName("Fn")
		apiName := props.Environment.GetStackName(folder)

		// Create the Lambda function
		lambdaFn := awslambda.NewFunction(stack, jsii.String(folderName+"LambdaFn"), &awslambda.FunctionProps{
			Code:         awslambda.Code_FromAsset(jsii.String("build/dist/"+folder+".zip"), &awss3assets.AssetOptions{}),
			Timeout:      awscdk.Duration_Seconds(jsii.Number(300)),
			Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
			Architecture: awslambda.Architecture_ARM_64(),
			Handler:      jsii.String("bootstrap"), // Must be "bootstrap" for provided.al2023
			FunctionName: jsii.String(lambdaName),
		})

		// Create API Gateway with Lambda integration
		api := awsapigateway.NewLambdaRestApi(stack, jsii.String(folder+"Endpoint"), &awsapigateway.LambdaRestApiProps{
			Handler:     lambdaFn,
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

	// Create props with environment information
	props := &StackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
		Environment: environment,
	}

	CoreStack(app, coreStackName, props)
	stack := LambdaStack(app, lambdaStackName, props)

	// Add stack outputs for PR environments
	if environment.IsPR {
		awscdk.NewCfnOutput(stack, jsii.String("EnvironmentType"), &awscdk.CfnOutputProps{
			Value: jsii.String("PR"),
		})
		awscdk.NewCfnOutput(stack, jsii.String("PRNumber"), &awscdk.CfnOutputProps{
			Value: jsii.String(environment.PRNumber),
		})
	}

	app.Synth(nil)
}

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
