package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/TrianaLab/remake/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

var (
	loginInsecure  bool
	loginUsername  string
	loginPassword  string
	newRegistry    = remote.NewRegistry
	saveConfig     = config.SaveConfig
	passwordReader = term.ReadPassword
	inputReader    = func() *bufio.Reader { return bufio.NewReader(os.Stdin) }
)

// loginCmd authenticates to an OCI registry and saves credentials under a normalized key.
var loginCmd = &cobra.Command{
	Use:   "login <oci_endpoint>",
	Short: "Authenticate to an OCI registry and save credentials",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		endpoint := args[0]
		if !regexp.MustCompile(`^[A-Za-z0-9\.\-]+(?::[0-9]+)?$`).MatchString(endpoint) {
			return fmt.Errorf("invalid registry %s", endpoint)
		}

		// Request credentials if missing
		if loginUsername == "" {
			fmt.Print("Username: ")
			u, err := inputReader().ReadString('\n')
			if err != nil {
				return err
			}
			loginUsername = strings.TrimSpace(u)
		}
		if loginPassword == "" {
			fmt.Print("Password: ")
			pw, err := passwordReader(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return err
			}
			loginPassword = string(pw)
		}

		// Perform login
		repo, err := newRegistry(endpoint)
		if err != nil {
			return fmt.Errorf("invalid registry %s: %w", endpoint, err)
		}
		if loginInsecure {
			repo.PlainHTTP = true
		}
		repo.Client = &auth.Client{
			Credential: auth.StaticCredential(endpoint, auth.Credential{
				Username: loginUsername,
				Password: loginPassword,
			}),
			Cache: auth.NewCache(),
		}
		if err := repo.Ping(context.Background()); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		// Load configuration
		cfg := viper.AllSettings()
		registries, _ := cfg["registries"].(map[string]interface{})
		if registries == nil {
			registries = map[string]interface{}{}
		}

		key := config.NormalizeKey(endpoint)
		registries[key] = map[string]interface{}{
			"username": loginUsername,
			"password": loginPassword,
		}
		cfg["registries"] = registries

		// Save configuration in disk
		configFile := filepath.Join(os.Getenv("HOME"), ".remake", "config.yaml")
		viper.Reset()
		viper.SetConfigFile(configFile)
		if err := viper.MergeConfigMap(cfg); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
		if err := saveConfig(); err != nil {
			return err
		}

		fmt.Printf("Connected to %s successfully\n", endpoint)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVarP(&loginUsername, "username", "u", "", "Registry username")
	loginCmd.Flags().StringVarP(&loginPassword, "password", "p", "", "Registry password")
}
