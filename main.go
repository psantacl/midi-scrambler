package main

import (
	"os"
	"com.github/psantacl/midi-scrambler/pkg/logging"
	"com.github/psantacl/midi-scrambler/cmd"
)


func main() {
	var logger = logging.InitLogging()
	defer logger.Sync() // flushes buffer, if any
	logging.Sugar.Infow("Welcome to midi-scrambler")

	cmd.RootCmd.Execute()
	os.Exit(0)

}
