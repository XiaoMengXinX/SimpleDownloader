# SimpleDownloader

[![Go Report Card](https://goreportcard.com/badge/github.com/XiaoMengXinX/SimpleDownloader)](https://goreportcard.com/report/github.com/XiaoMengXinX/SimpleDownloader)
[![](https://pkg.go.dev/badge/github.com/XiaoMengXinX/SimpleDownloader)](https://pkg.go.dev/github.com/XiaoMengXinX/SimpleDownloader)

A simple multi-thread downloader package for Go.

## Example:

```go
package main

import (
	"fmt"
	"time"

	downloader "github.com/XiaoMengXinX/SimpleDownloader"
)

func main() {
	d := downloader.NewDownloader().SetDownloadRoutine(4)

	url := "https://file-examples.com/storage/fe8c7eef0c6364f6c9504cc/2017/04/file_example_MP4_1920_18MG.mp4"
	task, _ := d.NewDownloadTask(url)
	ch := task.ForceMultiThread().DownloadWithChannel()

loop:
	for {
		select {
		case err := <-ch:
			if err != nil {
				panic(err)
			}
			break loop
		default:
			fmt.Printf("Download Speed: %s\n", task.CalculateSpeed(time.Millisecond*200))
		}
	}
}
```

## Documentation

See [pkg.go.dev](https://pkg.go.dev/github.com/XiaoMengXinX/SimpleDownloader)
