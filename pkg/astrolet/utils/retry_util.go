package utils

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

func isConflictOrServiceUnavailable(err error) bool {
	return errors.IsConflict(err) || errors.IsServiceUnavailable(err)
}

func RetryConflictOrServiceUnavailable(backoff wait.Backoff, fn func() error) error {
	return retry.OnError(backoff, isConflictOrServiceUnavailable, fn)
}
