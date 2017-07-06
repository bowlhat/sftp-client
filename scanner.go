package sftpClient

import (
	"fmt"
	"log"
)

// Find Remote Files
func (c *SFTPClient) FindRemoteFiles(path string) func() (r <-chan FileResponse) {
	return func() (r <-chan FileResponse) {
		responseChannel := make(chan FileResponse)

		go func() {
			defer close(responseChannel)
			stats, err := c.client.Lstat(path)
			if err != nil {
				responseChannel <- FileResponse{File: "", Err: fmt.Errorf("Cannot STAT 'remote:%s': %v", path, err)}
				return
			}
			if !stats.IsDir() {
				responseChannel <- FileResponse{File: "", Err: fmt.Errorf("'remote:%s' is not a directory", path)}
				return
			}

			var walker = *c.client.Walk(path)
			for walker.Step() {
				if err := walker.Err(); err != nil {
					responseChannel <- FileResponse{File: "", Err: err}
					continue
				}
				if walker.Path() == path {
					continue
				}
				responseChannel <- FileResponse{File: walker.Path(), Err: nil}
			}
		}()

		return responseChannel
	}
}

func findRemoteFilesAggregator(functions []func() (r <-chan FileResponse)) (r <-chan FileResponse) {
	responseChannel := make(chan FileResponse)
	done := make(chan bool)

	for _, function := range functions {
		go func(function func() (r <-chan FileResponse)) {
			intermediateChannel := function()
			for response := range intermediateChannel {
				responseChannel <- response
			}
			done <- true
		}(function)
	}

	go func() {
		for range functions {
			<-done
		}
		close(responseChannel)
		close(done)
	}()

	return responseChannel
}

// Find All Remote Files
func (c *SFTPClient) FindAllRemoteFiles(paths []string) ([]string, error) {
	var functions []func() (r <-chan FileResponse)
	var files []string

	for _, path := range paths {
		functions = append(functions, c.FindRemoteFiles(path))
	}

	responseChannel := findRemoteFilesAggregator(functions)
	encounteredErrors := 0
	for response := range responseChannel {
		if response.Err != nil {
			encounteredErrors++
			log.Println(response.Err)
		}
		if encounteredErrors == 0 {
			files = append(files, response.File)
		}
	}

	if encounteredErrors > 0 {
		return nil, fmt.Errorf("Encountered %d errors", encounteredErrors)
	}
	return files, nil
}
