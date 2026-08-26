package target

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testTargets = []Target{
	{ExecutionID: "event", TargetID: "event-global"},
	{ExecutionID: "event/foo.*", TargetID: "event-group"},
	{ExecutionID: "event/foo.bar", TargetID: "event-specific"},
	{ExecutionID: "function/Call", TargetID: "function-call-1"},
	{ExecutionID: "function/Call", TargetID: "function-call-2"},
	{ExecutionID: "request", TargetID: "request-global"},
	{ExecutionID: "request/nomen.test.TestService", TargetID: "request-service"},
	{ExecutionID: "request/nomen.test.TestService/TestMethod", TargetID: "request-method"},
}

func TestRouterGet(t *testing.T) {
	router := NewRouter(testTargets)
	targets, ok := router.Get("function/Call")
	assert.True(t, ok)
	assert.Equal(t, testTargets[3:5], targets)

	targets, ok = router.Get("event/foo.missing")
	assert.False(t, ok)
	assert.Nil(t, targets)
}

func TestRouterGetEventBestMatch(t *testing.T) {
	router := NewRouter(testTargets)

	targets, ok := router.GetEventBestMatch("event/foo.bar")
	assert.True(t, ok)
	assert.Equal(t, testTargets[2:3], targets)

	targets, ok = router.GetEventBestMatch("event/foo.baz")
	assert.True(t, ok)
	assert.Equal(t, testTargets[1:2], targets)

	targets, ok = router.GetEventBestMatch("event/other")
	assert.True(t, ok)
	assert.Equal(t, testTargets[0:1], targets)
}
