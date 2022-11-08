package downloader

import "time"

type Downloader struct {
	SavePath        string
	HttpProxy       HttpProxy
	TimeOut         time.Duration
	UserAgent       string
	DownloadRoutine int
	BreakPoint      bool
}

type HttpProxy struct {
	Host      string
	isNoProxy bool
}

// SetNoProxy disable the proxy settings from environment
func (d *Downloader) SetNoProxy() *Downloader {
	d.HttpProxy.isNoProxy = true
	return d
}

// SetProxy set the proxy settings
func (d *Downloader) SetProxy(host string) *Downloader {
	d.HttpProxy.Host = host
	d.HttpProxy.isNoProxy = false
	return d
}

// SetUserAgent set the user agent
func (d *Downloader) SetUserAgent(userAgent string) *Downloader {
	d.UserAgent = userAgent
	return d
}

// SetSavePath set the path to save files
func (d *Downloader) SetSavePath(path string) *Downloader {
	d.SavePath = path
	return d
}

// SetTimeOut set the context timeout
func (d *Downloader) SetTimeOut(timeout time.Duration) *Downloader {
	d.TimeOut = timeout
	return d
}

// SetDownloadRoutine set the threads to download
func (d *Downloader) SetDownloadRoutine(routine int) *Downloader {
	d.DownloadRoutine = routine
	return d
}

// SetBreakPoint set the break point download
func (d *Downloader) SetBreakPoint(isEnabled bool) *Downloader {
	d.BreakPoint = isEnabled
	return d
}
