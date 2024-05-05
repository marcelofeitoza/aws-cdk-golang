package main

import (
	"log"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsautoscaling"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type AwsCdkTestStackProps struct {
	awscdk.StackProps
}

func NewAwsCdkTestStack(scope constructs.Construct, id string, props *AwsCdkTestStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Create a VPC
	vpc := awsec2.NewVpc(stack, jsii.String("MyVPC"), &awsec2.VpcProps{
		Cidr:   jsii.String("10.1.0.0/16"),
		MaxAzs: jsii.Number(2),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				SubnetType: awsec2.SubnetType_PUBLIC,
				Name:       jsii.String("Public"),
				CidrMask:   jsii.Number(24),
			},
			{
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
				Name:       jsii.String("Private"),
				CidrMask:   jsii.Number(24),
			},
		},
	})

	// Create an RDS Database
	awsrds.NewDatabaseInstance(stack, jsii.String("MyDatabase"), &awsrds.DatabaseInstanceProps{
		Engine: awsrds.DatabaseInstanceEngine_Postgres(&awsrds.PostgresInstanceEngineProps{
			Version: awsrds.PostgresEngineVersion_VER_12_15(),
		}),
		InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_BURSTABLE3, awsec2.InstanceSize_MICRO),
		Vpc:          vpc,
	})

	// Create EC2 Instance
	_ = awsautoscaling.NewAutoScalingGroup(stack, jsii.String("ASG"), &awsautoscaling.AutoScalingGroupProps{
		Vpc:          vpc,
		InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_BURSTABLE2, awsec2.InstanceSize_MICRO),
		MachineImage: awsec2.NewAmazonLinuxImage(nil),
		MinCapacity:  jsii.Number(1),
		MaxCapacity:  jsii.Number(2),
	})

	// Create an Elastic Load Balancer
	lb := awselasticloadbalancingv2.NewApplicationLoadBalancer(stack, jsii.String("LB"), &awselasticloadbalancingv2.ApplicationLoadBalancerProps{
		Vpc:            vpc,
		InternetFacing: jsii.Bool(true),
	})

	// Add listeners and targets
	listener := lb.AddListener(jsii.String("Listener"), &awselasticloadbalancingv2.BaseApplicationListenerProps{
		Port: jsii.Number(8080),
	})

	listener.AddTargets(jsii.String("Target"), &awselasticloadbalancingv2.AddApplicationTargetsProps{
		Port: jsii.Number(8080),
		Targets: &[]awselasticloadbalancingv2.IApplicationLoadBalancerTarget{
			awsautoscaling.NewAutoScalingGroup(stack, jsii.String("ASGTargetGroup"), &awsautoscaling.AutoScalingGroupProps{
				Vpc:          vpc,
				InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_BURSTABLE2, awsec2.InstanceSize_MICRO),
				MachineImage: awsec2.NewAmazonLinuxImage(nil),
				MinCapacity:  jsii.Number(1),
				MaxCapacity:  jsii.Number(2),
			}),
		},
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewAwsCdkTestStack(app, "AwsCdkTestStack", &AwsCdkTestStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	account := os.Getenv("CDK_DEFAULT_ACCOUNT")
	region := os.Getenv("CDK_DEFAULT_REGION")

	if account == "" || region == "" {
		log.Fatal("Both CDK_DEFAULT_ACCOUNT and CDK_DEFAULT_REGION must be set")
	}

	return &awscdk.Environment{
		Account: jsii.String(account),
		Region:  jsii.String(region),
	}
}
