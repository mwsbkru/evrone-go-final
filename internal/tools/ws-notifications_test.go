package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserStreamName(t *testing.T) {
	result := GetUserStreamName("john")
	assert.Equal(t, "notifications:john", result)
}

func TestGetUserLastReadedNotificationID(t *testing.T) {
	result := GetUserLastReadedNotificationID("john")
	assert.Equal(t, "last-readed-notification--john", result)
}
