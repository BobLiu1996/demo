package log

import (
	"context"
	"testing"
)

func TestNewZapLogger(t *testing.T) {
	NewLogger("cg-agent", "agent.log", 100, 10, 10, WithLevel("debug"))
	var ctx = context.Background()
	Infof(ctx, "Info Printf")
	Errorf(ctx, "Error Printf")
	Debug(ctx, "Debug Printf")
}
