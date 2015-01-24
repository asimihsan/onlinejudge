## Purpose

Run untrusted code in a sandbox that prevents it from harming the host machine, other processes, and unauthorised use of the network.

## TODO

-   Use HTML5 localstorage to store last program in case browser dies
-   Allow stdin and expected stdout in call. Return actual stdout and correct or not. Easier for clients to use. 
-   Persist code/stdin/actual stdout etc to DynamoDB, browsing to it retrieves it
-   Use Solarized theme for CodeMirror (play.elevatorsaga.com)
-   chosen combo box not working on mobile
    -   ah. chosen isn't supported on mobile. try not using optgroup.
-   Don't think run/run-output files are getting deleting from /tmp, even though there's a defer to delete them.
-   allow stdin as input and expected stdout as output.
    -   runner passes stdin
    -   runner still returns stdout, but additionally returns some boolean saying
        whether stdout is correct or not
    -   but don't show a diff, or indicate what the correct stdout it! hide it.
-   setuid for runner in upstart conf not working, won't run, why?
-   create a fizz buzz example, hard code it for now
    -   given a single number input print out the correct fizz buzz output for
        1 to n inclusive.
    -   better idea is to come up with a single self-container file that defines the problem in some format
        -   so the runner can accept this as a file then prepare it. pump in stdin, create temporary files (cache?)
        -   investment of time, so it's easy to add problems later.
        -   single text file...or JAR-like ZIP file with a certain format?
    -   then add a dropdown / list for problems to attempt.
-   Add seccomp whitelist to LXC provisioned by packer
    -   Use Docker perl script to generate (https://github.com/docker/docker/blob/master/contrib/mkseccomp.pl)
    -   Even if it's a 100% whitelist fine for now (bottom of https://github.com/lxc/lxc)
-   Provision default LXC config as file using packer, not heredoc
-   Route 53 -> one Digital Ocean load balancer in NA
    -   Each Digital Ocean load balancer to one Digital Ocean droplet for now
    -   Provision droplet image using packer
    -   Run using sandbox. Run as nobody user.
-   Test trying to break out of sandbox

## TODO done

-   Add a proper Java mode for CodeMirror from here: http://codemirror.net/1/contrib/java/
    -   Use clike, but specify mode as "text/x-java".
-   prevent a "rm -rf /" from a program by dropping to a new user
    -   Did a "chown nobody:nogroup sandbox; chown +s sandbox"
    -   This means that whoever runs sandbox will get the equivalent of nobody's access.
    -   Can try e.g. "os.unlink("/var/foo")" after touching it by root
-   Fix javac and java to work; memory limits getting hit
    -   memory limits hit during javac but seccomp in sandbox somehow suppressed
        this
-   Add lots of other languages.
-   Use curl to test it works.
-   Rebuild Digital Ocean image
    -   Had to fix nginx conf, frontend main.js URL, reduce LXC mem to 256MB
-   When binding /tmp/foo then writing to foo.py, need to chmod a+rx foo.py!
    -   Fix runner.go
-   Sandbox
    -   Fix seccomp to support Ruby. Add all possible system calls, then remove until it fails.
    -   However you do it, don't allow network calls (prevent SYS_socketcall, or don't provision network in LXC)
-   On startup the runner should stop a running container then always start it.
    -   Rather than only starting it if it isn't running it.
    -   If the runner crashed, since only one process runs per LXC container, may as well restart it.
-   Add simple daemon wrapper to restart it on failure, e.g. supervisord or daemontools whatever.
    -   Used upstart, see "runner_upstart.conf". Copy to /etc/init/runner.conf.
-   Add file-based logging to server, it crashes
    -   By using Upstart log output goes to /var/log/upstart/runner.log
-   Use nginx to serve static content and runner from port 80.
    -   If runner crashes nginx can return the 500.
    -   Don't need CORS for the runner after this.
    -   Also won't work in environments where port 8080 for web servers is blocked, e.g. Amazon.

## TODO rejected

