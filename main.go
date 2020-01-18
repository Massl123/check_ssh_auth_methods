package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func usage() {
	fmt.Println()
	fmt.Println("check_ssh_auth_methods Copyright (C) 2020 Marcel Freundl <https://github.com/Massl123>")
	fmt.Println("Licensed under GNU General Public License v3.0, see <https://www.gnu.org/licenses/gpl-3.0>")
	fmt.Println("Version v0.1.0 <https://github.com/Massl123/check_ssh_auth_methods>")

	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("./check_ssh_auth_methods -host <host> -u root -u admin -password a")
	fmt.Println()

	fmt.Println("allow:  authentication method must be allowed")
	fmt.Println("forbid: authentication method must not be allowed (default if not stated otherwise)")
	fmt.Println("ignore: authentication method is not checked")
	fmt.Println("Arguments:")
	flag.PrintDefaults()
}

// UNKNOWN state for nagios/icinga
const UNKNOWN = 3

// CRITICAL state for nagios/icinga
const CRITICAL = 2

// WARNING state for nagios/icinga
const WARNING = 1

// OK state for nagios/icinga
const OK = 0

func checkErr(pretext string, err error) {
	if err != nil {
		fmt.Printf("%s%s\n", pretext, err)
		os.Exit(UNKNOWN)
	}
}

type sshAuthMethod struct {
	Name            string
	DisplayName     string
	Allowed         bool
	Ignored         bool
	OfferedFromSSHD bool
}

// Set is used from Flag to set allowed and ignored states
func (s *sshAuthMethod) Set(state string) error {
	if len(state) < 1 {
		return fmt.Errorf("argument to short, requires at least one character")
	}

	st := state[0]

	if st != 'a' && st != 'f' && st != 'i' {
		return fmt.Errorf("unknown state %q, need a[llow], f[orbid], i[gnore]", state)
	}

	switch st {
	case 'i':
		s.Allowed = false
		s.Ignored = true
	case 'a':
		s.Allowed = true
		s.Ignored = false
	case 'f':
		s.Allowed = false
		s.Ignored = false
	}
	return nil
}

// Set states from ssh authentication methods line
func (s *sshAuthMethod) SetFromAuthLine(sshAuthLine string) {
	if strings.Contains(sshAuthLine, s.Name) {
		s.OfferedFromSSHD = true
	}
}

// String returns the default value for flag
func (s *sshAuthMethod) String() string {
	if s.Allowed {
		return "allowed"
	}
	if s.Ignored {
		return "ignored"
	}
	if !s.Allowed {
		return "forbidden"
	}
	return ""
}

func (s *sshAuthMethod) generateStatus() (bool, string) {
	if s.Ignored {
		return true, "ignored"
	}

	if s.Allowed && s.OfferedFromSSHD {
		return true, "allowed"
	}

	if !s.Allowed && !s.OfferedFromSSHD {
		return true, "forbidden"
	}

	if s.Allowed && !s.OfferedFromSSHD {
		return false, "forbidden but should be allowed"
	}

	if !s.Allowed && s.OfferedFromSSHD {
		return false, "allowed but should be forbidden"
	}
	panic("Unknown case!")
}

// true if check is ok, false if something doesnt match
func (s *sshAuthMethod) IsOK() bool {
	status, _ := s.generateStatus()
	return status
}

// Generate output for status lines
func (s *sshAuthMethod) GetOutput() string {
	_, text := s.generateStatus()
	return fmt.Sprintf("%s: %s", s.DisplayName, text)
}

// Store ssh users in this type
type users []string

func (i *users) String() string {
	return strings.Join(*i, ", ")
}

func (i *users) Slice() []string {
	return *i
}

