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
    "github.com/XiaoMengXinX/SimpleDownloader"
)

func main() {
    d := downloader.NewDownloader().SetDownloadRoutine(4)

    url := "https://xve.me/DemoVideo"
    task, _ := d.NewDownloadTask(url)
    ch := task.ForceMultiThread().SetFileName("demo.mp4").DownloadWithChannel()

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
