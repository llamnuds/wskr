# wskr
A simple command line program for scanning a range of machines for being alive, logged on user, etc.

wskr usage :-
wskr -n:ABC123..ABC999 [-s=start][-e=end][-p=PaddingString][-x=PrefixString][-w=1|0][-d=DelaySeconds][-y] -f|-r|-g|-m Some Thing To Check

MANDATORY - You must have one, and only one, of these :-

(But do NOT use = after any of these.)

--file     -f		Search for a file.

--registry -r		Search for a registry value.	

--ping	   -g		Search for LIVE machines.

--free     -3       Search for machines with no active user.

--wmic	   -m		Run your WMIC your command.

                    For an HTML formatted output postfix this:- /format:hform
		    
					For a LIST output use this :- /format:list
					
          
MANDATORY - You will of course need to state a RANGE of computers to look at.

--range:   -n=string[..string]    FirstMachine[.. LastMachine] (Or you could use the -p -x -s and -e options.)

--range:   -n='filename.txt'       Name of text file to read in, it should end in .txt.

The text file must be in the same directory that WSKR.EXE is run from.

Each line of the text file should start with a machine name, then a space; everything after the space is ignored.

Blank lines are ignored, as are any lines starting with a space or hash symbol.


OPTIONAL :-

--pad=	  -p=String 	Pad Computer name number with up to this many zeros.	Default = 000

--show=	  -w=String	Return successes(1), Failures(0).			Default = 1 i.e. Only successes (-w=10 to show all)

--delay=  -d=Integer	Seconds of Delay between machines. 			Default = 0 Seconds

--prefix= -x=String	Prefix of machine name.

--start=   -s=Integer	First machine number. 

--end=	  -e=Integer	Last machine number.

--save=   -v='String'     File name, to save in same location as EXE. Use single quotes.

--summary -y		Just give final counts.



EXAMPLES :-

To search PC0001 through PC1234, finding machines that do NOT have "c:\data\some file.txt" use :-

	wskr -x=PC -s=0 -e=1234 -p0000 -f c:\data\some file.txt
	
	 ...equivalent to...
	 
	wskr --range=pc0001..pc1234 --file c:\data\some file.txt
	
To search for a registry Value on a single computer :-

	wskr -n=comp456 -r HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon\Shell
	
To see various things such as :-

   Logged in users, saving result:  wskr.exe --range=WS123 --wmic computersystem get username --save='output.txt'
   
   OS version:                      wskr.exe --range=WS123 --wmic os get version
   
   Installed software MSI's:        wskr.exe --range=WS123 --wmic product get name,vendor,version
   
   System serial number:            wskr.exe --range=WS123 --wmic bios get serialnumber	
   
   Installed printers:              wskr.exe --range=WS123 --wmic printerconfig list
   
   The IP,DHCPserver, Gateway:      wskr.exe --range=WS123 --wmic nicconfig get IPAddress,dhcpserver,defaultipgateway
   
   AssetTag (not the SerialNumber): wskr.exe --range=WS123 --wmic systemenclosure get SMBIOSAssetTag
   
   HTML for all COMPUTERSYSTEM:     wskr.exe --range=WS123 --wmic computersystem get /format:hform --save='cs-output.html'
   
Oviously the above ranges could be in the :-

	* Multiple machine format: --range=SSnnn..SSmmm
	
	* File name format:        --range=myMachines.txt
	
  
  
v0.1 - Copyright 2023

Author - Shaun Dunmall

