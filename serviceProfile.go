package main

import (
	"fmt"
	"strings"
	"strconv"

	"github.com/lindsaybb/gopon"
)

func displayServiceProfiles(olt *gopon.LumiaOlt) error {
	// the top level data structure for provisioning services on an OLT is represented by a "Service Profile"
	// we will perform a GET Request to retrieve all currently configured Service Profiles on the OLT
	// a separate object holds a list of the individual profile objects to allow group tabwrite methods
	var err error
	var spl *gopon.ServiceProfileList
	spl, err = olt.GetServiceProfiles()
	if err != nil {
		return err
	}
	spl.Tabwrite()
	return nil
}

func modifyServiceProfiles(olt *gopon.LumiaOlt, arg string) error {

	var err error
	var spl *gopon.ServiceProfileList
	spl, err = olt.GetServiceProfiles()
	if err != nil {
		return err
	}
	spl.TabwriteFull()

	fmt.Print(">> Which Service Profile would you like to Modify?\n>> ")
	spName := sanitizeInput(readFromStdin())
	if spName == "" {
		return gopon.ErrNotInput
	}
	var sp *gopon.ServiceProfile
	sp, err = olt.GetServiceProfileByName(spName)
	if err != nil {
		return err
	}
	fmt.Print(">> Would you like to delete this profile? (y/N)\n>> ")
	input := strings.ToLower(sanitizeInput(readFromStdin()))
	if input == "y" {
		if sp.IsUsed() {
			// this needs to be made generic, able to search profiles and their sub-profile contents
			fmt.Println(">> Cannot delete in-use profile. Fetching the list of devices using this profile...")
			err = olt.UpdateOnuRegistry()
			if err != nil {
				return err
			}
			onuList := olt.GetOnuRegistryProfileUsage(sp.Name)
			printList(onuList)
			fmt.Print(">> Would you like to copy this Profile to a new name to be able to modify it? (y/N)\n>> ")
			rnBool := strings.ToLower(sanitizeInput(readFromStdin()))
			if rnBool == "y" {
				sp, err = modifyServiceProfileHandler(olt, sp, 0)
				if err != nil {
					return err
				}
			} else {
				return nil
			}
		} else {
			return olt.DeleteServiceProfile(sp.Name)
		}
	}
	if sp.IsUsed() {
		fmt.Println("!! Cannot modify in-use profile")
		sp, err = modifyServiceProfileHandler(olt, sp, 0)
		if err != nil {
			return err
		}
	}
	if arg == "" {
		arg = getArgFromSelection(gopon.ServiceProfileHeaders)
	}
	for {
		modVal := getIntFromArg(arg, gopon.ServiceProfileHeaders)
		sp, err = modifyServiceProfileHandler(olt, sp, modVal)
		if err != nil {
			return err
		}
		fmt.Println(">> Modified Service Profile:")
		sp.TabwriteFull()
		fmt.Print(">> Make further modifications? (y/N)\n>> ")
		modBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if modBool == "y" {
			arg = getArgFromSelection(gopon.ServiceProfileHeaders)
		} else {
			break
		}
	}
	fmt.Print(">> Post this modification? (Y/n)\n>> ")
	postBool := strings.ToLower(sanitizeInput(readFromStdin()))
	if postBool == "y" || postBool == "" {
		err = olt.DeleteServiceProfile(sp.Name)
		if err != nil {
			// if the profile has been renamed it can't be deleted and this not 200 OK is expected
			// if the profile is in-use it shouldn't have made it thus far without new name
			if err != gopon.ErrNotStatusOk {
				// other errors will prevent posting
				return err
			}
		}
		return olt.PostServiceProfile(sp.GenerateJson())
	}
	return nil
}

