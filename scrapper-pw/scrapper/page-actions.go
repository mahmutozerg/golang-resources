package scrapper

import (
	"scrapper/constants"
	"scrapper/helper"

	"github.com/playwright-community/playwright-go"
)

func (s *Scrapper) NewPage() {

	page, err := s.Context.NewPage()
	helper.AssertErrorToNil(err, constants.FailureNewPage)
	s.Page = page
}

func (s *Scrapper) GoTo(url string) {

	if err := helper.IsValidURL(url); err != nil {
		panic(constants.GeneralFailure)
	}

	_, err := s.Page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})

	helper.AssertErrorToNil(err, constants.FailedGoTo)
}
