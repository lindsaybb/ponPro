package main

import (
	"fmt"
	"strings"
	"strconv"

	"github.com/lindsaybb/gopon"
)

func displayFlowProfiles(olt *gopon.LumiaOlt) error {
	var err error
	var fpl *gopon.FlowProfileList
	fpl, err = olt.GetFlowProfiles()
	if err != nil {
		return err
	}
	fpl.Tabwrite()
	return nil
}

func modifyFlowProfiles(olt *gopon.LumiaOlt, arg string) error {
	profType := "Flow"
	err := displayFlowProfiles(olt)
	if err != nil {
		return err
	}
	fmt.Printf(">> Which %s Profile would you like to Modify?\n>> ", profType)
	fpName := sanitizeInput(readFromStdin())
	if fpName == "" {
		return gopon.ErrNotInput
	}
	var fp *gopon.FlowProfile
	fp, err = olt.GetFlowProfileByName(fpName)
	if err != nil {
		return err
	}
	fmt.Print(">> Would you like to delete this profile? (y/N)\n>> ")
	input := strings.ToLower(sanitizeInput(readFromStdin()))
	if input == "y" {
		if fp.IsUsed() {
			fmt.Println("!! Cannot delete in-use profile.")
			//
			//	Want this to read "Here is a list of Service Profiles using this sub-profile"
			//	Then, here is a list of devices using those Service Profiles, like above
			//
			return nil
		} else {
			return olt.DeleteFlowProfile(fp.Name)
		}
	}
	if fp.IsUsed() {
		fmt.Println("!! Cannot modify in-use profile")
		fp, err = modifyFlowProfileHandler(olt, fp, 0)
		if err != nil {
			return err
		}
	}
	if arg == "" {
		arg = getArgFromSelection(gopon.FlowProfileHeaders)
	}
	for {
		modVal := getIntFromArg(arg, gopon.FlowProfileHeaders)
		fp, err = modifyFlowProfileHandler(olt, fp, modVal)
		if err != nil {
			return err
		}
		fmt.Printf(">> Modified %s Profile:\n", profType)
		fp.Tabwrite()
		fmt.Print(">> Make further modifications? (y/N)\n>> ")
		modBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if modBool == "y" {
			arg = getArgFromSelection(gopon.FlowProfileHeaders)
		} else {
			break
		}
	}
	fmt.Print(">> Post this modification? (Y/n)\n>> ")
	postBool := strings.ToLower(sanitizeInput(readFromStdin()))
	if postBool == "y" || postBool == "" {
		err = olt.DeleteFlowProfile(fp.Name)
		if err != nil {
			// if the profile has been renamed it can't be deleted and this not 200 OK is expected
			// if the profile is in-use it shouldn't have made it thus far without new name
			if err != gopon.ErrNotStatusOk {
				// other errors will prevent posting
				return err
			}
		}
		return olt.PostFlowProfile(fp.GenerateJson())
	}
	return nil
}

