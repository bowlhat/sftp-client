package sftpClient

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kr/fs"
)

// Upload upload/synchronize folder trees
func (c *SFTPClient) Upload(folders []FolderMapping) (r <-chan ErrorResponse, count <-chan bool, copied <-chan bool) {
	responseChannel := make(chan ErrorResponse)
	fileStatResponseChannel := make(chan FileResponse)
	countResponseChannel := make(chan bool)
	copiedCountResponseChannel := make(chan bool)

	doneStat := make(chan bool)
	doneCopy := make(chan bool)

	for _, folder := range folders {
		go func(f FolderMapping) {
			defer func() {
				doneStat <- true
			}()

			if _, err := os.Lstat(f.Local); err != nil {
				fileStatResponseChannel <- FileResponse{file: f.Local, Err: fmt.Errorf("Count not access local file: %v", err)}
				return
			}

			var fullLocalPath string
			p, err := filepath.EvalSymlinks(f.Local)
			if err != nil {
				fileStatResponseChannel <- FileResponse{file: f.Local, Err: fmt.Errorf("Could not access local file: %v", err)}
				return
			}

			fullLocalPath = p
			walker := fs.Walk(fullLocalPath)

			for walker.Step() {
				if err := walker.Err(); err != nil {
					fileStatResponseChannel <- FileResponse{file: "", Err: fmt.Errorf("Error: %v", err)}
					continue
				}
				p := walker.Path()
				if p == fullLocalPath {
					continue
				}
				if !strings.HasPrefix(p, f.Remote) {
					continue
				}
				fileStatResponseChannel <- FileResponse{file: p, Err: nil}
				countResponseChannel <- true
			}
		}(folder)

		go func(f FolderMapping) {
			defer func() {
				doneCopy <- true
			}()

			for response := range fileStatResponseChannel {
				if response.Err != nil {
					responseChannel <- ErrorResponse{response.Err}
					continue
				}

				remoteFileName := strings.TrimPrefix(response.file, f.Local)
				remoteFileName = strings.TrimPrefix(remoteFileName, "/")
				remoteFileName = strings.Join([]string{f.Remote, remoteFileName}, "/")

				if err := c.PutFile(remoteFileName, response.file); err != nil {
					responseChannel <- ErrorResponse{err}
				}
				copiedCountResponseChannel <- true
			}
		}(folder)
	}

	go func() {
		defer func() {
			close(doneStat)
			close(countResponseChannel)
			close(fileStatResponseChannel)
		}()
		for range folders {
			<-doneStat
		}
	}()
	go func() {
		defer func() {
			close(doneCopy)
			close(responseChannel)
			close(copiedCountResponseChannel)
		}()
		for range folders {
			<-doneCopy
		}
	}()

	return responseChannel, countResponseChannel, copiedCountResponseChannel
}
