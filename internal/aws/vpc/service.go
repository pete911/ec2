package vpc

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/pete911/ec2/internal/errs"
	"log/slog"
	"sort"
)

type Service struct {
	logger *slog.Logger
	svc    *ec2.Client
}

func NewService(logger *slog.Logger, cfg aws.Config) Service {
	return Service{
		logger: logger.With("component", "aws.vpc.service"),
		svc:    ec2.NewFromConfig(cfg),
	}
}

// GetVpcs returns VPCs containing subnets and route tables
func (s Service) GetVpcs(ctx context.Context) ([]Vpc, error) {
	vpcs, err := s.describeVpcs(ctx)
	if err != nil {
		return nil, err
	}
	routeTables, err := s.describeRouteTables(ctx)
	if err != nil {
		return nil, err
	}
	subnets, err := s.describeSubnets(ctx)
	if err != nil {
		return nil, err
	}

	// set route tables on subnets and sort them by VPC id
	subnetsByVpcId := make(map[string][]Subnet)
	for _, subnet := range subnets {
		rtb := getRouteTableBySubnet(routeTables, subnet)
		subnet.RouteTable = rtb
		subnetsByVpcId[subnet.VpcId] = append(subnetsByVpcId[subnet.VpcId], subnet)
	}

	// set subnets on VPCs
	for i := range vpcs {
		vpcs[i].Subnets = subnetsByVpcId[vpcs[i].Id]
	}
	// sort vpcs
	sort.Slice(vpcs, func(i, j int) bool {
		return vpcs[i].Id < vpcs[j].Id
	})
	s.logger.DebugContext(ctx, fmt.Sprintf("found %d vpcs", len(vpcs)))
	return vpcs, nil
}

// getRouteTableBySubnet rtbs key has to be in "<vpc-id>" (main route table) or "<vpc-id><subnet-id>" format
func getRouteTableBySubnet(rtbs map[string]RouteTable, subnet Subnet) RouteTable {
	if v, ok := rtbs[subnet.VpcId+subnet.Id]; ok {
		return v
	}
	// using default vpc route table
	if v, ok := rtbs[subnet.VpcId]; ok {
		return v
	}
	return RouteTable{}
}

func (s Service) describeSubnets(ctx context.Context) ([]Subnet, error) {
	in := &ec2.DescribeSubnetsInput{
		Filters: []ec2types.Filter{
			{Name: aws.String("state"), Values: []string{"available"}},
		},
	}

	var subnets []Subnet
	for {
		out, err := s.svc.DescribeSubnets(ctx, in)
		if err != nil {
			return nil, errs.FromAwsApi(err, "ec2 describe-subnets")
		}
		for _, v := range out.Subnets {
			subnet := toSubnet(v)
			subnets = append(subnets, subnet)
		}

		if aws.ToString(out.NextToken) == "" {
			break
		}
		in.NextToken = out.NextToken
	}
	s.logger.DebugContext(ctx, fmt.Sprintf("found %d subnets", len(subnets)))
	return subnets, nil
}

// describeRouteTables returns map of route tables where key is either <vpc-id> (main route table) or <vpc-id><subnet-id>
func (s Service) describeRouteTables(ctx context.Context) (map[string]RouteTable, error) {
	in := &ec2.DescribeRouteTablesInput{}

	var routeTablesCount int
	routeTables := make(map[string]RouteTable)
	for {
		out, err := s.svc.DescribeRouteTables(ctx, in)
		if err != nil {
			return nil, errs.FromAwsApi(err, "ec2 describe-route-tables")
		}
		for _, v := range out.RouteTables {
			routeTablesCount++
			routeTable := toRouteTable(v)
			if routeTable.Main {
				routeTables[routeTable.VpcId] = routeTable
			}
			for _, subnetId := range routeTable.SubnetIds {
				routeTables[routeTable.VpcId+subnetId] = routeTable
			}
		}

		if aws.ToString(out.NextToken) == "" {
			break
		}
		in.NextToken = out.NextToken
	}
	s.logger.DebugContext(ctx, fmt.Sprintf("found %d route tables", routeTablesCount))
	return routeTables, nil
}

// describeVpcs returns list of VPCs. Vpc does NOT have 'subnets' field set yet
func (s Service) describeVpcs(ctx context.Context) ([]Vpc, error) {
	in := &ec2.DescribeVpcsInput{
		Filters: []ec2types.Filter{
			{Name: aws.String("state"), Values: []string{"available"}},
		},
	}

	var vpcs []Vpc
	for {
		out, err := s.svc.DescribeVpcs(ctx, in)
		if err != nil {
			return nil, errs.FromAwsApi(err, "ec2 describe-vpcs")
		}
		for _, v := range out.Vpcs {
			vpcs = append(vpcs, toVpc(v))
		}

		if aws.ToString(out.NextToken) == "" {
			break
		}
		in.NextToken = out.NextToken
	}
	s.logger.DebugContext(ctx, fmt.Sprintf("found %d vpcs", len(vpcs)))
	return vpcs, nil
}
