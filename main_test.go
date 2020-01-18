package main

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Setup stuff can go here

	// Run actual tests
	flag.Parse()
	exitCode := m.Run()

	// Cleanup

	os.Exit(exitCode)
}

// Test for sshAuthMethod.Set()
func TestSshAuthMethodSet(t *testing.T) {
	type testCase struct {
		argument                          string
		isAllowed, isIgnored, shouldError bool
	}
	testCases := []testCase{
		testCase{argument: "a", isAllowed: true, isIgnored: false},
		testCase{argument: "f", isAllowed: false, isIgnored: false},
		testCase{argument: "i", isAllowed: false, isIgnored: true},
		testCase{argument: "x", shouldError: true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Testing argument %s", tc.argument), func(t *testing.T) {
			s := sshAuthMethod{}
			err := s.Set(tc.argument)

			if err != nil && !tc.shouldError {
				t.Errorf("Got error but shouldn't, %s", err)
			}

			if s.Allowed != tc.isAllowed {
				t.Errorf("Allowed is %t, should be %t", s.Allowed, tc.isAllowed)
			}
			if s.Ignored != tc.isIgnored {
				t.Errorf("Ignored is %t, should be %t", s.Ignored, tc.isIgnored)
			}
		})
	}
}

// Test for sshAuthMethod.SetFromAuthLine()
func TestSshAuthMethodSetFromAuthLine(t *testing.T) {
	type testCase struct {
		name, argument    string
		isOfferedFromSSHD bool
	}
	testCases := []testCase{
		testCase{name: "none", argument: "debug1: randomText: none", isOfferedFromSSHD: true},
		testCase{name: "none", argument: "debug1: randomText: password", isOfferedFromSSHD: false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Testing argument %s", tc.argument), func(t *testing.T) {
			s := sshAuthMethod{Name: tc.name}
			s.SetFromAuthLine(tc.argument)

			if s.OfferedFromSSHD != tc.isOfferedFromSSHD {
				t.Errorf("OfferedFromSSHD is %t, should be %t", s.OfferedFromSSHD, tc.isOfferedFromSSHD)
			}
		})
	}
}

// Test for sshAuthMethod.IsOK()
// includes sshAuthMethod.generateStatus()
func TestSshAuthMethodIsOK(t *testing.T) {
	type testCase struct {
		isOfferedFromSSHD, isAllowed, isIgnored, isOK bool
	}
	testCases := []testCase{
		testCase{isOfferedFromSSHD: true, isAllowed: true, isIgnored: true, isOK: true},
		testCase{isOfferedFromSSHD: true, isAllowed: true, isIgnored: false, isOK: true},
		testCase{isOfferedFromSSHD: true, isAllowed: false, isIgnored: true, isOK: true},
		testCase{isOfferedFromSSHD: true, isAllowed: false, isIgnored: false, isOK: false},
		testCase{isOfferedFromSSHD: false, isAllowed: true, isIgnored: true, isOK: true},
		testCase{isOfferedFromSSHD: false, isAllowed: true, isIgnored: false, isOK: false},
		testCase{isOfferedFromSSHD: false, isAllowed: false, isIgnored: true, isOK: true},
		testCase{isOfferedFromSSHD: false, isAllowed: false, isIgnored: false, isOK: true},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Testing #%d", i), func(t *testing.T) {
			s := sshAuthMethod{Allowed: tc.isAllowed, Ignored: tc.isIgnored, OfferedFromSSHD: tc.isOfferedFromSSHD}
			status := s.IsOK()

			if status != tc.isOK {
				t.Errorf("Status is %t, should be %t", status, tc.isOK)
			}
		})
	}
}