func modifyFlowProfileHandler(olt *gopon.LumiaOlt, fp *gopon.FlowProfile, modVal int) (*gopon.FlowProfile, error) {
	var err error
	switch modVal {
	case 0:
		fmt.Print(">> Provide new name for Flow Profile\n>> ")
		newFpName := sanitizeInput(readFromStdin())
		fp, err = fp.Copy(newFpName)
		if err != nil {
			return nil, err
		}
	case 1:
		// UsMatchVlanProfile
		fmt.Printf(">> Current setting is [%v]. Reverse bool? (Y/n)\n>> ", fp.GetMatchUsVlanProfile())
		flipBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if flipBool == "y" || flipBool == "" {
			if fp.GetMatchUsVlanProfile() {
				fp.MatchUsVlanProfile = 2
			} else {
				fp.MatchUsVlanProfile = 1
			}
		}
	case 2:
		// DsMatchVlanProfile
		fmt.Printf(">> Current setting is [%v]. Reverse bool? (Y/n)\n>> ", fp.GetMatchDsVlanProfile())
		flipBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if flipBool == "y" || flipBool == "" {
			if fp.GetMatchDsVlanProfile() {
				fp.MatchDsVlanProfile = 2
			} else {
				fp.MatchDsVlanProfile = 1
			}
		}
	case 3:
		// UsMatchOther
		fmt.Printf(">> Current UsMatchOther parameters are: [%v]\n", fp.GetMatchUsOther())
		arg := getArgFromSelection(gopon.FlowProfileUsOther)
		modVal := getIntFromArg(arg, gopon.FlowProfileUsOther)
		fp, err = modifyUsOther(olt, fp, modVal)
		if err != nil {
			return nil, err
		}
	case 4:
		// DsMatchOther
		fmt.Printf(">> Current DsMatchOther parameters are: [%v]\n", fp.GetMatchDsOther())
		arg := getArgFromSelection(gopon.FlowProfileDsOther)
		modVal := getIntFromArg(arg, gopon.FlowProfileDsOther)
		fp, err = modifyDsOther(olt, fp, modVal)
		if err != nil {
			return nil, err
		}
	case 5:
		// UsHandling
		fmt.Printf(">> Current UsHandling parameters are: [%v]\n", fp.GetUsHandling())
		arg := getArgFromSelection(gopon.FlowProfileUsHandling)
		modVal := getIntFromArg(arg, gopon.FlowProfileUsHandling)
		fp, err = modifyUsHandling(olt, fp, modVal)
		if err != nil {
			return nil, err
		}
	case 6:
		// DsHandling
		fmt.Printf(">> Current DsHandling parameters are: [%v]\n", fp.GetDsHandling())
		arg := getArgFromSelection(gopon.FlowProfileDsHandling)
		modVal := getIntFromArg(arg, gopon.FlowProfileDsHandling)
		fp, err = modifyDsHandling(olt, fp, modVal)
		if err != nil {
			return nil, err
		}
	case 7:
		// QueuingPriority
		fmt.Printf(">> Current Queuing Priority is [%s]. Provide new value: (0-7)\n>> ", fp.GetQueueingPriority())
		qpInput := sanitizeInput(readFromStdin())
		if qpInput == "" {
			return nil, gopon.ErrNotInput
		}
		qp, err := strconv.Atoi(qpInput)
		if err != nil {
			return nil, gopon.ErrNotInput
		}
		if qp < 0 || qp > 7 {
			fmt.Println("!! Settable range is 0-7, reverting input to default value")
			qp = 0
		}
		fp.DsQueuingPriority = qp
	case 8:
		// SchedulingMode
		fmt.Printf(">> Current Scheduling Mode is [%s]\n", fp.GetSchedulingMode())
		arg := getArgFromSelection(gopon.FlowProfileSchedulingModes)
		modVal := getIntFromArg(arg, gopon.FlowProfileSchedulingModes)
		if modVal > 0 {
			fp.DsSchedulingMode = modVal
		}
	default:
		fmt.Println("!! Unexpected input, nothing to modify.")
	}
	return fp, nil
}