-   HTTP server that builds/runs python/ruby/java, return stdout/stderr
    -   Put HTTP server in LXC container.
    -   Will not run outside LXC container and do on every run clone/start/stop/destroy LXC container (measure latency)
        -   Around 5 seconds, too high
-   Defer gzip'ing of responses to nginx, let runner focus on running.

## Snippets

Refresh runner code:

```
watchmedo shell-command -c \
    'clear && date && GOOS=linux GOARCH=amd64 go build -o runner.linux && \
    ssh -i ~/.ssh/digitalocean root@104.236.136.8 "service runner stop" && \
    scp -i ~/.ssh/digitalocean runner.linux root@104.236.136.8:~/runner.linux && \
    ssh -i ~/.ssh/digitalocean root@104.236.136.8 "service runner start"' \
    -w -p '*.go' .
```

## Requirements

Tested on Ubuntu LTS 14.04.1 x64.

## How to use

### 1. Creating new EC2 AMI

-   Update `packer/packer.json`
    -   Set the AWS credentials you want to use for EC2 AMI creation.
    -   Set `source_ami` to what you want as the base image. Probably want the latest Ubuntu LTS, e.g. https://cloud-images.ubuntu.com/releases/trusty/release/
    -   Set `region` to where you want the AMI to live.
-   Run `./packer/build.sh` to build a new base AMI.

Currently the output AMIs we use are:

-   `us-west-2`: `-`

## 2. Running an EC2 instance for developing and playing around

`./run_ec2.rb`

## Details

-   Bind to an "incoming" AWS SQS queue for jobs of new untrusted code to run.
    -   This makes it easy and cheap to scale up/down, by spinning up more/less EC2 spot instances.
        -   Need at least one reserved instances bound at all times to make sure work can always be done.
        -   TODO make spin up/down based on a Cloudwatch alarm on the SQS queue size.
-   Put results onto an "outgoing" AWS SQS queue.

## Learnings

### Sandbox

TODO

### LXC

-   https://help.ubuntu.com/lts/serverguide/lxc.html

#### 1. Privileged example, create to destroy

```
# Create a privileged (root) container. Not recommended for security.
sudo lxc-create -t download -n u1 -- --dist ubuntu --release trusty --arch amd64

# Start it.
sudo lxc-start --name u1 --daemon

# Connect to it
sudo lxc-attach --name u1

# Stop it
sudo lxc-stop --name u1

# Destroy it.
sudo lxc-destroy --name u1
```

#### 2. Unprivileged example

Create a default unprivileged container config file to configure the user id mappings. This maps root in the container to an unprivileged host userid, and hence prevents access to e.g. /proc and /sys files representing host resources, and other files owner by root on the host.

```
mkdir -p ~/.config/lxc
echo "lxc.id_map = u 0 100000 65536" > ~/.config/lxc/default.conf
echo "lxc.id_map = g 0 100000 65536" >> ~/.config/lxc/default.conf
```

If you want the container to have external network access, additionally configure:

```
echo "lxc.network.type = veth" >> ~/.config/lxc/default.conf
echo "lxc.network.link = lxcbr0" >> ~/.config/lxc/default.conf
echo "$USER veth lxcbr0 2" | sudo tee -a /etc/lxc/lxc-usernet
```

Or e.g. no networking, `lxc.network.type = empty`. More examples in `/usr/share/doc/lxc/examples/`.

Then the container lifecycle is the same as the privileged example, except without sudo:

```
lxc-create -t download -n u1 -- -d ubuntu -r trusty -a amd64
lxc-start -n u1 -d
lxc-attach --name u1 --clear-env
lxc-stop -n u1
lxc-destroy -n u1
```

Get info on all LXC containers:

```
lxc-ls --fancy
```

Lifecycle

```
lxc-start-ephemeral -d -o ubase -n u1
lxc-attach -n u1 -- su - ubuntu -c bash -c '/usr/bin/env python -c "print(\"hi\")"'
lxc-stop -k -n u1
```

### AppArmor

TODO