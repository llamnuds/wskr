package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mu sync.Mutex // Declare a Mutex
var countGood = 0
var countBad = 0
var argSummary = false
var usingFile = false
var computerList []string
var fileOfCompNames = ""

type result struct {
	rTime   time.Time
	rResult bool
}

var results []result

func main() {
	wg := new(sync.WaitGroup)

	// Set DEFAULT argument values
	var argStart int = -1
	var argEnd int = -1
	var argShow string = "1"
	var argShowGood bool = true
	var argShowBad bool = false
	var argPad string = "000"
	var argDelay int = 0
	var argPrefix string = ""
	var argAction string = "Nothing"
	var argItem string = ""
	var argSave string = ""
	var argTimings bool = false

	// Used letters - 3defgmnprstuvwxy

	if len(os.Args) == 1 {
		printHelp()
	} else {

		// Determine arguments
		var finalArg []string
		for _, v := range os.Args[1:] {
			if v[0] == "-"[0] {
				v3 := strings.Split(strings.ToLower(v), "=")
				// An argument has been found
				// Let's see what it is and set values appropriately

				if v3[0] == "--range" || v3[0] == "-n" {
					// split on "-" into start and end args
					argRange := strings.Split(v3[1], "..")
					switch len(argRange) {
					case 0:
						// No parameters
						// That's not correct!
						// Ignore it.
					case 1:
						// One parameter, just one machine to check, OR this is a filename.

						// Is this a filename?
						if strings.Contains(argRange[0], ".txt") {
							// It IS a filename, so make a note of that...
							usingFile = true

							// ...and we don't need these...
							argPrefix = "N/A"
							argPad = "N/A"
							argStart = -1
							argEnd = -1

							// Get the list of Names from the file
							fileOfCompNames = argRange[0]
							computerList = readRangeFromFile(fileOfCompNames)
						} else {
							// This is NOT a filename so extract the single machine name.
							argPrefix, argStart, argPad = splitMachineName(argRange[0])
							argEnd = argStart
						}

					case 2:
						// Two parameteres, a START and an END hopefully.
						argPrefix, argStart, argPad = splitMachineName(argRange[0])
						_, argEnd, _ = splitMachineName(argRange[1])

					default:
						// More that 2 parameters
						// That's not correct!
						// Ignore it.
					}
				}

				if v3[0] == "--help" || v3[0] == "-?" {
					printHelp()
				}

				if v3[0] == "--start" || v3[0] == "-s" {
					argStart, _ = strconv.Atoi(v3[1])
				}

				if v3[0] == "--end" || v3[0] == "-e" {
					argEnd, _ = strconv.Atoi(v3[1])
				}

				if argEnd < argStart {
					argEnd = argStart
				}

				if v3[0] == "--prefix" || v3[0] == "-x" {
					argPrefix = v3[1]
				}

				if v3[0] == "--show" || v3[0] == "-w" {
					argShow = v3[1]
				}

				if v3[0] == "--timings" || v3[0] == "-t" {
					argTimings = true
				}

				if v3[0] == "--delay" || v3[0] == "-d" {
					argDelay, _ = strconv.Atoi(v3[1])
				}

				if v3[0] == "--pad" || v3[0] == "-p" {
					argPad = v3[1]
				}

				if v3[0] == "--file" || v3[0] == "-f" {
					argAction = "File"
				}

				if v3[0] == "--userfile" || v3[0] == "-u" {
					argAction = "UserFile"
				}

				if v3[0] == "--registry" || v3[0] == "-r" {
					argAction = "Registry"
				}

				if v3[0] == "--ping" || v3[0] == "-g" {
					argAction = "Ping"
				}

				if v3[0] == "--wmic" || v3[0] == "-m" {
					argAction = "WMIC"
				}

				if v3[0] == "--summary" || v3[0] == "-y" {
					argSummary = true
				}

				if v3[0] == "--save" || v3[0] == "-v" {
					argSave = v3[1]
				}

				if v3[0] == "--free" || v3[0] == "-3" {
					argAction = "Free"
				}
			} else {
				finalArg = append(finalArg, v)
			}
		}
		// Final argument has been found :- strings.Join(finalArg, " ")
		// This is for the file/folder to search for.
		argItem = strings.Join(finalArg, " ")
	}

	if strings.Contains(argShow, "1") {
		argShowGood = true
	} else {
		argShowGood = false
	}
	if strings.Contains(argShow, "0") {
		argShowBad = true
	} else {
		argShowBad = false
	}

	if argAction == "File" {
		argItem = strings.Replace(argItem, "\\", "\\\\", -1)
		argItem = strings.Replace(argItem, ":", "$", -1)
	}

	if argAction == "Nothing" {
		fmt.Println("No action specified, exiting..")
		fmt.Println()
		os.Exit(1)
	}

	fmt.Println()
	if usingFile {
		fmt.Println("Using Data File :-", fileOfCompNames)
	} else {
		fmt.Println("Launching", argEnd-argStart+1, "actions.")
		fmt.Println("Prefix =", argPrefix)
		fmt.Println("Start =", argStart)
		fmt.Println("End =", argEnd)
		fmt.Println("Pad =", argPad)
		fmt.Println("Delay =", argDelay, "seconds.")
	}
	fmt.Println()

	if argAction == "Free" {
		// Search for Computers with no active user.
		fmt.Println("Will look for machines with no active logged on user.")
	}
	if argAction == "Ping" {
		fmt.Println("Will PING the machines.")
	}
	if argAction == "Registry" {
		fmt.Print("Will look for this ", strings.ReplaceAll(argItem, "\\\\", "\\"))
		fmt.Println(" REGISTRY VALUE")
	}
	if argAction == "File" {
		fmt.Print("Will look for this ", strings.ReplaceAll(argItem, "\\\\", "\\"))
		fmt.Println(" FILE/FOLDER")
	}
	if argAction == "UserFile" {
		fmt.Print("Will look for this ", strings.ReplaceAll(argItem, "\\\\", "\\"))
		fmt.Println(" USER FILE/FOLDER")
	}
	if argAction == "WMIC" {
		fmt.Println("Will run WMIC against the machines.")
	}

	fmt.Println()

	if !usingFile && argStart == -1 {
		fmt.Println("No START specified, exiting.")
		fmt.Println()
		os.Exit(2)
	}
	if !usingFile && argEnd == -1 {
		fmt.Println("No END specified, exiting..")
		fmt.Println()
		os.Exit(3)
	}
	if !usingFile && argPrefix == "" {
		fmt.Println("No PREFIX specified, exiting..")
		fmt.Println()
		os.Exit(4)
	}

	fmt.Println("Type Y to continue :-")
	var dummy string
	fmt.Scan(&dummy)
	if strings.ToUpper(dummy) != "Y" {
		os.Exit(0)
	}
	fmt.Println()

	// Check we are not using WMIC with delete, call, uninstall,create,jscript.dll,vbscript.dll,shadowcopy
	// If we are then the user is trying to change something rather than just view
	// so we warn and then EXIT
	if strings.Contains(strings.ToUpper(argAction), "WMIC") {
		if strings.Contains(strings.ToUpper(argItem), "DELETE") {
			fmt.Println("WMIC call with disallowed option :- DELETE")
			os.Exit(5)
		}
	}
	if strings.Contains(strings.ToUpper(argAction), "WMIC") {
		if strings.Contains(strings.ToUpper(argItem), "CALL") {
			fmt.Println("WMIC call with disallowed option :- CALL")
			os.Exit(6)
		}
	}
	if strings.Contains(strings.ToUpper(argAction), "WMIC") {
		if strings.Contains(strings.ToUpper(argItem), "UNINSTALL") {
			fmt.Println("WMIC call with disallowed option :- UNINSTALL")
			os.Exit(7)
		}
	}
	if strings.Contains(strings.ToUpper(argAction), "WMIC") {
		if strings.Contains(strings.ToUpper(argItem), "CREATE") {
			fmt.Println("WMIC call with disallowed option :- CREATE")
			os.Exit(8)
		}
	}
	if strings.Contains(strings.ToUpper(argAction), "WMIC") {
		if strings.Contains(strings.ToUpper(argItem), "JSCRIPT.DLL") {
			fmt.Println("WMIC call with disallowed option :- JSCRIPT.DLL")
			os.Exit(9)
		}
	}
	if strings.Contains(strings.ToUpper(argAction), "WMIC") {
		if strings.Contains(strings.ToUpper(argItem), "VBSCRIPT.DLL") {
			fmt.Println("WMIC call with disallowed option :- VBSCRIPT.DLL")
			os.Exit(10)
		}
	}
	if strings.Contains(strings.ToUpper(argAction), "WMIC") {
		if strings.Contains(strings.ToUpper(argItem), "SHADOWCOPY") {
			fmt.Println("WMIC call with disallowed option :- SHADOWCOPY")
			os.Exit(11)
		}
	}

	// Start NOW !
	startTime := time.Now()

	if usingFile {
		// Iterate through computerList
		for _, pc := range computerList {
			performAction(wg, &mu, argAction, pc, argItem, argShowGood, argShowBad, argSave)

			// Delay before next searcher launched
			time.Sleep(time.Duration(argDelay) * time.Second)
		}
	} else {
		for i := argStart; i <= argEnd; i++ {
			// Work out what the next computer is called
			pc := strconv.Itoa(i)
			if len(pc) < len(argPad) {
				pc = (argPad + pc)[len(argPad+pc)-len(argPad):]
			}
			pc = argPrefix + pc
			// Perform the action
			performAction(wg, &mu, argAction, pc, argItem, argShowGood, argShowBad, argSave)

			// Delay before next searcher launched
			time.Sleep(time.Duration(argDelay) * time.Second)

		}
	}
	fmt.Println("Waiting for searchers to finish...")
	fmt.Println()
	wg.Wait()
	fmt.Println()
	fmt.Println("ALL DONE")
	fmt.Println()
	fmt.Println("Failures :", countBad)
	fmt.Println("Successes :", countGood)
	fmt.Println("Total :", countBad+countGood)

	// Print out the timing stats
	const bucketCount = 20
	var goodBuckets [bucketCount + 1]int
	var badBuckets [bucketCount + 1]int

	// Maximum number of characters in graph
	const maxCharCount = 120

	// Find the range of result times
	// And the lowest time.
	bucketRange := 0.0
	bucketLow := 99999.0
	bucketHigh := 0.0
	for _, r := range results {
		t := r.rTime.Sub(startTime)
		if t.Seconds() < bucketLow {
			bucketLow = t.Seconds()
		}

		if t.Seconds() > bucketHigh {
			bucketHigh = t.Seconds()
		}
	}
	bucketRange = bucketHigh - bucketLow

	fmt.Println()
	fmt.Printf("Time to complete = %.2f Seconds\n", bucketHigh)
	fmt.Println()

	// Quit if there was only one machine tested,
	// As there won't be any interesting stats for it.
	if argStart == argEnd {
		os.Exit(0)
	}

	// Quit if we don't want to print the timings, otherwise carry on.
	if !argTimings {
		os.Exit(0)
	}

	// Fill the Buckets with the results
	for _, r := range results {
		t := (r.rTime.Sub(startTime)).Seconds()
		bucketIndex := int((t - bucketLow) / (bucketRange + 1) * float64(bucketCount))
		if r.rResult {
			goodBuckets[bucketIndex]++
		} else {
			badBuckets[bucketIndex]++
		}
	}

	// Find the biggest bucket count
	bucketMaximum := 0
	for _, j := range goodBuckets {
		if j > bucketMaximum {
			bucketMaximum = j
		}
		for _, j := range badBuckets {
			if j > bucketMaximum {
				bucketMaximum = j
			}
		}
	}

	// Print the buckets
	bucketWidth := bucketRange / bucketCount

	fmt.Println()
	fmt.Println("Succeses :-")
	for i, j := range goodBuckets[:len(goodBuckets)-1] {
		thisBucketStart := float64(i)*bucketWidth + bucketLow
		thisBucketEnd := thisBucketStart + bucketWidth
		fmt.Printf("%2d %5.2f %5.2f |%-120s|\n", i+1, thisBucketStart, thisBucketEnd, strings.Repeat("O", j*maxCharCount/bucketMaximum))
	}

	fmt.Println()
	fmt.Println("Failures :-")
	for i, j := range badBuckets[:len(badBuckets)-1] {
		thisBucketStart := float64(i)*bucketWidth + bucketLow
		thisBucketEnd := thisBucketStart + bucketWidth
		fmt.Printf("%2d %5.2f %5.2f |%-120s|\n", i+1, thisBucketStart, thisBucketEnd, strings.Repeat("X", j*maxCharCount/bucketMaximum))
	}
	fmt.Println()

}

