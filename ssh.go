package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"

	"golang.org/x/crypto/ssh"
)

type (
	grapeSSH struct {
		keySigner ssh.Signer
	}
	grapeSSHClient struct {
		*ssh.Client
	}
	std struct {
		Out string
		Err string
	}
	sshOutput struct {
		Command command
		Std     std
	}
	sshOutputArray []*sshOutput
	sshError       error
)

func (gSSH *grapeSSH) newError(errMsg string) sshError {
	return errors.New(errMsg)
}

func (gSSH *grapeSSH) setKey(keyPath keyPath) sshError {
	privateBytes, err := ioutil.ReadFile(string(keyPath))
	if err != nil {
		return gSSH.newError("Could not open identity file.")
	}
	privateKey, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return gSSH.newError(fmt.Sprint("Could not parse identity file."))
	}
	gSSH.keySigner = privateKey
	return nil
}

func (gSSH *grapeSSH) newClient(server server, hostKeyCallback ssh.HostKeyCallback) (*grapeSSHClient, sshError) {
	client, err := ssh.Dial("tcp", server.Host, &ssh.ClientConfig{
		User: server.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(gSSH.keySigner),
		},
		HostKeyCallback: hostKeyCallback,
	})
	if err != nil {
		return nil, gSSH.newError(err.Error() + " - could not establish ssh connection")
	}
	return &grapeSSHClient{client}, nil
}

func (client *grapeSSHClient) execCommand(cmd command) *sshOutput {
	output := &sshOutput{
		Command: cmd,
	}
	session, err := client.NewSession()
	if err != nil {
		output.Std.Err = "could not establish ssh session"
	} else {
		var stderr, stdout bytes.Buffer
		session.Stdout, session.Stderr = &stdout, &stderr
		session.Run(string(cmd))
		session.Close()
		output.Std = std{
			Out: stdout.String(),
			Err: stderr.String(),
		}
	}
	return output
}

func (client *grapeSSHClient) execCommands(commands commands) sshOutputArray {
	output := sshOutputArray{}
	for _, command := range commands {
		output = append(output, client.execCommand(command))
	}
	return output
}
