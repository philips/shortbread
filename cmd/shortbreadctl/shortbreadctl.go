package main

import "github.com/coreos/cobra"

var (
	shortbreadCtl *cobra.Command
	user          string
	key           string
	baseUrl       string
)

func init() {
	shortbreadCtl = &cobra.Command{
		Use:   "shortbreadctl",
		Short: "A command line tool to interact with the CA server and issue/revoke/modify user and host certificates",
	}
	shortbreadCtl.PersistentFlags().StringVarP(&user, "username", "u", "", "username of the entity to whom the certificate is issued")
	shortbreadCtl.PersistentFlags().StringVarP(&key, "key", "k", "", "bears the path to the public key that will be signed by the CA's private key")
	shortbreadCtl.PersistentFlags().StringVarP(&baseUrl, "server", "s", "", "base url for the CA server")
}

func main() {
	shortbreadCtl.AddCommand(updateUser)
	shortbreadCtl.AddCommand(revokeUser)
	shortbreadCtl.Execute()
}

// bin/shortbreadctl adduser -k /Users/shantanu/.ssh/id_rsa.pub -p /Users/shantanu/.ssh/users_ca -u shantanu -e permit-pty -c USER
