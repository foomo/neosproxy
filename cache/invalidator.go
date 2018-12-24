package cache

import (
	"errors"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cloudfoundry/bytefmt"
	"github.com/foomo/neosproxy/logging"
)

// Invalidate cache maybe invalidates cache, but skips requests if invalidation queue is already full
func (c *Cache) Invalidate() bool {

	// logger
	log := logging.GetDefaultLogEntry().WithField(logging.FieldWorkspace, c.Workspace)

	select {
	case c.invalidationChannel <- time.Now():
		log.Info("invalidation request added to queue")
		return true
	default:
		log.Info("invalidation request ignored, queue seems to be full")
		return false
	}

}

// cacheNeosContentServerExport ...
func (c *Cache) cacheNeosContentServerExport() error {

	start := time.Now()

	// logger
	log := logging.GetDefaultLogEntry().WithField(logging.FieldWorkspace, c.Workspace)

	// download new export
	downloadFilename := c.file + ".download"
	neosContentServerExportURL := c.neos.URL.String() + "/contentserver/export?workspace=" + c.Workspace
	if err := downloadNeosContentServerExport(downloadFilename, neosContentServerExportURL); err != nil {
		return err
	}

	// calc hash of existing file
	hashOld, errHashOld := hashFile(c.file)
	if errHashOld != nil {
		return errHashOld
	}

	// calc hash of new file
	hashNew, errHashNew := hashFile(downloadFilename)
	if errHashNew != nil {
		return errHashNew
	}

	// logging
	log.WithFields(logrus.Fields{
		"hashNew": hashNew,
		"hashOld": hashOld,
	}).Debug("content server export hashed")

	if hashNew == hashOld {
		return ErrorNoNewExort
	}

	if errRename := os.Rename(downloadFilename, c.file); errRename != nil {
		return errRename
	}

	fileInfo, errFileInfo := os.Stat(c.file)
	if errFileInfo != nil {
		return errFileInfo
	}

	// empty file => remove cached file
	if fileInfo.Size() == 0 {
		errRemove := os.Remove(c.file)
		if errRemove != nil {
			return errRemove
		}
		return errors.New("cache: download failed, empty file")
	}

	// notify broker
	c.broker.NotifyOnSitemapChange(c.Workspace) // user ???

	log.WithDuration(start).WithField("size", bytefmt.ByteSize(uint64(fileInfo.Size()))).Debug("cached a new contentserver export")
	return nil
}
