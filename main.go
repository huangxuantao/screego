package main

import (
	"github.com/huangxuantao/screego/cmd"
	pmode "github.com/huangxuantao/screego/config/mode"
)

var (
	version    = "unknown"
	commitHash = "unknown"
	mode       = pmode.Dev
)

func main() {
	pmode.Set(mode)
	cmd.Run(version, commitHash)
}