// printHelp function prints some helpful text.
func printHelp() {

	fmt.Println(`
	wskr usage :-
	
	wskr -n=ABC123..ABC999 [-w=1|0][-d=DelaySeconds][-y] -f|-r|-g|-m Some Thing To Check
	
	MANDATORY - You must have one, and only one, of these :-
	(But do NOT use = after any of these.)
	--file|-f		Search for a file.
	--userfile|-u   Show files in a specified folder for all users.
	--registry|-r	Search for a registry value.	
	--ping|-g		Search for LIVE machines.
	--free|-3		Search for machines with no active user.
	--wmic|-m		Run your WMIC your command.
					For an HTML formatted output postfix this:- /format:hform
					For a LIST output use this :- /format:list
	
	MANDATORY - You will of course need to state a RANGE of computers to look at.
	--range=|-n=	string[..string]    FirstMachine[.. LastMachine]
	--range=|-n=	'filename.txt'       Name of text file to read in, it should end in .txt.
					The text file must be in the same directory that WSKR.EXE is run from.
					Each line of the text file should start with a machine name, then a space; everything after the space is ignored.
					Blank lines are ignored, as are any lines starting with a space or hash symbol.
	
	OPTIONAL :-
	[--show=|-w=]	String		1,Return successes, 0,Failures.		(-w=10 to show all) *Note.
	[--delay=|-d=]	Integer		Seconds of Delay between machines.
	[--save=|-v=]	'String'	File name, to save in same location as EXE. Use single quotes.
	[--summary|-y]				Just give final counts.
	[--timings|-t]              Display the timings of the results coming back.
	[--help|-?]					This help page.

	*Note :- Finding something may be more meaningful that NOT being able to find something.
				Because you may be prevented from finding things for multiple reasons,
				e.g. rights, firewalls, remote services off etc.
	
	To search PC0001 through PC1234, finding machines that do NOT have "c:\data\some file.txt" use :-
				wskr --show=0 --range=pc0001..pc1234 --file 'c:\data\some file.txt'
				(Note the --show=0, to see only the failures.)
	
	To search PC00 through PC99, showing the files present for each user on each machine in a specific folder try something like :-
				wskr --range-pc00..pc99 --userfile 'AppData\roaming\icaclient'

	To search for a registry Value on a single computer :-
				wskr -n=comp456 -r 'HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon\Shell'
	
	To see various things such as :-
	   Logged in users, saving result:  wskr.exe --range=WS123 --wmic computersystem get username --save='output.txt'
	   OS version:                      wskr.exe --range=WS123 --wmic os get version
	   Installed software MSI's:        wskr.exe --range=WS123 --wmic product get name,vendor,version
	   System serial number:            wskr.exe --range=WS123 --wmic bios get serialnumber	
	   Installed printers:              wskr.exe --range=WS123 --wmic printerconfig list
	   The some IP related stuff:       wskr.exe --range=WS123 --wmic nicconfig get ipaddress,defaultipgateway,dhcpserver,dnsdomainsuffixsearchorder
	   AssetTag (not the SerialNumber): wskr.exe --range=WS123 --wmic systemenclosure get SMBIOSAssetTag
	   HTML for all COMPUTERSYSTEM:     wskr.exe --range=WS123 --wmic computersystem get /format:hform --save='cs-output.html'
	
	Oviously the above ranges could be in the :-
		* Multiple machine format: --range=SSnnn..SSmmm
		* File name format:        --range=myMachines.txt

	Dependencies :-
		1) The machine you are running this on must be running Windows.
		2) --ping is reliant on Windows PING.EXE
		3) --wmic is reliant on Windows WMIC.EXE
		4) --registry is reliant on Windows REG.EXE

	Assumptions :-
		1) Your machine names have at least one character at the start, followed by at least one digit.
		2) The machines you are scanning are running Windows.
		3) You have admin rights on the remote machines.
		4) Ensure that WMI service is enabled and running on the remote machines.
		5) Ensure any required firewall ports are open between your machine and the remote machines.

	v0.1 - Copyright 2023
	Author -- llamnuds.
	
			`)
	os.Exit(0)
}

