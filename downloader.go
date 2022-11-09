package downloader

import (
	"context"
	"net/url"
	"time"
)

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36"

// NewDownloader is to create a new downloader
func NewDownloader() *Downloader {
	return &Downloader{
		SavePath:        "./",
		HttpProxy:       HttpProxy{},
		TimeOut:         60 * time.Second,
		DownloadRoutine: 4,
		UserAgent:       defaultUserAgent,
		BreakPoint:      false,
	}
}

// NewDownloadTask is to create a new download task belongs to the downloader
func (d *Downloader) NewDownloadTask(URL string) (*DownloadTask, error) {
	urlData, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}
	if urlData.Scheme == "" {
		urlData.Scheme = "http"
	}
	return &DownloadTask{
		Downloader: d,
		Scheme:     urlData.Scheme,
		Host:       urlData.Host,
		Path:       urlData.Path,
		Query:      urlData.RawQuery,
		resolvedIP: "",
		fileName:   "",
		client:     nil,
	}, nil
}

// DownloadWithChannel returns a channel to receive the error
func (d *DownloadTask) DownloadWithChannel() (ch chan error) {
	ch = make(chan error, 1)
	go func() {
		ch <- d.Download()
	}()
	return ch
}

// Download starts the download task
func (d *DownloadTask) Download() (err error) {
	if err = d.initClient(); err != nil {
		return err
	}
	if err = d.initFiles(); err != nil {
		return err
	}

	d.writtenBytes = 0

	var ranges [][]int64
	var errChan chan error
	var threads int
	if d.acceptRange && d.Downloader.DownloadRoutine > 1 {
		errChan = make(chan error, d.Downloader.DownloadRoutine)
		ranges = d.splitBytes()
		threads = d.Downloader.DownloadRoutine
	} else {
		errChan = make(chan error, 1)
		ranges = [][]int64{{0, d.fileSize - 1}}
		threads = 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Downloader.TimeOut)

	for i, ranges := range ranges {
		go func(i int, start, end int64) {
			errChan <- d.start(ctx, i, start, end)
		}(i, ranges[0], ranges[1])
	}
	for i := 0; i < threads; i++ {
		select {
		case err := <-errChan:
			if err != nil {
				cancel()
				if !d.Downloader.BreakPoint {
					d.CleanTempFiles()
				} else {
					d.closeTempFiles()
				}
				return err
			}
		}
	}
	defer cancel()

	return d.mergeFiles()
}
