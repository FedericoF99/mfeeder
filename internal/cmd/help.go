package cmd

import "fmt"

func help() {
	fmt.Println("Metric Feeder help")
	fmt.Println("available commands:")
	fmt.Println("	mfeeder day <date>")
	fmt.Println("		prints the feed for the given date, if no date is given, prints the feed for today")
	fmt.Println("")
	fmt.Println("	mfeeder ex <option> <value>")
	fmt.Println("		options available:")
	fmt.Println("			- add: adds an exclusion to the config file")
	fmt.Println("			- rm: removes an exclusion from the config file")
	fmt.Println("		exluded processes will not be recorder and printed in the feed")
	fmt.Println("")
	fmt.Println("	mfeeder sys")
	fmt.Println("		toggles the includeSystem flag in the config file, this includes to the watcher the system processes")
	fmt.Println("")
	fmt.Println("	mfeeder get <option>")
	fmt.Println("		available options:")
	fmt.Println("			- ex: prints the current exclusions")
	fmt.Println("			- sys: prints the current includeSystem flag")
	fmt.Println("")
	fmt.Println("	mfeeder --help")
	fmt.Println("		prints this help message")
	fmt.Println("")
}
