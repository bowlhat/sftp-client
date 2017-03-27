package sftpClient

import (
	"fmt"
	"os"
	"strings"
)

// CreateDir create a folder on remote location
func (c *SFTPClient) CreateDir(remoteFolderName string) error {
	if _, err := c.client.Lstat(remoteFolderName); err != nil {
		if os.IsNotExist(err) {
			if err := c.client.Mkdir(remoteFolderName); err != nil {
				return fmt.Errorf("Could not create folder 'remote:%s': %v", remoteFolderName, err)
			}
			if err := c.client.Chmod(remoteFolderName, 0755); err != nil {
				return fmt.Errorf("Could not set folder permissions on 'remote:%s': %v", remoteFolderName, err)
			}
		} else {
			return fmt.Errorf("Could not create folder 'remote:%s': %v", remoteFolderName, err)
		}
	}
	return nil
}

// CreateDirHierarchy create all folders in tree on remote location
func (c *SFTPClient) CreateDirHierarchy(remoteFolderTree string) error {
	tree := strings.Split(remoteFolderTree, "/")
	parent := "."
	for _, dir := range tree {
		if err := c.CreateDir(fmt.Sprintf("%s/%s", parent, dir)); err != nil {
			return err
		}
	}
	return nil
}
