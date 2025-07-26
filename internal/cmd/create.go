package cmd

import (
	"fmt"
	"github.com/pete911/ec2/internal/cmd/prompt"
	"github.com/spf13/cobra"
	"os"
)

var (
	createCmd = &cobra.Command{
		Use:   "create <name>",
		Short: "create EC2 instance",
		Long:  "",
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run:   runCreate,
	}
)

func init() {
	Root.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) {
	name := args[0]
	logger := NewLogger()
	client := NewClient(logger)
	subnet := SelectSubnet(client)
	if !prompt.Prompt(fmt.Sprintf("create %s EC2 instance in %s region %s - %q subnet", name, client.Region, subnet.Id, subnet.Name)) {
		return
	}

	instance, err := client.Create(name, subnet)
	if err != nil {
		fmt.Printf("create %s EC2: %v\n", name, err)
		os.Exit(1)
	}

	fmt.Printf("EC2 instance %s created\n", instance.Id)
	fmt.Printf("    public dns %s\n", instance.PublicDnsName)
	fmt.Printf("    public IP  %s\n", instance.PublicIp)
}
