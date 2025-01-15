package main

import (
	"flag"
	"time"

	"github.com/NickP005/go_mcminterface"
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
	solo_node := ""

	flag.StringVar(&SETTINGS_PATH, "settings", "interface_settings.json", "Path to the settings file")
	flag.StringVar(&TFILE_PATH, "tfile", "mochimo/bin/d/tfile.dat", "Path to node's tfile.dat file")
	flag.Float64Var(&SUGGESTED_FEE_PERC, "fp", 0.25, "The percentile of the minimum fee")
	flag.DurationVar(&REFRESH_SYNC_INTERVAL, "refresh_interval", 5*time.Second, "The interval in seconds to refresh the sync")
	flag.IntVar(&Globals.LogLevel, "ll", 5, "Log level (1-5). Most to least verbose")
	flag.StringVar(&solo_node, "solo", "", "Bypass settings and use a single node ip (e.g. 0.0.0.0")
	flag.IntVar(&Globals.APIPort, "p", 8080, "Port to listen to")

	flag.Parse()

	if flag.Lookup("help") != nil {
		flag.PrintDefaults()
		return false
	}

	if solo_node != "" {
		go_mcminterface.Settings.StartIPs = []string{solo_node}
		go_mcminterface.Settings.ForceQueryStartIPs = true
	}

	return true
}