// performAction function checks to see which single function is required
// and executes it.
func performAction(wg *sync.WaitGroup, mu *sync.Mutex, argAction string, pc string, argItem string, argShowGood bool, argShowBad bool, argSave string) {

	wg.Add(1)

	if argAction == "File" {
		go checkFile(wg, mu, pc, argItem, argShowGood, argShowBad, argSave)
	}
	if argAction == "UserFile" {
		go checkUserFile(wg, mu, pc, argItem, argShowGood, argShowBad, argSave)
	}
	if argAction == "Registry" {
		go checkRegistry(wg, mu, pc, argItem, argShowGood, argShowBad, argSave)
	}
	if argAction == "Ping" {
		go checkPing(wg, mu, pc, argShowGood, argShowBad, argSave)
	}
	if argAction == "WMIC" {
		go checkWMI(wg, mu, pc, argItem, argShowGood, argShowBad, argSave)
	}
	if argAction == "Free" {
		go checkFree(wg, mu, pc, argItem, argShowGood, argShowBad, argSave)
	}
}

func badResult() {
	mu.Lock()
	results = append(results, result{rTime: time.Now(), rResult: false})
	mu.Unlock()
}

func goodResult() {
	mu.Lock()
	results = append(results, result{rTime: time.Now(), rResult: true})
	mu.Unlock()
}

