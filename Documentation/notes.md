### Proposed Command Line Interface ###

shortbreadctl is a command line tool for sysadmins to interact with a remote CA. 

Available Flags:
```
  -a, --after=0: number of days before the certificate becomes valid
  -b, --before=0: number of days the certificate is valid
  -c, --cert="": choose from "USER" or "HOST"
  -e, --extensions=[]: comma separated list of permissions(extesions) to bestow upon the user
      --help=false: help for shortbreadCtl
  -k, --key="": bears the path to the public key that will be signed by the CA's private key
  -p, --private="": specify the path of the private key to be used in creating the certificate
  -r, --restrictions=[]: comma separated list of permissions(restrictions) to place on the user
  -s, --server="": base url for the CA server
  -u, --user="": username of the entity to whom the certificate is issued

Use "shortbreadctl help [command]" for more information about that command.
```

#### Illustrative Examples ####

To add a new user certificate

```
shortbreadctl update -k $HOME/.ssh/id_rsa.pub -p users_ca -u shantanu -e permit-pty -c USER -b 31-December-2014
```

To revoke a certificate

```
shortbreadctl revoke -u core 
```




### Build ###

To build the server and cli and client simply execute:

```
./build 
```

### TODO ###

**Build systems**

* Add Godeps and modify build scripts accordingly 

-------------------------------------------------------------------------------

### Detailed Theory of Operation ###

**Interaction of Sys-Admin**

* Generate N pairs of public private keys on the CA (Certifying Authority)
* Get access to id_rsa.pub of the user/users he wants to provide access to
* Use the cmd line interface to specify which private keys to use to sign the public key
    * Have a map with server-names as keys and values as path to private keys on the CA 
    * provide facility in cli to list all servers that sys-admin can provide access to. 
    * Potentially, more than one private key can be associated with providing access to a set of servers
    * This depends on the setup server side, not discussed here. Focussing on making it work for users first.

* This will generate multiple certs for one public key, stored in datastructure `map[PublicKey][]*ssh.Certificate`
  * Possibly, wrap the cert in a  data-struct that also keeps track of server/private key used to sign it 
  * While `revoking` a cert, specify the PublicKey of the user, and the servers/private key to revoke, for fine grained deletion, just specifying public key will delete all certs.

**User's Interaction**

* Hand over Public Key to sys-admin.
* shortbreadctl daemon running will download diff certs, but also create copies of his id_rsa private key and add that to ssh-agent via ssh-add. 
* this will add the cert and hence he has to do nothing.
* I also deleted the copied keys and the connection seems to work fine, data now in ssh-agent. 