func modifyUsOther(olt *gopon.LumiaOlt, fp *gopon.FlowProfile, modVal int) (*gopon.FlowProfile, error) {

	switch modVal {
	case 0:
		//	"MatchUsAny",
		fmt.Println("++ Match every upstream packet frame")
		fmt.Printf(">> Current value is [%v], toggle status? (Y/n)\n>> ", fp.IsMatchUsAny())
		togBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if togBool == "y" || togBool == "" {
			if fp.IsMatchUsAny() {
				fp.MatchUsAny = 2
			} else {
				fp.MatchUsAny = 1
			}
		}
	case 1:
		//	"MatchUsMacDestAddr",
		fmt.Println("++ Match upstream packet frame with specified destination MAC address. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsMacDestAddr)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchUsMacDestAddr = newStr
		}
	case 2:
		//	"MatchUsMacDestMask",
		fmt.Println("++ This mask value identifies the portion of MatchUsMacDestAddr that is compared with upstream packet. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsMacDestMask)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchUsMacDestMask = newStr
		}
	case 3:
		//	"MatchUsMacSrcAddr",
		fmt.Println("++ Match upstream packet frame with specified source MAC address. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsMacSrcAddr)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchUsMacSrcAddr = newStr
		}
	case 4:
		//	"MatchUsMacSrcMask",
		fmt.Println("++ This mask value identifies the portion of MatchUsMacSrcAddr that is compared with upstream packet. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsMacSrcMask)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchUsMacSrcMask = newStr
		}
	case 5:
		//	"MatchUsCPcp",
		fmt.Println("++ Match upstream packet frame with specified Customer PCP (Priority Code Point) which is also known as class of service (CoS) bits. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsCPcp)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 8 {
				fp.MatchUsCPcp = i
			} else {
				fp.MatchUsCPcp = -1
			}
		}
	case 6:
		//	"MatchUsSPcp",
		fmt.Println("++ Match upstream packet frame with specified Service PCP (Priority Code Point) which is also known as class of service (CoS) bits. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsSPcp)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 8 {
				fp.MatchUsSPcp = i
			} else {
				fp.MatchUsSPcp = -1
			}
		}
	case 7:
		//	"MatchUsCVlanIDRange",
		fmt.Println("++ Match upstream packet frame with specified list (bitmask) of Customer VLAN Id. An empty string indicates that parameter has not been defined.")
		fmt.Printf(">> Current value is [%v], this section does not allow direct modification\n", fp.MatchUsCVlanIDRange)
	case 8:
		//	"MatchUsSVlanIDRange",
		fmt.Println("++ Match upstream packet frame with specified list (bitmask) of Service VLAN Id. An empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], this section does not allow direct modification\n", fp.MatchUsSVlanIDRange)
	case 9:
		// "MatchUsEthertype",
		fmt.Println("++ Match upstream packet frame with specified EtherType value (int range -1...65535). A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsEthertype)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 65536 {
				fp.MatchUsEthertype = i
			} else {
				fp.MatchUsEthertype = -1
			}
		}
	case 10:
		//	"MatchUsIPProtocol",
		fmt.Println("++ Match upstream packet frame with specified IP protocol value. A value of -1 indicates that parameter has not been defined. Some of standard protocol values: icmp : 1, igmp : 2, ip: 4 (ip in ip encapsulation), tcp: 6, udp: 17")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsIPProtocol)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 256 {
				fp.MatchUsIPProtocol = i
			} else {
				fp.MatchUsIPProtocol = -1
			}
		}
	case 11:
		//	"MatchUsIPSrcAddr",
		fmt.Println("++ Match upstream packet frame with specified source IP address. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsIPSrcAddr)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchUsIPSrcAddr = newStr
		}
	case 12:
		//	"MatchUsIPSrcMask",
		fmt.Println("++ This mask value identifies the portion of MatchUsIpSrcAddr that is compared with upstream packet. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsIPSrcMask)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchUsIPSrcMask = newStr
		}
	case 13:
		//	"MatchUsIPDestAddr",
		fmt.Println("++ Match upstream packet frame with specified destination IP address. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsIPDestAddr)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchUsIPDestAddr = newStr
		}
	case 14:
		//	"MatchUsIPDestMask",
		fmt.Println("++ This mask value identifies the portion of MatchUsIpDestAddr that is compared with upstream packet. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsIPDestMask)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchUsIPDestMask = newStr
		}
	case 15:
		//	"MatchUsIPDscp",
		fmt.Println("++ Match upstream packet frame with specified CSC (Class Selector Code Point) = IP precedence (part of TOS field) value. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsIPDscp)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 64 {
				fp.MatchUsIPDscp = i
			} else {
				fp.MatchUsIPDscp = -1
			}
		}
	case 16:
		//	"MatchUsIPCsc",
		fmt.Println("++ Match upstream packet frame with specified IP precedence (part of TOS field) value. A value of -1 indicates that parameter has not been defined.")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsIPCsc)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 8 {
				fp.MatchUsIPCsc = i
			} else {
				fp.MatchUsIPCsc = -1
			}
		}
	case 17:
		//	"MatchUsIPDropPrecedence",
		fmt.Println("++ Match upstream packet frame with specified Drop precedence two bits value: noDrop(0): 00, lowDrop(1): 01, mediumDrop(2): 10, highDrop(3): 11. A value of -1 indicates that parameter has not been defined.")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsIPDropPrecedence)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 4 {
				fp.MatchUsIPDropPrecedence = i
			} else {
				fp.MatchUsIPDropPrecedence = -1
			}
		}
	case 18:
		//	"MatchUsTCPSrcPort",
		fmt.Println("++ Match upstream packet frame with specified source TCP port number. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsTCPSrcPort)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 65536 {
				fp.MatchUsTCPSrcPort = i
			} else {
				fp.MatchUsTCPSrcPort = -1
			}
		}
	case 19:
		//	"MatchUsTCPDestPort",
		fmt.Println("++ Match upstream packet frame with specified destination TCP port number. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsTCPDestPort)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 65536 {
				fp.MatchUsTCPDestPort = i
			} else {
				fp.MatchUsTCPDestPort = -1
			}
		}
	case 20:
		//	"MatchUsUDPSrcPort",
		fmt.Println("++ Match upstream packet frame with specified source UDP port number. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsUDPSrcPort)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 65536 {
				fp.MatchUsUDPSrcPort = i
			} else {
				fp.MatchUsUDPSrcPort = -1
			}
		}
	case 21:
		//	"MatchUsUDPDstPort",
		fmt.Println("++ Match upstream packet frame with specified destination UDP port number. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsUDPDstPort)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 65536 {
				fp.MatchUsUDPDstPort = i
			} else {
				fp.MatchUsUDPDstPort = -1
			}
		}
	case 22:
		//	"MatchUsIpv6SrcAddr",
		fmt.Println("++ Match upstream packet frame with specified source IPv6 address. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsIpv6SrcAddr)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchUsIpv6SrcAddr = newStr
		}
	case 23:
		//	"MatchUsIpv6SrcAddrMaskLen",
		fmt.Println("++ This mask value identifies the portion of MatchUsIpv6SrcAddr that is compared with upstream packet")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsIpv6DstAddrMaskLen)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i > 0 && i < 129 {
				fp.MatchUsIpv6DstAddrMaskLen = i
			} else {
				fp.MatchUsIpv6DstAddrMaskLen = 0
			}
		}
	case 24:
		//	"MatchUsIpv6DstAddr",
		fmt.Println("++ Match upstream packet frame with specified destination IPv6 address. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsIpv6DstAddr)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchUsIpv6DstAddr = newStr
		}
	case 25:
		//	"MatchUsIpv6SrcAddrMaskLen",
		fmt.Println("++ This mask value identifies the portion of MatchUsIpv6DestAddr that is compared with upstream packet")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchUsIpv6SrcAddrMaskLen)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i > 0 && i < 129 {
				fp.MatchUsIpv6SrcAddrMaskLen = i
			} else {
				fp.MatchUsIpv6SrcAddrMaskLen = 0
			}
		}
	default:
		fmt.Println("!! No matching option")
	}
	return fp, nil
}

