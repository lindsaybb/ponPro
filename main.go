package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"strconv"

	"github.com/lindsaybb/gopon"
)

var (
	helpFlag      = flag.Bool("h", false, "Show this help")
	showSpDetails = flag.Bool("sp", false, "View Detailed Information about Service Profiles")
	modifyProfile = flag.Bool("mp", false, "Modify Service Profiles and the Profiles they contain, interactively")
)

// purpose: modify service profiles on the fly based on a template from a file

const usage = "`gopon sp demo` [options] <olt_ip>"

func main() {
	flag.Parse()

	if *helpFlag || flag.NArg() < 1 {
		fmt.Println(usage)
		flag.PrintDefaults()
		return
	}
	var err error
	host := flag.Args()[0]
	olt := gopon.NewLumiaOlt(host)
	if !olt.HostIsReachable() {
		fmt.Printf("!! Host %s is not reachable\n", host)
		return
	}
	if *showSpDetails {
		fmt.Println(">> Show Service Profile Details called [-sp]")
		err = displayProfilesHandler(olt, -1)
		if err != nil {
			fmt.Printf("!! Error running demo: %v\n", err)
		}
		if promptRerun() {
			main()
		}
	}
	if *modifyProfile {
		fmt.Println(">> Modify Service Profile Details called [-mp]")
		err = modifyProfileHandler(olt)
		if err != nil {
			fmt.Printf("!! Error running demo: %v\n", err)
		}
		if promptRerun() {
			main()
		}
	}
}

var ProfileHandlerList = []string{
	"Service Profiles",
	"Flow Profiles",
	"VLAN Profiles",
	"ONU Flow Profiles",
	"ONU TCONT Profiles",
	"ONU VLAN Profiles",
	"IGMP Profiles",
	"ONU IGMP Profiles",
	"Security Profiles",
}

// use this to decide which displays are called by adding a flag to the function call
func displayProfilesHandler(olt *gopon.LumiaOlt, modVal int) error {
	var err error
	switch modVal {
	case -1:
		for i := 0; i < len(ProfileHandlerList); i++ {
			err = displayProfilesHandler(olt, i)
			if err != nil {
				return err
			}
		}
	case 0:
		// Service Profiles
		err = displayServiceProfiles(olt)
		if err != nil {
			return err
		}
	case 1:
		// Flow Profiles
		err = displayFlowProfiles(olt)
		if err != nil {
			return err
		}
	case 2:
		// VLAN Profiles
		err = displayVlanProfiles(olt)
		if err != nil {
			return err
		}
	case 3:
		// ONU Flow Profiles
		err = displayOnuFlowProfiles(olt)
		if err != nil {
			return err
		}
	case 4:
		// ONU T-CONT Profiles
		err = displayOnuTcontProfiles(olt)
		if err != nil {
			return err
		}
	case 5:
		// ONU VLAN Profiles 
		err = displayOnuVlanProfiles(olt)
		if err != nil {
			return err
		}
	case 6:
		// IGMP Profiles
		err = displayIgmpProfiles(olt)
		if err != nil {
			return err
		}
	case 7:
		// ONU IGMP Profiles
		err = displayOnuIgmpProfiles(olt)
		if err != nil {
			return err
		}
	case 8:
		// Security Profiles
		err = displaySecurityProfiles(olt)
		if err != nil {
			return err
		}
	}
	return nil
}

// allow this function to supplied with arg on call instead of handling the arg
func modifyProfileHandler(olt *gopon.LumiaOlt) error {

	arg := getArgFromSelection(gopon.ServiceProfileHeaders)
	modVal := getIntFromArg(arg, gopon.ServiceProfileHeaders)
	switch modVal {
	case 0:
		return modifyServiceProfiles(olt, "")
	case 1:
		return modifyFlowProfiles(olt, "")
	case 2:
		return modifyVlanProfiles(olt, "")
	case 3:
		return modifyOnuFlowProfiles(olt, "")
	case 4:
		return modifyOnuTcontProfiles(olt, "")
	case 5:
		return modifyOnuVlanProfiles(olt)
	case 6:
		return modifyServiceProfiles(olt, "gem")
	case 7:
		return modifyServiceProfiles(olt, "tp")
	case 8:
		return modifySecurityProfiles(olt, "")
	case 9:
		return modifyIgmpProfiles(olt)
	case 10:
		return modifyOnuIgmpProfiles(olt)
	case 11:
		fmt.Println("!! L2CP not implemented currently")
		return nil
	case 12:
		return modifyServiceProfiles(olt, "dhcp")
	case 13:
		return modifyServiceProfiles(olt, "ppp")
	default:
		return gopon.ErrNotInput
	}
}

func printList(list []string) {
	for i, v := range list {
		fmt.Printf("[%3d]\t%s\n", i, v)
	}
}

func stringListToIntList(str []string) (l []int, err error) {
	for _, v := range str {
		i, err := strconv.Atoi(v)
		if err == nil {
			if i < 4095 && i > 1 {
				l = append(l, i)
			}
		}
	}
	if len(l) == 0 {
		return l, gopon.ErrNotExists
	}
	if len(l) != len(str) {
		return l, gopon.ErrNotInput
	}
	return l, nil
}

func promptRerun() bool {
	fmt.Print(">> Re-Run? (Y/n)")
	input := readFromStdin()
	if input == "" || strings.ToLower(input) == "y" {
		return true
	} else if strings.ToLower(input) == "n" {
		return false
	} else {
		promptRerun()
	}
	return false
}

func getIntFromArg(arg string, list []string) int {
	if arg == "" {
		return -1
	}
	for i, v := range list {
		// simple check is already sanitized
		// worst case error is a false positive
		if strings.Contains(strings.ToLower(sanitizeInput(v)), arg) {
			// for example, input is ONU TCONT
			// contains won't provide longest match, but input string can be longer
			return i
		}
		// number can be supplied instead of name, preferred short-cut
		if arg == fmt.Sprintf("%d", i) {
			return i
		}
	}
	fmt.Printf("!! Input does not match any items in list: %s\n", arg)
	return -1
}

func getArgFromSelection(list []string) string {
	fmt.Println(">> Which Element would you like to modify?")
	printList(list)
	fmt.Print(">> ")
	return strings.ToLower(sanitizeInput(readFromStdin()))
}

func readFromStdin() string {
	r := bufio.NewReaderSize(os.Stdin, 1024*1024)
	a, _, err := r.ReadLine()
	if err == io.EOF {
		return ""
	} else if err != nil {
		panic(err)
	}

	return strings.TrimRight(string(a), "\r\n")
}

func sanitizeInput(input string) string {
	var allowedChar = []rune{
		'-', '_', '/', '.', '[', ']', '(', ')',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	}
	var isOk bool
	var output string
	for _, c := range input {
		for _, f := range allowedChar {
			if f == c {
				isOk = true
			}
		}
		if isOk {
			output += string(c)
			isOk = false
		}
	}
	return output
}
