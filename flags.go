package main

import (
	"flag"
	"fmt"
	"os"
)

// args contains the command line arguments
var args []string
var flagCacheBytes int64
var flagCacheTimeout int64
var flagTimeout int
var bindAddress string

func init() {
	flag.Int64Var(&flagCacheBytes, "cache-size", 25*1000, "maximum in-memory cache size in bytes")
	flag.Int64Var(&flagCacheBytes, "cache-age", 60*60, "maximum number of seconds after which cache entries should expire")
	flag.IntVar(&flagTimeout, "timeout", 1, "timeout in seconds after which to expire akhttpd")
	flag.Parse()

	// read command line arguments
	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("Missing 'bindAdress'")
		os.Exit(1)
	}

	// read the bind address
	bindAddress = args[0]
}
