package downloader

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

func (d *DownloadTask) initFiles() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "HEAD", d.url, nil)
	if err != nil {
		return err
	}

	if d.Downloader.UserAgent != "" {
		req.Header.Set("User-Agent", d.Downloader.UserAgent)
	}
	if d.headerHost != "" {
		req.Host = d.headerHost
	}
	for k, v := range d.headers {
		if s := strings.ToLower(k); s == "user-agent" || s == "host" {
			continue
		}
		req.Header.Set(k, v)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("%s", resp.Status)
	}

	d.setFileInfo(resp)
	d.savePath = fmt.Sprintf("%s/%s", d.Downloader.SavePath, d.fileName)

	return d.openTempFiles()
}

func (d *DownloadTask) setFileInfo(r *http.Response) {
	if d.fileName == "" {
		disposition := r.Header.Get("Content-Disposition")
		if disposition != "" {
			var re = regexp.MustCompile(`(?m)filename="(.*)"`)
			list := re.FindAllStringSubmatch(disposition, 100)
			if len(list) > 0 && len(list[0]) >= 1 {
				d.fileName = list[0][1]
			}
		} else {
			d.fileName = path.Base(d.Path)
		}
	}
	if r.Header.Get("Accept-Ranges") != "" {
		d.acceptRange = true
	}
	d.fileSize = r.ContentLength
}

func (d *DownloadTask) openTempFiles() (err error) {
	var tempFiles []*os.File
	if d.acceptRange {
		for i := 0; i < d.Downloader.DownloadRoutine; i++ {
			tempFile, err := os.OpenFile(fmt.Sprintf("%s.%d", d.savePath, i), os.O_CREATE|os.O_RDWR, 0666)
			if err != nil {
				return err
			}
			tempFiles = append(tempFiles, tempFile)
		}
	} else {
		tempFile, err := os.Create(d.savePath)
		if err != nil {
			return err
		}
		tempFiles = append(tempFiles, tempFile)
	}
	d.tempFiles = tempFiles
	return nil
}

func (d *DownloadTask) mergeFiles() (err error) {
	if !d.acceptRange {
		return nil
	}

	file, err := os.Create(d.savePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, f := range d.tempFiles {
		_, err = f.Seek(0, 0)
		if err != nil {
			return err
		}
		_, err = file.ReadFrom(f)
		if err != nil {
			return err
		}
	}
	d.CleanTempFiles()

	return
}

// CleanTempFiles delete the temp files of the task
func (d *DownloadTask) CleanTempFiles() {
	d.closeTempFiles()
	for _, f := range d.tempFiles {
		os.Remove(f.Name())
	}
}

func (d *DownloadTask) closeTempFiles() {
	for _, f := range d.tempFiles {
		f.Close()
	}
}
