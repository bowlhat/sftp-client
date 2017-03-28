package sftpClient

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// BackupFiles backup remote files
func (c *SFTPClient) BackupFiles(destination string, files []string) (saved <-chan savedCountChannel, r <-chan ErrorResponse) {
	responseChannel := make(chan ErrorResponse)
	savedChannel := make(chan bool)

	go func() {
		defer func() {
			close(responseChannel)
			close(savedChannel)
		}()
		var err error
		if _, err = os.Stat(destination); err != nil {
			if err2 := os.Mkdir(destination, 755); err2 != nil {
				responseChannel <- ErrorResponse{Err: fmt.Errorf("Cannot access folder: %s; Failed to create folder: %s", err, err2)}
			}
		}

		// create a file and get a handle to write gzipped data to
		tarfile := fmt.Sprintf("%s/%s.tar.gz", destination, time.Now().UTC().Format("20060102-150405"))
		var zbuf *os.File
		if zbuf, err = os.Create(tarfile); err != nil {
			responseChannel <- ErrorResponse{err}
		}
		defer func() {
			if err := zbuf.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		// set up the gzip intermediary backing onto the file above
		gzw := gzip.NewWriter(zbuf)
		defer func() {
			if err := gzw.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		// get a handle to a tar writer instance using the gzip intermediate buffer above
		tw := tar.NewWriter(gzw)
		defer func() {
			if err := tw.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		for _, filename := range files {
			if err := c.TarFile(tw, filename); err != nil {
				responseChannel <- ErrorResponse{err}
			}
			savedChannel <- true
		}
	}()

	return responseChannel
}

// TarFile add a remote file into tar archive
func (c *SFTPClient) TarFile(w *tar.Writer, filename string) error {
	f, err := c.client.Open(filename)
	if err != nil {
		return fmt.Errorf("Could not open file 'remote:%s': %v", filename, err)
	}
	defer f.Close()

	// create a tar header for this file
	stats, err := f.Stat()
	if err != nil {
		return fmt.Errorf("Could not stat file 'remote:%s': %v", filename, err)
	}
	hdr, err := tar.FileInfoHeader(stats, "")
	if err != nil {
		return fmt.Errorf("Could not create tar header for 'remote:%s': %v", filename, err)
	}
	name := strings.TrimPrefix(filename, "htdocs/sites/all")
	name = strings.TrimPrefix(name, "htdocs/sites/default")
	hdr.Name = name

	// write the header into our tar stream
	if err := w.WriteHeader(hdr); err != nil {
		return fmt.Errorf("Could not write tar header for 'remote:%s': %v", filename, err)
	}

	if !stats.IsDir() {
		// write the actual file to the tar stream
		if _, err := io.Copy(w, f); err != nil {
			return fmt.Errorf("Could not write tar data for 'remote:%s': %v", filename, err)
		}
	}

	return nil
}
