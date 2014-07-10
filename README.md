# shortbread

OpenSSH CA signing and publishing Daemon.

## Problem

Managing SSH keys has two problems:

1. Onboarding new users to login to a box usually requires scp'ing keys to the
   target. Revoking requires removing the key (and remembering to do so).

2. Users generally just blindly trust hosts the first time they connect. This
   opens users up to MitM attacks.

To fix these two problems OpenSSH has implemented an SSH CA system. However, it
is a command line tool that is rather hard to use. See https://www.digitalocean.com/community/tutorials/how-to-create-an-ssh-ca-to-validate-hosts-and-clients-with-ubuntu

## Solution

Put Go and HTTP on it! All of these features should work from the go.ssh
library: https://godoc.org/code.google.com/p/go.crypto/ssh#Certificate

### Onboarding New Users

User story: Alice the admin needs to give access to the prod cluster to Ian the
intern.

Alice would post to a URL to sign a public key with constraints like time, or
commands.

```
POST /v1/sign
{
	'certificate': 'prod-servers',
	'username': 'core',
	'validityInterval': '201506231248',
	'rsaPubkey': 'ssh-rsa AAUw==',
}
```

Then a daemon living on the users laptop would pull down their certificates
from the signing machine.

```
GET /v1/certificates/fingerprint

$CERT_BODY
```

Unfortunatly Ian's laptop was stolen and he didn't encrypt the disk! We had
better revoke his keys:

```
POST /v1/revoke
{
	'rsaPubkey': 'ssh-rsa AAUw==',
```

### Onboarding New Hosts

Host verificiation would work very similarly. Lets get the user case down
first.
