package utils

import (
	"testing"

	"github.com/stretchr/testify/require"

	"transcode-demo/pkg/cerrors"
)

func AsCError(t testing.TB, err error) cerrors.CError {
	var cErr *cerrors.CError
	require.ErrorAs(t, err, &cErr)
	return *cErr
}
