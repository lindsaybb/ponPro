package main

import (
	"fmt"
	"strings"
	"strconv"

	"github.com/lindsaybb/gopon"
)

func displayVlanProfiles(olt *gopon.LumiaOlt) error {
	var err error
	var vpl *gopon.VlanProfileList
	vpl, err = olt.GetVlanProfiles()
	if err != nil {
		return err
	}
	vpl.Tabwrite()
	return nil
}

func modifyVlanProfiles(olt *gopon.LumiaOlt, arg string) error {
	profType := "Vlan"
	err := displayVlanProfiles(olt)
	if err != nil {
		return err
	}
	fmt.Printf(">> Which %s Profile would you like to Modify?\n>> ", profType)
	vpName := sanitizeInput(readFromStdin())
	if vpName == "" {
		return gopon.ErrNotInput
	}
	var vp *gopon.VlanProfile
	vp, err = olt.GetVlanProfileByName(vpName)
	if err != nil {
		return err
	}
	fmt.Print(">> Would you like to delete this profile? (y/N)\n>> ")
	input := strings.ToLower(sanitizeInput(readFromStdin()))
	if input == "y" {
		if vp.IsUsed() {
			fmt.Println("!! Cannot delete in-use profile.")
			//
			//	Want this to read "Here is a list of Service Profiles using this sub-profile"
			//	Then, here is a list of devices using those Service Profiles, like above
			//
			return nil
		} else {
			return olt.DeleteVlanProfile(vp.Name)
		}
	}
	if vp.IsUsed() {
		fmt.Println("!! Cannot modify in-use profile")
		vp, err = modifyVlanProfileHandler(olt, vp, 0)
		if err != nil {
			return err
		}
	}
	if arg == "" {
		arg = getArgFromSelection(gopon.VlanProfileHeaders)
	}
	for {
		modVal := getIntFromArg(arg, gopon.VlanProfileHeaders)
		vp, err = modifyVlanProfileHandler(olt, vp, modVal)
		if err != nil {
			return err
		}
		fmt.Printf(">> Modified %s Profile:\n", profType)
		vp.Tabwrite()
		fmt.Print(">> Make further modifications? (y/N)\n>> ")
		modBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if modBool == "y" {
			arg = getArgFromSelection(gopon.VlanProfileHeaders)
		} else {
			break
		}
	}
	fmt.Print(">> Post this modification? (Y/n)\n>> ")
	postBool := strings.ToLower(sanitizeInput(readFromStdin()))
	if postBool == "y" || postBool == "" {
		err = olt.DeleteVlanProfile(vp.Name)
		if err != nil {
			// if the profile has been renamed it can't be deleted and this not 200 OK is expected
			// if the profile is in-use it shouldn't have made it thus far without new name
			if err != gopon.ErrNotStatusOk {
				// other errors will prevent posting
				return err
			}
		}
		return olt.PostVlanProfile(vp.GenerateJson())
	}
	return nil
}

func modifyVlanProfileHandler(olt *gopon.LumiaOlt, vp *gopon.VlanProfile, modVal int) (*gopon.VlanProfile, error) {
	var err error
	switch modVal {
	case 0:
		// Name
		fmt.Print(">> Provide new name for Vlan Profile\n>> ")
		newVpName := sanitizeInput(readFromStdin())
		vp, err = vp.Copy(newVpName)
		if err != nil {
			return nil, err
		}
	case 1:
		// C-Vid
		fmt.Println("++ Customer VLANs Identification (bit mask)")
		fmt.Printf(">> Current value is [%v], provide new space-separated list of VLAN IDs to use:\n>> ", vp.GetCVid())
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
			err = vp.SetCVid(vIntList)
			if err != nil {
				return nil, err
			}
		}
	case 2:
		// C-Vid Native
		fmt.Println("++ Native Customer VLAN Identifier. A value of -1 indicates that parameter has not been defined. This value must be included in C-VID Range")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", vp.CVidNative)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 4095 {
				vp.CVidNative = i
			} else {
				vp.CVidNative = -1
			}
		}
	case 3:
		// S-Vid
		fmt.Println("++ Service VLAN Identifier. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", vp.SVid)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 4095 {
				vp.SVid = i
			} else {
				vp.SVid = -1
			}
		}
	case 4:
		// S-Ethertype
		fmt.Println("++ S-Tag Ethertype value (decimal). Default value is 34984: 0x88a8 (S-Tag on Q-in-Q). Common values are 33024: 0x8100 (Single-Tag) and 37124: 0x9100 (Double-Tag). See https://en.wikipedia.org/wiki/EtherType and convert to Decimal")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", vp.SEtherType)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			// lowest value 0x0800 (IPv4) = 2048, further checking a good idea here
			// 0x9100 (double-tagged) = 37120
			if i >= 2048 && i < 38000 {
				vp.SEtherType = i
			} else {
				vp.SEtherType = 34984
			}
		}
	default:
		fmt.Println("!! Unpexpected input, nothing to modify")
	}
	return vp, nil
}


