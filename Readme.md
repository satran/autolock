## Requirements:

- Go
- rfcomm
- hcitools

## Installation

    go get github.com/tgulacsi/autolock

If you want to compile in your phone's address, then you can do it with

	go build -ldflags '-X main.addr the:addr' github.com/tgulacsi/autolock

If you don't do this, you'll have to add the address for each invocation of autolock.


## Enabling Bluetooth
You can use `rfkill` to enable your bluetooth device if it is blocked:
`sudo rfkill list` to list the devices, and `sudo rfkill unblock hci0` tu unblock.

If `hciconfig` says `DOWN` in the third line, the device can be started with
`sudo hciconfig hci0 up`. After this `hcitool dev` shall list the device.

## Getting phone's bluetooth address
Use ``hcitool scan` AFTER you've *enabled* scanning on your phone.

## Errors
`Can't open RFCOMM control socket: Protocol not supported`: `sudo modprobe -v rfcomm`
