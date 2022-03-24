package logger

import (
	"github.com/hashicorp/go-hclog"
)

// AppLogger defines the global application logger
var AppLogger = hclog.New(&hclog.LoggerOptions{
	DisableTime: true,
	Level:       hclog.Debug,
})
