package service

import (
	swire "demo/internal/server/wire"
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	NewCronService,
	NewGreeterService,
	wire.Bind(new(swire.CronService), new(*CronService)),
)