func (i *users) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	// Vars for output
	var statusOutput []string
	var criticalSSHUsers []string
	exitCode := OK

	// Initialize sshAuthMethods and set default values
	var authMethods []*sshAuthMethod

	// None has to be the 0th element
	authMethods = append(authMethods, &sshAuthMethod{Name: "none", DisplayName: "None", Allowed: false, Ignored: false})
	authMethods = append(authMethods, &sshAuthMethod{Name: "hostbased", DisplayName: "Hostbased", Allowed: false, Ignored: false})
	authMethods = append(authMethods, &sshAuthMethod{Name: "password", DisplayName: "Password", Allowed: false, Ignored: false})
	authMethods = append(authMethods, &sshAuthMethod{Name: "keyboardinteractive", DisplayName: "KeyboardInteractive", Allowed: false, Ignored: false})
	authMethods = append(authMethods, &sshAuthMethod{Name: "publickey", DisplayName: "PublicKey", Allowed: true, Ignored: false})
	authMethods = append(authMethods, &sshAuthMethod{Name: "gssapikeyex", DisplayName: "GssapiKeyex", Allowed: false, Ignored: true})
	authMethods = append(authMethods, &sshAuthMethod{Name: "gssapiwithmic", DisplayName: "GssapiWithMic", Allowed: false, Ignored: true})

	// Setup basic flag arguments
	flag.Usage = usage

	sshServer := flag.String("host", "", "Host to connect to (required)")
	sshPort := flag.String("p", "22", "SSH port")
	var sshUsers users
	flag.Var(&sshUsers, "u", "SSH users to check, repeat argument for multiple users (default: root)")
	sshTimeout := flag.String("t", "10", "SSH timeout in seconds")

	// Setup authentication arguments
	for _, a := range authMethods {
		am := a
		flag.Var(am, am.Name, fmt.Sprintf("%s authentication, set to a[llow], f[orbid], i[gnore]", am.DisplayName))
	}

	flag.Parse()

	// Check that required arguments are set
	if *sshServer == "" {
		fmt.Println("-host has to be set!")
		fmt.Println("See -h for more details.")
		os.Exit(UNKNOWN)
	}

	if len(sshUsers) == 0 {
		sshUsers.Set("root")
	}

	// Try authentication with ssh -v and catch line
	// debug1: Authentications that can continue: publickey,gssapi-keyex,gssapi-with-mic,password
	// Parse string and check options

	sshBin, err := exec.LookPath("ssh")
	checkErr("Can't find ssh binary: ", err)

	// Check for every user
	for _, sshUser := range sshUsers.Slice() {
		var userAuthMethods = make([]*sshAuthMethod, len(authMethods))
		copy(userAuthMethods, authMethods)

		// Run ssh in batch mode to disable any prompt
		// Disable PublickeyAuthentication so we dont't login with SSH-Keys on the system - it is still shown in available methods
		ssh := exec.Command(sshBin, "-v", "-o", "BatchMode yes",
			"-o", "PubkeyAuthentication no",
			"-o", "StrictHostKeyChecking no",
			"-o", fmt.Sprintf("ConnectTimeout %s", *sshTimeout),
			"-l", sshUser, "-p", *sshPort, *sshServer, "exit")

		var sshOut []byte
		sshOut, err = ssh.CombinedOutput()

		// Normalize output for easier handling
		sshOut = bytes.ToLower(sshOut)
		sshLines := bufio.NewScanner(bytes.NewReader(sshOut))
		var sshAuthLine string

		if err == nil {
			// No error encountered -> successfull login
			// This means that authentication method "none" is enabled
			userAuthMethods[0].OfferedFromSSHD = true
		} else {
			// 255 means error ocurred, so auth was unsuccessful (unfortunately a generic exit code)
			// But if auth line is included all is good
			if err.Error() == "exit status 255" && bytes.Contains(sshOut, []byte("debug1: authentications that can continue:")) {
			} else {
				fmt.Printf("Error running ssh: %s\n%s\n", err, sshOut)
				os.Exit(UNKNOWN)
			}
		}

		// Ensure output is from openssh
		if !bytes.Contains(sshOut, []byte("openssh")) {
			fmt.Println("SSH binary is not OpenSSH - only OpenSSH is supported!")
			os.Exit(UNKNOWN)
		}

		// Get the authentication line for later use
		for sshLines.Scan() {
			line := sshLines.Text()
			if strings.HasPrefix(line, "debug1: authentications that can continue:") {
				sshAuthLine = line
				// Stop after first match
				break
			}
		}

		// Parse sshAuthLine and set states for userAuthMethods
		for _, a := range userAuthMethods {
			a.SetFromAuthLine(sshAuthLine)
		}

		// Generate output and exit code
		lineStatus := "OK"

		// Check every sshAuthMethod if it is ok
		for _, a := range userAuthMethods {
			ok := a.IsOK()
			if !ok {
				exitCode = CRITICAL
				lineStatus = "CRITICAL"
			}
		}

		// Append user to criticalSSHUsers if current user is not set up right
		if lineStatus == "CRITICAL" {
			criticalSSHUsers = append(criticalSSHUsers, sshUser)
		}

		// Generate status line for current user
		var outputStatus []string
		for _, a := range userAuthMethods {
			outputStatus = append(outputStatus, a.GetOutput())
		}

		statusOutput = append(statusOutput, fmt.Sprintf("%s: %5s (%s)", lineStatus, sshUser, strings.Join(outputStatus, ", ")))

	}

	// Print status line
	if exitCode == OK {
		fmt.Printf("OK, checked user(s) %s\n", strings.Join(sshUsers, ", "))
	} else {
		fmt.Printf("CRITICAL for user(s) %s\n", strings.Join(criticalSSHUsers, ", "))
	}

	// Print additional output with more info of what happened
	for _, line := range statusOutput {
		fmt.Println(line)
	}
	os.Exit(exitCode)
}
