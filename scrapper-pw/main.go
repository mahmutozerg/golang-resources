package main

import (
	"scrapper/constants"
	"scrapper/helper"
	"scrapper/scrapper"
)

func main() {

	abs := constants.GetAvailableBrowsers()

	helper.AssertNotEmpty(abs[0], "Available browser")
	scr, err := scrapper.NewScrapper(abs[0])

	helper.AssertErrorToNil(err, constants.GeneralFailure)
	defer scr.Close()

}
