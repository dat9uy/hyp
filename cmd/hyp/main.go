package main

import (
	"os"

	"github.com/dat9uy/hyp/internal/channel"
)

func main() { os.Exit(channel.Run(os.Args[1:], os.Stdout, os.Stderr)) }
