package main

import (
	"fmt"
	"github.com/hnakamur/go-scp"
	"github.com/howeyc/gopass"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Server struct {
		User string `yaml:"user"`
		Port int8 `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	Path struct {
		Remote string `yaml:"remote"`
		Local string `yaml:"local"`
	} `yaml:"path"`
}

func main() {
	var cfg Config
	readConfig(&cfg)

	client, err := connect(cfg)
	if err != nil {
		print(err.Error())
		os.Exit(0)
	}

	remoteFolderList := findRemoteDirs(client, cfg.Path.Remote)
	for _, dir := range remoteFolderList {
		remoteFileList := findRemoteFiles(client, dir)
		for _, remoteFileName := range remoteFileList {
			localFileName := cfg.Path.Local + remoteFileName[len(cfg.Path.Remote):]
			if fileNotExists(localFileName) {
				testDir := filepath.Dir(localFileName)
				if fileNotExists(testDir) {
					fmt.Println("Create dir " + testDir)
					err = os.MkdirAll(testDir, 0644)
					if err != nil {
						fmt.Println(err)
					}
				}
				fmt.Println("Copy remote " + remoteFileName + " to local " + localFileName)
				err = scp.NewSCP(client).ReceiveFile(remoteFileName, localFileName)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}
	}
}

func readConfig(cfg *Config) {
	f, err := os.Open("config.yml")
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	//todo тут надо проверить все поля конфига
}

func connect(cfg Config) (*ssh.Client, error){
	println("password:")
	password, _ := gopass.GetPasswd()

	config := &ssh.ClientConfig{
		User: cfg.Server.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(string(password)),
		},
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	return ssh.Dial("tcp", addr, config)
}

func findRemoteFiles(client *ssh.Client, dir string) []string {
	session, err := client.NewSession()
	if err != nil {
		panic(err.Error())
	}
	defer session.Close()
	cmd := fmt.Sprintf("find %s  -maxdepth 1 -type f", dir)
	b, _ := session.CombinedOutput(cmd)
	return strings.Fields(string(b))
}

func findRemoteDirs(client *ssh.Client, remoteFolder string) []string {
	session, err := client.NewSession()
	if err != nil {
		panic(err.Error())
	}
	defer session.Close()
	cmd := fmt.Sprintf("find %s -type d", remoteFolder)
	b, _ := session.CombinedOutput(cmd)
	return strings.Fields(string(b))
}

func fileNotExists(filename string) bool {
	_, err := os.Stat(filename)
	return err != nil && os.IsNotExist(err)
}