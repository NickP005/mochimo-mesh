package main

import (
	"flag"
	"time"
)

/*
 * Creates if it doesn't exist the interface_settings.json file
 */
func Setup() {

}

/*
 * Loads the flags and prepares the program accordingly
 */
func SetupFlags() bool {
	flag.StringVar(&SETTINGS_PATH, "settings", "interface_settings.json", "Path to the settings file")
	flag.StringVar(&TFILE_PATH, "tfile", "mochimo/bin/d/tfile.dat", "Path to node's tfile.dat file")
	flag.Float64Var(&SUGGESTED_FEE_PERC, "fp", 0.25, "The percentile of the minimum fee")
	flag.DurationVar(&REFRESH_SYNC_INTERVAL, "refresh_interval", 1*time.Second, "The interval in seconds to refresh the sync")
	flag.IntVar(&Globals.LogLevel, "ll", 5, "Log level (1-5). Most to least verbose")

	flag.Parse()

	if flag.Lookup("help") != nil || flag.Lookup("h") != nil {
		flag.PrintDefaults()
		return false
	}

	return true
}
