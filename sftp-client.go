package sftpClient

import (
	"github.com/pkg/sftp"

	sshclient "github.com/bowlhat/ssh-client"
)

// SFTPClient a proxy to sftp.Client
type SFTPClient struct {
	client     *sftp.Client
	connection *sshclient.SSHConnection
}

// ErrorResponse a response which can only ever be an error
type ErrorResponse struct {
	Err error
}

// FileResponse ...
type FileResponse struct {
	file string
	Err  error
}

// FolderMapping map local to remote folders
type FolderMapping struct {
	Local  string
	Remote string
}

// New Make new SFTP Connection
func New(hostname string, port int, username string, password string) (client *SFTPClient, err error) {
	ssh, err := sshclient.New(hostname, port, username, password)
	if err != nil {
		return nil, err
	}

	sftpClient, err := sftp.NewClient(ssh.Client)
	if err != nil {
		ssh.Close()
		return nil, err
	}

	return &SFTPClient{client: sftpClient, connection: ssh}, nil
}

// Close close the SFTP session
func (c *SFTPClient) Close() error {
	return c.connection.Close()
}
