#!/usr/bin/env bash

# https://gist.github.com/emgee3/07bfec5d102012b9e47a
# 
set -e

sudo sed -i 's|[#]*$ModLoad imudp|$ModLoad imudp|g' /etc/rsyslog.conf
sudo sed -i 's|[#]*$UDPServerRun 514|$UDPServerRun 514|g' /etc/rsyslog.conf
sudo sed -i 's|[#]*$IncludeConfig /etc/rsyslog.d/\*.conf|#$IncludeConfig /etc/rsyslog.d/\*.conf|g' /etc/rsyslog.conf
sudo echo '$UDPServerAddress 127.0.0.1' >> /etc/rsyslog.conf
sudo echo '$IncludeConfig /etc/rsyslog.d/*.conf' >> /etc/rsyslog.conf
sudo tee /etc/rsyslog.d/49-haproxy.conf >/dev/null <<EOF
$AddUnixListenSocket /var/lib/haproxy/dev/log
local0.* /var/log/haproxy.log
local1.* /var/log/haproxy.log
notice.* /var/log/haproxy.log
EOF
mkdir -p /var/lib/haproxy/dev
sudo restart rsyslog

sudo tee /etc/logrotate.d/haproxy >/dev/null <<EOF
/var/log/haproxy.log {
    daily
    rotate 52
    missingok
    notifempty
    compress
    delaycompress
    postrotate
        invoke-rc.d rsyslog rotate >/dev/null 2>&1 || true
    endscript
}
EOF

sudo mkdir -p /etc/haproxy/
sudo tee /etc/haproxy/haproxy.cfg >/dev/null <<EOF
global
    maxconn 50000

    # run in background
    daemon

    # run as specific low-privilege user
    user haproxy
    group haproxy

    # Only one process. This is the default anyway.
    nbproc 1

    # for restarts
    pidfile /var/run/haproxy.pid

    # Logging to syslog facility
    # log format: http://cbonte.github.io/haproxy-dconv/configuration-1.6.html#8.2.3
    log 127.0.0.1 local0
    log 127.0.0.1 local1 notice

    # allow control by haproxyctl over socket
    stats socket /tmp/socket mode 660 level admin
    stats timeout 1s

    # allow seamless restarts by dumping server state into a file
    # http://blog.haproxy.com/2015/10/14/whats-new-in-haproxy-1-6/
    # yikes, after restarts haproxy is no longer responsive to requests.
    # don't use this!
    #server-state-file /tmp/server_state

    # distribute health checks with a bit of randomness
    spread-checks 5
    
    # safe SSL settings
    ssl-default-bind-options no-sslv3 no-tls-tickets force-tlsv12
    ssl-default-bind-ciphers AES128+EECDH:AES128+EDH

defaults
    log global
    mode http

    # if queue backs up and user clicks "stop" kill the HTTP request to free
    # up the request
    option abortonclose

    option httplog
    option dontlognull
    option forwardfor
    option redispatch

    # protects against slowloris attacks where the HTTP POST headers are sent
    # fast, but the POST body is sent slowly. without this option the
    # timeout http-request option is ineffective if the POST body is slow.
    option http-buffer-request

    # ignore 408's
    # http://blog.haproxy.com/2015/10/14/whats-new-in-haproxy-1-6/
    option http-ignore-probes

    # default is 'never'. reuse idle connections for sessions other than
    # the session that opened the connection. 
    # http://blog.haproxy.com/2015/10/14/whats-new-in-haproxy-1-6/
    http-reuse safe

    # if sending request to one server fails, retry before aborting
    retries 3
    
    timeout connect 10s
    timeout client 10s
    timeout http-request 10s
    timeout server 15s
    timeout queue 300s
    timeout http-keep-alive 60s
    timeout tarpit 10s
    
    default-server on-marked-down shutdown-sessions inter 1s
    compression algo gzip
    compression type text/html text/css text/javascript text/plain application/json

listen stats
    bind 127.0.0.1:8001
    mode http
    stats enable
    stats uri /

