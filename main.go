package main

import (
	"github.com/perfect6566/sftpclient/myconfig"
	"github.com/perfect6566/sftpclient/mysftp"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

func init() {
	if !checkendfix(myconfig.Remotefilespath) {
		myconfig.Remotefilespath += "/"
	}
	if !checkendfix(myconfig.Localfilespath) {
		myconfig.Localfilespath += "/"
	}

}
func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	var (
		pk  []byte
		err error
	)

	if len(myconfig.Rsaprivatekey) != 0 {
		pk, err = ioutil.ReadFile(myconfig.Rsaprivatekey) // required only if private key authentication is to be used
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		pk = []byte("")
	}

	config := mysftp.Config{
		Username:     myconfig.Username,
		Password:     myconfig.Password, // required only if password authentication is to be used
		PrivateKey:   string(pk),        // required only if private key authentication is to be used
		Server:       myconfig.Server,
		KeyExchanges: []string{"diffie-hellman-group-exchange-sha256", "diffie-hellman-group14-sha256"}, // optional
		Timeout:      time.Second * 30,                                                                  // 0 for not timeout
	}

	//wait until the network connection is ready
	ready := make(chan int, 1)
	go func() {

		for counter:=0;counter<10e2;counter++{
			_, err := mysftp.New(config)
			if err != nil {
				log.Println(err)
			}
			if err == nil {
				ready <- 1
				break
			}
			time.Sleep(time.Second)
		}
	}()
	log.Println("Waiting for connection ready...")
	<-ready
	log.Println("Connection is ready")

	client, err := mysftp.New(config)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	//For Upload file

	//// Open local file for reading.
	//source, err := os.Open("file.txt")
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//defer source.Close()

	//// Create remote file for writing.
	//destination, err := client.Create("tmp/file.txt")
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//defer destination.Close()
	//
	//// Upload local file to a remote location as in 1MB (byte) chunks.
	//if err := client.Upload(source, destination, 1000000); err != nil {
	//	log.Fatalln(err)
	//}

	//// For Download remote file.

	//file, err := client.Download("tmp/file.txt")
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//defer file.Close()
	//
	//// Read downloaded file.
	//data, err := ioutil.ReadAll(file)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//fmt.Println(string(data))

	// Get remote file stats.

	//info, err := client.Info("tmp")
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//fmt.Printf("%+v\n", info)

	remotefiles, err := client.SftpClient.ReadDir(myconfig.Remotefilespath)
	if err != nil {
		log.Println(err)
	}

	prefixtobereplace := make([]string, 0)
	prefixtobereplace = append(prefixtobereplace, "?", "\"")
	localfiles := make([]string, 0)
	fs, err := ioutil.ReadDir(myconfig.Localfilespath)
	if err != nil {
		log.Println(err)
	}
	for _, localfile := range fs {
		localfiles = append(localfiles, localfile.Name())
	}

	for _, file := range remotefiles {
		filename := file.Name()

		if IsExist(localfiles, filename) || checkprefix(prefixtobereplace, localfiles, filename) {
			continue
		}

		log.Println("processing " + filename)

		remotefile, err := client.Download(myconfig.Remotefilespath + file.Name())
		if err != nil {
			log.Println(err)
		}
		if remotefile != nil {
			defer remotefile.Close()
		}

		if strings.Contains(filename, "\"") || strings.Contains(filename, "?") {
			filename = strings.ReplaceAll(filename, "?", "_")
			filename = strings.ReplaceAll(filename, "\"", "_")

		}

		newfile, err := os.Create(myconfig.Localfilespath + filename)
		if err != nil {
			log.Println(err)
		}
		if newfile != nil {
			defer newfile.Close()
		}

		io.Copy(newfile, remotefile)
	}
}
func IsExist(flist []string, fname string) bool {
	for _, file := range flist {
		if file == fname {
			return true
		}
	}
	return false
}

func checkprefix(prefixtobereplace, localfiles []string, filename string) bool {
	for _, prefix := range prefixtobereplace {

		if IsExist(localfiles, strings.ReplaceAll(filename, prefix, "_")) {
			return true
		}
	}
	return false
}
func checkendfix(path string) bool {
	l := len(path)
	if l == 0 {
		return false
	}
	if string(path[l-1]) == "/" {

		return true
	}
	return false
}
