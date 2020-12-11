package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rakyll/portmidi"
)

var (
	oDevice = flag.Int("device", 1, "MIDI device id for output")
	oPort   = flag.Int("port", 9000, "UDP listening port")
	oDebug  = flag.Bool("d", false, "debug logging")
)

// start for a specific outputdevice
func main() {
	if len(os.Args) == 0 {
		showAvailable()
		return
	}
	flag.Parse()

	// midi interface
	err := portmidi.Initialize()
	if err != nil {
		log.Fatalln(err)
	}
	defer portmidi.Terminate()

	// report what we have
	showAvailable()

	// wait for control-C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)
	go func() {
		for sig := range c {
			fmt.Println("\nstopped because:", sig)
			os.Exit(0)
		}
	}()

	// accept
	log.Printf("\033[1;33mwaiting\033[0m for client on :%d", *oPort)
	lis, err := newUDPToMIDIListener(*oPort, *oDevice)
	if err != nil {
		log.Fatalln(err)
	}
	defer lis.close()

	// forward
	info := portmidi.Info(portmidi.DeviceID(*oDevice))
	log.Printf("\033[1;33mlistening\033[0m for MIDI, forwarding to device %s/%s (press CTR-C to abort)\n", info.Interface, info.Name)
	lis.start()
}

func showAvailable() {
	fmt.Println("\033[1;33mavailable devices:\033[0m")
	var midiDeviceInfo *portmidi.DeviceInfo
	for i := 0; i < portmidi.CountDevices(); i++ {
		midiDeviceInfo = portmidi.Info(portmidi.DeviceID(i)) // returns info about a MIDI device
		fmt.Printf("[midi] device %d: ", i)
		usage := "output"
		if midiDeviceInfo.IsInputAvailable {
			usage = "input"
		}
		oc := "open"
		if !midiDeviceInfo.IsOpened {
			oc = "closed"
		}
		fmt.Print("\"", midiDeviceInfo.Interface, "/", midiDeviceInfo.Name, "\"",
			", is ", oc, " for ", usage)
		fmt.Println()
	}
}
