package vpc

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type Subnet struct {
	VpcId                   string
	Id                      string
	SubnetArn               string
	Name                    string
	AvailabilityZone        string
	AvailabilityZoneId      string
	AvailableIpAddressCount int
	CidrBlock               string
	DefaultForAz            bool
	RouteTable              RouteTable // added additionally in the client
	State                   string
	Tags                    map[string]string
}

func toSubnet(in types.Subnet) Subnet {
	tags := fromTags(in.Tags)
	return Subnet{
		VpcId:                   aws.ToString(in.VpcId),
		Id:                      aws.ToString(in.SubnetId),
		SubnetArn:               aws.ToString(in.SubnetArn),
		Name:                    tags["Name"],
		AvailabilityZone:        aws.ToString(in.AvailabilityZone),
		AvailabilityZoneId:      aws.ToString(in.AvailabilityZoneId),
		AvailableIpAddressCount: int(aws.ToInt32(in.AvailableIpAddressCount)),
		CidrBlock:               aws.ToString(in.CidrBlock),
		DefaultForAz:            aws.ToBool(in.DefaultForAz),
		State:                   string(in.State),
		Tags:                    tags,
	}
}

func (s Subnet) IsPubic() bool {
	return s.RouteTable.HasPublicRoute()
}
