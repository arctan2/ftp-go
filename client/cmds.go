package client

import (
	"github.com/fatih/color"
)

type envs interface {
	remoteEnv
	localEnv
}

var blue = color.New(color.FgBlue).PrintfFunc()
