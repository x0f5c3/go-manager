package cmd

import (
	"os"
	"os/signal"

	"github.com/pterm/pcli"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "go-manager",
	Short:   "This tool will download the latest version of Go",
	Version: "v0.0.1", // <---VERSION---> Updating this version, will also create a new GitHub release.
	// Uncomment the following lines if your bare application has an action associated with it:
	// RunE: func(cmd *cobra.Command, args []string) error {
	// 	// Your code here
	//
	// 	return nil
	// },
}

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
	rootCmd.PersistentFlags().BoolVarP(&pterm.PrintDebugMessages, "debug", "", false, "enable debug messages")
	rootCmd.PersistentFlags().BoolVarP(&pterm.RawOutput, "raw", "", false, "print unstyled raw output (set it if output is written to a file)")
	rootCmd.PersistentFlags().BoolVarP(&pcli.DisableUpdateChecking, "disable-update-checks", "", false, "disables update checks")

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
}
