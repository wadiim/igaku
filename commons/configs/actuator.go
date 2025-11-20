package configs

import (
	actuator "github.com/sinhashubham95/go-actuator"
)

var ActuatorConfig = &actuator.Config{
	Endpoints: []int{
		actuator.Env,
		actuator.Info,
		actuator.Metrics,
		actuator.Ping,
		actuator.Shutdown,
		actuator.ThreadDump,
	},
	Env: "dev",
	Name: "Igaku",
	Port: 8080,
	Version: "0.1.0",
}
