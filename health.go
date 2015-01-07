package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"strings"
)

type Config struct {
	Recipients []string `json:"recipients"`
	Sender     string   `json:"sender"`
	Smtp       Smtp     `json:"smtp"`
}

type Smtp struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type Result struct {
	Hostname string
	Uname    string
	Uptime   string
	Disk     string
	Memory   string
}

func cmd(cmd string, args ...string) ([]byte, error) {
	path, err := exec.Command("/usr/bin/which", cmd).Output()
	if err != nil {
		return []byte("Unknown"), err
	}
	response, err := exec.Command(strings.TrimSuffix(string(path), "\n"), args...).Output()
	if err != nil {
		response = []byte("Unknown")
	}
	return response, err
}

func conf() (Config, error) {
	if _, err := os.Stat("config.json"); err != nil {
		if os.IsNotExist(err) {
			return Config{}, err
		}
	}

	config, err := ioutil.ReadFile("config.json")
	if err != nil {
		return Config{}, err
	}

	conf := Config{}
	err = json.Unmarshal([]byte(config), &conf)
	if err != nil {
		return Config{}, err
	}
	return conf, err
}

func mail(result Result, conf Config) error {
	tpl, err := template.ParseFiles("template.html")
	if err != nil {
		return err
	}

	var body bytes.Buffer
	err = tpl.ExecuteTemplate(&body, "template.html", result)
	if err != nil {
		return err
	}

	header := make(map[string]string)
	header["From"] = conf.Sender
	header["Subject"] = "Server Status Report"
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString(body.Bytes())

	auth := smtp.PlainAuth(
		"",
		conf.Smtp.User,
		conf.Smtp.Password,
		conf.Smtp.Host,
	)

	err = smtp.SendMail(
		conf.Smtp.Host+":"+conf.Smtp.Port,
		auth,
		conf.Sender,
		conf.Recipients,
		[]byte(message),
	)
	return err
}

func main() {

	printMode := os.Args[1]

	//conf, err := conf()
	//if err != nil {
	//	log.Fatal(err)
	//}

	result := Result{}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "Unknown"
	}
	result.Hostname = hostname

	uname, _ := cmd("uname", "-a")
	result.Uname = string(uname)

	uptime, _ := cmd("uptime")
	result.Uptime = string(uptime)

	disk, _ := cmd("df", "-hlT")
	result.Disk = string(disk)

	memory, _ := cmd("free", "-mo")
	result.Memory = string(memory)

	//err = mail(result, conf)
	//if err != nil {
	//	log.Fatal(err)
	//}
	if printMode == "json" {
		b, err := json.MarshalIndent(result, "", "	")
		if err != nil {
			log.Fatal(err)
			os.Exit(-1)
		} else if b!= nil {
			fmt.Println(string(b))
			os.Exit(0)
		}
	} else {
		fmt.Println("Hostname: ", hostname)
		fmt.Print("\nUname: ", result.Uname)
		fmt.Print("\nUptime: ", result.Uptime)
		fmt.Println("\nDisk: ")
		os.Stdout.Write(disk)
		fmt.Println("\nMemory: ")
		os.Stdout.Write(memory)
		os.Exit(0)
	}

}
