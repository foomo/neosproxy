package content

import "errors"

// ErrorNotFound error in case of no cache hit
var ErrorNotFound = errors.New("cache item not found")

// ErrorInvalidationRejectedQueueExhausted error in case invalidation queue is full
var ErrorInvalidationRejectedQueueExhausted = errors.New("invalidation request rejected: invalidation queue capacity exhausted")
