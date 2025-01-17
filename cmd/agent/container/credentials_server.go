package container

import (
	"context"
	"fmt"
	"os"

	"github.com/loft-sh/devpod/cmd/flags"
	"github.com/loft-sh/devpod/pkg/agent"
	"github.com/loft-sh/devpod/pkg/agent/tunnel"
	"github.com/loft-sh/devpod/pkg/credentials"
	"github.com/loft-sh/devpod/pkg/log"
	"github.com/loft-sh/devpod/pkg/port"
	"github.com/loft-sh/devpod/pkg/random"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// CredentialsServerCmd holds the cmd flags
type CredentialsServerCmd struct {
	*flags.GlobalFlags

	User string

	ConfigureGitHelper    bool
	ConfigureDockerHelper bool
}

// NewCredentialsServerCmd creates a new command
func NewCredentialsServerCmd(flags *flags.GlobalFlags) *cobra.Command {
	cmd := &CredentialsServerCmd{
		GlobalFlags: flags,
	}
	credentialsServerCmd := &cobra.Command{
		Use:   "credentials-server",
		Short: "Starts a credentials server",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			return cmd.Run(context.Background(), args)
		},
	}
	credentialsServerCmd.Flags().BoolVar(&cmd.ConfigureGitHelper, "configure-git-helper", false, "If true will configure git helper")
	credentialsServerCmd.Flags().BoolVar(&cmd.ConfigureDockerHelper, "configure-docker-helper", false, "If true will configure docker helper")
	credentialsServerCmd.Flags().StringVar(&cmd.User, "user", "", "The user to use")
	_ = credentialsServerCmd.MarkFlagRequired("user")
	return credentialsServerCmd
}

// Run runs the command logic
func (cmd *CredentialsServerCmd) Run(ctx context.Context, _ []string) error {
	log := log.NewFileLogger("/tmp/devpod-credentials-server.log", logrus.InfoLevel)
	if cmd.Debug {
		log.SetLevel(logrus.DebugLevel)
	}
	log.Infof("Start credentials server")

	// create a grpc client
	tunnelClient, err := agent.NewTunnelClient(os.Stdin, os.Stdout, true)
	if err != nil {
		return fmt.Errorf("error creating tunnel client: %w", err)
	}

	// this message serves as a ping to the client
	_, err = tunnelClient.Ping(ctx, &tunnel.Empty{})
	if err != nil {
		return errors.Wrap(err, "ping client")
	}

	// find available port
	port, err := port.FindAvailablePort(random.InRange(12000, 18000))
	if err != nil {
		return errors.Wrap(err, "find port")
	}

	// run the credentials server
	return credentials.RunCredentialsServer(ctx, cmd.User, port, cmd.ConfigureGitHelper, cmd.ConfigureDockerHelper, tunnelClient, log)
}
