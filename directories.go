package sftpClient

import (
	"fmt"
	"os"
	"strings"
)

// Create a folder on remote location
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
			return fmt.Errorf("Error finding 'remote:%s': %v", remoteFolderName, err)
		}
	}
	return nil
}

// Create a folder hierarchy on remote location
func (c *SFTPClient) CreateDirHierarchy(remoteFolderTree string) error {
	tree := strings.Split(remoteFolderTree, "/")
	parent := "."
	for _, dir := range tree {
		parent = strings.Join([]string{parent, dir}, "/")
		if err := c.CreateDir(parent); err != nil {
			return err
		}
	}
	return nil
}
