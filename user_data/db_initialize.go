package main

import (
    "log"

    "github.com/smugmug/godynamo/conf"
    "github.com/smugmug/godynamo/conf_file"
    "github.com/smugmug/godynamo/conf_iam"
    keepalive "github.com/smugmug/godynamo/keepalive"
)

func Initialize() {
    // Read in the conf file, panic if it hasn't been initialized correctly.
    conf_file.Read()

    conf.Vals.ConfLock.RLock()
    defer conf.Vals.ConfLock.RUnlock()

    if conf.Vals.Initialized == false {
        log.Panicf("the conf.Vals global conf struct has not been initialized")
    }

    // launch a background poller to keep conns to aws alive
    if conf.Vals.Network.DynamoDB.KeepAlive {
        log.Printf("launching background keepalive")
        go keepalive.KeepAlive([]string{conf.Vals.Network.DynamoDB.URL})
    }

    // deal with iam, or not
    if conf.Vals.UseIAM {
        iam_ready_chan := make(chan bool)
        go conf_iam.GoIAM(iam_ready_chan)
        _ = <-iam_ready_chan
    }
}
