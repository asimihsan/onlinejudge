## Purpose

Run untrusted code in a sandbox that prevents it from harming the host machine, other processes, and unauthorised use of the network.

## Evaluator schema

A service that allows people to:

-   get problems (statement, initial starting code)
-   attempt a problem (accept a proposed solution, then executes it with the unit tests and returns result)
    -   do not expose unit tests to end-user.
-   create problems
-   update problems

### problem

-   problem_summary
    -   attributes:
        -   id (URL-friendly short name) (string)
        -   version (number)
        -   title (string)
        -   supported_languages (JSON list for what languages are supported)
        -   creation_date (ISO 8601 datetime) (string)
        -   last_updated_date (ISO 8601 datetime) (string)
    -   hash key: id
    -   range key: title
    -   local secondary indexes: <none>
    -   global secondary indexes: <none>
-   problem_details
    -   attributes:
        -   id (<problem id>#<language>) (string)
        -   version (number)
        -   description (compress using GZIP) (binary)
        -   initial_code (compress using GZIP) (single initial code file) (binary)
    -   hash key: id
    -   range key: <none>
    -   local secondary indexes: <none>
    -   global secondary indexes: <none>
-   unit_test
    -   attributes:
        -   id (<problem id>#<language>) (string)
        -   version (number)
        -   unit_test (compress using GZIP) (single test file) (binary)

Notes

-   Problem descriptions may be language specific. Hence problem_details and unit_test are keyed using language.


## TODO

-   evaluator
    -   need to put problems onto server if you want to upload them (or don't! maybe don't allow uploads from server)
    -   rather make a command line argument to recreate/upload problems OR start server
-   Make some automated tests.
    -   evaluator
        -   GET/OPTIONS work
        -   unknown problem ID gives 404, even when you evaluate
-   Fix suid on sandbox. It seems to be root again, able to e.g. delete all files.
-   I think on boot the image can't ssh into ubuntu@localhost. Fix image.
-   After running the following infinite print in Java can't run Java programs any more

```
public class Solution {
    public static void main(String[] args) {
        while (true)
            System.out.println("foo");
    }
}
```

-   Add tabs, "Code" and "Tests".
    -   Tests get appended to code then run as one unit.
    -   It's optional, so use a checkbox and grey out text box etc.
-   Add decent description to frontend about how the site works
-   Persist code/stdin/actual stdout etc to DynamoDB, browsing to it retrieves it
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

-   Do graceful restarts of server.
    -   Test by running e.g. sleep then restart service, check you get a response.
-   Fix restart of service, can't restart because can't kill runner since it runs it via ssh?
-   Still using privileged containers, since you were root.
    -   `sudo useradd -d /home/ubuntu -m ubuntu -p password`
    -   `su - ubuntu -c ...` (put config, create container)
-   Bug fix - the 'output' in JSON is a massive byte array full of nulls.
    -   It compresses well but is wasting time.
    -   Can see this in Chrome or Firefox inspector.
-   Bug fix
    -   Firefox CORS doesn't work on runsomecode.com, just www.runsomecode.com?
    -   Fixed by changing POST to hit "/run/", not "http://www.runsomecode.com/run"
-   chosen combo box not working on mobile
    -   ah. chosen isn't supported on mobile. try not using optgroup.
-   Use Solarized theme for CodeMirror (play.elevatorsaga.com)
-   Use HTML5 localstorage to store last program in case browser dies
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

-   Add reCAPTCHA to site to prevent bots/APIs from using it.
    -   When you start using it use API keys to authorize yourself.
    -   By design Google reCAPTCHA will fail the second time for the same authorized response. Hence you need to do something else to authenticate the client once confirmed as human, e.g. give them an HMAC with a server secret. Or set cookie / use localstorage? Make it last e.g. 24 hours.
-   HTTP server that builds/runs python/ruby/java, return stdout/stderr
    -   Put HTTP server in LXC container.
    -   Will not run outside LXC container and do on every run clone/start/stop/destroy LXC container (measure latency)
        -   Around 5 seconds, too high
-   Defer gzip'ing of responses to nginx, let runner focus on running.
-   Allow stdin and expected stdout in call. Return actual stdout and correct or not. Easier for clients to use. 
    -   No, needs to be like Codewars where you run test code that exercises the submitted code. That way submitted code samples look concise and more relevant without boilerplate.

## Snippets

Refresh runner code:

```
watchmedo shell-command -c \
    'clear && date && GOOS=linux GOARCH=amd64 go build -o runner.linux && \
    ssh -i ~/.ssh/digitalocean root@www.runsomecode.com "service runner stop" ; \
    ssh -i ~/.ssh/digitalocean root@www.runsomecode.com "pkill runner.linux" ; \
    scp -i ~/.ssh/digitalocean runner.linux root@www.runsomecode.com:/usr/local/bin/runner.linux && \
    ssh -i ~/.ssh/digitalocean root@www.runsomecode.com "service runner start"' \
    -w -p '*.go' .
```

Refresh frontend

```
watchmedo shell-command \
    -c 'rsync -avz -e "ssh -i /Users/ai/.ssh/digitalocean" frontend/ root@www.runsomecode.com:/usr/share/nginx/html' \
    -w frontend
```

Refresh sandbox

```
watchmedo shell-command -c \
    'clear && date && \
    scp -r -i ~/.ssh/digitalocean . root@www.runsomecode.com:~/sandbox && \
    ssh -i ~/.ssh/digitalocean root@www.runsomecode.com "rm -rf /tmp/foo/sandbox && cp -r ~/sandbox /tmp/foo && chmod -R 777 /tmp/foo/sandbox"' \
    -w -p '*.cpp' .
```

Refresh evaluator:

```
watchmedo shell-command -c \
    'clear && date && GOOS=linux GOARCH=amd64 go build -o evaluator.linux && \
    ssh -i ~/.ssh/digitalocean root@www.runsomecode.com "service evaluator stop" ; \
    ssh -i ~/.ssh/digitalocean root@www.runsomecode.com "pkill evaluator.linux" ; \
    scp -i ~/.ssh/digitalocean evaluator.linux root@www.runsomecode.com:/usr/local/bin/evaluator.linux && \
    ssh -i ~/.ssh/digitalocean root@www.runsomecode.com "service evaluator start"' \
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