package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/flant/werf/cmd/werf/build"
	"github.com/flant/werf/cmd/werf/build_and_publish"
	"github.com/flant/werf/cmd/werf/cleanup"
	"github.com/flant/werf/cmd/werf/deploy"
	"github.com/flant/werf/cmd/werf/dismiss"
	"github.com/flant/werf/cmd/werf/publish"
	"github.com/flant/werf/cmd/werf/purge"
	"github.com/flant/werf/cmd/werf/run"

	helm_secret_decrypt "github.com/flant/werf/cmd/werf/helm/secret/decrypt"
	helm_secret_encrypt "github.com/flant/werf/cmd/werf/helm/secret/encrypt"
	helm_secret_file_decrypt "github.com/flant/werf/cmd/werf/helm/secret/file/decrypt"
	helm_secret_file_edit "github.com/flant/werf/cmd/werf/helm/secret/file/edit"
	helm_secret_file_encrypt "github.com/flant/werf/cmd/werf/helm/secret/file/encrypt"
	helm_secret_generate_secret_key "github.com/flant/werf/cmd/werf/helm/secret/generate_secret_key"
	helm_secret_rotate_secret_key "github.com/flant/werf/cmd/werf/helm/secret/rotate_secret_key"
	helm_secret_values_decrypt "github.com/flant/werf/cmd/werf/helm/secret/values/decrypt"
	helm_secret_values_edit "github.com/flant/werf/cmd/werf/helm/secret/values/edit"
	helm_secret_values_encrypt "github.com/flant/werf/cmd/werf/helm/secret/values/encrypt"

	"github.com/flant/werf/cmd/werf/ci_env"
	"github.com/flant/werf/cmd/werf/slugify"

	images_cleanup "github.com/flant/werf/cmd/werf/images/cleanup"
	images_publish "github.com/flant/werf/cmd/werf/images/publish"
	images_purge "github.com/flant/werf/cmd/werf/images/purge"

	stages_build "github.com/flant/werf/cmd/werf/stages/build"
	stages_cleanup "github.com/flant/werf/cmd/werf/stages/cleanup"
	stages_purge "github.com/flant/werf/cmd/werf/stages/purge"

	host_cleanup "github.com/flant/werf/cmd/werf/host/cleanup"
	host_purge "github.com/flant/werf/cmd/werf/host/purge"

	helm_deploy_chart "github.com/flant/werf/cmd/werf/helm/deploy_chart"
	helm_generate_chart "github.com/flant/werf/cmd/werf/helm/generate_chart"
	helm_get_service_values "github.com/flant/werf/cmd/werf/helm/get_service_values"
	helm_lint "github.com/flant/werf/cmd/werf/helm/lint"
	helm_render "github.com/flant/werf/cmd/werf/helm/render"

	meta_get_helm_release "github.com/flant/werf/cmd/werf/meta/get_helm_release"
	meta_get_namespace "github.com/flant/werf/cmd/werf/meta/get_namespace"

	"github.com/flant/werf/cmd/werf/completion"
	"github.com/flant/werf/cmd/werf/docs"
	"github.com/flant/werf/cmd/werf/version"

	"github.com/flant/werf/cmd/werf/common"
	"github.com/flant/werf/cmd/werf/common/templates"
	"github.com/flant/werf/pkg/logging"
	"github.com/flant/werf/pkg/process_exterminator"
)

func main() {
	trapTerminationSignals()

	if err := logging.Init(); err != nil {
		common.LogErrorF(fmt.Sprintf("logger initialization error: %s\n", err))

		os.Exit(1)
	}

	if err := process_exterminator.Init(); err != nil {
		common.LogErrorF(fmt.Sprintf("process exterminator initialization error: %s\n", err))

		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:   "werf",
		Short: "Werf helps to implement and support Continuous Integration and Continuous Delivery",
		Long: common.GetLongCommandDescription(`Werf helps to implement and support Continuous Integration and Continuous Delivery.

Find more information at https://werf.io`),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	groups := templates.CommandGroups{
		{
			Message: "Main Commands:",
			Commands: []*cobra.Command{
				build.NewCmd(),
				publish.NewCmd(),
				build_and_publish.NewCmd(),
				run.NewCmd(),
				deploy.NewCmd(),
				dismiss.NewCmd(),
				cleanup.NewCmd(),
				purge.NewCmd(),
			},
		},
		{
			Message: "Toolbox:",
			Commands: []*cobra.Command{
				slugify.NewCmd(),
				ci_env.NewCmd(),
				metaCmd(),
			},
		},
		{
			Message: "Lowlevel Management Commands:",
			Commands: []*cobra.Command{
				stagesCmd(),
				imagesCmd(),
				helmCmd(),
				hostCmd(),
			},
		},
	}
	groups.Add(rootCmd)

	templates.ActsAsRootCommand(rootCmd, groups...)

	rootCmd.AddCommand(
		completion.NewCmd(rootCmd),
		version.NewCmd(),
		docs.NewCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		common.LogErrorF("Error: %s\n", err)

		os.Exit(1)
	}
}

func imagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "images",
		Short: "Work with images",
	}
	cmd.AddCommand(
		images_publish.NewCmd(),
		images_cleanup.NewCmd(),
		images_purge.NewCmd(),
	)

	return cmd
}

func stagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stages",
		Short: "Work with stages, which are cache for images",
	}
	cmd.AddCommand(
		stages_build.NewCmd(),
		stages_cleanup.NewCmd(),
		stages_purge.NewCmd(),
	)

	return cmd
}

func helmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "helm",
		Short: "Manage application deployment with helm",
	}
	cmd.AddCommand(
		helm_get_service_values.NewCmd(),
		helm_generate_chart.NewCmd(),
		helm_deploy_chart.NewCmd(),
		helm_lint.NewCmd(),
		helm_render.NewCmd(),
		secretCmd(),
	)

	return cmd
}

func hostCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "host",
		Short: "Work with werf cache and data of all projects on the host machine",
	}
	cmd.AddCommand(
		host_cleanup.NewCmd(),
		host_purge.NewCmd(),
	)

	return cmd
}

func secretCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Work with secrets",
	}

	fileCmd := &cobra.Command{
		Use:   "file",
		Short: "Work with secret files",
	}

	fileCmd.AddCommand(
		helm_secret_file_encrypt.NewCmd(),
		helm_secret_file_decrypt.NewCmd(),
		helm_secret_file_edit.NewCmd(),
	)

	valuesCmd := &cobra.Command{
		Use:   "values",
		Short: "Work with secret values files",
	}

	valuesCmd.AddCommand(
		helm_secret_values_encrypt.NewCmd(),
		helm_secret_values_decrypt.NewCmd(),
		helm_secret_values_edit.NewCmd(),
	)

	cmd.AddCommand(
		fileCmd,
		valuesCmd,
		helm_secret_generate_secret_key.NewCmd(),
		helm_secret_encrypt.NewCmd(),
		helm_secret_decrypt.NewCmd(),
		helm_secret_rotate_secret_key.NewCmd(),
	)

	return cmd
}

func metaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meta",
		Short: "Work with werf project meta configuration",
	}
	cmd.AddCommand(
		meta_get_helm_release.NewCmd(),
		meta_get_namespace.NewCmd(),
	)

	return cmd
}

func trapTerminationSignals() {
	c := make(chan os.Signal, 1)
	signals := []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT}
	signal.Notify(c, signals...)
	go func() {
		<-c

		common.LogErrorF("interrupted\n")

		os.Exit(17)
	}()
}