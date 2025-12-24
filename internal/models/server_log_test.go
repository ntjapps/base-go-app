package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerLog_TableName(t *testing.T) {
	s := ServerLog{}
	assert.Equal(t, "log", s.TableName())
}
