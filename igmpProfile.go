package main

import (
	"fmt"

	"github.com/lindsaybb/gopon"
)

func displayIgmpProfiles(olt *gopon.LumiaOlt) error {
	var err error
	var ipl *gopon.IgmpProfileList
	ipl, err = olt.GetMulticastProfiles()
	if err != nil {
		return err
	}
	ipl.Tabwrite()
	return nil
}

func modifyIgmpProfiles(olt *gopon.LumiaOlt) error {
	fmt.Println("Modify the IGMP Profile Details placeholder")
	return nil
}