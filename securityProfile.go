package main

import (
	"fmt"
	"strings"
	"strconv"

	"github.com/lindsaybb/gopon"
)

func displaySecurityProfiles(olt *gopon.LumiaOlt) error {
	var err error
	var secpl *gopon.SecurityProfileList
	secpl, err = olt.GetSecurityProfiles()
	if err != nil {
		return err
	}
	secpl.Tabwrite()
	return nil
}

func modifySecurityProfiles(olt *gopon.LumiaOlt, arg string) error {
	profType := "Security"
	err := displaySecurityProfiles(olt)
	if err != nil {
		return err
	}
	fmt.Printf(">> Which %s Profile would you like to Modify?\n>> ", profType)
	secpName := sanitizeInput(readFromStdin())
	if secpName == "" {
		return gopon.ErrNotInput
	}
	var secp *gopon.SecurityProfile
	secp, err = olt.GetSecurityProfileByName(secpName)
	if err != nil {
		return err
	}
	fmt.Print(">> Would you like to delete this profile? (y/N)\n>> ")
	input := strings.ToLower(sanitizeInput(readFromStdin()))
	if input == "y" {
		if secp.IsUsed() {
			fmt.Println("!! Cannot delete in-use profile.")
			//
			//	Want this to read "Here is a list of Service Profiles using this sub-profile"
			//	Then, here is a list of devices using those Service Profiles, like above
			//
			return nil
		} else {
			return olt.DeleteSecurityProfile(secp.Name)
		}
	}
	if secp.IsUsed() {
		fmt.Println("!! Cannot modify in-use profile")
		secp, err = modifySecurityProfileHandler(olt, secp, 0)
		if err != nil {
			return err
		}
	}
	if arg == "" {
		arg = getArgFromSelection(gopon.SecurityProfileHeaders)
	}
	for {
		modVal := getIntFromArg(arg, gopon.SecurityProfileHeaders)
		secp, err = modifySecurityProfileHandler(olt, secp, modVal)
		if err != nil {
			return err
		}
		fmt.Printf(">> Modified %s Profile:\n", profType)
		secp.Tabwrite()
		fmt.Print(">> Make further modifications? (y/N)\n>> ")
		modBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if modBool == "y" {
			arg = getArgFromSelection(gopon.SecurityProfileHeaders)
		} else {
			break
		}
	}
	fmt.Print(">> Post this modification? (Y/n)\n>> ")
	postBool := strings.ToLower(sanitizeInput(readFromStdin()))
	if postBool == "y" || postBool == "" {
		err = olt.DeleteSecurityProfile(secp.Name)
		if err != nil {
			// if the profile has been renamed it can't be deleted and this not 200 OK is expected
			// if the profile is in-use it shouldn't have made it thus far without new name
			if err != gopon.ErrNotStatusOk {
				// other errors will prevent posting
				return err
			}
		}
		return olt.PostSecurityProfile(secp.GenerateJson())
	}
	return nil
}