// checkFree function checks to see if a device has no active user.
func checkFree(wg *sync.WaitGroup, mu *sync.Mutex, pc string, argItem string, argShowGood, argShowBad bool, argSave string) {
	defer wg.Done()
	// Launch an EXE and keep the results
	out, err := exec.Command("cmd", "/c", "wmic /node:"+pc+" computersystem get username").Output()

	// Check 'out' to see if this machine really is FREE
	var isNotFree bool = true
	if err == nil {
		out2 := strings.Replace(string(out), pc, "", 1)
		out2 = strings.TrimSpace(out2)
		if strings.ToLower(out2) == "username" {
			isNotFree = false
		} else {
			isNotFree = true
		}
	}

	if err != nil || isNotFree {
		if !argSummary {
			if argShowBad {
				print(pc, err.Error())
				maybeSaveToFile("0-"+argSave, pc, err.Error())
			}
		}
		mu.Lock()
		countBad++
		mu.Unlock()
		badResult()
	} else {
		if !argSummary {
			if argShowGood {
				print(pc, "is Free.")
				maybeSaveToFile("1-"+argSave, pc, "is Free")
			}
		}
		mu.Lock()
		countGood++
		mu.Unlock()
		goodResult()
	}
}

// readRangeFromFile function returns a slice of computer names
// discarding extraneous information from the input text file.
func readRangeFromFile(myFile string) []string {

	// Slice of computer names
	computers := []string{}

	file, err := os.Open(myFile)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := scanner.Text()
		if data == "" {
			// Ignore this blank line
		} else {
			// Not a blank line
			if data[0] == ' ' || data[0] == '#' {
				// DO NOTHING
			} else {
				//  Should be data so split into computername and comment
				computer, _, found := strings.Cut(data, " ")
				if found {
					computers = append(computers, computer)
				} else {
					// No space found, so use the whole thing
					computers = append(computers, data)
				}
			}
		}

	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
	fmt.Println(computers)
	return computers

}

