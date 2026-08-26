package zerrors_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shippinAI/nomen/internal/zerrors"
)

func TestAlreadyExists(t *testing.T) {
	parentErr := errors.New("parent error")
	id := "test_id"
	message := "test message"

	t.Run("ThrowAlreadyExists", func(t *testing.T) {
		err := zerrors.ThrowAlreadyExists(parentErr, id, message)
		assert.NotNil(t, err)

		nomenErr, ok := zerrors.AsNomenError(err)
		assert.True(t, ok)
		assert.Equal(t, zerrors.KindAlreadyExists, nomenErr.Kind)

		nomenError := new(zerrors.NomenError)
		if errors.As(err, &nomenError) {
			assert.Equal(t, parentErr, nomenError.Unwrap())
			assert.Equal(t, id, nomenError.ID)
			assert.Equal(t, message, nomenError.Message)
		} else {
			t.Errorf("error is not of type NomenError")
		}
	})

	t.Run("ThrowAlreadyExistsf", func(t *testing.T) {
		format := "formatted %s"
		arg := "message"
		expectedMessage := "formatted message"

		err := zerrors.ThrowAlreadyExistsf(parentErr, id, format, arg)
		assert.NotNil(t, err)

		nomenErr, ok := zerrors.AsNomenError(err)
		assert.True(t, ok)
		assert.Equal(t, zerrors.KindAlreadyExists, nomenErr.Kind)

		nomenError := new(zerrors.NomenError)
		if errors.As(err, &nomenError) {
			assert.Equal(t, parentErr, nomenError.Unwrap())
			assert.Equal(t, id, nomenError.ID)
			assert.Equal(t, expectedMessage, nomenError.Message)
		} else {
			t.Errorf("error is not of type NomenError")
		}
	})

	t.Run("IsErrorAlreadyExists", func(t *testing.T) {
		err := zerrors.ThrowAlreadyExists(parentErr, id, message)
		isAlreadyExists := zerrors.IsErrorAlreadyExists(err)
		assert.True(t, isAlreadyExists)

		otherErr := errors.New("some other error")
		isAlreadyExists = zerrors.IsErrorAlreadyExists(otherErr)
		assert.False(t, isAlreadyExists)
	})
}
