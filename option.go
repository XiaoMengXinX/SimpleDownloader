package downloader

import "time"

type Downloader struct {
	SavePath        string
	HttpProxy       HttpProxy
	TimeOut         time.Duration
	UserAgent       string
	DownloadRoutine int
	//BreakPoint    bool
}

type HttpProxy struct {
	Host      string
	isNoProxy bool
}

func (d *Downloader) SetNoProxy() *Downloader {
	d.HttpProxy.isNoProxy = true
	return d
}

func (d *Downloader) SetProxy(host string) *Downloader {
	d.HttpProxy.Host = host
	d.HttpProxy.isNoProxy = false
	return d
}

func (d *Downloader) SetSavePath(path string) *Downloader {
	d.SavePath = path
	return d
}

func (d *Downloader) SetTimeOut(timeout time.Duration) *Downloader {
	d.TimeOut = timeout
	return d
}

func (d *Downloader) SetDownloadRoutine(routine int) *Downloader {
	d.DownloadRoutine = routine
	return d
}

/*
	//TODO: implement
	func (d *Downloader) SetBreakPoint(breakPoint bool) *Downloader {
		d.BreakPoint = breakPoint
		return d
	}
*/
