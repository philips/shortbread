## How to get Shortbread deployed and Running 

There are 3 distinct pieces to shortbread. The CA server, commandline tool and client, with each running on a different machine.

If you are not familiar with the concept of SSH certificates [this](https://www.digitalocean.com/community/tutorials/how-to-create-an-ssh-ca-to-validate-hosts-and-clients-with-ubuntu) tutorial is a great place to start. 

----

## Build Instructions

The Shortbread client can be built by simply executing the build-client script as shown below : 

~~~shell
./build-client
~~~

-----

If you want to build both the client and the command line tool run the build script as shown below: 

~~~shell
./build
~~~

-----

If you want to make changes to the server and create an executable to test your changes you will need to execute the following commands: 

~~~shell
cd Godeps/_workspace/src/github.com/libgit2/git2go
make install
~~~

You only need to execute the above commands once to build the `libgit2.a` file. After that you can just build the server by executing the build-server script: 

~~~shell
./build-server
~~~ 

All the build scripts will place the resulting executable in `shortbread/bin` 


---- 

### CA Server 

Getting the shortbread server running couldn't be simpler. When starting your coreOS instance use the `cloud-config` provided in the `script` folder. 

~~~YAML
#cloud-config
coreos:
  etcd:
    # generate a new token for each unique cluster from https://discovery.etcd.io/new
    discovery: https://discovery.etcd.io/<token>
    # multi-region and multi-cloud deployments need to use $public_ipv4
    addr: $private_ipv4:4001
    peer-addr: $private_ipv4:7001
  units:
    - name: etcd.service
      command: start
    - name: fleet.service
      command: start
    - name: core-shortbread.service
      enable: true 
      command: start
      content: |
        [Unit]
        Description=shortbread CA server
        After=etcd.service
        After=docker.service
        [Service]
        TimeoutStartSec=0
        Environment="HOME=/root"
        ExecStartPre=-/usr/bin/docker kill shortbread_server
        ExecStartPre=-/usr/bin/docker rm shortbread_server
        ExecStartPre=/usr/bin/docker pull quay.io/shantanu/shortbread
        ExecStart=/usr/bin/docker run -p 8889:8080 -v /etc/ssh:/root/ssh --name shortbread_server quay.io/shantanu/shortbread
        ExecStop=/usr/bin/docker stop shortbread_server
        [Install]
        WantedBy=multi-user.target
write_files:
  - path: /root/.dockercfg
    owner: core:core
    permissions: 0644
    content: |
      {
       "https://quay.io/v1/": {
        "auth": "JHRva2VuOjA0MFUwTEZPS1cyUDgxSDNBTUUwTU1WN0RONzRLWkVYTjdUREozQTZRSkpDNEhKN1ZMSUc4UENJVzhDS0gzWDA=",
        "email": ""
        }
      }

~~~

The data required by the server is expected to be in `/etc/ssh`. An example file-system layout is shown below

~~~
/etc
   ssh/
      sshd_config
      ssh_host_key
      ssh_host_key.pub
      ssh_host_key_rsa
      ssh_host_key_rsa.pub
      production_server --> private rsa key used to sign a users public key ( to create the certificate) 
      production_server.pub --> public part of the rsa key, to be stored on the production server
   
      shortread/
         certs/
            .git/
            serverDirectory -> encoded file. contains backup of list of server names mapped to their url's  
            userDirectory -> encoded file. contains backup of list of user names mapped to their public keys. 
            2d714bbeb313606dc464296ceb95aabe/  --> fingerprint of users public key
               production_server-cert.pub --> certificate named after the private key used to sign it. 
   
~~~

The `shortbread/certs` folder in `etc/ssh` has a `git` repo that is used to backup the data stored in the server. At the time of initialization shortbread will look for this git repo, if none is found it will create a local repo. 

Shortbread also optionally supports cloning and pushing to a remote repo via SSH. To avail of this feature place the requiste public and private keys in `/etc/ssh/`. The public key should be registered with the remote host (eg: Github) 

----- 

##cli

shortbread ships with a command line tool: `shortbreadctl` 

The cli allows a sys-admin to interact with the server through an API. 

After using the build script, place the resulting `shortbreadctl` binary in a directory(Ideally `$GOBIN`) located in your `PATH`.

The user is required to set the `SHORTBREADCTL_URL` environment variable to point to the correct server instance. The default value is `http://localhost:8889/v1/`. If you want to use a different port other than `8889`, you can change the portmapping in the `cloud-config`

### Useful sub-commands 

**user-add:** The user-add commands takes a user name and a path to a public key, and transmits them to the shortbread server. This ensures that the user of the cli does not need to keep track of myriad public keys. 

**new** This sub-command is used to create new certificates on the CA. It has a number of flags that allow the user to create certificates with features to his/her satisfaction. 

~~~
  -a, --after="0": specify the initial date(DD-January-YYYY) from which the certificate will be valid
  -b, --before="INFINITY": specify the date(DD-January-YYYY) upto which the certificate is valid. Default value is  "INFINITY"
  -c, --cert="USER": choose from "USER" or "HOST"
  -e, --extensions=[]: comma separated list of permissions(extesions) to bestow upon the user
      --help=false: help for new
  -p, --private="": specify the name of the private key to be used
  -r, --restrictions=[]: comma separated list of criticalOptions(restrictions) to place on the user
  -u, --username="": username of the entity to whom the certificate is issued.Must be a valid username stored in the user directory
~~~


Thus creating a new certificate would involve the following steps: 

~~~shell
export SHORTBREADCTL_URL="http://ec2-ip@amazonaws.com:port/v1/" #ensure environment variable is set correctly
shortbreadctl user-add username path/to/publickey.pub 
shortbreadctl new -u username -e permit-pty -p name-of-privateKey-on-CA
~~~

Refer [here](http://openbsd.cs.toronto.edu/cgi-bin/cvsweb/src/usr.bin/ssh/PROTOCOL.certkeys?annotate=1.9) for a full list of permissions(extensions) and restrictions(criticialOptions) that can be set while creating a certificate

### Future Work

- Add a revoke command to allow a user to revoke a certificate.
- Secure the API

----- 

## client 

We will assume the client is using **OS X**. 

The build script will produce two  binaries of interest for the client: `generatePlist` and `client`. 

#### generatePlist

This binary is used to populate the `com.shortbread.plist.template` file stored in the `script` folder. It takes only one argument which is the url of the shortbread CA server. 

~~~
generatePlist http://shortbreadexample:8889/v1/
~~~

After executing this binary, it should create `3 files` in the `script` folder

~~~
shortbread/
  script/
    com.shortbread.client.plist
    shortbread.client_error.log
    shortbread.client_output.log 
~~~

The .plist file needs to be loaded into `launchd`, which can be accomplished using the folllowing commands. 
~~~shell 
cp script/com.shortbread.client.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.shortbread.client.plist
~~~

The binary will be run everytime the user logs in and after every 5 minutes. It will add any new certs to the ssh-agent if any. 


#### client 

The `client` binary is the one being run by launchd after every `5 minutes`. It makes `Get` requests to the CA server at the url provided as an argument to `generatePlist`. The resulting certificates are added to the `ssh-agent` and can be verified by running `ssh-add -l`. 

An example output is given below: 

~~~shell
ssh-add -l 
2048 94:cd:96:72:74:8e:08:6c:64:e9:d1:79:a7:7d:9b:c2 certificate added by shortbread (RSA-CERT)
~~~

----- 

## Example Overview of the system

* Use the cloud-config to spin up a new coreOS instance to serve as a CA (Certifying Authority) 
   * create public, private rsa keys for signing
   * send the public keys to the servers of your choice  
   * on the servers, that will accept the certs make sure to add the following `TrustedUserCAKeys /etc/ssh/*.pub` to the `sshd_config` 
* issue new certificates using the cli tool 
* user does nothing and his certs are automatically added to his agent. ( win ! ) 