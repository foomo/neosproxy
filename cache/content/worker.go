package content

import (
	"container/list"
	"sync"
	"time"

	"github.com/foomo/neosproxy/client/cms"
	"github.com/sirupsen/logrus"
)

var retryWorkerSingleton sync.Once

// runRetryWorker will run a singleton of a retry worker
// it will slow down a job retry and add it to the invalidation queue after some criterias are met
// but it will ensure that a job will be executed until it succeeds or failed more then 50 times in a row
func (c *Cache) runRetryWorker() {
	retryWorkerSingleton.Do(func() {
		go func() {
			tick := time.Tick(10 * time.Second)
			for {
				select {
				case <-tick:
					before := time.Now().Add(-5 * time.Minute)

					if c.retryQueue.Len() > 0 {
						var markedForDeletion bool
						var prev *list.Element
						last := c.retryQueue.Back() // last element

						// loop over the whole queue
						for e := c.retryQueue.Front(); e != nil; e = e.Next() {

							req := e.Value.(InvalidationRequest)

							prev = e.Prev()
							if prev == nil {
								prev = c.retryQueue.Front()
							}

							// remove previous element if marked for deletion
							// we cannot immediately remove it, otherwise we would saw on the branch we sit on
							if prev != nil && markedForDeletion {
								c.retryQueue.Remove(prev)
								markedForDeletion = false
							}

							// less then 5 executions, please try it again
							if req.ExecutionCounter < 5 {
								c.invalidationChannel <- req
								markedForDeletion = true
								continue
							}

							// older then 5 minutes => slow down, but retry
							if req.LastExecutedAt.Before(before) {
								c.invalidationChannel <- req
								markedForDeletion = true
								continue
							}

							// it's the end of the world ...
							if e == last {
								break
							}
						}
						// remove last item if marked for deletion
						if markedForDeletion {
							c.retryQueue.Remove(c.retryQueue.Back())
						}
					}

				case req := <-c.invalidationRetryChannel:
					// add a new job to the end of the line (retry queue)
					c.retryQueue.PushBack(req)
				}
			}
		}()
	})
}

// invalidationWorkers will take care of invalidating the jobs which are in the queue
func (c *Cache) invalidationWorker(id int) {
	for job := range c.invalidationChannel {

		// invalidate
		_, err := c.invalidate(job)

		// well done
		if err == nil {
			continue
		}

		// logger
		l := c.log.WithFields(logrus.Fields{
			"id":         job.ID,
			"dimension":  job.Dimension,
			"retry":      job.ExecutionCounter,
			"createdAt":  job.CreatedAt,
			"modifiedAt": job.LastExecutedAt,
			"waitTime":   time.Since(job.CreatedAt).Seconds(),
		}).WithError(err)

		// too many executions => cancel that job
		if job.ExecutionCounter >= 10 {
			// @todo: inform in slack channel?
			l.Warn("content cache invalidation failed to often - request will be ignored")
			continue
		}

		// unresolvable error => cancel that job
		if err == cms.ErrorNotFound || err == cms.ErrorBadRequest {
			// @todo: inform in slack channel?
			l.Warn("content cache invalidation failed - request will be ignored")
			continue
		}

		// retry
		c.retry(job)
		l.Warn("content cache invalidation failed, retry job added to queue")
	}
}

func (c *Cache) retry(job InvalidationRequest) {
	job.LastExecutedAt = time.Now()
	job.ExecutionCounter++
	c.invalidationRetryChannel <- job
}
