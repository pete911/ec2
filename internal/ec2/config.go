package ec2

import (
	"fmt"
	"github.com/pete911/ec2/internal/aws"
	"github.com/pete911/ec2/internal/aws/iam"
)

const NamePrefix = "ec2-"

type Config struct {
	meta      aws.MetadataInput
	accountId string
	region    string
}

func NewConfig(name, accountId, region string) Config {
	return Config{
		meta:      GetMetadataInput(name),
		accountId: accountId,
		region:    region,
	}
}

func (c Config) GetInstanceProfileInput() iam.InstanceProfileInput {
	return iam.InstanceProfileInput{
		Name: c.meta.Name,
		Tags: c.meta.Tags,
		Role: iam.RoleInput{
			RoleName:           fmt.Sprintf("%s-%s", c.meta.Name, c.region),
			ManagedPolicyNames: getSSMManagedPolicies(),
			InlinePolicies:     nil,
			Tags:               c.meta.Tags,
		},
	}
}

func GetMetadataInput(name string) aws.MetadataInput {
	name = NamePrefix + name
	return aws.MetadataInput{
		Name: name,
		Tags: map[string]string{
			"Name":       name,
			"Project":    "ec2",
			"Repository": "https://github.com/pete911/ec2",
		},
	}
}

func getSSMManagedPolicies() []string {
	return []string{"AmazonSSMManagedInstanceCore", "AmazonSSMPatchAssociation"}
}
