package vpc

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type Vpc struct {
	Id            string
	Name          string
	CidrBlock     string
	DhcpOptionsId string
	IsDefault     bool
	Subnets       []Subnet // added additionally in the client
	OwnerId       string
	State         string
	Tags          map[string]string
}

func toVpc(in types.Vpc) Vpc {
	tags := fromTags(in.Tags)
	return Vpc{
		Id:            aws.ToString(in.VpcId),
		Name:          tags["Name"],
		CidrBlock:     aws.ToString(in.CidrBlock),
		DhcpOptionsId: aws.ToString(in.DhcpOptionsId),
		IsDefault:     aws.ToBool(in.IsDefault),
		OwnerId:       aws.ToString(in.OwnerId),
		State:         string(in.State),
		Tags:          tags,
	}
}

func (v Vpc) HasPublicSubnet() bool {
	for _, v := range v.Subnets {
		if v.IsPubic() {
			return true
		}
	}
	return false
}
