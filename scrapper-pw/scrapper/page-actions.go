package scrapper

import (
	"fmt"
	"scrapper/constants"
	"scrapper/helper"

	"github.com/playwright-community/playwright-go"
)

func (s *Scrapper) NewPage() error {
	page, err := s.context.NewPage()
	if err != nil {
		return helper.WrapError(err, constants.FailureNewPage)
	}

	s.page = page
	return nil
}
func (s *Scrapper) GoTo(url string, po ...playwright.PageGotoOptions) error {
	if err := helper.ValidateURL(url); err != nil {
		return helper.WrapError(err, constants.InvalidUrl)
	}

	opt := playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}

	if len(po) > 0 {
		opt = po[0]
	}

	if _, err := s.page.Goto(url, opt); err != nil {
		return helper.WrapError(err, constants.FailedGoTo)
	}

	return nil
}

func (s *Scrapper) CollectUrls(sp string) error {

	if len(sp) == 0 {
		return fmt.Errorf("initial url can not be null")
	}

	var err error
	var locator playwright.Locator
	for i := 0; i < s.maxDepth; i++ {

		if err = s.GoTo(sp, playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateNetworkidle}); err != nil {
			return err
		}

		if locator, err = s.locateElement("a"); err != nil {
			return err
		}

		count, _ := locator.Count()
		fmt.Println("Total links:", count)

		for i := range count {
			href, _ := locator.Nth(i).GetAttribute("href")
			fmt.Println("href:", href)
		}

	}

	return nil
}

func (s *Scrapper) locateElement(e string) (playwright.Locator, error) {

	if len(e) == 0 {
		return nil, fmt.Errorf("initial url can not be null")
	}

	return s.page.Locator(e), nil

}
