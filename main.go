package main

import (
	"github.com/mja00/kami-chan-server-installer/cmd"
	"github.com/mja00/kami-chan-server-installer/paper"
)

var Version = "dev"
var Commit = "none"

func main() {
	paper.Version = Version
	paper.Commit = Commit
	cmd.Version = Version
	cmd.Commit = Commit
	cmd.Run()
}