func modifyServiceProfileHandler(olt *gopon.LumiaOlt, sp *gopon.ServiceProfile, modVal int) (*gopon.ServiceProfile, error) {
	var err error
	switch modVal {
	case 0:
		fmt.Print(">> Provide new name for Service Profile\n>> ")
		newSpName := sanitizeInput(readFromStdin())
		sp, err = sp.Copy(newSpName)
		if err != nil {
			return nil, err
		}
	case 1:
		err = displayFlowProfiles(olt)
		if err != nil {
			return nil, err
		}
		fmt.Print(">> Which Flow Profile would you like to assign to the Service Profile instead?\n>> ")
		newFp := sanitizeInput(readFromStdin())
		if newFp == "" {
			return nil, gopon.ErrNotInput
		}
		var fp *gopon.FlowProfile
		fp, err = olt.GetFlowProfileByName(newFp)
		if err != nil {
			return nil, err
		}
		sp.FlowProfileName = fp.Name
	case 2:
		err = displayVlanProfiles(olt)
		if err != nil {
			return nil, err
		}
		fmt.Print(">> Which VLAN Profile would you like to assign to the Service Profile instead?\n>> ")
		newVp := sanitizeInput(readFromStdin())
		if newVp == "" {
			return nil, gopon.ErrNotInput
		}
		var vp *gopon.VlanProfile
		vp, err = olt.GetVlanProfileByName(newVp)
		if err != nil {
			return nil, err
		}
		sp.VlanProfileName = vp.Name
	case 3:
		err = displayOnuFlowProfiles(olt)
		if err != nil {
			return nil, err
		}
		fmt.Print(">> Which ONU Flow Profile would you like to assign to the Service Profile instead?\n>> ")
		newOfp := sanitizeInput(readFromStdin())
		if newOfp == "" {
			return nil, gopon.ErrNotInput
		}
		var ofp *gopon.OnuFlowProfile
		ofp, err = olt.GetOnuFlowProfileByName(newOfp)
		if err != nil {
			return nil, err
		}
		sp.OnuFlowProfileName = ofp.Name
	case 4:
		err = displayOnuTcontProfiles(olt)
		if err != nil {
			return nil, err
		}
		fmt.Print(">> Which ONU T-CONT Profile would you like to assign to the Service Profile instead?\n>> ")
		newOtp := sanitizeInput(readFromStdin())
		if newOtp == "" {
			return nil, gopon.ErrNotInput
		}
		var otp *gopon.OnuTcontProfile
		otp, err = olt.GetOnuTcontProfileByName(newOtp)
		if err != nil {
			return nil, err
		}
		sp.OnuTcontProfileName = otp.Name
	case 5:
		fmt.Println("!! ONU VLAN Profile to be implemented")
	case 6:
		fmt.Printf(">> Current Virtual GEM Port is [%d]. Enter the desired value:\n>> ", sp.OnuVirtGemPortID)
		newVgem := sanitizeInput(readFromStdin())
		if newVgem == "" {
			return nil, gopon.ErrNotInput
		}
		vg, err := strconv.Atoi(newVgem)
		if err != nil {
			return nil, gopon.ErrNotInput
		}
		if vg > 32 || vg < 1 {
			fmt.Print("!! Virtual GEM allowed range between 1-32... correcting input to 1")
			vg = 1
		}
		sp.OnuVirtGemPortID = vg
	case 7:
		fmt.Printf(">> Current ONU Termination Port Type is [%s]\n", gopon.ConvertOnuTPToString(sp.OnuTpType))
		newOnuTpType := getArgFromSelection(gopon.OnuTpTypeList)
		tp := getIntFromArg(newOnuTpType, gopon.OnuTpTypeList)
		if tp < 1 {
			return nil, gopon.ErrNotSettable
		}
		sp.OnuTpType = tp
	case 8:
		fmt.Println("!! Security Profile to be implemented")
	case 9:
		fmt.Println("!! IGMP Profile to be implemented")
	case 10:
		fmt.Println("!! ONU IGMP Profile to be implemented")
	case 11:
		fmt.Println("!! L2CP Profile to be implemented")
	case 12:
		fmt.Println("!! DHCP-RA Settings to be implemented")
	case 13:
		fmt.Println("!! PPPoE-IA Settings to be implemented")
	default:
		fmt.Println("!! Unexpected value, no change made")
	}
	return sp, nil
}