func modifySecurityProfileHandler(olt *gopon.LumiaOlt, secp *gopon.SecurityProfile, modVal int) (*gopon.SecurityProfile, error) {
	var err error
	switch modVal {
	case 0:
		// Name
		fmt.Print(">> Provide new name for Security Profile\n>> ")
		newSecpName := sanitizeInput(readFromStdin())
		secp, err = secp.Copy(newSecpName)
		if err != nil {
			return nil, err
		}
	case 1:
		// Port-Protect
		fmt.Println("++ A protected port does not forward any traffic (unicast, multicast, or broadcast) to any other port that is also a protected port. All data traffic passing between protected ports must be forwarded through a Layer 3 device. Forwarding behavior between a protected port and a non-protected port proceeds as usual")
		fmt.Printf(">> Current value is [%v], toggle status? (Y/n)\n>> ", secp.GetProtectedPort())
		togBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if togBool == "y" || togBool == "" {
			secp.SetProtectedPort(!secp.GetProtectedPort())
		}
	case 2:
		// MAC-SG
		fmt.Println("++ MAC Source Guard prevents customers from creating an [unintentional] loop on the CPE equipment by connecting two or more CPE device together, connecting two or more CPE devices to a hub, or connecting two or more ports on a CPE together. A MAC-SG violation blocks one of the ports, or both of them if the MAC also appears on the uplink interface")
		fmt.Printf(">> Current value is [%v], toggle status? (Y/n)\n>> ", secp.GetMacSG())
		togBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if togBool == "y" || togBool == "" {
			secp.SetMacSG(!secp.GetMacSG())
		}
	case 3:
		// MAC-Limit
		fmt.Println("++ Limit the number of MAC addresses (0...64), where a value of 0 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", secp.GetMacLimit())
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i > 0 && i < 65 {
				// this function limits settable amount to 16 despite YANG model limit of 64
				secp.SetMacLimit(i)
			} else {
				secp.SetMacLimit(0)
			}
		}
	case 4:
		// Port-Security
		fmt.Println("++ Port-Security learns the MAC addresses of connected devices and limits the amount of devices based on MacLimit")
		fmt.Printf(">> Current value is [%v], toggle status? (Y/n)\n>> ", secp.GetPortSecurity())
		togBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if togBool == "y" || togBool == "" {
			secp.SetMacSG(!secp.GetMacSG())
		}
	case 5:
		// ARP Inspection
		fmt.Println("++ Address Resolution Protocol (ARP) assists Layer 2 network segments find Layer 3 resources. Dynamic ARP Inspection (DAI) validates ARPs through DHCP Snooping, creating a Trusted Database or IP-to-MAC bindings")
		fmt.Printf(">> Current value is [%v], toggle status? (Y/n)\n>> ", secp.GetArpInspect())
		togBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if togBool == "y" || togBool == "" {
			secp.SetArpInspect(!secp.GetArpInspect())
		}

	case 6:
		// IPv4-SG
		fmt.Println("++ IP Source-Guard (IPv4) is a security feature that restricts IP traffic on untrusted Layer 2 ports by filtering traffic based on the DHCP snooping binding database. This feature helps prevent IP spoofing where a host tries to use the IP address of another host")
		fmt.Printf(">> Current value is [%v], modify configuration? (Y/n)\n>> ", secp.GetIPv4SG())
		togBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if togBool == "y" || togBool == "" {
			arg := getArgFromSelection(gopon.SecIpSgList)
			modVal := getIntFromArg(arg, gopon.SecIpSgList)
			secp, err = modifyIpSgParameters(secp, modVal)
			if err != nil {
				return nil, err
			}
		}
	case 7:
		// IPv6-SG
		fmt.Println("++ IP Source-Guard (IPv6) is a security feature that restricts IP traffic on untrusted Layer 2 ports by filtering traffic based on the DHCP snooping binding database. This feature helps prevent IP spoofing where a host tries to use the IP address of another host")
		fmt.Printf(">> Current value is [%v], modify configuration? (Y/n)\n>> ", secp.GetIPv6SG())
		togBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if togBool == "y" || togBool == "" {
			arg := getArgFromSelection(gopon.SecIpSgList)
			modVal := getIntFromArg(arg, gopon.SecIpSgList)
			secp, err = modifyIpSgParameters(secp, modVal)
			if err != nil {
				return nil, err
			}
		}
	case 8:
		// Storm-Control
		fmt.Println("++ Storm Control is a max data rate in Packets Per Second (pps) from (0...65535) where a value of -1 is disabled. This function is essential to reduce the potential of Broadcast Storms, and Denial of Service (DoS) events related to the 'endless loop' vulnerability of Multicast and Unknown-Unicast frames")
		fmt.Printf(">> Current value is [%v], modify configuration? (Y/n)\n>> ", secp.GetStormControlString())
		togBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if togBool == "y" || togBool == "" {
			arg := getArgFromSelection(gopon.SecStmCtlList)
			modVal := getIntFromArg(arg, gopon.SecStmCtlList)
			secp, err = modifyStormControlParameters(secp, modVal)
			if err != nil {
				return nil, err
			}
		}
	case 9:
		// AppRateLimit
		fmt.Println("++ ")
		fmt.Printf(">> Current value is [%v], modify configuration? (Y/n)\n>> ", secp.GetAppRateLimitString())
		togBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if togBool == "y" || togBool == "" {
			arg := getArgFromSelection(gopon.SecArlList)
			modVal := getIntFromArg(arg, gopon.SecArlList)
			secp, err = modifyARLParameters(secp, modVal)
			if err != nil {
				return nil, err
			}
		}
	default:
		fmt.Println("!! Unpexpected input, nothing to modify")
	}
	return secp, nil
}

