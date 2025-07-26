package cmd

import (
	"context"
	"fmt"
	"github.com/pete911/ec2/internal/aws"
	"github.com/pete911/ec2/internal/aws/vpc"
	"github.com/pete911/ec2/internal/cmd/flag"
	"github.com/pete911/ec2/internal/cmd/prompt"
	"github.com/pete911/ec2/internal/ec2"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"strings"
	"time"
)

var (
	Root      = &cobra.Command{}
	logLevels = map[string]slog.Level{"debug": slog.LevelDebug, "info": slog.LevelInfo, "warn": slog.LevelWarn, "error": slog.LevelError}
	Version   string
)

func init() {
	flag.InitPersistentFlags(Root)
}

func NewLogger() *slog.Logger {
	if level, ok := logLevels[strings.ToLower(flag.LogLevel)]; ok {
		opts := &slog.HandlerOptions{Level: level}
		return slog.New(slog.NewTextHandler(os.Stderr, opts))
	}

	fmt.Printf("invalid log level %s", flag.LogLevel)
	os.Exit(1)
	return nil
}

func NewClient(logger *slog.Logger) ec2.Client {
	// prompt region if user did not select any and set it on client
	if flag.Region == "" {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		regions, region, err := aws.ListOptedInRegions(ctx, logger)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		i, _ := prompt.Select("region", regions.Names(), aws.RegionByCode(region).Name)
		selectedRegionCode := regions[i].Code
		flag.Region = selectedRegionCode
	}

	awsClient, err := aws.NewClient(logger, flag.Region)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return ec2.NewClient(logger, awsClient)
}

func SelectSubnet(client ec2.Client) vpc.Subnet {
	vpcs, err := client.GetVpcs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	i, _ := prompt.Select("vpc", vpcLabels(vpcs), "")
	selectedVpc := vpcs[i]

	k, _ := prompt.Select("subnet", subnetLabels(selectedVpc.Subnets), "")
	return selectedVpc.Subnets[k]
}

func vpcLabels(in []vpc.Vpc) []string {
	var out []string
	for _, v := range in {
		label := fmt.Sprintf("%s %s", v.CidrBlock, v.Id)
		if v.Name != "" {
			label = fmt.Sprintf("%s %s", label, v.Name)
		}
		if v.HasPublicSubnet() {
			label = fmt.Sprintf("%s [public]", label)
		}
		if v.IsDefault {
			label = fmt.Sprintf("%s [default]", label)
		}
		out = append(out, label)
	}
	return out
}

func subnetLabels(in []vpc.Subnet) []string {
	var out []string
	for _, v := range in {
		label := fmt.Sprintf("%s %s", v.CidrBlock, v.Id)
		if v.Name != "" {
			label = fmt.Sprintf("%s %s", label, v.Name)
		}
		if v.IsPubic() {
			label = fmt.Sprintf("%s [public]", label)
		}
		out = append(out, label)
	}
	return out
}

// SelectInstance either verifies if supplied instance name exists, or prompts user to select instance if argument is empty
func SelectInstance(client ec2.Client, instanceName string) aws.Instance {
	instances, err := client.List()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !strings.HasPrefix(instanceName, ec2.NamePrefix) {
		instanceName = ec2.NamePrefix + instanceName
	}

	// name has not been provided, we only have prefix
	if instanceName != ec2.NamePrefix {
		for _, i := range instances {
			if i.Name == instanceName {
				return i
			}
		}
		fmt.Printf("instance %s not found\n", instanceName)
		os.Exit(1)
	}

	i, _ := prompt.Select("instance", instances.Names(), "")
	return instances[i]
}
