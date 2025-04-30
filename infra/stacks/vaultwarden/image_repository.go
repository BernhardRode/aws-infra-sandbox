package vaultwarden

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type ImageRepositoryProps struct {
	ImageName string
	Version   string
}

// ImageRepository copies the official Vaultwarden container images from Docker Hub
// to a dedicated private registry in the AWS account
type ImageRepository struct {
	Repository awsecr.Repository
}

func NewImageRepository(scope constructs.Construct, id string, props *ImageRepositoryProps) *ImageRepository {
	construct := constructs.NewConstruct(scope, &id)

	environment := GetEnvironmentFromContext(scope)

	// Create an ECR repository to store the Vaultwarden image
	repoName := environment.GetStackName("VaultwardenImageRepository")
	repository := awsecr.NewRepository(construct, jsii.String(repoName), &awsecr.RepositoryProps{
		RepositoryName: jsii.String(props.ImageName),
	})

	// Note: In Go CDK, we don't have a direct equivalent to cdk-ecr-deployment
	// We would need to use a custom resource or Lambda to pull and push the image
	// For now, we'll add a comment explaining how to manually copy the image

	// Add a CloudFormation output with instructions
	// accountId := jsii.String(*awscdk.NewStack(scope, &id, nil).Account())
	// region := jsii.String(*awscdk.NewStack(scope, &id, nil).Region())
	// copyImageToECR(props.Version, props.ImageName, *accountId, *region)

	awscdk.NewCfnOutput(construct, jsii.String("ImagePullInstructions"), &awscdk.CfnOutputProps{
		Value: jsii.String("To copy the Vaultwarden image to this repository, run: " +
			"docker pull vaultwarden/server:" + props.Version + " && " +
			"docker tag vaultwarden/server:" + props.Version + " " +
			"${AWS::AccountId}.dkr.ecr.${AWS::Region}.amazonaws.com/" + props.ImageName + ":" + props.Version + " && " +
			"aws ecr get-login-password | docker login --username AWS --password-stdin ${AWS::AccountId}.dkr.ecr.${AWS::Region}.amazonaws.com && " +
			"docker push ${AWS::AccountId}.dkr.ecr.${AWS::Region}.amazonaws.com/" + props.ImageName + ":" + props.Version),
	})
	// Can we wait here and check if the image is copied?
	// Before we can use the image, we need to ensure it's copied
	// Check with the AWS SDK if the image exists in the ECR repository

	return &ImageRepository{
		Repository: repository,
	}
}

// func copyImageToECR(imageName, version, accountId, region string) error {
// 	ecrRepo := accountId + ".dkr.ecr." + region + ".amazonaws.com/" + imageName + ":" + version

// 	commands := [][]string{
// 		{"docker", "pull", "vaultwarden/server:" + version},
// 		{"docker", "tag", "vaultwarden/server:" + version, ecrRepo},
// 		{"aws", "ecr", "get-login-password", "--region", region},
// 		{"docker", "login", "--username", "AWS", "--password-stdin", accountId + ".dkr.ecr." + region + ".amazonaws.com"},
// 		{"docker", "push", ecrRepo},
// 	}

// 	for _, cmdArgs := range commands {
// 		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
// 		cmd.Stdout = log.Writer()
// 		cmd.Stderr = log.Writer()
// 		if cmdArgs[0] == "docker" && cmdArgs[1] == "login" {
// 			// Pipe password from previous aws ecr get-login-password
// 			// You'd need to handle this step separately
// 			continue
// 		}
// 		if err := cmd.Run(); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