// splitMachineName function takes an input string of characters and
// splits it into a textual 'prefix' and and padding of zeros ('0').
func splitMachineName(a string) (string, int, string) {
	j := 0
	for i := 0; i < len(a); i++ {
		b := string(a[i])
		if b <= "9" && b >= "0" {
			j = i
			break
		}
	}
	prefix := a[:j]
	pad := strings.Repeat("0", len(a[j:]))
	number, _ := strconv.Atoi(a[j:])
	return prefix, number, pad
}

// checkWMI function performs some user requested WMI check on a remote machine
func checkWMI(wg *sync.WaitGroup, mu *sync.Mutex, pc string, argItem string, argShowGood bool, argShowBad bool, argSave string) {
	defer wg.Done()
	// Launch an EXE and keep the results
	out, err := exec.Command("cmd", "/c", "wmic /node:"+pc+" "+argItem).Output()
	if err != nil {
		if !argSummary {
			if argShowBad {
				print(pc, err.Error())
				maybeSaveToFile("0-"+argSave, pc, err.Error())
			}
		}
		mu.Lock()
		countBad++
		mu.Unlock()
		badResult()
	} else {
		if !argSummary {
			if argShowGood {
				print(pc, string(out))
				maybeSaveToFile("1-"+argSave, pc, string(out))
			}
		}
		mu.Lock()
		countGood++
		mu.Unlock()
		goodResult()
	}
}

