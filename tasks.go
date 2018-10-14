package main

/*
   Copyright 2018 TheRedSpy15

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

import (
	"bufio"
	"compress/gzip"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"

	"github.com/daviddengcn/go-colortext"
	"github.com/jordan-wright/email"
)

// TODO: add sha-256 hash
// Takes a file path, and then prints the hash of the file
func hashFile(target string) {
	checkTarget(target)             // make sure target is valid
	ct.Foreground(ct.Yellow, false) // set text color to dark yellow

	file := readFileIntoByte(target)                          // get bytes of file to hash
	hash := sha1.New()                                        // create sha1 object
	hash.Write(file)                                          // hash file to object
	target = base64.URLEncoding.EncodeToString(hash.Sum(nil)) // encode hash sum into string

	println("SHA-1 hash :", target)
}

// ListTasks - lists all currently working tasks
func listTasks() {
	ct.Foreground(ct.Yellow, false)
	println("Available tasks:")
	println("Hash -r [file path]")
	println("encryptFile -r [file path]")
	println("decryptFile -r [file path]")
	println("Scrape -r [URL]")
	println("DOS -r [IP/URL]")
	println("Email")
	println("generatePassword")
	println("systemInfo")
	println("Audit -r [Online/Offline]")
	println("pwnAccount -r [Email]")

	println("About") // keep at bottom of print statements
}

// TODO: make & add 'printDisk'
// Prints extensive info about system
func systemInfoTask() {
	ct.Foreground(ct.Yellow, false) // set text color to dark yellow
	printCPU()                      // print cpu info
	printMemory()                   // print memory info
	printHost()                     // print host info
}

// TODO: break up into Util functions
// Check if an account has been pwned
func pwnAccount(target string) {
	checkTarget(target) // make sure target is valid

	pwnURL := fmt.Sprintf(`https://haveibeenpwned.com/api/v2/breachedaccount/%v`, target)
	response, err := http.Get(pwnURL) // make response object
	if err != nil {
		ct.Foreground(ct.Red, true) // set text color to bright red
		panic(err.Error)
	}

	defer response.Body.Close()                   // close on function end
	bodyBytes, _ := ioutil.ReadAll(response.Body) // read bytes from response

	if len(bodyBytes) == 0 { // nothing found - all good
		ct.Foreground(ct.Green, true) // set text color to bright green
		println("Good news — no pwnage found!")
	} else { // account found in breach
		ct.Foreground(ct.Red, true) // set text color to bright red
		println("Oh no — account has been pwned!")
	}
}

// Encrypts the target file
func encryptFileTask(target string) {
	checkTarget(target)             // make sure target is valid
	ct.Foreground(ct.Yellow, false) // set text color to dark yellow

	data := readFileIntoByte(target) // read file bytes
	print("Enter Password: ")
	password := getPassword() // get password securely

	encryptFile(target, data, password) // encrypt file
	println("\nFile encrypted!")
}

// BUG: decrypted file is unusable
// NOTE: decrypt file doesn't actually save as unencrypted
// Decrypts the target file
func decryptFileTask(target string) {
	checkTarget(target)             // make sure target is valid
	ct.Foreground(ct.Yellow, false) // set text color to dark yellow

	print("Enter Password: ")
	password := getPassword() // get password securely

	file, err := os.Create(target) // create file object
	if err != nil {
		ct.Foreground(ct.Red, true) // set text color to bright red
		panic(err.Error())
	}
	defer file.Close()                        // close file on function end
	file.Write(decryptFile(target, password)) // decrypt file
	println("\nFile decrypted!")
}

// TODO: run the right command that cleans "thumbs" & the system cache
// Clean cached files
func cleanTask() {
	ct.Foreground(ct.Red, true)
	println("Not a working feature yet!")
	cmd := exec.Command("rm", "-rf", "~/.thumbs/*") // don't think this is the right command
	cmd.Run()
}

// Prints details about the program
func about() {
	printBanner()

	ct.Foreground(ct.Yellow, false) // set text color to dark yellow
	println("Multi Go v0.6.1", "\nBy: TheRedSpy15")
	println("GitHub:", "https://github.com/TheRedSpy15")
	println("Project Page:", "https://github.com/TheRedSpy15/Multi-Go")
	println("\nMulti Go allows IT admins and Cyber Security experts")
	println("to conveniently perform all sorts of tasks.")
}

// Scrapes target website
func scapeTask(target string) {
	checkTarget(target)               // make sure target is valid
	collyAddress(target, true, false) // run colly - scraping happens here
}

// BUG: exit status 1
// Runs linuxScanner.py to audit system vulnerabilities
func auditTask(target string) {
	checkTarget(target)             // make sure target is valid
	ct.Foreground(ct.Yellow, false) // set text color to dark yellow

	if strings.TrimRight(target, "\n") == "Online" { // online audit
		runAuditOnline()
	} else if strings.TrimRight(target, "\n") == "Offline" { // offline audit - not started
		ct.Foreground(ct.Red, true) // set text color to bright red
		println("Not a feature yet!")
	} else { // not valid option
		ct.Foreground(ct.Red, true) // set text color to bright red
		println("Not a valid mode!")
		println(`Use "Online" or "Offline"`)
	}
}

func compressTask(target string) {
	checkTarget(target)

	file, err := os.Create(target)
	if err != nil {
		ct.Foreground(ct.Red, true)
		panic(err.Error())
	}
	defer file.Close()

	os.Rename(target, target+".gz")

	w := gzip.NewWriter(file)
	w.Write(readFileIntoByte(target))
	defer w.Close()

	ct.Foreground(ct.Green, true)
	println("finished!")
}

// TODO: if contains .gz
func decompressTask(target string) {
	ct.Foreground(ct.Red, true)
	println("Not a working feature yet!")
}

// TODO: use set length
// Generates a random string for use as a password
func generatePasswordTask() {
	ct.Foreground(ct.Yellow, false) // set text color to dark yellow
	println("Password:", randomString())
}

// TODO: add amplification - such as NTP monlist
// Indefinitely sends data to target
func dosTask(target string) {
	checkTarget(target) // make sure target is valid

	conn, err := net.Dial("udp", target) // setup connection object
	defer conn.Close()                   // make sure to close connection when finished
	if err != nil {
		ct.Foreground(ct.Red, true)
		panic(err.Error())
	} else { // nothing bad happened when connecting to target
		ct.Foreground(ct.Green, true)
		println("Checks passed!")
	}

	ct.Foreground(ct.Red, true)                                        // set text color to bright red
	println("\nWarning: you are solely responsible for your actions!") // disclaimer
	println("ctrl + c to cancel")
	println("\n10 seconds until DOS")
	ct.ResetColor() // reset text color to default

	time.Sleep(10 * time.Second) // 10 second delay - give chance to cancel

	threads, err := cpu.Counts(false) // get threads on system to set DOS thread limit
	if err != nil {
		ct.Foreground(ct.Red, true) // set text color to bright red
		panic(err.Error())
	}

	for i := 0; i < threads; i++ { // create DOS threads within limit
		go dos(conn)                   // create thread
		ct.Foreground(ct.Yellow, true) // set text color to dark yellow
		println("Thread created!")
	}
}

// BUG: mail: missing word in phrase: mail: invalid string
// TODO: use native go email
// TODO: break up into Util functions
// TODO: find out if attachment works with path, or just name
// Send email
func emailTask() {
	reader := bufio.NewReader(os.Stdin) // make reader object
	e := email.NewEmail()               // make email object
	ct.Foreground(ct.Yellow, false)     // set text color to dark yellow
	println("Prepare email")
	ct.ResetColor() // reset text color to default

	// email setup
	print("From: ")
	e.From, _ = reader.ReadString('\n') // from

	print("To: ")
	To, _ := reader.ReadString('\n') // to
	e.To = []string{To}

	print("Bcc (leave blank if none): ") // bcc
	Bcc, _ := reader.ReadString('\n')
	e.Bcc = []string{Bcc}

	print("Cc (leave blank if none): ") // cc
	Cc, _ := reader.ReadString('\n')
	e.To = []string{Cc}

	print("Subject: ")
	e.Subject, _ = reader.ReadString('\n') // subject

	print("Text: ")
	Text, _ := reader.ReadString('\n') // text
	e.Text = []byte(Text)

	print("File path (if sending one): ") // attachment
	Path, _ := reader.ReadString('\n')
	if Path != "" {
		e.AttachFile(Path)
	}

	// authentication
	print("Provider (example: smtp.gmail.com): ") // provider
	provider, _ := reader.ReadString('\n')
	print("Port (example: 587): ") // port
	port, _ := reader.ReadString('\n')
	print("Password (leave blank if none): ") // password
	password := getPassword()

	// confirmation
	print("Confirm send? (yes/no): ")
	confirm, _ := reader.ReadString('\n')          // get string of user confirm choice
	if strings.TrimRight(confirm, "\n") == "yes" { // yes - confirm send
		// sending
		err := e.Send(provider+":"+port, smtp.PlainAuth("", e.From, password, provider)) // send & get error
		if err != nil {
			ct.Foreground(ct.Red, true) // set text color to bright red
			println("error sending email -", err.Error())
		}
	} else { // cancelled
		ct.Foreground(ct.Red, true)
		println("Cancelled!")
	}
}