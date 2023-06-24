/*
[watchdog command manual]
watchdog notifies silent stoppages of the Power Automate + DSView laboratory setup used to record digital PSE frameware logs for debugging.
The watchdog expects the path used to log DSView (digital logic analyzer) output (.dsl files). Crashes will be pushed to the #power-automate-slackbot channel in Presidio Medical Slack.

Usage:
watchdog [PATH]

Options:

	PATH	Absolute or relative path to DSView output directory.
			Default: "$HOME"
*/
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

/*
TODO:
	stop signal : stop signal should be a struct containing the reason for stoppage, enabling main goroutine to generate informative Slack message before exiting
	external SIGTERM : handle termination of main goroutine with special Slack message for catching unforseen crash of this watchdog
	^C exit : handle manual termination/interruption of main goroutine through terminal with special Slack message informing "user quit the program on PM-XXX/USER"
*/

// poller : efficient channel-based polling loop which calls atomic watchdog functions
func poller(target string, stopCh chan struct{}) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		// non-blocking deployment of checks, concurrent interrup of all goroutines if stoppage found
		// TODO: verify deployment of goroutines here yields deterministic behavior
		go freshFiles(target, stopCh)
		go storage(stopCh)
		go dsview(stopCh)
		go powerautomate(stopCh)
		select {
		case <-stopCh:
			return
		default:
		}
	}
}

// age : return the (integer) number of minutes since the modification of the newest file in TARGET
func age(path string) (int, error) {
	var recentModTime time.Time

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return 0, err
	}

	for _, file := range files {
		if file.Mode().IsRegular() {
			modTime := file.ModTime()
			if modTime.After(recentModTime) {
				recentModTime = modTime
			}
		}
	}

	if recentModTime.IsZero() {
		return 0, fmt.Errorf("no files found in the directory")
	}

	minutesAgo := int(time.Since(recentModTime).Round(time.Minute).Minutes())
	return minutesAgo, nil
}

// freshFiles : timeout if new data is not logged in target directory after 1.5 logging cycles (15min) by checking time since most recent file creation in path
func freshFiles(target string, stopCh chan struct{}) {
	if fileAge, err := age(target); err != nil || fileAge > 10 {
		defer close(stopCh)
		return
	}
}

// getFreeSpace : return the (uint64) number of free bytes on the windows system
func getFreeSpace(path string) (uint64, error) {
	fs := syscall.MustLoadDLL("kernel32.dll")
	getDiskFreeSpaceEx := fs.MustFindProc("GetDiskFreeSpaceExW")

	freeBytesAvailable := uint64(0)
	totalBytes := uint64(0)
	totalFreeBytes := uint64(0)

	utfPath, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		log.Fatal("Failed to convert 'path' to UTF16 String in 'getFreeSpace'")
	}

	_, _, err = getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(utfPath)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)

	if err != nil {
		return 0, err
	}

	return freeBytesAvailable, nil
}

// storage : check for remaining storage and throw err if less than 10G is free
func storage(stopCh chan struct{}) {
	tenG := uint64(1000000000)
	if space, err := getFreeSpace("C:\\"); err != nil || space < tenG {
		defer close(stopCh)
		return
	}
}

// dsview : check status of DSView and throw err if DSView is closed
func dsview(stopCh chan struct{}) {
	// TODO
}

// powerautomate : check status of Power Automate and throw err if Power Automate is closed
func powerautomate(stopCh chan struct{}) {
	// TODO
}

// notify : awaits forseen stoppage to notify Slack
func notify(stopCh chan struct{}) {
	<-stopCh
	fmt.Println("Encountered a silent stoppage!")
}

// main : indefinitely awaits the stop signal or external termination and notifies Slack before exiting.
func main() {
	// Decide the directory to watch
	path := os.Args[1]
	TARGET, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get user home directory")
	}
	if path != "" {
		TARGET = os.Args[1]
	}

	// Get current process handle (Windows Process) to address parameters
	handle := windows.CurrentProcess()

	// Set low system priority (Windows PriorityClass) of main goroutine to prevent reentrant errors
	if err := windows.SetPriorityClass(handle, windows.BELOW_NORMAL_PRIORITY_CLASS); err != nil {
		log.Fatal("Failed to set priority class:", err)
	}

	// Create a channel for intentional interrupts and deploy poller on this signal
	stopCh := make(chan struct{})
	go poller(TARGET, stopCh)
	go notify(stopCh)

	// Create a channel to receive os interrupt signals
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt)

	// Wait for an os interrupt signal
	<-osSignals

	// Interrupt the poller goroutine and wait before exiting
	close(stopCh)
	time.Sleep(time.Second)
	fmt.Println("Program interrupted by system.")

}
