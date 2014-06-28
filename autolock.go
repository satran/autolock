package main

import (
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

const (
	// To establish initial connection
	rfcomm string = "/usr/bin/rfcomm"

	// To ping for the connection signal
	hcitool string = "/usr/bin/hcitool"

	// To lock the system
	lock string = "/usr/local/bin/slock"

	// Phone/Bluetooth device address
	addr = "B0:EC:71:E1:DA:49"

	channel = "2"
	device  = "0"

	PING_TIME = 10
	CONN_TIME = 10
	
	PROXIMITY = -1
)

var connected = false

func main() {
	go connect()

	var lock_cmd *exec.Cmd
	locked := false

	r, _ := regexp.Compile("-?[0-9]+")

	for {
		command := exec.Command(hcitool, "rssi", addr)
		if !connected {
			// We wait for the connection to become active again.
			// Most likely the phone went out of range.
			time.Sleep(CONN_TIME * time.Second)
			continue
		}

		res, err := command.Output()
		if err != nil {
			log.Println(err)
			// Ignore error and try again
			continue
		}

		out := string(res)
		if out == "" {
			log.Println("Could not find device.")
			continue
		}

		proximity, _ := strconv.Atoi(r.FindString(out))

		if proximity >= PROXIMITY && locked {
			if lock_cmd == nil {
				// Most likely the command finished executing.
				// I might have unlocked the machine manually.
				locked = false
				time.Sleep(PING_TIME * time.Second)
				continue
			}

			err := lock_cmd.Process.Kill()
			if err != nil {
				log.Println(err)
			} else {
				locked = false
			}
		
		} else if !locked {
			lock_cmd = exec.Command(lock)
			err := lock_cmd.Start()
			if err != nil {
				log.Println(err)
			}
			locked = true
		}

		time.Sleep(PING_TIME * time.Second)
	}
}

func connect() {
	for {
		command := exec.Command(rfcomm, "connect", device, addr, channel)
		err := command.Start()
		if err != nil {
			log.Fatal(err)
		}
		connected = true
		log.Println("Connected...")
		
		err = command.Wait()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Disconnected")
		
		connected = false
		time.Sleep(CONN_TIME * time.Second)
	}
}
