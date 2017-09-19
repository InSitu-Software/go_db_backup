package main

import (
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
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
	pflag.String("logfile", "", "Logile path")
	pflag.String("config", "", "Configfile")
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)

	if viper.GetString("config") != "" {
		viper.SetConfigName("go_db_backup")
		viper.AddConfigPath("config")
		viper.AddConfigPath("config/default")
		err := viper.ReadInConfig()
		if err != nil {
			log.Fatal(err)
		}

	}

	if viper.GetString("logfile") != "" {
		f, err := os.OpenFile(viper.GetString("logfile"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		log.SetOutput(f)
	}
}

func main() {
	var now = time.Now().Local()

	log.WithFields(
		log.Fields{
			"time": now.Format("2006-01-02_15:04:02"),
		},
	).Info("Start Backup Retrival")

	var backupFileName = viper.GetString("dbName") + "_" + now.Format("2006-01-02") + ".7z"
	var backupFilePath = viper.GetString("remoteBasePath") + backupFileName

	privateKey, err := ioutil.ReadFile(viper.GetString("keyFile"))
	if err != nil {
		log.Fatal(err)
	}
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

	log.WithField(
		"host",
		viper.GetString("host")+":"+viper.GetString("port"),
	).Debug("Try to establish connection")
	client, err := ssh.Dial(
		"tcp",
		viper.GetString("host")+":"+viper.GetString("port"),
		config,
	)
	if err != nil {
		log.WithField(
			"host",
			viper.GetString("host")+":"+viper.GetString("port"),
		).Fatal(err)
	}
	log.WithField(
		"time",
		time.Now().Local().Format("2006-01-02_15:04:02"),
	).Debug("Connection established")

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

	log.Debug("Connections / Interfaces created")

	// Create the destination file
	dstFile, err := os.Create(viper.GetString("localBasePath") + backupFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer dstFile.Close()
	log.Debug("Local File touched")

	// Copy the file
	written, err := srcFile.WriteTo(dstFile)
	if err != nil {
		log.Fatal(err)
	}
	log.WithFields(log.Fields{
		"time":      time.Now().Local().Format("2006-01-02_15:04:02"),
		"bytes":     written,
		"localPath": viper.GetString("localBasePath") + backupFileName,
	}).Info("Backup successfully written")
}

func keyCallBack(hostname string, remote net.Addr, key ssh.PublicKey) error {
	return nil
}
