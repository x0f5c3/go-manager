package cmd

import (
	"os"
	"os/signal"

	"github.com/pterm/pcli"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/x0f5c3/go-manager/internal/fsutil"
	"github.com/x0f5c3/go-manager/pkg"
)

var rootCmd = &cobra.Command{
	Use:     "go-manager",
	Short:   "This tool will download the latest version of Go",
	Args:    cobra.ExactArgs(1),
	Version: "v0.0.4", // <---VERSION---> Updating this version, will also create a new GitHub release.
	// Uncomment the following lines if your bare application has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		dlSettings.OutDir = args[0]
		return pkg.DownloadLatest(&dlSettings)
	},
}

var dlSettings = pkg.DownloadSettings{}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Fetch user interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		pterm.Warning.Println("user interrupt")
		checkUpdate()
		os.Exit(0)
	}()

	// Execute cobra
	if err := rootCmd.Execute(); err != nil {
		checkUpdate()
		os.Exit(1)
	}
	checkUpdate()
}

func checkUpdate() {
	// Check for updates
	err := pcli.CheckForUpdates()
	if err != nil {
		pterm.Fatal.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Adds global flags for PTerm settings.
	// Fill the empty strings with the shorthand variant (if you like to have one).
	rootCmd.PersistentFlags().BoolVarP(&pterm.PrintDebugMessages, "debug", "d", false, "enable debug messages")
	rootCmd.PersistentFlags().BoolVar(&pterm.RawOutput, "raw", false, "print unstyled raw output (set it if output is written to a file)")
	rootCmd.PersistentFlags().BoolVar(&pcli.DisableUpdateChecking, "disable-update-checks", false, "disables update checks")
	rootCmd.Flags().StringVarP(&dlSettings.Arch, "arch", "a", pkg.CurrentKind.Arch, "architecture")
	rootCmd.Flags().StringVarP(&dlSettings.Os, "os", "o", pkg.CurrentKind.Os, "operating system")
	rootCmd.Flags().StringVarP(&dlSettings.Kind, "kind", "k", pkg.CurrentKind.Kind, "kind")
	cobra.OnInitialize(mustInitConfig)

	// Use https://github.com/pterm/pcli to style the output of cobra.
	err := pcli.SetRepo("x0f5c3/go-manager")
	if err != nil {
		pterm.Fatal.Println(err)
		os.Exit(1)
	}
	pcli.SetRootCmd(rootCmd)
	pcli.Setup()

	// Change global PTerm theme
	pterm.ThemeDefault.SectionStyle = *pterm.NewStyle(pterm.FgCyan)
	rootCmd.AddCommand(fsutil.InstallCmd)
}
