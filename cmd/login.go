package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/TrianaLab/remake/internal/registry"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var loginUsername string
var loginPassword string

var loginCmd = &cobra.Command{
	Use:   "login <oci_endpoint>",
	Short: "Authenticate to an OCI registry and store credentials locally.",
	Long: `Login prompts for username and password (unless supplied via flags) 
and saves them in the local config for the specified registry.
These credentials are used by pull and push commands to authenticate.`,
	Example: ` # Interactive login
  remake login registry.example.com

  # Login using flags
  remake login -u myuser -p mypass registry.example.com`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpoint := args[0]
		if !regexp.MustCompile(`^[A-Za-z0-9\.\-]+(?::[0-9]+)?$`).MatchString(endpoint) {
			return fmt.Errorf("invalid registry %s", endpoint)
		}
		if loginUsername == "" {
			fmt.Print("Username: ")
			u, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return err
			}
			loginUsername = strings.TrimSpace(u)
		}
		if loginPassword == "" {
			fmt.Print("Password: ")
			pw, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return err
			}
			loginPassword = string(pw)
		}
		if err := registry.Login(endpoint, loginUsername, loginPassword); err != nil {
			return fmt.Errorf("Login failed: %w", err)
		}
		fmt.Println("Login successful")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVarP(&loginUsername, "username", "u", "", "Registry username")
	loginCmd.Flags().StringVarP(&loginPassword, "password", "p", "", "Registry password")
}