func modifyDsOther(olt *gopon.LumiaOlt, fp *gopon.FlowProfile, modVal int) (*gopon.FlowProfile, error) {

	switch modVal {
	case 0:
		//	"MatchDsAny",
		fmt.Println("++ Match every downstream packet frame")
		fmt.Printf(">> Current value is [%v], toggle status? (Y/n)\n>> ", fp.IsMatchDsAny())
		togBool := strings.ToLower(sanitizeInput(readFromStdin()))
		if togBool == "y" || togBool == "" {
			if fp.IsMatchDsAny() {
				fp.MatchDsAny = 2
			} else {
				fp.MatchDsAny = 1
			}
		}
	case 1:
		//	"MatchDsMacDestAddr",
		fmt.Println("++ Match downstream packet frame with specified destination MAC address. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsMacDestAddr)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchDsMacDestAddr = newStr
		}
	case 2:
		//	"MatchDsMacDestMask",
		fmt.Println("++ This mask value identifies the portion of MatchDsMacDestAddr that is compared with downstream packet. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsMacDestMask)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchDsMacDestMask = newStr
		}
	case 3:
		//	"MatchDsMacSrcAddr",
		fmt.Println("++ Match downstream packet frame with specified source MAC address. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsMacSrcAddr)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchDsMacSrcAddr = newStr
		}
	case 4:
		//	"MatchDsMacSrcMask",
		fmt.Println("++ This mask value identifies the portion of MatchDsMacSrcAddr that is compared with downstream packet. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsMacSrcMask)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchDsMacSrcMask = newStr
		}
	case 5:
		//	"MatchDsCPcp",
		fmt.Println("++ Match downstream packet frame with specified Customer PCP (Priority Code Point) which is also known as class of service (CoS) bits. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsCPcp)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 8 {
				fp.MatchDsCPcp = i
			} else {
				fp.MatchDsCPcp = -1
			}
		}
	case 6:
		//	"MatchDsSPcp",
		fmt.Println("++ Match downstream packet frame with specified Service PCP (Priority Code Point) which is also known as class of service (CoS) bits. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsSPcp)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 8 {
				fp.MatchDsSPcp = i
			} else {
				fp.MatchDsSPcp = -1
			}
		}
	case 7:
		//	"MatchDsCVlanIDRange",
		fmt.Println("++ Match downstream packet frame with specified list (bitmask) of Customer VLAN Id. An empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], this section does not allow direct modification\n", fp.MatchDsCVlanIDRange)
	case 8:
		//	"MatchDsSVlanIDRange",
		fmt.Println("++ Match downstream packet frame with specified list (bitmask) of Service VLAN Id. An empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], this section does not allow direct modification\n", fp.MatchDsSVlanIDRange)
	case 9:
		//	"MatchDsEthertype",
		fmt.Println("++ Match downstream packet frame with specified EtherType value. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsEthertype)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 65536 {
				fp.MatchDsEthertype = i
			} else {
				fp.MatchDsEthertype = -1
			}
		}
	case 10:
		//	"MatchDsIPProtocol",
		fmt.Println("++ Match downstream packet frame with specified IP protocol value. A value of -1 indicates that parameter has not been defined. Some of standard protocol values: icmp : 1, igmp : 2, ip: 4 (ip in ip encapsulation), tcp: 6, udp: 17")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsIPProtocol)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 256 {
				fp.MatchDsIPProtocol = i
			} else {
				fp.MatchDsIPProtocol = -1
			}
		}
	case 11:
		//	"MatchDsIPSrcAddr",
		fmt.Println("++ Match downstream packet frame with specified source IP address. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsIPSrcAddr)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchDsIPSrcAddr = newStr
		}
	case 12:
		//	"MatchDsIPSrcMask",
		fmt.Println("++ This mask value identifies the portion of MatchDsIpSrcAddr that is compared with downstream packet. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsIPSrcMask)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchDsIPSrcMask = newStr
		}
	case 13:
		//	"MatchDsIPDestAddr",
		fmt.Println("++ Match downstream packet frame with specified destination IP address. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsIPDestAddr)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchDsIPDestAddr = newStr
		}
	case 14:
		//	"MatchDsIPDestMask",
		fmt.Println("++ This mask value identifies the portion of MatchDsIpDestAddr that is compared with downstream packet. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsIPDestMask)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchDsIPDestMask = newStr
		}
	case 15:
		//	"MatchDsIPDscp",
		fmt.Println("++ Match downstream packet frame with specified IP DSCP value. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsIPDscp)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 64 {
				fp.MatchDsIPDscp = i
			} else {
				fp.MatchDsIPDscp = -1
			}
		}
	case 16:
		//	MatchDsIpCsc
		fmt.Println("++ Match downstream packet frame with specified CSC (Class Selector Code Point) = IP precedence (part of TOS field) value. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsIPCsc)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 8 {
				fp.MatchDsIPCsc = i
			} else {
				fp.MatchDsIPCsc = -1
			}
		}
	case 17:
		//	MatchDsIpDropPrecedence
		fmt.Println("++ Match downstream packet frame with specified Drop precedence two bits value: noDrop(0): 00, lowDrop(1): 01, mediumDrop(2): 10, highDrop(3): 11. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsIPDropPrecedence)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 4 {
				fp.MatchDsIPDropPrecedence = i
			} else {
				fp.MatchDsIPDropPrecedence = -1
			}
		}
	case 18:
		//	"MatchDsTCPSrcPort",
		fmt.Println("++ Match downstream packet frame with specified source TCP port number. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsTCPSrcPort)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 65536 {
				fp.MatchDsTCPSrcPort = i
			} else {
				fp.MatchDsTCPSrcPort = -1
			}
		}
	case 19:
		//	"MatchDsTCPDestPort",
		fmt.Println("++ Match downstream packet frame with specified destination TCP port number. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsTCPDestPort)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 65536 {
				fp.MatchDsTCPDestPort = i
			} else {
				fp.MatchDsTCPDestPort = -1
			}
		}
	case 20:
		//	"MatchDsUDPSrcPort",
		fmt.Println("++ Match downstream packet frame with specified source UDP port number. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsUDPSrcPort)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 65536 {
				fp.MatchDsUDPSrcPort = i
			} else {
				fp.MatchDsUDPSrcPort = -1
			}
		}
	case 21:
		//	"MatchDsUDPDstPort",
		fmt.Println("++ Match downstream packet frame with specified destination UDP port number. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsUDPDstPort)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 65536 {
				fp.MatchDsUDPDstPort = i
			} else {
				fp.MatchDsUDPDstPort = -1
			}
		}
	case 22:
		//	"MatchDsIpv6SrcAddr",
		fmt.Println("++ Match downstream packet frame with specified source IPv6 address. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsIpv6SrcAddr)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchDsIpv6SrcAddr = newStr
		}
	case 23:
		//	"MatchDsIpv6SrcAddrMaskLen",
		fmt.Println("++ This mask value identifies the portion of MatchDsIpv6SrcAddr that is compared with downstream packet")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsIpv6DstAddrMaskLen)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i > 0 && i < 129 {
				fp.MatchDsIpv6DstAddrMaskLen = i
			} else {
				fp.MatchDsIpv6DstAddrMaskLen = 0
			}
		}
	case 24:
		//	"MatchDsIpv6DstAddr",
		fmt.Println("++ Match downstream packet frame with specified destination IPv6 address. Empty string indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.MatchDsIpv6DstAddr)
		newStr := sanitizeInput(readFromStdin())
		if newStr != "" {
			fp.MatchDsIpv6DstAddr = newStr
		}
	case 25:
		//	"MatchDsIpv6DstAddrMaskLen",
		fmt.Println("++ This mask value identifies the portion of MatchDsIpv6DestAddr that is compared with downstrem packet")
	default:
		fmt.Println("!! No matching option")
	}
	return fp, nil
}

