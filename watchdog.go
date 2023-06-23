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

// fileTimeout : timeout if new data is not logged in target directory after 1.5 logging cycles (15min)
func fileTimeout() {

}

// dsvOpen : poll status of DSView and throw err if DSView is closed
func dsvOpen() {

}

// paOpen : poll status of Power Automate and throw err if Power Automate is closed
func paOpen() {

}

func main() {

}
