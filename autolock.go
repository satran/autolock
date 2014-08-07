package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var (
	connected = false
	connMtx   sync.Mutex

	// Phone/Bluetooth device address
	addr = ""
)

func main() {
	fRfcomm := flag.String("rfcomm", "/usr/bin/rfcomm", "rfcomm executable")
	fHcitool := flag.String("hcitool", "/usr/bin/hcitool", "hcitool executable")
	fLock := flag.String("lock", "slock", "locker executable")
	fChannel := flag.String("channel", "2", "channel")
	fDevice := flag.String("device", "0", "device")
	fPingTime := flag.Duration("ping", 10*time.Second, "ping time")
	fConnTime := flag.Duration("conn", 10*time.Second, "conn time")
	fProximity := flag.Int("proximity", -1, "proximity")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: "+os.Args[0]+" [options] <device:mac:addr>")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() == 1 {
		addr = flag.Arg(0)
	} else if addr == "" {
		flag.Usage()
		os.Exit(1)
	}

	go connect(*fRfcomm, *fDevice, addr, *fChannel, *fConnTime)

	var lockCmd *exec.Cmd
	locked := false

	r := regexp.MustCompile("-?[0-9]+")

	for {
		command := exec.Command(*fHcitool, "rssi", addr)
		{
			connMtx.Lock()
			connected := connected
			connMtx.Unlock()
			if !connected {
				// We wait for the connection to become active again.
				// Most likely the phone went out of range.
				time.Sleep(*fConnTime)
				continue
			}
		}

		res, err := command.CombinedOutput()
		if err != nil {
			log.Printf("%s: %v (%s)", command.Args, err, res)
			// Ignore error and try again
			time.Sleep(1 * time.Second)
			continue
		}

		if len(res) == 0 {
			log.Println("Could not find device.")
			continue
		}
		out := string(res)

		proximity, _ := strconv.Atoi(r.FindString(out))

		if proximity >= *fProximity && locked {
			if lockCmd == nil {
				// Most likely the command finished executing.
				// I might have unlocked the machine manually.
				locked = false
				time.Sleep(*fPingTime)
				continue
			}

			err := lockCmd.Process.Kill()
			if err != nil {
				log.Println(err)
			} else {
				locked = false
			}
		} else if !locked {
			lockCmd = exec.Command(*fLock)
			err := lockCmd.Start()
			if err != nil {
				log.Println(err)
			}
			locked = true
		}

		time.Sleep(*fPingTime)
	}
}

func connect(rfcomm, device, addr, channel string, connTime time.Duration) {
	for {
		command := exec.Command(rfcomm, "connect", device, addr, channel)
		connMtx.Lock()
		connected = true
		connMtx.Unlock()
		res, err := command.CombinedOutput()
		if err != nil {
			log.Fatalf("%s: %v (%s)", command.Args, err, res)
		}
		log.Println("Disconnected")

		connMtx.Lock()
		connected = false
		connMtx.Unlock()
		time.Sleep(connTime)
	}
}