func modifyUsHandling(olt *gopon.LumiaOlt, fp *gopon.FlowProfile, modVal int) (*gopon.FlowProfile, error) {

	switch modVal {
	case 0:
		//	"UsCdr",
		fmt.Println("++ Upstream committed data rate (E-CDR) in kbps (0...1000000)")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.UsCdr)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i > 0 && i < 1000001 {
				fp.UsCdr = i
			} else {
				fp.UsCdr = 0
			}
		}
	case 1:
		//	"UsCdrBurstSize",
		fmt.Println("++ Upstream committed data rate burst size in kB (0...16384). When parameter is set to 0 (default), it's automatically updated to default burst size in according with current QoSProfileInCdr value")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.UsCdrBurstSize)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i > 0 && i < 16385 {
				fp.UsCdrBurstSize = i
			} else {
				fp.UsCdrBurstSize = 0
			}
		}
	case 2:
		//	"UsPdr",
		fmt.Println("++ Upstream peak data rate (E-PDR) in kbps (0...1000000)")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.UsPdr)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i > 0 && i < 1000001 {
				fp.UsPdr = i
			} else {
				fp.UsPdr = 0
			}
		}
	case 3:
		//	"UsPdrBurstSize",
		fmt.Println("++ Upstream peak data rate burst size in kB (0...16384). When parameter is set to 0 (default), it's automatically updated to default burst size in according with current msanQoSProfileInPdr value")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.UsPdrBurstSize)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i > 0 && i < 16385 {
				fp.UsPdrBurstSize = i
			} else {
				fp.UsPdrBurstSize = 0
			}
		}
	case 4:
		//	"UsMarkPcp",
		fmt.Println("++ Type of upstrem PCP marking. If set to userValue(3), parameter UsMarkPcpValue is used. A value of copyFromCsc(2) is an option. A value of none(1) indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.UsMarkPcp)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i > 0 && i < 4 {
				fp.UsMarkPcp = i
			} else {
				fp.UsMarkPcp = 1
			}
		}
	case 5:
		//	"UsMarkPcpValue",
		fmt.Println("++ Mark upstream packets with specified PCP (Priority Code Point) value (0-7) = CoS. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.UsMarkPcpValue)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 8 {
				fp.UsMarkPcpValue = i
			} else {
				fp.UsMarkPcpValue = -1
			}
		}
	case 6:
		//	"UsMarkDscp",
		fmt.Println("++ Type of upstream DSCP marking. If set to userValue(3), parameter msanServiceFlowProfileUsMarkDscpValue is used. A value of copyFromPcp(2) is an option. A value of none(1) indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.UsMarkDscp)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i > 0 && i < 4 {
				fp.UsMarkDscp = i
			} else {
				fp.UsMarkDscp = 1
			}
		}
	case 7:
		//	"UsMarkDscpValue",
		fmt.Println("++ Mark upstream packets with specified DSCP (Diffserv Code Point) value (0-63). A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.UsMarkDscpValue)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 64 {
				fp.UsMarkDscpValue = i
			} else {
				fp.UsMarkDscpValue = -1
			}
		}
	default:
		fmt.Println("!! No matching option")
	}
	return fp, nil
}