// maybeSaveToFile function saves some text to a user named file if they so wish.
func maybeSaveToFile(filename string, pc string, data string) {
	if len(filename) < 3 {
		return
	}
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	var data2 string = "\n" + strings.TrimSpace(pc) + "\n"
	for _, line := range strings.Split(data, "\n") {
		datum := strings.TrimSpace(line)
		if datum != "" {
			data2 += " " + datum + "\n"
		}
	}
	if _, err := f.WriteString(data2); err != nil {
		log.Println(err)
	}
}

// print function will print the computer name on one line (pc)
// then the remainder of the information (data) on further lines
// each indented by one space.
func print(pc string, data string) {
	var data2 string = strings.TrimSpace(pc) + "\n"
	for _, line := range strings.Split(data, "\n") {
		datum := strings.TrimSpace(line)
		if datum != "" {
			data2 += " " + datum + "\n"
		}
	}
	fmt.Println(data2)
}

// checkPing function will see if a machine is alive or not.
func checkPing(wg *sync.WaitGroup, mu *sync.Mutex, pc string, argShowGood bool, argShowBad bool, argSave string) {
	defer wg.Done()

	var buffer bytes.Buffer
	cmd := exec.Command("ping", pc, "-n", "1", "-4")
	cmd.Stdout = &buffer
	_ = cmd.Run()
	result := buffer.String()

	if len(result) < 100 {
		// FAILED TO PING
		mu.Lock()
		countBad++
		mu.Unlock()
		badResult()
		if argShowBad { // We want to see the failures
			if !argSummary { // But not if we only want to see the summary counts
				print(pc, "NOT-Alive")
				maybeSaveToFile("0-"+argSave, pc, "NOT-Alive")
			}
		}
	} else {
		// Something, was returned
		success := false
		for _, value := range strings.Split(result, "\n") {
			if strings.Contains(string(value), "Received = 1") {
				success = true
			}
		}

		if success {
			mu.Lock()
			countGood++
			mu.Unlock()
			goodResult()
			if !argSummary {
				if argShowGood {
					print(pc, "Alive")
					maybeSaveToFile("1-"+argSave, pc, "Alive")
				}
			}
		} else {
			mu.Lock()
			countBad++
			mu.Unlock()
			badResult()
			if argShowBad { // We want to see the failures
				if !argSummary { // Unless we only want to see the summaries
					print(pc, "NOT-Alive")
					maybeSaveToFile("0-"+argSave, pc, "NOT-Alive")
				}
			}
		}

	}
}

// getRegData function will attempt to get a registry value from a remote computer's registry.
func getRegData(pc, key, value string, argShowGood bool, argShowBad bool, argSave string) {

	// Launch an EXE and keep the results
	var buffer bytes.Buffer
	cmd := exec.Command("c:\\windows\\system32\\reg.exe", "query", "\\\\"+pc+"\\"+key, "/v", value)
	cmd.Stdout = &buffer
	_ = cmd.Run()

	if buffer.String() == "" {
		// Nothing was returned
		mu.Lock()
		countBad++
		mu.Unlock()
		badResult()
		if argShowBad {
			// We want to see the failures
			if !argSummary {
				print(pc, buffer.String())
				maybeSaveToFile("0-"+argSave, pc, buffer.String())
			}
		}
	} else {
		// Something was returned
		mu.Lock()
		countGood++
		mu.Unlock()
		goodResult()
		if argShowGood {
			// We want to see the successes
			if !argSummary {
				print(pc, buffer.String())
				maybeSaveToFile("1-"+argSave, pc, buffer.String())
			}
		}
	}
}

