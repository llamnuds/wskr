# WSKr
Do you manage hundreds or thousands of machines?  Then this simple tool could be of use to you.  Just download and run the EXE on your Windows machine. Scan a single machine or thousands in just a few seconds.  It basically fronts calls to Windows PING, REG and WMIC executables, but we get the benefit of parrallelism with Go's go-routines.


## FEATURES
* Scan a range of Windows machines quickly
* Look for a registry value
* Look for the presence or absence of a file or folder
* See which machines are alive, or not
* Use WMI to find machines ...
  * IP address
  * DHCP or DNS settings
  * See who is logged on
  * Find machines with no-one logged on
  * Installed software or patches
  * Serial or Asset Tag numbers
  * Installed printers
  * OS version
  * ...and more...
* No installation required
* No configuration files


## GETTING STARTED
To get started, download the WSKr.exe file from the Github repository and run it on your Windows machine.
For example, if you have a thousand machines, named ```WS000``` through ```WS999```,
then you can then use the following command to run a simple Ping scan on them all.
You'll know which are on in just a few seconds :-
```
wskr --range=WS000..WS999  --ping
```


## MANDATORY PARAMETER - Tell WSKr what to do.
You must have one, and only one, of these.
*(But do NOT use = after any of these.)*
```
--file|-f      some-file           Search for a file.
--registry|-r  some-reg-value      Search for a registry value.	
--wmic|-m      some-wmic-command   Run your WMIC your command.
--ping|-g                          Search for LIVE machines.
--free|-3                          Search for machines with no active user.
```
* With ```--wmic```, For an HTML formatted output postfix this ```/format:hform``` ...or for a LIST output use this ```/format:list```
          
          
## MANDATORY PARAMETER - Tell WSKr on which machines to operate.
You will need to state a RANGE of computers to look at.
```
--range=|-n=   string[..string]    FirstMachine[.. LastMachine]
--range=|-n=   'filename.txt'      Name of text file to read in, it should end in .txt.
```

* The text file must be in the same directory that WSKR.EXE is run from.
* Each line of the text file should start with a machine name, then a space; everything after the space is ignored.
* Blank lines are ignored, as are any lines starting with a space or hash symbol.


## OPTIONAL PARAMETERS
```
[--show=|-w=]     String    Return successes(1), Failures(0).             Default = 1 i.e. Only successes (-w=10 to show all)
[--delay=|-d=]    Integer   Seconds of Delay between machines.            Default = 0 Seconds
[--save=|-v=]     'String'  File name, to save in same location as EXE.   Use single quotes.
[--summary|-y]              Just give final counts.
```


## SAVING THE RESULTS
You can save the results of your scan to a file using the following option:
```--save=``` or ```-v=``` followed by the filename to save to. The file will be saved to the same location as EXE.
Successes are saved to the file that you specified but prefixed with a ```1-```, and the failure file is prefixed with ```0-```.


## EXAMPLES

To search ```PC0001``` through ```PC1234```, finding machines that do NOT have ```c:\data\some file.txt``` use :-

*(Note the --show=0, to see only the failures.)*
```
wskr --show=0 --range=pc0001..pc1234 --file c:\data\some file.txt
```	
To search for a REGISTRY Value on a single computer :-
```
wskr -n=comp456 -r HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon\Shell
```	
WMIC is useful to see a whole bunch of different things, such as :-
```
Logged in users, saving result:  wskr.exe --range=WS123 --wmic computersystem get username --save='output.txt'
OS version:                      wskr.exe --range=WS123 --wmic os get version
Installed software MSI's:        wskr.exe --range=WS123 --wmic product get name,vendor,version
System serial number:            wskr.exe --range=WS123 --wmic bios get serialnumber	
Installed printers:              wskr.exe --range=WS123 --wmic printerconfig list
The IP,DHCPserver, Gateway:      wskr.exe --range=WS123 --wmic nicconfig get IPAddress,dhcpserver,defaultipgateway
AssetTag (not the SerialNumber): wskr.exe --range=WS123 --wmic systemenclosure get SMBIOSAssetTag
HTML for all COMPUTERSYSTEM:     wskr.exe --range=WS123 --wmic computersystem get /format:hform --save='cs-output.html'
```

The above ranges could be in the machine name range format:
```--range=SSnnn..SSmmm```
,or file name format:
```--range=myMachines.txt```	
  

## DEPENDENCIES
1) The machine you are running this on must be running Windows.
2) ```--ping``` is reliant on Windows ```PING.EXE```
3) ```--wmic``` is reliant on Windows ```WMIC.EXE```
4) ```--registry``` is reliant on Windows ```REG.EXE```

## ASSUMPTIONS
1) Your machine names have at least one character at the start, followed by at least one digit.
2) The machines you are scanning are running Windows.

## v0.1 - Copyright 2023

Author - Shaun Dunmall

