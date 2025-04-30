package core_test

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
	
	"aws-infra-sandbox/lib"
	"aws-infra-sandbox/stacks/core"
)

func TestCoreStackSynthesizes(t *testing.T) {
	// GIVEN
	app := awscdk.NewApp(nil)
	
	// Create a test environment
	env := lib.Environment{
		Name:     "test",
		Username: "tester",
	}
	
	// WHEN
	stack := core.NewCoreStack(app, "TestCoreStack", &core.CoreStackProps{
		StackProps: awscdk.StackProps{
			Env: &awscdk.Environment{
				Account: jsii.String("123456789012"),
				Region:  jsii.String("us-east-1"),
			},
		},
		Environment: env,
	})
	
	// THEN - the stack should synthesize without errors
	if stack == nil {
		t.Fatal("Stack should not be nil")
	}
}
