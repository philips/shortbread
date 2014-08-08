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
shortbreadctl useradd --username="abc" -k path/to/PublicKey -p path/to/PrivateKeyOnCA --extensions="ext1,ext2,..." --cert USER 
```

To revoke a certificate

```
shortbreadctl revoke -u core 
```

To modify an existing certificate

```
shortbreadctl update -username core -e ext1,ext2 --before 21089765 
```

####Design Decisions####

* Do we want separate sub-commands or do we want to proceed with a single root command as implemented in the prototype ? (My preference is for sub-commands)
* Enforce strict POSIX flags or stick to the modified cobra/pflags library
* etcd integration ? ( stick to simple in-memory + local disk storage for now ? )
* User enters value of `validBefore` and `validAfter` in days/weeks/ (or some other format) 
* Sysadmin enters url of CA everytime ? (or environment variables ?)

### Build ###

To build the server and cli simply execute:

```
./build 
```

### TODO ###

**server/server.go**

* Supports only signing for now, add support to revoke and modify
* Run it on amazon servers instead of localhost
    * write client side daemon to pull up to date certs from CA
* Change the key on the global map to be ssh fingerprint (?)
* Warn user about not specifying a certType value
* move from hardcoded validBefore and validAfter to user-specified values
* use disk to populate datastructure on startup

**cli/shortbreadctl.go**

* add sub-commands and appropriate run functions
* update/improve command descriptions

**client/shortbreadclient.json**

* add methods for `update` and `revoke`

**Build systems**

* Add Godeps and modify build scripts accordingly 

