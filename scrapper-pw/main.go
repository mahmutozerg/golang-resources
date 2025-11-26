package main

import (
	"scrapper/constants"
	"scrapper/helper"
	"scrapper/scrapper"
)

func main() {

	abs := constants.GetAvailableBrowsers()

	scr, err := scrapper.NewScrapper(abs)

	helper.AssertErrorToNil(err, constants.GeneralFailure)
	defer scr.Close()

}
