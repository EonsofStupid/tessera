package zerrors_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shippinAI/nomen/internal/zerrors"
)

func TestPermissionDenied(t *testing.T) {
	parentErr := errors.New("parent error")
	id := "PERMISSION_DENIED"
	message := "you do not have permission"

	t.Run("ThrowPermissionDenied", func(t *testing.T) {
		err := zerrors.ThrowPermissionDenied(parentErr, id, message)
		assert.NotNil(t, err)

		nomenErr, ok := zerrors.AsNomenError(err)
		assert.True(t, ok)
		assert.Equal(t, zerrors.KindPermissionDenied, nomenErr.Kind)

		nomenError := new(zerrors.NomenError)
		if errors.As(err, &nomenError) {
			assert.Equal(t, parentErr, nomenError.Unwrap())
			assert.Equal(t, id, nomenError.ID)
			assert.Equal(t, message, nomenError.Message)
		} else {
			t.Errorf("error is not of type NomenError")
		}
	})

	t.Run("ThrowPermissionDeniedf", func(t *testing.T) {
		format := "formatted %s"
		arg := "message"
		expectedMessage := "formatted message"

		err := zerrors.ThrowPermissionDeniedf(parentErr, id, format, arg)
		assert.NotNil(t, err)

		nomenErr, ok := zerrors.AsNomenError(err)
		assert.True(t, ok)
		assert.Equal(t, zerrors.KindPermissionDenied, nomenErr.Kind)

		nomenError := new(zerrors.NomenError)
		if errors.As(err, &nomenError) {
			assert.Equal(t, parentErr, nomenError.Unwrap())
			assert.Equal(t, id, nomenError.ID)
			assert.Equal(t, expectedMessage, nomenError.Message)
		} else {
			t.Errorf("error is not of type NomenError")
		}
	})

	t.Run("IsPermissionDenied", func(t *testing.T) {
		err := zerrors.ThrowPermissionDenied(parentErr, id, message)
		isPermissionDenied := zerrors.IsPermissionDenied(err)
		assert.True(t, isPermissionDenied)
	})
}
