package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lindsaybb/gopon"
)

func displayOnuTcontProfiles(olt *gopon.LumiaOlt) error {
	var err error
	var otpl *gopon.OnuTcontProfileList
	otpl, err = olt.GetOnuTcontProfiles()
	if err != nil {
		return err
	}
	otpl.Tabwrite()
	return nil
}

func modifyOnuTcontProfiles(olt *gopon.LumiaOlt, arg string) error {
	profType := "ONU T-CONT"
	err := displayOnuTcontProfiles(olt)
	if err != nil {
		return err
	}
	fmt.Printf(">> Which %s Profile would you like to Modify?\n>> ", profType)
	otpName := sanitizeInput(readFromStdin())
	if otpName == "" {
		return gopon.ErrNotInput
	}
	var otp *gopon.OnuTcontProfile
	otp, err = olt.GetOnuTcontProfileByName(otpName)
	if err != nil {
		return err
	}
	fmt.Print(">> Would you like to delete this profile? (y/N)\n>> ")
	input := strings.ToLower(sanitizeInput(readFromStdin()))
	if input == "y" {
		if otp.IsUsed() {
			fmt.Println("!! Cannot delete in-use profile.")
			//
			//	Want this to read "Here is a list of Service Profiles using this sub-profile"
			//	Then, here is a list of devices using those Service Profiles, like above
			//
			return nil
		} else {
			return olt.DeleteOnuTcontProfile(otp.Name)
		}
	}
	if otp.IsUsed() {
		fmt.Println("!! Cannot modify in-use profile")
		otp, err = modifyOnuTcontProfileHandler(olt, otp, 0)
		if err != nil {
			return err
		}
	}
	if arg == "" {
		arg = getArgFromSelection(gopon.OnuTcontProfileHeaders)
	}
	for {
		modVal := getIntFromArg(arg, gopon.OnuTcontProfileHeaders)
		otp, err = modifyOnuTcontProfileHandler(olt, otp, modVal)
		if err != nil {
			return err
		}
		fmt.Printf(">> Modified %s Profile:\n", profType)
		otp.Tabwrite()
		fmt.Print(">> Make further modifications? (y/N)\n>> ")
		modBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if modBool == "y" {
			arg = getArgFromSelection(gopon.OnuTcontProfileHeaders)
		} else {
			break
		}
	}
	fmt.Print(">> Post this modification? (Y/n)\n>> ")
	postBool := strings.ToLower(sanitizeInput(readFromStdin()))
	if postBool == "y" || postBool == "" {
		err = olt.DeleteOnuTcontProfile(otp.Name)
		if err != nil {
			// if the profile has been renamed it can't be deleted and this not 200 OK is expected
			// if the profile is in-use it shouldn't have made it thus far without new name
			if err != gopon.ErrNotStatusOk {
				// other errors will prevent posting
				return err
			}
		}
		return olt.PostOnuTcontProfile(otp.GenerateJson())
	}
	return nil
}

func modifyOnuTcontProfileHandler(olt *gopon.LumiaOlt, otp *gopon.OnuTcontProfile, modVal int) (*gopon.OnuTcontProfile, error) {
	var err error
	switch modVal {
	case 0:
		// Name
		fmt.Print(">> Provide new name for ONU T-CONT Profile or Supply -1 to use Auto-Generated\n>> ")
		newOtpName := sanitizeInput(readFromStdin())
		if newOtpName == "" {
			return nil, gopon.ErrNotInput
		} else if newOtpName == "-1" {
			newOtpName = otp.GenerateTcontName()
		}
		otp, err = otp.Copy(newOtpName)
		if err != nil {
			return nil, err
		}
	case 1:
		// Description
		desc := otp.GetTcontDescription()
		fmt.Printf("++ %s\n", desc)
	case 2:
		// Type
		fmt.Println("++ T-Cont Types are a value from 1-5 identifying the handling of committed and burst rates")
		otp.PrintTcontInfo()
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", otp.TcontType)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			t, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			// this function handles the error check and defaults to type 5 on error
			// this function also preserves and attempts to re-populate FAM rates into new type
			otp.SetTcontType(t)
		}
	case 3:
		// ID
		fmt.Println("++ T-Cont IDs are a value from 1-6 that allow stacking multiple T-Conts on the same ONU, by providing non-overlapping values")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", otp.TcontID)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			// this function handles the error check and defaults to id 1 on error
			otp.SetTcontID(i)
		}
	case 4:
		// Fixed
		fmt.Println("++ ONU T-CONT Fixed data rate. Any rate value can be entered, but is rounded up to the multiple of 64 kbps. Limitation: Maximum rate cannot be lower than the sum of fixed and assured rates")
		fmt.Printf(">> Current value is [%v] and ability to set with this T-CONT Type is [%v].\n", otp.FixedDataRate, otp.CanSetFixed())
		if otp.CanSetFixed() {
			fmt.Print(">> Provide a new value:\n>> ")
			newInt := sanitizeInput(readFromStdin())
			if newInt != "" {
				// could allow input like 10M and translate, but keep it simple for now
				r, err := strconv.Atoi(newInt)
				if err != nil {
					return nil, gopon.ErrNotInput
				}
				// this function handles the settable bounds
				otp.SetFixedRate(r)
			}
		}
	case 5:
		// Assured
		fmt.Println("++ ONU T-CONT Assured data rate  (256 - 2500000 kbps). Default value is 0, meaning that no rate is configured. Any rate value can be entered, but is rounded up to the multiple of 64 kbps. Limitation: Maximum rate cannot be lower than the sum of fixed and assured rates")
		fmt.Printf(">> Current value is [%v] and ability to set with this T-CONT Type is [%v].\n", otp.AssuredDataRate, otp.CanSetAssured())
		if otp.CanSetAssured() {
			fmt.Print(">> Provide a new value:\n>> ")
			newInt := sanitizeInput(readFromStdin())
			if newInt != "" {
				// could allow input like 10M and translate, but keep it simple for now
				r, err := strconv.Atoi(newInt)
				if err != nil {
					return nil, gopon.ErrNotInput
				}
				// this function handles the settable bounds
				otp.SetAssuredRate(r)
			}
		}
	case 6:
		// Max
		fmt.Println("++ ONU T-CONT Maximum data rate. Any rate value can be entered, but is rounded up to the multiple of 64 kbps. Limitation: Maximum rate cannot be lower than the sum of fixed and assured rates")
		fmt.Printf(">> Current value is [%v] and ability to set with this T-CONT Type is [%v].\n", otp.MaxDataRate, otp.CanSetMax())
		if otp.CanSetMax() {
			fmt.Print(">> Provide a new value:\n>> ")
			newInt := sanitizeInput(readFromStdin())
			if newInt != "" {
				// could allow input like 10M and translate, but keep it simple for now
				r, err := strconv.Atoi(newInt)
				if err != nil {
					return nil, gopon.ErrNotInput
				}
				// this function handles the settable bounds
				otp.SetMaxRate(r)
			}
		}
	default:
		fmt.Println("!! Unpexpected input, nothing to modify")
	}
	return otp, nil
}