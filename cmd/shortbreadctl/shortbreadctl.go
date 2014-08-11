package main

import (
	"github.com/coreos/cobra"
	"github.com/coreos/shortbread/cmd/shortbreadctl/command"
)

var (
	shortbreadCtl *cobra.Command
)

func init() {
	shortbreadCtl = &cobra.Command{
		Use:   "shortbreadctl",
		Short: "A command line tool to interact with the CA server and issue/revoke/modify user and host certificates",
	}

}

func main() {
	shortbreadCtl.AddCommand(command.GetAddUser())
	shortbreadCtl.Execute()
}

// bin/shortbreadctl adduser -k /Users/shantanu/.ssh/id_rsa.pub -p /Users/shantanu/.ssh/users_ca -u shantanu -e permit-pty -c USER
