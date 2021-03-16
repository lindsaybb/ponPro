package main

import (
	"fmt"
	"strings"
	"strconv"

	"github.com/lindsaybb/gopon"
)

func displayOnuFlowProfiles(olt *gopon.LumiaOlt) error {
	var err error
	var ofpl *gopon.OnuFlowProfileList
	ofpl, err = olt.GetOnuFlowProfiles()
	if err != nil {
		return err
	}
	ofpl.Tabwrite()
	return nil
}

func modifyOnuFlowProfiles(olt *gopon.LumiaOlt, arg string) error {
	profType := "ONU Flow"
	err := displayOnuFlowProfiles(olt)
	if err != nil {
		return err
	}
	fmt.Printf(">> Which %s Profile would you like to Modify?\n>> ", profType)
	ofpName := sanitizeInput(readFromStdin())
	if ofpName == "" {
		return gopon.ErrNotInput
	}
	var ofp *gopon.OnuFlowProfile
	ofp, err = olt.GetOnuFlowProfileByName(ofpName)
	if err != nil {
		return err
	}
	fmt.Print(">> Would you like to delete this profile? (y/N)\n>> ")
	input := strings.ToLower(sanitizeInput(readFromStdin()))
	if input == "y" {
		if ofp.IsUsed() {
			fmt.Println("!! Cannot delete in-use profile.")
			//
			//	Want this to read "Here is a list of Service Profiles using this sub-profile"
			//	Then, here is a list of devices using those Service Profiles, like above
			//
			return nil
		} else {
			return olt.DeleteOnuFlowProfile(ofp.Name)
		}
	}
	if ofp.IsUsed() {
		fmt.Println("!! Cannot modify in-use profile")
		ofp, err = modifyOnuFlowProfileHandler(olt, ofp, 0)
		if err != nil {
			return err
		}
	}
	if arg == "" {
		arg = getArgFromSelection(gopon.OnuFlowProfileHeaders)
	}
	for {
		modVal := getIntFromArg(arg, gopon.OnuFlowProfileHeaders)
		ofp, err = modifyOnuFlowProfileHandler(olt, ofp, modVal)
		if err != nil {
			return err
		}
		fmt.Printf(">> Modified %s Profile:\n", profType)
		ofp.Tabwrite()
		fmt.Print(">> Make further modifications? (y/N)\n>> ")
		modBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if modBool == "y" {
			arg = getArgFromSelection(gopon.OnuFlowProfileHeaders)
		} else {
			break
		}
	}
	fmt.Print(">> Post this modification? (Y/n)\n>> ")
	postBool := strings.ToLower(sanitizeInput(readFromStdin()))
	if postBool == "y" || postBool == "" {
		err = olt.DeleteOnuFlowProfile(ofp.Name)
		if err != nil {
			// if the profile has been renamed it can't be deleted and this not 200 OK is expected
			// if the profile is in-use it shouldn't have made it thus far without new name
			if err != gopon.ErrNotStatusOk {
				// other errors will prevent posting
				return err
			}
		}
		return olt.PostOnuFlowProfile(ofp.GenerateJson())
	}
	return nil
}

func modifyOnuFlowProfileHandler(olt *gopon.LumiaOlt, ofp *gopon.OnuFlowProfile, modVal int) (*gopon.OnuFlowProfile, error) {
	var err error
	switch modVal {
	case 0:
		// Name
		fmt.Print(">> Provide new name for ONU Flow Profile\n>> ")
		newOfpName := sanitizeInput(readFromStdin())
		ofp, err = ofp.Copy(newOfpName)
		if err != nil {
			return nil, err
		}
	case 1:
		// MatchUsC-VidRange
		fmt.Println("++ Match ONU upstream packet frame with specified list (bitmask) of Customer VLAN Id. An empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new space-separated list of VLAN IDs to use:\n>> ", ofp.GetMatchUsCVlanIDRange())
		// not sanitizing this input but error-check it during processing to int
		vList := strings.Fields(readFromStdin())
		if len(vList) != 0 {
			vIntList, err := stringListToIntList(vList)
			if err != nil {
				if err == gopon.ErrNotInput {
					fmt.Println("Not all input was accepted, creating a partial list")
				} else {
					return nil, err
				}
			}
			err = ofp.SetMatchUsCVlanIDRange(vIntList)
			if err != nil {
				return nil, err
			}
		}
	case 2:
		// MatchUsCPcp
		fmt.Println("++ Match ONU upstream packet frame with specified Customer PCP (Priority Code Point) which is also known as class of service (CoS) bits. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", ofp.MatchUsCPcp)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 8 {
				ofp.MatchUsCPcp = i
			} else {
				ofp.MatchUsCPcp = -1
			}
		}
	case 3:
		// UsCdr
		fmt.Println("++ ONU upstream committed data rate (E-CDR) in kbps. Any rate value can be entered, but is rounded up to the multiple of 64 kbps. Limitation: Commited rate cannot be higher than peak rate")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", ofp.UsCdr)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 128 && i <= ofp.UsPdr {
				ofp.UsCdr = i
			} else {
				fmt.Println("!! Not settable")
			}
		}
	case 4:
		// UsPdr
		fmt.Println("++ ONU upstream peak data rate (E-PDR) in kbps. Any rate value can be entered, but is rounded up to the multiple of 64 kbps. Limitation: Peak rate cannot be lower than guaranteed rate")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", ofp.UsPdr)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= ofp.UsCdr && i < 2500001 {
				ofp.UsPdr = i
			} else {
				fmt.Println("!! Not settable")
			}
		}
	case 5:
		// UsFlowPriority
		fmt.Println("++ ONU upstream flow priority (0...7)")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", ofp.UsFlowPriority)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 8 {
				ofp.UsFlowPriority = i
			} else {
				ofp.UsFlowPriority = 0
			}
		}
	case 6:
		// DsFlowPriority
		fmt.Println("++ ONU downstream flow priority (0...7)")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", ofp.DsFlowPriority)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 8 {
				ofp.DsFlowPriority = i
			} else {
				ofp.DsFlowPriority = 0
			}
		}
	default:
		fmt.Println("!! Unpexpected input, nothing to modify")
	}
	return ofp, nil
}