// checkRegistry function will attempt to get a registry value from a remote computer's registry.
func checkRegistry(wg *sync.WaitGroup, mu *sync.Mutex, pc string, registry string, argShowGood bool, argShowBad bool, argSave string) {
	defer wg.Done()
	registrySplit := strings.Split(registry, `\`)
	regSplitLengthMinusOne := len(registrySplit) - 1
	regKey := strings.Join(registrySplit[:regSplitLengthMinusOne], `\`) // all up till the last word
	regValue := registrySplit[regSplitLengthMinusOne]                   // the last word
	getRegData(pc, regKey, regValue, argShowGood, argShowBad, argSave)
}

// checkFile function will check the existence of a file or folder on a remote machine.
func checkFile(wg *sync.WaitGroup, mu *sync.Mutex, pc string, file string, argShowGood bool, argShowBad bool, argSave string) {
	defer wg.Done()
	searchForThis := "\\\\" + pc + "\\" + file
	if fileStat, err := os.Stat(searchForThis); err == nil {
		mu.Lock()
		countGood++
		mu.Unlock()
		goodResult()
		if argShowGood {
			if !argSummary {
				print(pc, searchForThis+" , "+fileStat.ModTime().Format(time.UnixDate))
				maybeSaveToFile("1-"+argSave, pc, searchForThis+" , "+fileStat.ModTime().Format(time.UnixDate))
			}
		}
	} else {
		mu.Lock()
		countBad++
		mu.Unlock()
		badResult()
		if argShowBad {
			if !argSummary {
				print(pc, strings.Replace(err.Error(), "CreateFile ", "", -1))
				maybeSaveToFile("0-"+argSave, pc, strings.Replace(err.Error(), "CreateFile ", "", -1))
			}
		}
	}
}

// checkUserFile function will check the existence of a file or folder on a remote machine, in the USER folders.
func checkUserFile(wg *sync.WaitGroup, mu *sync.Mutex, pc string, userfile string, argShowGood bool, argShowBad bool, argSave string) {
	defer wg.Done()
	wg2 := new(sync.WaitGroup)
	// Determine user folders on this machine
	remoteDir := "\\\\" + pc + "\\c$\\users\\"
	userFolders, err := ioutil.ReadDir(remoteDir)
	if err != nil {
		if !argSummary && argShowBad {
			print(pc, "Error reading "+remoteDir)
			maybeSaveToFile("0-"+argSave, pc, "Error reading "+remoteDir)
		}
	} else {
		// Check each User Folder in turn
		for _, userFolder := range userFolders {
			if userFolder.IsDir() {

				// Compile the full path to be checked
				folderToCheck := `c$\users\` + userFolder.Name() + `\` + userfile

				// Double up the slashes
				// folderToCheck = strings.ReplaceAll(folderToCheck, `\`, `\\`)

				// Powershell removes the single quotes, CMD doesn't, so we do it here - CHECK to see if this needs to go back in.   XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
				// folderToCheck = strings.ReplaceAll(folderToCheck, "'", "")

				// Launch it
				wg2.Add(1)
				go checkFilePS(wg2, mu, pc, folderToCheck, argShowGood, argShowBad, argSave)
			}
		}
	}
	wg2.Wait()
}

// Use PowerShell to check a File's stats (version, last mod etc)
func checkFilePS(wg *sync.WaitGroup, mu *sync.Mutex, pc string, userfile string, argShowGood bool, argShowBad bool, argSave string) {
	defer wg.Done()

	// Powershell Command to run
	psCommand := `write-output $(gci "\\` + pc + `\` + userfile + `")`

	// Construct the PowerShell command with the required arguments
	cmd := exec.Command("powershell", "-Command", psCommand)

	// Execute the command and capture the output
	output, err := cmd.Output()
	results := string(output)
	success := false

	if err == nil {
		// No err but, blanks are bad too!
		if results != "" {
			success = true
		}
	}

	if success {
		mu.Lock()
		countGood++
		mu.Unlock()
		goodResult()
		if !argSummary && argShowGood {
			print(pc, userfile+" = "+results)
			maybeSaveToFile("1-"+argSave, pc, userfile+" = "+results)
		}
	} else {
		mu.Lock()
		countBad++
		mu.Unlock()
		badResult()
		if !argSummary && argShowBad {
			if results == "" {
				print(pc, "Not found.")
				maybeSaveToFile("0-"+argSave, pc, "Not found.")
			} else {
				print(pc, "Error = "+string(err.Error()))
				maybeSaveToFile("0-"+argSave, pc, "Error = "+string(err.Error()))
			}
		}
	}
}

/*

Errors :-
1  - No ACTION specified.
2  - No START specified.
3  - No END specified.
4  - No PREFIX specified.
5  - WMIC call with disallowed option - DELETE.
6  - WMIC call with disallowed option - CALL.
7  - WMIC call with disallowed option - UNINSTALL.
8  - WMIC call with disallowed option - CREATE.
9  - WMIC call with disallowed option - JSCRIPT.DLL.
10 - WMIC call with disallowed option - VBSCRIPT.DLL.
11 - WMIC call with disallowed option - SHADOWCOPY.
*/
