# Securely executing untrusted code

## Introduction

How would you run a service that lets random strangers attempt to code a correct solution to a given problem, where the service then executes their code to verify its correctness? What types of security threats would the service face and how can we counter them in a cheap and efficient way?

This article discusses a solution that I've developed for this problem. It is far from novel, nor is this a new problem. Please see the numerous references in the right margin for prior art or more information about concepts.

![](architecture1.pdf)

## Overview

Referring to the diagram above, the request lifecycle for our service is:

-   A user submits a piece of code to the service, which is running in a virtual machine (VM) such as Digital Ocean ^[[http://www.digitalocean.com](http://www.digitalocean.com)] or Amazon EC2 ^[[http://aws.amazon.com/ec2](http://aws.amazon.com/ec2)] running Ubuntu 14.04 LTS.
-   nginx ^[[http://www.nginx.org](http://www.nginx.org)], a reverse web proxy, forwards the code to another internal web service called `runner`.
-   `runner` (a custom service written in Go ^[[http://golang.org](http://golang.org)]) makes the code available to another distinct, light-weight LXC ^[[https://help.ubuntu.com/lts/serverguide/lxc.html](https://help.ubuntu.com/lts/serverguide/lxc.html)] container.
-   `runner` delegates the execution of the code to another program called `sandbox`. `sandbox` runs within the LXC container.
-   `sandbox` (a custom executable written in C) erects restrictions around the code, executes it, then returns the result.

We're going to discuss each of these stages in turn, but in reverse order because to be frank the most interesting parts are towards the end. After we'll step through examples of attackers attempting to violate the security of our service and what components defend against these attacks. Finally I'll discuss limitations, and plans for the future.

## sandbox

Recall that our use case for `sandbox` is "We have some arbtirary code we want to run using e.g. `/usr/bin/python /tmp/code.py`, but please reduce the code's ability to harm the system it's running on". What kinds of harm can code cause, and how can we prevent it? [^1]

The most common "attackers" of our service will be accidental. It's reasonable to expect students to write infinite loops, use up large amounts of memory, or attempt to create very large files for computations. Each of these represents the starvation of a given resource (in turn: CPU, virtual memory, and file system capacity).

On Linux, we can enforce resource limits by calls to `setrlimit()` ^[Secure Programming Cookbook for C and C++ (2003), Section 13.9 "Guarding Against Resource Starvation Attacks on Unix"]. We simply set up the following rlimit's before the code [^2]:

-   `RLIMIT_CPU`, soft and hard limits of 5 seconds
-   `RLIMIT_FSIZE`, soft and hard limits of 10MB
-   `RLIMIT_LOCKS`, soft and hard limits of 0.
-   `RLIMIT_MEMLOCK`, soft and hard limits of 0.
-   `RLIMIT_NPROC`, soft and hard limits of 25.

It's important to note what these rlimit's do and do not protect against:

-   `RLIMIT_CPU` is for time actually spent running on the CPU. This rlimit does not protect against a user sleeping forever. To protect against an infinite sleep the `runner` server enforces a timeout when it executes `sandbox` (more on this later).
-   Setting memory limits using `RLIMIT_AS` below the maximum memory available in the LXC container always causes Java to crash. This is because Java attempts to memory map all available virtual memory to it as some sort of initialization step.
-   Setting memory limits using `RLIMIT_RSS` isn't useful because it only affects `madvise()` calls and only effects older Linux kernels. ^[[http://man7.org/linux/man-pages/man2/setrlimit.2.html](http://man7.org/linux/man-pages/man2/setrlimit.2.html)].
-   Hence we cannot rely on `sandbox` for preventing code from allocating infinite amounts of memory. Instead we rely on the LXC container's cgroups to enforce this limit (more on this later).

However, what about malicious attackers? By running `sandbox` as an unprivileged user we can prevent the direct ability to e.g. `rm -rf /`, `reboot`, etc. However there are, and always will be, known and unknown ways for users to escalate their privileges, i.e. exploit the system into granting them administrator rights even though they're running as an unprivileged user. Ways of doing this include executing vulnerable setuid programs or exploiting buggy kernel functions (aka syscalls).

Our method of limiting the likelihood of privilege escalation attacks is to prevent the code from calling certain syscalls. As an example, given a specific exploit affecting `vmsplice()` ^[[`vmsplice()`: the making of a local root exploit, LWN, (2008) (http://lwn.net/Articles/268783/)](http://lwn.net/Articles/268783/)], and given we prevent code from executing `vmsplice()`, we prevent exploits that use this syscall.

How does one set up restrictions for what syscalls can be made? The Linux kernel has supported "mode 2 seccomp" since 3.5 (released 2012-07-21), which is a way for non-root users to set up these restrictions. The "mode 2" here means that we can use the Berkley Packet Filter (BPF) system to filter syscalls based on argument, and execute different actions depending on each syscall. [^3]

However it is prudent to accept that any form of sandboxing using user privileges, rlimits, and kernel syscall filtering will fail to prevent a motivated attacker. Hence we must run both the sandbox and the untrusted code in another containment to limit the ability of exploits to affect the host VM.

## LXC containers

LXC is "chroot on steroids". It is a userspace interface to a variety of Linux kernel containment features (LXC relies on 2.6.32, released 2009-12-03). LXC lies in between chroot and a full virtual machine, offering more security features without sacrificing memory and CPU efficiency.

TODO more details

## `runner`

`runner` receives untrusted code over HTTP, makes it available to the LXC container, then uses `sandbox` to execute the code within the LXC container.

TODO more details

It was imperative that `runner` have as small a memory footprint as possible, because I want to host the service on very cheap 512MB RAM VMs. Hence I used Go for the server; during load testing the private resident set size (RSS) memory occupancy of `runner` never exceeds around 2MB. It is also pleasantly CPU performant, easy to write, and easy to deploy. 

`runner` will recreate the LXC container whenever the untrusted code fails to run in any way (returns a non-0 return code, times out, violates `sandbox`'s restrictions, hits the LXC container's cgroup memory limit, etc). Ideally we would simply recreate the ephemeral LXC container on any code execution. However since LXC ephemeral containers take around 5 seconds to recreate from a base image this would make our service a little too unresponsive to users.

## nginx

I use nginx to protect the `runner` service from accidental or malicious denial of service attacks. For example:

-   Setting `client_body_timeout` (for the request HTTP body) ^[[http://nginx.org/en/docs/http/ngx_http_core_module.html#client_body_timeout](http://nginx.org/en/docs/http/ngx_http_core_module.html#client_body_timeout)], `client_header_timeout` (for the request HTTP header) ^[[http://nginx.org/en/docs/http/ngx_http_core_module.html#client_header_timeout](http://nginx.org/en/docs/http/ngx_http_core_module.html#client_header_timeout)] and `send_timeout` (for  sending the HTTP response) ^[[http://nginx.org/en/docs/http/ngx_http_core_module.html#send_timeout](http://nginx.org/en/docs/http/ngx_http_core_module.html#send_timeout)] all to 10 seconds to prevent slow clients from using up all available connections. This, for example, prevents the Slowloris ^[[https://en.wikipedia.org/wiki/Slowloris_%28software%29](https://en.wikipedia.org/wiki/Slowloris_%28software%29)] attack from succeeding.

However this isn't particularly interesting or novel for this article.

## The Host VM 

Like any other public-facing Linux server we harden our server by:

-   Setting up a firewall to only leave ports 22 and 80 open, and bind nginx to port 80 where nginx then forwards traffic to the internal `runner` service bound as a unprivileged user to port 8080.
-   Disabling password authentication over SSH then installing `fail2ban` ^[[https://en.wikipedia.org/wiki/Fail2ban](https://en.wikipedia.org/wiki/Fail2ban)]to at least slow down the annoying SSH brute forcers out there.
-   Enabling security features in `/etc/sysctl.conf`. ^[[http://www.cyberciti.biz/faq/linux-kernel-etcsysctl-conf-security-hardening/](http://www.cyberciti.biz/faq/linux-kernel-etcsysctl-conf-security-hardening/)].

Also I prepare the host VM as a static image using Packer [^4], with all the hardening and setup of the LXC containers done ahead of time. Instead of leaving servers running for a long time and periodically upgrading them I periodically recreate static base images and replace all running servers with new servers. This reduces the change of untested changes being made to running servers, and makes it easier to set up a comprehensive continuous integration pipeline with end to end testing of my service. [^5]

## Example attacks

TODO

## Limitations

TODO

## Future work

TODO

[^1]:  Of course note that with compiled languages, such as C, C++, and Java, that we want to sandbox both the compilation and execution of untrusted code. We want to avoid e.g. infinite C++ compilation due to well craft recursive templates ([http://stackoverflow.com/questions/6079603/infinite-compilation-with-templates](http://stackoverflow.com/questions/6079603/infinite-compilation-with-templates)).

[^2]: The code I used is almost identical to Geordi [https://github.com/Eelis/geordi/blob/master/src/lockdown.cpp#L97](https://github.com/Eelis/geordi/blob/master/src/lockdown.cpp#L97), licensed as Public Domain.

[^3]: The code I used is almost identical to "Using simple seccomp filters" ([http://outflux.net/teach-seccomp/](http://outflux.net/teach-seccomp/)), The Chromium OS Authors (2012), licensed under BSD-style license. However the trick here is to find the minimum set of syscalls supported for each type of interpreter/compiler, to reduce the exposed kernel surface area as much as possible.

[^4]: [https://www.packer.io](https://www.packer.io)

[^5]: This is commonly known as the "immutable server" pattern. [http://martinfowler.com/bliki/ImmutableServer.html](http://martinfowler.com/bliki/ImmutableServer.html)