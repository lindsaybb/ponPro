package main

import (
	"fmt"

	"github.com/lindsaybb/gopon"
)

func displayOnuIgmpProfiles(olt *gopon.LumiaOlt) error {
	var err error
	var oipl *gopon.OnuIgmpProfileList
	oipl, err = olt.GetOnuMulticastProfiles()
	if err != nil {
		return err
	}
	oipl.Tabwrite()
	return nil
}

func modifyOnuIgmpProfiles(olt *gopon.LumiaOlt) error {
	fmt.Println("Modify the ONU IGMP Profile Details placeholder")
	return nil
}