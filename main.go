package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/pkg/sftp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

func init() {
	pflag.String("keyFile", "~/.ssh/id_rsa", "Path to KeyFile used")
	pflag.String("host", "", "Host to connect to")
	pflag.String("port", "22", "SSH Port to use")
	pflag.String("remoteBasePath", "", "Remote Base Path to backup folder")
	pflag.String("localBasePath", "", "Local Base apth for backup")
	pflag.String("user", "", "Username")
	pflag.String("dbName", "insitu", "DBName of Backup")
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)

	viper.SetConfigName("go_db_backup")
	viper.AddConfigPath("config")
	viper.AddConfigPath("config/default")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	var now = time.Now().Local()
	var backupFileName = viper.GetString("dbName") + "_" + now.Format("2006-01-02") + ".7z"
	var backupFilePath = viper.GetString("remoteBasePath") + backupFileName

	privateKey, err := ioutil.ReadFile(viper.GetString("keyFile"))
	signer, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		log.Fatal(err)
	}
	config := &ssh.ClientConfig{
		User:            viper.GetString("user"),
		HostKeyCallback: keyCallBack,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	client, err := ssh.Dial(
		"tcp",
		viper.GetString("host")+":"+viper.GetString("port"),
		config,
	)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}
	fmt.Println("Successfully connected to ssh server.")

	// open an SFTP session over an existing ssh connection.
	sftp, err := sftp.NewClient(client)
	if err != nil {
		log.Fatal(err)
	}
	defer sftp.Close()

	// Open the source file
	srcFile, err := sftp.Open(backupFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(viper.GetString("localBasePath") + backupFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer dstFile.Close()

	// Copy the file
	_, err = srcFile.WriteTo(dstFile)
	if err != nil {
		log.Fatal(err)
	}
}

func keyCallBack(hostname string, remote net.Addr, key ssh.PublicKey) error {
	return nil
}
