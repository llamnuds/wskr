package main

import (
	"bufio"
	"bytes"
	"fmt"
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
	var argDebug bool = false
	var argPad string = "000"
	var argDelay int = 0
	var argPrefix string = ""
	var argAction string = "Nothing"
	var argItem string = ""
	var argSave string = ""
	var argConfirm bool = false
	var argTimings bool = false

	// Used letters - 3abcdefgimnprstuvwxy / / / / / / / / / / / / / / / / / / / / / / / / / / / / /

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
					if len(v3) <= 1 {
						fmt.Println("Usage error: --range requires a value")
						os.Exit(12)
					}
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
							if len(computerList) == 0 {
								fmt.Println("No computer names found in", fileOfCompNames)
								os.Exit(12)
							}
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
					if len(v3) <= 1 {
						fmt.Println("Usage error: --start requires a value")
						os.Exit(12)
					}
					argStart, _ = strconv.Atoi(v3[1])
				}

				if v3[0] == "--confirm" || v3[0] == "-c" {
					argConfirm = true
				}

				if v3[0] == "--end" || v3[0] == "-e" {
					if len(v3) <= 1 {
						fmt.Println("Usage error: --end requires a value")
						os.Exit(12)
					}
					argEnd, _ = strconv.Atoi(v3[1])
				}

				if argEnd < argStart {
					argEnd = argStart
				}

				if v3[0] == "--prefix" || v3[0] == "-x" {
					if len(v3) <= 1 {
						fmt.Println("Usage error: --prefix requires a value")
						os.Exit(12)
					}
					argPrefix = v3[1]
				}

				if v3[0] == "--show" || v3[0] == "-w" {
					if len(v3) <= 1 {
						fmt.Println("Usage error: --show requires a value")
						os.Exit(12)
					}
					argShow = v3[1]
				}

				if v3[0] == "--timings" || v3[0] == "-t" {
					argTimings = true
				}

				if v3[0] == "--debug" || v3[0] == "-a" {
					argDebug = true
				}

				if v3[0] == "--delay" || v3[0] == "-d" {
					if len(v3) <= 1 {
						fmt.Println("Usage error: --delay requires a value")
						os.Exit(12)
					}
					argDelay, _ = strconv.Atoi(v3[1])
				}

				if v3[0] == "--pad" || v3[0] == "-p" {
					if len(v3) <= 1 {
						fmt.Println("Usage error: --pad requires a value")
						os.Exit(12)
					}
					argPad = v3[1]
				}

				if v3[0] == "--bitlocker" || v3[0] == "-b" {
					argAction = "Bitlocker"
				}

				if v3[0] == "--file" || v3[0] == "-f" {
					argAction = "File"
				}

				if v3[0] == "--dir" || v3[0] == "-i" {
					argAction = "Dir"
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
					if len(v3) <= 1 {
						fmt.Println("Usage error: --save requires a value")
						os.Exit(12)
					}
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
		fmt.Println("Delay =", argDelay, "seconds.")
		fmt.Println("Timings = ", argTimings)
	} else {
		fmt.Println("Launching", argEnd-argStart+1, "actions.")
		fmt.Println("Prefix =", argPrefix)
		fmt.Println("Start =", argStart)
		fmt.Println("End =", argEnd)
		fmt.Println("Pad =", argPad)
		fmt.Println("Delay =", argDelay, "seconds.")
		fmt.Println("Timings = ", argTimings)
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
	if argAction == "Dir" {
		fmt.Print("Will display a directory of this ", strings.ReplaceAll(argItem, "\\\\", "\\"))
		fmt.Println(" FOLDER")
	}
	if argAction == "UserFile" {
		fmt.Print("Will look for this ", strings.ReplaceAll(argItem, "\\\\", "\\"))
		fmt.Println(" USER FILE/FOLDER")
	}
	if argAction == "WMIC" {
		fmt.Println("Will run WMIC against the machines.")
	}
	if argAction == "Bitlocker" {
		fmt.Println("Will extract Bitlocker key from the machines.")
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

	if !argConfirm {
		fmt.Println("Type Y to continue :-")
		var dummy string
		fmt.Scan(&dummy)
		if strings.ToUpper(dummy) != "Y" {
			os.Exit(0)
		}
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
			performAction(wg, &mu, argAction, pc, argItem, argShowGood, argShowBad, argSave, argDebug)

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
			performAction(wg, &mu, argAction, pc, argItem, argShowGood, argShowBad, argSave, argDebug)

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
	--dir|-i		Display the contents of a directory.
	--userfile|-u   Show files in a specified folder for all users.
	--registry|-r	Search for a registry value.	
	--ping|-g		Search for LIVE machines.
	--free|-3		Search for machines with no active user.
	--bitlocker|-b  Retrieve Bitlocker Recovery key.
	--wmic|-m		Run your WMIC your command.
					For an HTML formatted output postfix this:- /format:hform
					For a LIST output use this :- /format:list
	
	MANDATORY - You will of course need to state a RANGE of computers to look at.
	--range=|-n=	string[..string]    FirstMachine[.. LastMachine]
	--range=|-n=	'filename.txt'       Name of text file to read in, it should end in .txt.
					The text file must be in the same directory that WSKR.EXE is run from.
					Each line of the text file should start with a machine name, then a space; everything after the space is ignored.
					Blank lines are ignored, as are any lines starting with a space, hash symbol or tab.
	
	OPTIONAL :-
	[--show=|-w=]	String		1,Return successes, 0,Failures.		(-w=10 to show all) *Note.
	[--delay=|-d=]	Integer		Seconds of Delay between machines.
	[--save=|-v=]	'String'	File name, to save in same location as EXE. Use single quotes.
	[--summary|-y]				Just give final counts.
	[--timings|-t]              Display the timings of the results coming back.
	[--help|-?]					This help page.

	*Note :- Finding something may be more meaningful than NOT being able to find something.
				Because you may be prevented from finding things for multiple reasons,
				e.g. rights, firewalls, remote services off etc.
	
	To search PC0001 through PC1234, finding machines that do NOT have "c:\data\some file.txt" use :-
				wskr --show=0 --range=pc0001..pc1234 --file 'c:\data\some file.txt'
				(Note the --show=0, to see only the failures.)
	
	To search PC00 through PC99, showing the files present for each user on each machine in a specific folder try something like :-
				wskr --range-pc00..pc99 --userfile 'AppData\roaming\icaclient'

	To search for a registry Value on a single computer :-
				wskr -n=comp456 -r 'HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon\Shell'

	To show the contents of a given directory :-
				wskr -n=comp456 --dir 'c:\windows\program files'
	
	To see various things such as :-
	   Logged in users, saving result:  wskr.exe --range=WS123 --wmic computersystem get username --save='output.txt'
	   OS version:                      wskr.exe --range=WS123 --wmic os get version
	   Installed software MSI's:        wskr.exe --range=WS123 --wmic product get name,vendor,version
	   System serial number:            wskr.exe --range=WS123 --wmic bios get serialnumber	
	   Installed printers:              wskr.exe --range=WS123 --wmic printerconfig list
	   The some IP related stuff:       wskr.exe --range=WS123 --wmic nicconfig get ipaddress,defaultipgateway,dhcpserver,dnsdomainsuffixsearchorder
	   AssetTag (not the SerialNumber): wskr.exe --range=WS123 --wmic systemenclosure get SMBIOSAssetTag
	   HTML for all COMPUTERSYSTEM:     wskr.exe --range=WS123 --wmic computersystem get /format:hform --save='cs-output.html'
	   EFS running as a service:        wskr.exe --range=ws123 --wmic service "where name='efs'" get Started
	
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
		3) You have sufficient rights on the remote machines.
		4) Ensure that WMI service is enabled and running on the remote machines.
		5) Ensure any required firewall ports are open between your machine and the remote machines.

	Restrictions :-
		The following are not allowed in conjunction with --WMIC
		1) CALL
		2) CREATE
		3) UNINSTALL
		4) DELETE
		5) JSCRIPT.DLL
		6) VBSCRIPT.DLL
		7) SHADOWCOPY

	v0.1 - Copyright 2023 - llamnuds
	
			`)
	os.Exit(0)
}

// performAction function checks to see which single function is required
// and executes it.
func performAction(wg *sync.WaitGroup, mu *sync.Mutex, argAction string, pc string, argItem string, argShowGood bool, argShowBad bool, argSave string, argDebug bool) {

	wg.Add(1)

	if argAction == "File" {
		go checkFile(wg, mu, pc, argItem, argShowGood, argShowBad, argSave, argDebug)
	}
	if argAction == "Dir" {
		go checkDir(wg, mu, pc, argItem, argShowGood, argShowBad, argSave, argDebug)
	}
	if argAction == "UserFile" {
		go checkUserFile(wg, mu, pc, argItem, argShowGood, argShowBad, argSave, argDebug)
	}
	if argAction == "Registry" {
		go checkRegistry(wg, mu, pc, argItem, argShowGood, argShowBad, argSave, argDebug)
	}
	if argAction == "Ping" {
		go checkPing(wg, mu, pc, argShowGood, argShowBad, argSave, argDebug)
	}
	if argAction == "WMIC" {
		go checkWMI(wg, mu, pc, argItem, argShowGood, argShowBad, argSave, argDebug)
	}
	if argAction == "Free" {
		go checkFree(wg, mu, pc, argItem, argShowGood, argShowBad, argSave, argDebug)
	}
	if argAction == "Bitlocker" {
		go checkBitlocker(wg, mu, pc, argShowGood, argShowBad, argSave, argDebug)
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
func checkFree(wg *sync.WaitGroup, mu *sync.Mutex, pc string, argItem string, argShowGood, argShowBad bool, argSave string, argDebug bool) {
	defer wg.Done()
	// Launch an EXE and keep the results
	out, err := exec.Command("cmd", "/c", "wmic /node:"+pc+" computersystem get username").Output()

	if argDebug {
		maybeSaveToFile("debug.log", pc, string(out))
		if err != nil {
			maybeSaveToFile("debug.log", pc, string(err.Error()))
		}
	}

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
		return []string{}
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := scanner.Text()
		if data == "" {
			// Ignore this blank line
		} else {
			// Not a blank line
			if data[0] == ' ' || data[0] == '#' || data[0] == '\t' {
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
func checkWMI(wg *sync.WaitGroup, mu *sync.Mutex, pc string, argItem string, argShowGood bool, argShowBad bool, argSave string, argDebug bool) {
	defer wg.Done()
	// Launch an EXE and keep the results
	out, err := exec.Command("cmd", "/c", "wmic /node:"+pc+" "+argItem).Output()
	if err != nil {
		if argDebug {
			maybeSaveToFile("debug.log", pc, err.Error())
		}
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
		if argDebug {
			maybeSaveToFile("debug.log", pc, string(out))
		}
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

// checkBitlocker function tries to read Bitlocker Recovery-ID key on a remote machine
func checkBitlocker(wg *sync.WaitGroup, mu *sync.Mutex, pc string, argShowGood bool, argShowBad bool, argSave string, argDebug bool) {
	defer wg.Done()
	// Launch an EXE and keep the results

	// Powershell Command to run
	psCommand := `invoke-command -computername ` + pc + ` -scriptblock {$BitlockerVolumers = Get-BitLockerVolume;$BitlockerVolumers|ForEach-Object {$MountPoint=$_.MountPoint;$RecoveryKey=[string]($_.KeyProtector).RecoveryPassword;if ($RecoveryKey.Length -gt 5) {Write-Output ($RecoveryKey)}}}`

	// Construct the PowerShell command with the required arguments
	out, err := exec.Command("powershell", "-Command", psCommand).Output()

	if argDebug {
		if err != nil {
			maybeSaveToFile("debug.log", pc, err.Error())
		} else {
			maybeSaveToFile("debug.log", pc, string(out))
		}
	}

	if (err != nil) || (len(string(out)) < 55) {
		if !argSummary {
			if argShowBad {
				print(pc, "No Bitlocker Key retrieved.")
				maybeSaveToFile("0-"+argSave, pc, "No Bitlocker Key retrieved.")
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
		return
	}
	defer f.Close()

	var data2 string = "\n" + strings.TrimSpace(pc) + "\n"
	for _, line := range strings.Split(data, "\n") {
		datum := strings.TrimSpace(line)
		if datum != "" {
			data2 += "\t" + datum + "\n"
		}
	}
	if _, err := f.WriteString(data2); err != nil {
		log.Println(err)
	}
}

// print function will print the computer name on one line (pc)
// then the remainder of the information (data) on further lines
// each indented by one tab.
func print(pc string, data string) {
	var data2 string = strings.TrimSpace(pc) + "\n"
	for _, line := range strings.Split(data, "\n") {
		datum := strings.TrimSpace(line)
		if datum != "" {
			data2 += "\t" + datum + "\n"
		}
	}
	fmt.Println(data2)
}

// checkPing function will see if a machine is alive or not.
func checkPing(wg *sync.WaitGroup, mu *sync.Mutex, pc string, argShowGood bool, argShowBad bool, argSave string, argDebug bool) {
	defer wg.Done()

	var buffer bytes.Buffer
	cmd := exec.Command("ping", pc, "-n", "1", "-4")
	cmd.Stdout = &buffer
	_ = cmd.Run()
	result := buffer.String()

	if argDebug {
		maybeSaveToFile("debug.log", pc, result)
	}

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
			if strings.Contains(string(value), "(0%") {
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
func getRegData(pc, key, value string, argShowGood bool, argShowBad bool, argSave string, argDebug bool) {

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

				if argDebug {
					maybeSaveToFile("debug.log", pc, buffer.String())
				}
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
				if argDebug {
					maybeSaveToFile("debug.log", pc, buffer.String())
				}
			}
		}
	}
}

// checkRegistry function will attempt to get a registry value from a remote computer's registry.
func checkRegistry(wg *sync.WaitGroup, mu *sync.Mutex, pc string, registry string, argShowGood bool, argShowBad bool, argSave string, argDebug bool) {
	defer wg.Done()
	registrySplit := strings.Split(registry, `\`)
	regSplitLengthMinusOne := len(registrySplit) - 1
	regKey := strings.Join(registrySplit[:regSplitLengthMinusOne], `\`) // all up till the last word
	regValue := registrySplit[regSplitLengthMinusOne]                   // the last word
	getRegData(pc, regKey, regValue, argShowGood, argShowBad, argSave, argDebug)
}

// checkFile function will check the existence of a file or folder on a remote machine.
func checkFile(wg *sync.WaitGroup, mu *sync.Mutex, pc string, file string, argShowGood bool, argShowBad bool, argSave string, argDebug bool) {
	defer wg.Done()
	searchForThis := "\\\\" + pc + "\\" + file
	// print("checkFile : "+pc, searchForThis)
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

		if argDebug {
			maybeSaveToFile("debug.log", pc, err.Error())
		}

		if argShowBad {
			if !argSummary {
				print(pc, strings.Replace(err.Error(), "CreateFile ", "", -1))
				maybeSaveToFile("0-"+argSave, pc, strings.Replace(err.Error(), "CreateFile ", "", -1))
			}
		}
	}
}

// checkDir function will display the contents of a folder on a remote machine.
func checkDir(wg *sync.WaitGroup, mu *sync.Mutex, pc string, file string, argShowGood bool, argShowBad bool, argSave string, argDebug bool) {
	defer wg.Done()
	file = strings.Replace(file, ":", "$", -1)
	remoteDir := "\\\\" + pc + "\\" + file
	// print("checkDir : "+pc, remoteDir)
	entries, err := os.ReadDir(remoteDir)

	if argDebug && err != nil {
		maybeSaveToFile("debug.log", pc, err.Error())
	}

	var dirCount = 0
	var fileCount = 0
	if err != nil {
		mu.Lock()
		countBad++
		mu.Unlock()
		badResult()
		if !argSummary && argShowBad {
			print(pc, "Error reading "+remoteDir)
			maybeSaveToFile("0-"+argSave, pc, "Error reading "+remoteDir)
		}
	} else {
		mu.Lock()
		countGood++
		mu.Unlock()
		goodResult()
		var result string = ""
		for _, entry := range entries {
			if entry.IsDir() {
				info, err := entry.Info()
				if err != nil {
					continue
				}
				result += fmt.Sprintf("%s\tDir\t%10d bytes\t%s", remoteDir+"\\"+entry.Name(), info.Size(), info.ModTime()) + "\n"
				dirCount++
			} else {
				info, err := entry.Info()
				if err != nil {
					// handle the error, perhaps continue to the next entry
					continue
				}
				result += fmt.Sprintf("%s\tFile\t%10d bytes\t%s", remoteDir+"\\"+entry.Name(), info.Size(), info.ModTime()) + "\n"
				fileCount++
			}
		}
		result += fmt.Sprintf("File Count = %d\nDir Count = %d", fileCount, dirCount)
		print(pc, result)
		maybeSaveToFile("1-"+argSave, pc, result)
	}
}

// checkUserFile function will check the existence of a file or folder on a remote machine, in the USER folders.
func checkUserFile(wg *sync.WaitGroup, mu *sync.Mutex, pc string, userfile string, argShowGood bool, argShowBad bool, argSave string, argDebug bool) {
	defer wg.Done()
	wg2 := new(sync.WaitGroup)
	// Determine user folders on this machine
	remoteDir := "\\\\" + pc + "\\c$\\users\\"
	userFolders, err := os.ReadDir(remoteDir)
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
				folderToCheck := `c:\users\` + userFolder.Name() + `\` + userfile

				// Double up the slashes
				// folderToCheck = strings.ReplaceAll(folderToCheck, `\`, `\\`)

				// Powershell removes the single quotes, CMD doesn't, so we do it here - CHECK to see if this needs to go back in.   XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
				// folderToCheck = strings.ReplaceAll(folderToCheck, "'", "")

				// Launch it
				wg2.Add(1)
				//go checkFilePS(wg2, mu, pc, folderToCheck, argShowGood, argShowBad, argSave)
				go checkDir(wg2, mu, pc, folderToCheck, argShowGood, argShowBad, argSave, argDebug)
			}
		}
	}
	wg2.Wait()
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



DEBUG - DEBUG - DEBUG
	DONE - Ping
	DONE - Free
	TODO - Bitlocker
	TODO - Dir
	TODO - File
	TODO - Registry
	TODO - UserFile
	TODO - WMI
	TODO - Update Github README
	TODO - Update In-program Help
	TODO - Update CHANGELOG File
	TODO - Update Github Feature Requests

*/
