package cmd

import (
	"io"
	"os"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/cobra"

	"github.com/giantswarm/apptestctl/cmd/bootstrap"
	"github.com/giantswarm/apptestctl/cmd/version"
	"github.com/giantswarm/apptestctl/pkg/project"
)

type Config struct {
	Logger micrologger.Logger
	Stderr io.Writer
	Stdout io.Writer

	Version    string
	BinaryName string
	GitCommit  string
	Source     string
}

func New(config Config) (*cobra.Command, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Stderr == nil {
		config.Stderr = os.Stderr
	}
	if config.Stdout == nil {
		config.Stdout = os.Stdout
	}

	if config.Version == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Version must not be empty", config)
	}
	if config.GitCommit == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GitCommit must not be empty", config)
	}
	if config.Source == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Source must not be empty", config)
	}

	var err error

	var bootstrapCmd *cobra.Command
	{
		c := bootstrap.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,
		}

		bootstrapCmd, err = bootstrap.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionCmd *cobra.Command
	{
		c := version.Config{
			Logger: config.Logger,
			Stderr: config.Stderr,
			Stdout: config.Stdout,

			GitCommit: config.GitCommit,
			Source:    config.Source,
			Version:   config.Version,
		}

		versionCmd, err = version.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	f := &flag{}

	r := &runner{
		flag:   f,
		logger: config.Logger,
		stderr: config.Stderr,
		stdout: config.Stdout,
	}

	c := &cobra.Command{
		Use:          project.Name(),
		Short:        project.Description(),
		Long:         project.Description(),
		RunE:         r.Run,
		SilenceUsage: true,
	}

	f.Init(c)

	c.AddCommand(bootstrapCmd)
	c.AddCommand(versionCmd)

	return c, nil
}
