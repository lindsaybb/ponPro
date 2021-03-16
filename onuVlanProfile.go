package main

import (
	"fmt"
	"strings"

	"github.com/lindsaybb/gopon"
)

// data structure not made available yet, placeholder

func displayOnuVlanProfiles(olt *gopon.LumiaOlt) error {
	var err error
	var ovpl *gopon.OnuVlanProfileList
	var ovrl *gopon.OnuVlanRuleList
	ovpl, ovrl, err = olt.GetOnuVlanProfiles()
	// need to merge this table with the rules subset and sort accordingly
	// use this display function call to complete implementation of gopon library
	if err != nil {
		return err
	}
	ovpl.Tabwrite()
	ovrl.Tabwrite()
	return nil
}

func modifyOnuVlanProfiles(olt *gopon.LumiaOlt) error {
	profType := "ONU VLAN"
	err := displayOnuVlanProfiles(olt)
	if err != nil {
		return err
	}
	fmt.Printf(">> Which %s Profile would you like to Modify?\n>> ", profType)
	ovpName := sanitizeInput(readFromStdin())
	if ovpName == "" {
		return gopon.ErrNotInput
	}
	var ovp *gopon.OnuVlanProfile
	ovp, err = olt.GetOnuVlanProfileByName(ovpName)
	if err != nil {
		return err
	}
	fmt.Print(">> Would you like to delete this profile? (y/N)\n>> ")
	input := strings.ToLower(sanitizeInput(readFromStdin()))
	if input == "y" {
		if ovp.IsUsed() {
			fmt.Println("!! Cannot delete in-use profile.")
			//
			//	Want this to read "Here is a list of Service Profiles using this sub-profile"
			//	Then, here is a list of devices using those Service Profiles, like above
			//
			return nil
		} else {
			return olt.DeleteOnuVlanProfile(ovp.Name)
		}
	}

	return nil
}