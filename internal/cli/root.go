package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const KEYCHAIN_PATH = ".colonies"

var Verbose bool
var ServerHost string
var ServerPort int
var KubeColonyID string
var KubeColonyPrvKey string
var TargetColonyID string
var TargetColonyPrvKey string

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}

var rootCmd = &cobra.Command{
	Use:   "kolony",
	Short: "Kolony CLI tool",
	Long:  "Kolony CLI tool",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func CheckError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}