frontend http-in
    bind *:80
    mode http
    monitor-uri /ping

    # -------------------------------------------------------------------------
    # DDoS / rate limiting
    # https://github.com/jvehent/haproxy-aws
    # -------------------------------------------------------------------------
    # Define a table that will store IPs associated with counters
    stick-table type ip size 10m expire 30s store conn_cur,conn_rate(10s),http_req_rate(10s),http_err_rate(10s)

    # Enable tracking of src IP in the stick-table
    tcp-request content track-sc0 src

    # Reject the new connection if the client already has 20 opened
    http-request add-header X-Haproxy-ACL %[req.fhdr(X-Haproxy-ACL,-1)]over-10-active-connections, if { src_conn_cur ge 20 }

    # Reject the new connection if the client has opened more than 20 connections in 10 seconds
    http-request add-header X-Haproxy-ACL %[req.fhdr(X-Haproxy-ACL,-1)]over-20-connections-in-10-seconds, if { src_conn_rate ge 20 }

    # Reject the connection if the client has passed the HTTP error rate
    http-request add-header X-Haproxy-ACL %[req.fhdr(X-Haproxy-ACL,-1)]high-error-rate, if { sc0_http_err_rate() gt 5 }

    # Reject the connection if the client has passed the HTTP request rate
    http-request add-header X-Haproxy-ACL %[req.fhdr(X-Haproxy-ACL,-1)]high-request-rate, if { sc0_http_req_rate() gt 20 }

    # block PHP probes from bots
    http-request add-header X-Haproxy-ACL %[req.fhdr(X-Haproxy-ACL,-1)]bad-path, if { path_end -i .php }

    # if previous ACL didn't pass, tarpit the request
    # use tarpit to slow down attackers
    acl fail-validation req.fhdr(X-Haproxy-ACL) -m found
    http-request tarpit if fail-validation
    # -------------------------------------------------------------------------

    acl site_dead nbsrv(run) eq 0
    monitor fail if site_dead

    acl requires_runner path_beg -i /run
    acl requires_runner path_beg -i /evaluator/evaluate

    use_backend run if requires_runner
    use_backend other if !requires_runner


backend run
    mode http
    option httplog

    # hit the first server with available connections. since we also put
    # 'maxconn 2' in the server config (using administer.rb) we always
    # hit the same-region host first, and then spill over to other regions.
    # this is because right now each run server only supports one concurrent
    # request (only one LXC container).
    #
    # need maxconn 2 because right now we call /evaluator/evaluate, which
    # then hits the LB again to call /run.
    # http://cbonte.github.io/haproxy-dconv/configuration-1.6.html#4.2-balance
    # http://stackoverflow.com/questions/8750518/difference-between-global-maxconn-and-server-maxconn-haproxy
    balance first

    option httpchk GET /ping
# --- server block run start ---
#    server run.sfo1 104.131.152.160:80 check
#    server run.lon1 178.62.92.142:80 check
#    server run.sgp1 128.199.180.212:80 check
# --- server block run end ---

backend other
    mode http
    option httplog
    balance first
    option httpchk GET /ping
# --- server block other start ---
# foo
# --- server block other end ---

EOF

# Test for haproxy user and create it if needed. Chroot it and prevent it from 
# getting shell access
groupadd --system haproxy
useradd -g haproxy -d /var/lib/haproxy -s /bin/false haproxy

sudo /etc/init.d/haproxyctl configcheck
sudo /etc/init.d/haproxyctl reload

sudo tee /etc/init.d/haproxy >/dev/null <<"EOF"
#!/bin/bash
### BEGIN INIT INFO
# Provides:          haproxy
# Required-Start:    $local_fs $network $remote_fs
# Required-Stop:     $local_fs $remote_fs
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: fast and reliable load balancing reverse proxy
# Description:       This file should be used to start and stop haproxy.
### END INIT INFO

# Author: Arnaud Cornet <acornet@debian.org>

PATH=/sbin:/usr/sbin:/bin:/usr/bin
PIDFILE=/var/run/haproxy.pid
CONFIG=/etc/haproxy/haproxy.cfg
HAPROXY=/usr/local/sbin/haproxy
CONFIGTEST_LOG=/var/log/haproxy_configtest.log
USER=root
GROUP=root
EXTRAOPTS=
CHROOT_DIR=/var/lib/haproxy/dev

test -x $HAPROXY || exit 0
test -f "$CONFIG" || exit 0

if [ -e /etc/default/haproxy ]; then
    . /etc/default/haproxy
fi

