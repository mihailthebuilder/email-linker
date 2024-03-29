package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *Controller) Redirect(c *gin.Context) {
	path := c.Param("path")
	if lengthOfString(path) == 0 {
		log.Println("no path given")
		c.Redirect(http.StatusTemporaryRedirect, r.ErrorRedirectUrl)
		return
	}

	url, err := r.Database.GetRedirectUrl(c, path)
	if err != nil {
		log.Printf("error getting redirect url for path %s: %s", path, err)
		c.Redirect(http.StatusTemporaryRedirect, r.ErrorRedirectUrl)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, url)

	record, err := r.Database.GetRedirectRecord(c, path)
	if err != nil {
		log.Printf("error getting redirect data for path %s: %s", path, err)
		return
	}

	isPreview, err := r.BotChecker.IsBotRequest(c.Request)
	if err != nil {
		log.Printf("error checking if request for link path %s is for a preview: %s", record.Path, err)
		return
	}

	if isPreview {
		log.Printf("not tracking redirect for link path %s as it is preview request", record.Path)
		return
	}

	err = r.Database.AddLinkClick(c, record.LinkId)
	if err != nil {
		log.Printf("error adding link click for link path %s: %s", record.Path, err)
		return
	}

	if record.NumberOfTimesClicked > r.MaxNumberOfEmailAlerts {
		log.Printf("not sending email notification for link path %s as number of clicks exceeded", record.Path)
		return
	}

	err = r.sendEmail(record)
	if err != nil {
		log.Printf("error notifying that link path %s was clicked: %s", record.Path, err)
		return
	}

	log.Printf("sent email notification for link path %s", record.Path)
}

func (r *Controller) sendEmail(record RedirectRecord) error {
	subject_template := "LinkUp link id %s clicked"
	subject := fmt.Sprintf(subject_template, record.Path)

	var content_template string
	if lengthOfString(record.Tag) > 0 {
		content_template = fmt.Sprintf("LinkUp link with tag '%s' has been clicked!", record.Tag)
		content_template = content_template + " The link's URL is %s and redirects to %s."
	} else {
		content_template = "LinkUp link %s that redirects to %s has been clicked!"
	}

	content := fmt.Sprintf(content_template, fmt.Sprintf("%s/%s", r.RedirectUri, record.Path), record.RedirectUrl)

	if record.NumberOfTimesClicked == r.MaxNumberOfEmailAlerts-1 {
		content = content + " This is the last email alert you'll receive for this link. Please contact mihailthebuilder@gmail.com if you wish to receive more alerts."
	}

	ser := SendEmailRequest{Email: record.UserEmail, Subject: subject, Content: content}
	err := r.Emailer.SendEmail(ser)
	if err != nil {
		return fmt.Errorf("error sending email: %s", err)
	}

	return nil
}
