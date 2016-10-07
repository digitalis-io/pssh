package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
)

func main() {
	cmd := os.Args[1]
	user := os.Args[2]
	hosts := os.Args[3:]
	results := make(chan string, 1000)

	fmt.Println(os.Getenv("HOME") + "/.ssh/id_rsa.pub")
	config := &ssh.ClientConfig{
		User: user, //os.Getenv("LOGNAME"),
		Auth: []ssh.AuthMethod{PublicKeyFile(os.Getenv("HOME") + "/.ssh/id_rsa")},
	}

	for _, hostname := range hosts {
		go func(hostname string) {
			results <- executeCmd(cmd, hostname, config)
		}(hostname)
	}

	for i := 0; i < len(hosts); i++ {
		select {
		case res := <-results:
			fmt.Print(res)
		}
	}
}

func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func executeCmd(cmd, hostname string, config *ssh.ClientConfig) string {
	conn, err := ssh.Dial("tcp", hostname+":22", config)
	if err != nil {
		return fmt.Sprintf("ERROR: failed to Dial: %s", err)
	}
	session, err := conn.NewSession()
	if err != nil {
		return fmt.Sprintf("ERROR: failed to create session: %s", err)
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		session.Close()
		return fmt.Sprintf("ERROR: request for pseudo terminal failed: %s", err)
	}

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Run(cmd)

	return hostname + ": " + stdoutBuf.String()
}
