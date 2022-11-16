package cmd

import (
	"os"

	"github.com/pterm/pterm"
)

func checkErr(err error) {
	if err != nil {
		pterm.Fatal.Println(err)
		os.Exit(1)
	}
}
