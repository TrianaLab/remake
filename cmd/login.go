/*
Copyright © 2025 TrianaLab - Eduardo Diaz <edudiazasencio@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/TrianaLab/remake/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

var (
	loginUsername string
	loginPassword string
)

var loginCmd = &cobra.Command{
	Use:   "login [oci_endpoint]",
	Short: "Autentica en un registro OCI y guarda las credenciales",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.InitConfig(); err != nil {
			return err
		}

		endpoint := viper.GetString("default_registry")
		if len(args) == 1 {
			endpoint = args[0]
		}

		// Prompt interactively if flags are missing
		reader := bufio.NewReader(os.Stdin)
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
			pw, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return err
			}
			loginPassword = string(pw)
		}

		ctx := context.Background()
		reg, err := remote.NewRegistry(endpoint)
		if err != nil {
			return fmt.Errorf("registro inválido %s: %w", endpoint, err)
		}

		reg.Client = &auth.Client{
			Credential: auth.StaticCredential(endpoint, auth.Credential{
				Username: loginUsername,
				Password: loginPassword,
			}),
			Cache: auth.NewCache(),
		}

		if err := reg.Ping(ctx); err != nil {
			return fmt.Errorf("login fallido: %w", err)
		}

		viper.Set(fmt.Sprintf("registries.%s.username", endpoint), loginUsername)
		viper.Set(fmt.Sprintf("registries.%s.password", endpoint), loginPassword)
		if err := config.SaveConfig(); err != nil {
			return err
		}

		fmt.Printf("✅ Conectado a %s exitosamente\n", endpoint)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVarP(&loginUsername, "username", "u", "", "Usuario del registro")
	loginCmd.Flags().StringVarP(&loginPassword, "password", "p", "", "Contraseña del registro")

	viper.BindPFlag("username", loginCmd.Flags().Lookup("username"))
	viper.BindPFlag("password", loginCmd.Flags().Lookup("password"))
}