func modifyDsHandling(olt *gopon.LumiaOlt, fp *gopon.FlowProfile, modVal int) (*gopon.FlowProfile, error) {

	switch modVal {
	case 0:
		//	"DsCdr",
		fmt.Println("++ Downstream committed data rate (E-CDR) in kbps (0...1000000)")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.DsCdr)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i > 0 && i < 1000001 {
				fp.DsCdr = i
			} else {
				fp.DsCdr = 0
			}
		}
	case 1:
		//	"DsCdrBurstSize",
		fmt.Println("++ Downstream committed data rate burst size in kB (0...16384). When parameter is set to 0 (default), it's automatically updated to default burst size in according with current QoSProfileOutCdr value")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.DsCdrBurstSize)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i > 0 && i < 16385 {
				fp.DsCdrBurstSize = i
			} else {
				fp.DsCdrBurstSize = 0
			}
		}
	case 2:
		//	"DsPdr",
		fmt.Println("++ Downstream peak data rate (E-PDR) in kbps (0...1000000)")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.DsPdr)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i > 0 && i < 1000001 {
				fp.DsPdr = i
			} else {
				fp.DsPdr = 0
			}
		}
	case 3:
		//	"DsPdrBurstSize",
		fmt.Println("++ Downstream peak data rate burst size in kB (0...16384). When parameter is set to 0 (default), it's automatically updated to default burst size in according with current msanQoSProfileOutCdr value")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.DsPdrBurstSize)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i > 0 && i < 16385 {
				fp.DsPdrBurstSize = i
			} else {
				fp.DsPdrBurstSize = 0
			}
		}
	case 4:
		//	"DsMarkPcp",
		fmt.Println("++ Type of downstream PCP marking. If set to userValue(3), parameter msanServiceFlowProfileDsMarkPcpValue is used. A value of copyFromCsc(2) is an option. A value of none(1) indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.DsMarkPcp)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i > 0 && i < 4 {
				fp.DsMarkPcp = i
			} else {
				fp.DsMarkPcp = 1
			}
		}
	case 5:
		//	"DsMarkPcpValue",
		fmt.Println("++ Mark downstream packets with specified PCP (Priority Code Point) value (0-7) = CoS. A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.DsMarkPcpValue)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 8 {
				fp.DsMarkPcpValue = i
			} else {
				fp.DsMarkPcpValue = -1
			}
		}
	case 6:
		//	"DsMarkDscp",
		fmt.Println("++ Type of downstrem DSCP marking. If set to userValue(3), parameter DsMarkDscpValue is used. A value of copyFromPcp(2) is an option. A value of none(1) indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.DsMarkDscp)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i > 0 && i < 4 {
				fp.DsMarkDscp = i
			} else {
				fp.DsMarkDscp = 1
			}
		}
	case 7:
		//	"DsMarkDscpValue",
		fmt.Println("++ Mark downstream packets with specified DSCP (Diffserv Code Point) value (0-63). A value of -1 indicates that parameter has not been defined")
		fmt.Printf(">> Current value is [%v], provide new value:\n>> ", fp.DsMarkDscpValue)
		newInt := sanitizeInput(readFromStdin())
		if newInt != "" {
			i, err := strconv.Atoi(newInt)
			if err != nil {
				return nil, gopon.ErrNotInput
			}
			if i >= 0 && i < 64 {
				fp.DsMarkDscpValue = i
			} else {
				fp.DsMarkDscpValue = -1
			}
		}
	default:
		fmt.Println("!! No matching option")
	}
	return fp, nil
}