[ -f /etc/default/rcS ] && . /etc/default/rcS
. /lib/lsb/init-functions


haproxy_start()
{
    mkdir -p "$CHROOT_DIR"
    start-stop-daemon --start --pidfile "$PIDFILE" --chuid $USER:$GROUP \
        --exec $HAPROXY -- -f "$CONFIG" -D -p "$PIDFILE" \
        $EXTRAOPTS || return 2
    return 0
}

haproxy_stop()
{
    if [ ! -f $PIDFILE ] ; then
        # This is a success according to LSB
        return 0
    fi
    for pid in $(cat $PIDFILE) ; do
        /bin/kill $pid || return 4
    done
    rm -f $PIDFILE
    return 0
}

haproxy_reload()
{
    socat /tmp/socket - <<< "show servers state" > /tmp/server_state
    $HAPROXY -f "$CONFIG" -p $PIDFILE -D $EXTRAOPTS -sf $(cat $PIDFILE) \
        || return 2
    return 0
}

haproxy_status()
{
    if [ ! -f $PIDFILE ] ; then
        # program not running
        return 3
    fi

    for pid in $(cat $PIDFILE) ; do
        if ! ps --no-headers p "$pid" | grep haproxy > /dev/null ; then
            # program running, bogus pidfile
            return 1
        fi
    done

    return 0
}

haproxy_configtest()
{
    $HAPROXY -f "$CONFIG" -c > "$CONFIGTEST_LOG" 2>&1
    ret=$?
    if [ $ret -eq 0 ]; then
        # Valid config - remove $CONFIGTEST_LOG
        rm "$CONFIGTEST_LOG"
    fi

    return $ret
}


case "$1" in
start)
    log_daemon_msg "Starting haproxy" "haproxy"
    haproxy_start
    ret=$?
    case "$ret" in
    0)
        log_end_msg 0
        ;;
    1)
        log_end_msg 1
        echo "pid file '$PIDFILE' found, haproxy not started."
        ;;
    2)
        log_end_msg 1
        ;;
    esac
    exit $ret
    ;;
stop)
    log_daemon_msg "Stopping haproxy" "haproxy"
    haproxy_stop
    ret=$?
    case "$ret" in
    0|1)
        log_end_msg 0
        ;;
    2)
        log_end_msg 1
        ;;
    esac
    exit $ret
    ;;
reload|force-reload)
    log_daemon_msg "Reloading haproxy" "haproxy"
    haproxy_reload
    case "$?" in
    0|1)
        log_end_msg 0
        ;;
    2)
        log_end_msg 1
        ;;
    esac
    ;;
restart)
    log_daemon_msg "Checking haproxy configuration" "haproxy"
    haproxy_configtest
    ret=$?
    case "$ret" in
    0)
        log_end_msg 0
        ;;
    1)
        log_end_msg 1
        echo "Restart process aborted."
        echo "Check $CONFIGTEST_LOG for details."
        # Abort restart
        exit $ret
        ;;
    esac
    log_daemon_msg "Restarting haproxy" "haproxy"
    haproxy_stop
    haproxy_start
    case "$?" in
    0)
        log_end_msg 0
        ;;
    1)
        log_end_msg 1
        ;;
    2)
        log_end_msg 1
        ;;
    esac
    ;;
status)
    haproxy_status
    ret=$?
    case "$ret" in
    0)
        echo "haproxy is running."
        ;;
    1)
        echo "haproxy dead, but $PIDFILE exists."
        ;;
    *)
        echo "haproxy not running."
        ;;
    esac
    exit $ret
    ;;
configtest)
    haproxy_configtest
    ret=$?
    case "$ret" in
    0)
        echo "haproxy configuration is valid."
        ;;
    1)
        echo "haproxy configuration is NOT valid. Check $CONFIGTEST_LOG for details."
        ;;
    esac
    exit $ret
    ;;
*)
    echo "Usage: /etc/init.d/haproxy {start|stop|reload|restart|status|configtest}"
    exit 2
    ;;
esac

:
EOF
sudo chmod 755 /etc/init.d/haproxy
sudo chown root:root /etc/init.d/haproxy
sudo touch /var/run/haproxy.pid
sudo chmod a+rw /var/run/haproxy.pid
sudo update-rc.d haproxy defaults
sudo update-rc.d haproxy enable
