package scrapper

import (
	"scrapper/constants"
	"scrapper/helper"

	"github.com/playwright-community/playwright-go"
)

func (s *Scrapper) NewPage() error {
	page, err := s.Context.NewPage()
	if err != nil {
		return helper.WrapError(err, constants.FailureNewPage)
	}

	s.Page = page
	return nil
}

func (s *Scrapper) GoTo(url string) error {
	if err := helper.ValidateURL(url); err != nil {
		return helper.WrapError(err, constants.InvalidUrl)
	}

	_, err := s.Page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})

	if err != nil {
		return helper.WrapError(err, constants.FailedGoTo)
	}

	return nil
}
