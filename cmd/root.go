package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/alligator/gdqgo/internal/persist"
	"github.com/alligator/gdqgo/internal/statsfile"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gdqgo",
	Short: "A brief description of your application",
	Long:  ``,
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "print the config file path",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		edit, err := cmd.Flags().GetBool("edit")
		if err != nil {
			return err
		}

		path := persist.GetPath()
		if !edit {
			fmt.Println(path)
			return nil
		}

		editor, ok := os.LookupEnv("EDITOR")
		if !ok {
			return fmt.Errorf("EDITOR environment variable is unset!")
		}

		ecmd := exec.Command(editor, path)
		ecmd.Stdin = os.Stdin
		ecmd.Stderr = os.Stderr
		ecmd.Stdout = os.Stdout
		return ecmd.Run()
	},
}

var testParseCmd = &cobra.Command{
	Use:   "test-parse [file]",
	Short: "test parsing a stats file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sf, err := statsfile.Read(args[0])
		if err != nil {
			return err
		}

		firstgame, err := json.MarshalIndent(sf.Games[0], "  ", "  ")
		if err != nil {
			return err
		}

		firstviewers, err := json.MarshalIndent(sf.Viewers[0], "  ", "  ")
		if err != nil {
			return err
		}

		fmt.Printf("sucessfully parsed %s\n", args[0])
		fmt.Printf("  marathon name: %s\n", sf.MarathonName)
		fmt.Printf("  marathon type: %s\n", sf.MarathonType)
		fmt.Printf("  games: %d\n", len(sf.Games))
		fmt.Printf("  games[0]: %s\n", firstgame)
		fmt.Printf("  viewers: %d\n", len(sf.Viewers))
		fmt.Printf("  viewers[0]: %s\n", firstviewers)

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(configCmd)
	configCmd.Flags().BoolP("edit", "e", false, "edit the config file")

	rootCmd.AddCommand(testParseCmd)
}
