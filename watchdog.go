/*
[watchdog command manual]
watchdog notifies silent stoppages of the Power Automate + DSView laboratory setup used to record digital PSE frameware logs for debugging.
The watchdog expects the path used to log DSView (digital logic analyzer) output (.dsl files). Crashes will be pushed to the #power-automate-slackbot channel in Presidio Medical Slack.

Usage:
watchdog [--path PATH]

Options:

	--path	Absolute or relative path to DSView output directory.
			Default: "$HOME"
*/
package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"
)

/*
TODO:
	stop signal : stop signal should be a struct containing the reason for stoppage, enabling main goroutine to generate informative Slack message before exiting
	external SIGTERM : handle termination of main goroutine with special Slack message for catching unforseen crash of this watchdog
	^C exit : handle manual termination/interruption of main goroutine through terminal with special Slack message informing "user quit the program on PM-XXX/USER"
*/

// poller : efficient channel-based polling loop which calls atomic watchdog functions
func poller(stopCh chan struct{}) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		freshFiles()
		storage()
		dsview()
		powerautomate()
		select {
		case <-stopCh:
			return
		default:
		}
	}
}

// freshFiles : timeout if new data is not logged in target directory after 1.5 logging cycles (15min) by checking time since most recent file creation in path
func freshFiles() {

}

// storage : check for remaining storage and throw err if less than 10G is free
func storage() {

}

// dsview : check status of DSView and throw err if DSView is closed
func dsview() {

}

// powerautomate : check status of Power Automate and throw err if Power Automate is closed
func powerautomate() {

}

// main : indefinitely awaits the stop signal or external termination and notifies Slack before exiting.
func main() {
	stopCh := make(chan struct{})
	go poller(stopCh)

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
