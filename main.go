package main

import (
	"os"
	"com.github/psantacl/midi-scrambler/pkg/logging"
	"com.github/psantacl/midi-scrambler/cmd"
)


func main() {
	var logger = logging.InitLogging()
	defer logger.Sync() // flushes buffer, if any
	logging.Sugar.Infow("in logger",
		// Structured context as loosely typed key-value pairs.
		"cool", true,
	)

	cmd.RootCmd.Execute()
	os.Exit(0)

}
