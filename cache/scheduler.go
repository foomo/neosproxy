package cache

import (
	"time"

	"github.com/foomo/neosproxy/logging"
)

func (c *Cache) scheduleInvalidation() {

	skipped := 0

	// logger
	log := logging.GetDefaultLogEntry().WithField(logging.FieldWorkspace, c.Workspace)

	for {

		select {
		case requestTime := <-c.invalidationChannel:
			log.Info("handle invalidation request from queue")
			time.Sleep(invalidationSleepTime)

			if len(c.invalidationChannel) > 0 && skipped < maxSkipInvalidations {
				skipped++
				log.Info("skip invalidation due to more requests in queue")
				continue
			}

			skipped = 0
			if errInvalidation := c.cacheNeosContentServerExport(); errInvalidation != nil {

				if errInvalidation == ErrorNoNewExort {
					log.WithDuration(requestTime).Info("cache invalidation request processed - but export hash matches old one")
					continue
				}

				log.WithError(errInvalidation).Error("cache invalidation failed")
				continue
			}

			// @todo: notify observers

			log.WithDuration(requestTime).Info("cache invalidation request processed")
		}
	}

}
