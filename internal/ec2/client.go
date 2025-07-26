package ec2

import (
	"context"
	"fmt"
	"github.com/pete911/ec2/internal/aws"
	"github.com/pete911/ec2/internal/aws/vpc"
	"log/slog"
	"time"
)

type Client struct {
	Region    string
	logger    *slog.Logger
	awsClient aws.Client
}

func NewClient(logger *slog.Logger, awsClient aws.Client) Client {
	return Client{
		Region:    awsClient.Region,
		logger:    logger.With("component", "ec2.client"),
		awsClient: awsClient,
	}
}

func (c Client) GetVpcs() ([]vpc.Vpc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*300)
	defer cancel()

	return c.awsClient.GetVpcs(ctx)
}

func (c Client) Delete(instance aws.Instance) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*300)
	defer cancel()

	return c.awsClient.TerminateInstance(ctx, instance)
}

func (c Client) List() (aws.Instances, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// we don't care about name in the tags (it will be stripped anyway), so providing just empty string to get tags
	return c.awsClient.DescribeInstancesByNamePrefix(ctx, NamePrefix, GetMetadataInput("").Tags)
}

func (c Client) Create(name string, subnet vpc.Subnet) (aws.Instance, error) {
	// TODO - add option to supply custom user data
	instance, err := c.runInstance(name, subnet, "")
	if err != nil {
		return aws.Instance{}, err
	}
	c.logger.Info(fmt.Sprintf("starting instance %s in subnet %s AZ %s", instance.Id, subnet.Id, subnet.AvailabilityZone))
	c.logger.Info("waiting 60 seconds for instance to initialize")
	time.Sleep(60 * time.Second)

	// wait for instance to start
	for x := 0; x < 30; x++ {
		time.Sleep(15 * time.Second)
		status, err := c.describeInstanceStatus(instance.Id)
		if err != nil {
			return aws.Instance{}, err
		}

		c.logger.Info(fmt.Sprintf("instance %s - %s", instance.Id, status))
		if status.IsReady() {
			// get fresh initialized instance with public IP and dns set
			return c.describeInstanceById(instance.Id)
		}
		c.logger.Info("retry in 15 seconds")
	}
	return aws.Instance{}, fmt.Errorf("instance %s not ready", instance.Id)
}

func (c Client) describeInstanceStatus(id string) (aws.InstanceStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return c.awsClient.DescribeInstanceStatus(ctx, id)
}

func (c Client) describeInstanceById(id string) (aws.Instance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return c.awsClient.DescribeInstanceById(ctx, id)
}

func (c Client) runInstance(name string, subnet vpc.Subnet, userData string) (aws.Instance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	config := NewConfig(name, c.awsClient.AccountId, c.awsClient.Region)
	input := aws.RunInstancesInput{
		Metadata:        config.meta,
		Subnet:          subnet,
		UserData:        userData,
		InstanceProfile: config.GetInstanceProfileInput(),
	}
	return c.awsClient.RunInstance(ctx, input)
}
