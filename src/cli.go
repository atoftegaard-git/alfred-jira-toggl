package main

import "flag"

var (
    opts = &options{}
    cli  = flag.NewFlagSet("alfred-toggl-jira", flag.ContinueOnError)
)

type options struct {
    Update      bool
}

func init() {
    cli.BoolVar(&opts.Update, "update", false, "check for updates")
}
