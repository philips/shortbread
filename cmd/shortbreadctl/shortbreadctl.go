package main

<<<<<<< HEAD
import "github.com/coreos/cobra"
=======
import (
	"fmt"
	"os"

	"github.com/coreos/cobra"
	"github.com/coreos/shortbread/client"
)
>>>>>>> 58afb88... Corrected formatting errors from PR, added revoke and list sub-commands,using the usernames as keys, one command to add and modify an user.

var (
	shortbreadCtl *cobra.Command
	user          string
	key           string
	baseUrl       string
<<<<<<< HEAD
=======
	list          bool
>>>>>>> 58afb88... Corrected formatting errors from PR, added revoke and list sub-commands,using the usernames as keys, one command to add and modify an user.
)

func init() {
	shortbreadCtl = &cobra.Command{
		Use:   "shortbreadctl",
		Short: "A command line tool to interact with the CA server and issue/revoke/modify user and host certificates",
<<<<<<< HEAD
=======
		Run:   getUsers,
>>>>>>> 58afb88... Corrected formatting errors from PR, added revoke and list sub-commands,using the usernames as keys, one command to add and modify an user.
	}
	shortbreadCtl.PersistentFlags().StringVarP(&user, "username", "u", "", "username of the entity to whom the certificate is issued")
	shortbreadCtl.PersistentFlags().StringVarP(&key, "key", "k", "", "bears the path to the public key that will be signed by the CA's private key")
	shortbreadCtl.PersistentFlags().StringVarP(&baseUrl, "server", "s", "", "base url for the CA server")
<<<<<<< HEAD
=======
	shortbreadCtl.Flags().BoolVarP(&list, "list", "l", false, "list all usernames in the system")
}

func getUsers(c *cobra.Command, args []string) {
	if !list {
		return
	}

	svc, err := getHTTPClientService()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		return
	}

	getSvc := client.NewCertService(svc)
	users, err := getSvc.List().Do()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		return
	}

	for _, u := range users.List {
		fmt.Fprintf(os.Stdout, "%s\n", u)
	}

>>>>>>> 58afb88... Corrected formatting errors from PR, added revoke and list sub-commands,using the usernames as keys, one command to add and modify an user.
}

func main() {
	shortbreadCtl.AddCommand(updateUser)
	shortbreadCtl.AddCommand(revokeUser)
	shortbreadCtl.Execute()
}

// bin/shortbreadctl update -k /Users/shantanu/.ssh/id_rsa.pub -p /Users/shantanu/.ssh/users_ca -u shantanu -e permit-pty -c USER
