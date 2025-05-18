package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
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
	loginInsecure   bool
	loginUsername   string
	loginPassword   string
	loginInitConfig = config.InitConfig
	newRegistry     = remote.NewRegistry
	saveConfig      = config.SaveConfig
	passwordReader  = term.ReadPassword
	inputReader     = func() *bufio.Reader { return bufio.NewReader(os.Stdin) }
)

var loginCmd = &cobra.Command{
	Use:   "login [oci_endpoint]",
	Short: "Authenticate to an OCI registry and save credentials",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := loginInitConfig(); err != nil {
			return err
		}

		endpoint := viper.GetString("default_registry")
		if len(args) == 1 {
			endpoint = args[0]
		}

		if !regexp.MustCompile(`^[A-Za-z0-9\.\-]+(?::[0-9]+)?$`).MatchString(endpoint) {
			return fmt.Errorf("invalid registry %s", endpoint)
		}

		reader := inputReader()
		if loginUsername == "" {
			fmt.Print("Username: ")
			u, err := reader.ReadString('\n')
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

		reg, err := newRegistry(endpoint)
		if err != nil {
			return fmt.Errorf("invalid registry %s: %w", endpoint, err)
		}
		if loginInsecure {
			reg.PlainHTTP = true
		}
		reg.Client = &auth.Client{
			Credential: auth.StaticCredential(endpoint, auth.Credential{
				Username: loginUsername,
				Password: loginPassword,
			}),
			Cache: auth.NewCache(),
		}

		if err := reg.Ping(context.Background()); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		viper.Set(fmt.Sprintf("registries.%s.username", endpoint), loginUsername)
		viper.Set(fmt.Sprintf("registries.%s.password", endpoint), loginPassword)
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
	loginCmd.Flags().BoolVar(&loginInsecure, "insecure", false, "Allow insecure HTTP registry (plain HTTP)")
}
