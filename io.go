package downloader

import (
	"fmt"
	"io"
	"time"
)

func (d *DownloadTask) copy(w io.Writer, r io.Reader) (int64, error) {
	from := func(b []byte) (n int, err error) {
		return r.Read(b)
	}
	to := func(p []byte) (n int, err error) {
		n, err = w.Write(p)
		if n != 0 {
			d.writtenBytes += int64(n)
		}
		return n, err
	}
	return io.Copy(WriterF(to), ReaderF(from))
}

type ReaderF func(b []byte) (n int, err error)

func (f ReaderF) Read(b []byte) (n int, err error) { return f(b) }

type WriterF func(b []byte) (n int, err error)

func (f WriterF) Write(b []byte) (n int, err error) { return f(b) }

// CalculateSpeed returns the download speed of the current task.
func (d *DownloadTask) CalculateSpeed(elapse time.Duration) string {
	writtenBytes := d.GetWrittenBytes()
	time.Sleep(elapse)
	size := d.GetWrittenBytes() - writtenBytes
	res := int64((float64(size) / 1024) / elapse.Seconds())
	if res > 1024 {
		return fmt.Sprintf("%.2f MB/s", float64(res)/1024)
	}
	return fmt.Sprintf("%d KB/s", res)
}
