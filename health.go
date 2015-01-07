package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

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

func main() {

	printMode := "text"

	if len(os.Args) > 1 {
		printMode = os.Args[1]
	}

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