// a composite data model nested underneath whether the service is enabled (def: not)
func modifyIpSgParameters(secp *gopon.SecurityProfile, modVal int) (*gopon.SecurityProfile, error) {
	// get arg from list
	switch modVal {
	case 0:
		// v4Enable
		fmt.Printf(">> IPv4 Source-Guard Enabled: [%v], toggle status? (Y/n)\n>> ", secp.GetIPv4SG())
		togBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if togBool == "y" || togBool == "" {
			secp.SetIPv4SG(!secp.GetIPv4SG())
		}
	case 1:
		// v6Enable
		fmt.Printf(">> IPv6 Source-Guard Enabled: [%v], toggle status? (Y/n)\n>> ", secp.GetIPv6SG())
		togBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if togBool == "y" || togBool == "" {
			secp.SetIPv6SG(!secp.GetIPv6SG())
		}
	case 2:
		// FilterMode
		fmt.Println("++ IP Source-Guard can filter based on IP Source Address (false state) or IP and MAC Source Address (true state/default).")
		fmt.Printf(">> IP-SG Filter Mode is currently: [%v], toggle status? (Y/n)\n>> ", secp.GetFilterModeString())
		togBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if togBool == "y" || togBool == "" {
			secp.SetFilterMode(!secp.GetFilterMode())
		}
	case 3:
		// v4BindingLimit
		fmt.Println("++ IPv4 Source-Guard binding limit defines the number of addresses to track (0...15), where 0 means no limit")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", secp.IPSgBindingLimit)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			l, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if l > 0 && l < 16 {
				secp.IPSgBindingLimit = l
			} else {
				secp.IPSgBindingLimit = 0
			}
		}
	case 4:
		// v6BindingLimitDHCP
		fmt.Println("++ IPv6 Source-Guard DHCP binding limit defines the number of addresses to track (0...15), where 0 means no limit")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", secp.IPSgBindingLimitDhcpv6)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			l, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if l > 0 && l < 16 {
				secp.IPSgBindingLimitDhcpv6 = l
			} else {
				secp.IPSgBindingLimitDhcpv6 = 0
			}
		}
	case 5:
		// v6BindingLimitND
		fmt.Println("++ IPv6 Source-Guard Neighbor Discovery (ND) binding limit defines the number of addresses to track (0...15), where 0 means no limit")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", secp.IPSgBindingLimitND)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			l, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if l > 0 && l < 16 {
				secp.IPSgBindingLimitND = l
			} else {
				secp.IPSgBindingLimitND = 0
			}
		}
	default:
		fmt.Println("!! Unpexpected input, nothing to modify")
	}
	return secp, nil
}

// a composite value expressed as BUM: Broadcast, Unknown-Unicast, Multicast
func modifyStormControlParameters(secp *gopon.SecurityProfile, modVal int) (*gopon.SecurityProfile, error) {
	switch modVal {
	case 0:
		// Broadcast
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", secp.StormControlBroadcast)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			l, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if l >= 0 && l < 65536 {
				secp.StormControlBroadcast = l
			} else {
				secp.StormControlBroadcast = -1
			}
		}
	case 1:
		// Unknown-Unicast
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", secp.StormControlUnicast)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			l, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if l >= 0 && l < 65536 {
				secp.StormControlUnicast = l
			} else {
				secp.StormControlUnicast = -1
			}
		}
	case 2:
		// Multicast
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", secp.StormControlMulticast)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			l, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if l >= 0 && l < 65536 {
				secp.StormControlMulticast = l
			} else {
				secp.StormControlMulticast = -1
			}
		}
	default:
		fmt.Println("!! Unpexpected input, nothing to modify")
	}
	return secp, nil
}

// a composite value expressed as an ordered 5-int list of Applications and their Rate limits
func modifyARLParameters(secp *gopon.SecurityProfile, modVal int) (*gopon.SecurityProfile, error) {
	switch modVal {
	case 0:
		// DHCP
		fmt.Println("++ Max data rate (pps) of DHCP (Dynamic Host Configuration Protocol) packets on a port (0...1000), where -1 disables the Rate-limiting functionality")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", secp.AppRateLimitDhcp)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			l, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if l >= 0 && l < 1001 {
				secp.AppRateLimitDhcp = l
			} else {
				secp.AppRateLimitDhcp = -1
			}
		}
	case 1:
		// IGMP
		fmt.Println("++ Max data rate (pps) of IGMP (Internet Group Messaging Protocol) packets on a port (0...1000), where -1 disables the Rate-limiting functionality")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", secp.AppRateLimitIgmp)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			l, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if l >= 0 && l < 1001 {
				secp.AppRateLimitIgmp = l
			} else {
				secp.AppRateLimitIgmp = -1
			}
		}
	case 2:
		// PPPOE
		fmt.Println("++ Max data rate (pps) of PPPOE (Point-to-Point Protocol Over Ethernet) packets on a port (0...1000), where -1 disables the Rate-limiting functionality")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", secp.AppRateLimitPppoe)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			l, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if l >= 0 && l < 1001 {
				secp.AppRateLimitPppoe = l
			} else {
				secp.AppRateLimitPppoe = -1
			}
		}
	case 3:
		// STP
		fmt.Println("++ Max data rate (pps) of STP (Spanning Tree Protocol) packets on a port (0...1000), where -1 disables the Rate-limiting functionality")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", secp.AppRateLimitStp)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			l, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if l >= 0 && l < 1001 {
				secp.AppRateLimitStp = l
			} else {
				secp.AppRateLimitStp = -1
			}
		}
	case 4:
		// MN
		fmt.Println("++ Max data rate (pps) of MN (Management Network) packets on a port (0...1000), where -1 disables the Rate-limiting functionality")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", secp.AppRateLimitStp)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			l, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if l >= 0 && l < 1001 {
				secp.AppRateLimitStp = l
			} else {
				secp.AppRateLimitStp = -1
			}
		}
	default:
		fmt.Println("!! Unpexpected input, nothing to modify")
	}
	
	return secp, nil
}
