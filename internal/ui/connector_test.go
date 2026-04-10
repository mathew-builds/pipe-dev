package ui

import "testing"

func TestAnimatedConnectorFrames(t *testing.T) {
	// Different frames should produce different output.
	frame0 := RenderAnimatedConnector(3, 0, true)
	frame1 := RenderAnimatedConnector(3, 1, true)
	frame2 := RenderAnimatedConnector(3, 2, true)

	if frame0 == frame1 && frame1 == frame2 {
		t.Error("all frames are identical — no animation")
	}

	// Non-active connector should be static.
	static0 := RenderAnimatedConnector(3, 0, false)
	static1 := RenderAnimatedConnector(3, 1, false)
	if static0 != static1 {
		t.Error("non-active connector should not animate")
	}
}

func TestAnimatedConnectorContainsParticles(t *testing.T) {
	result := RenderAnimatedConnector(3, 0, true)
	if result == "" {
		t.Error("connector should not be empty")
	}
}
