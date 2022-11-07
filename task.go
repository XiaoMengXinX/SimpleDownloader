package downloader

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type DownloadTask struct {
	Downloader   *Downloader
	Scheme       string
	Host         string
	Path         string
	Query        string
	fileName     string
	url          string
	headerHost   string
	resolvedIP   string
	savePath     string
	fileSize     int64
	writtenBytes int64
	acceptRange  bool
	ignoreCert   bool
	client       *http.Client
	tempFiles    []*os.File
}

// SetFileName set the file name of the task
func (d *DownloadTask) SetFileName(fileName string) *DownloadTask {
	d.fileName = fileName
	return d
}

// ForceMultiThread force to use multi-thread download without checking the
// Accept-Ranges header
func (d *DownloadTask) ForceMultiThread() *DownloadTask {
	d.acceptRange = true
	return d
}

// ForceHttps set the scheme to https
func (d *DownloadTask) ForceHttps() *DownloadTask {
	d.Scheme = "https"
	return d
}

// ForceHttp set the scheme to http
func (d *DownloadTask) ForceHttp() *DownloadTask {
	d.Scheme = "http"
	return d
}

// WithResolvedIp will force to make http request with the specific IP address
// like the "--resolve" option in curl.
//
// For example:
//
//	d := downloader.NewDownloader()
//	task, _ := d.NewDownloadTask("https://www.example.com/1.jpg")
//	task.WithResolvedIp("114.5.1.4").Download()
//
// This is similar to:
//
//	curl --resolve www.example.com:443:114.5.1.4 https://www.example.com/1.jpg
func (d *DownloadTask) WithResolvedIp(ip string) *DownloadTask {
	d.resolvedIP = ip
	return d
}

// WithResolvedIpOnHost is similar to WithResolvedIp, but instead of replacing
// the dns record, it replaces the host name in url and add the original host
// name to the header.
//
// For example:
//
//	d := downloader.NewDownloader()
//	task, _ := d.NewDownloadTask("https://www.example.com/1.jpg")
//	task.WithResolvedIp("114.5.1.4:23333").Download()
//
// It is similar to:
//
//	curl -H "Host: www.example.com" https://114.5.1.4:23333/1.jpg
//
// This can be used to download a file through your own hosted reverse proxy
// server.
//
// Notice that if you want to request with a specific port, there is only
// WithResolvedIpOnHost seems to work. And this option will default ignore
// the SSL certificate verification.
func (d *DownloadTask) WithResolvedIpOnHost(ip string) *DownloadTask {
	d.AddHostNameToHeader(d.Host)
	d.ReplaceHostName(ip)
	return d
}

// ReplaceHostName replace the host name in url
func (d *DownloadTask) ReplaceHostName(host string) *DownloadTask {
	d.Host = host
	return d
}

// AddHostNameToHeader add the host name to the header
func (d *DownloadTask) AddHostNameToHeader(host string) *DownloadTask {
	d.headerHost = host
	return d
}

// IgnoreCertificateVerify ignore the SSL certificate verification
func (d *DownloadTask) IgnoreCertificateVerify() *DownloadTask {
	d.ignoreCert = true
	return d
}

// GetHostName return the real host name of the task
func (d *DownloadTask) GetHostName() string {
	if d.headerHost != "" {
		return d.headerHost
	}
	return d.Host
}

// GetWrittenBytes return the downloaded bytes number of the task
func (d *DownloadTask) GetWrittenBytes() int64 {
	return d.writtenBytes
}

// GetFileSize return the content length of the task
func (d *DownloadTask) GetFileSize() int64 {
	return d.fileSize
}

func (d *DownloadTask) initClient() (err error) {
	if d.client == nil {
		d.client = &http.Client{}
	}
	transport := http.Transport{}

	if !d.Downloader.HttpProxy.isNoProxy {
		transport.Proxy = http.ProxyFromEnvironment
	}
	if d.Downloader.HttpProxy.Host != "" {
		proxyURL, err := url.Parse(d.Downloader.HttpProxy.Host)
		if err != nil {
			return err
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	if d.resolvedIP != "" {
		dialFuncWithCtx := func(ctx context.Context, network, addr string) (net.Conn, error) {
			var dialer net.Dialer
			s := strings.Split(addr, ":")
			s[0] = strings.Split(d.resolvedIP, ":")[0]
			return dialer.DialContext(ctx, network, strings.Join(s, ":"))
		}
		transport.DialContext = dialFuncWithCtx
		transport.DialTLSContext = dialFuncWithCtx
	}
	if d.headerHost != "" {
		d.IgnoreCertificateVerify()
	} else {
		d.headerHost = d.Host
	}
	if d.ignoreCert {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	d.client.Transport = &transport
	d.url = d.Scheme + "://" + d.Host + d.Path
	if d.Query != "" {
		d.url += "?" + d.Query
	}

	return nil
}

func (d *DownloadTask) splitBytes() [][]int64 {
	var ranges [][]int64

	threads := int64(d.Downloader.DownloadRoutine)
	blockSize := d.fileSize / threads

	for i := int64(0); i < threads; i++ {
		var start = i * blockSize
		var end = (i+1)*blockSize - 1
		if i == threads-1 {
			end = d.fileSize - 1
		}
		ranges = append(ranges, []int64{start, end})
	}

	return ranges
}

func (d *DownloadTask) start(ctx context.Context, i int, start, end int64) (err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", d.url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	if d.Downloader.UserAgent != "" {
		req.Header.Set("User-Agent", d.Downloader.UserAgent)
	}
	if d.headerHost != "" {
		req.Host = d.headerHost
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = d.copy(d.tempFiles[i], resp.Body)
	return err
}
