package webui

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateUsername(t *testing.T) {
	require.NoError(t, validateUsername("abc012"))
	require.Error(t, validateUsername("abc-012"))
	require.Error(t, validateUsername(""))
	require.Error(t, validateUsername("Abc012"))
	require.Error(t, validateUsername("Abc.012"))
}
