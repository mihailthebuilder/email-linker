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

	ua := c.Request.Header.Get("User-Agent")
	log.Printf("redirect request for link with id %s has User-Agent: %s", record.LinkId, ua)

	isPreview, err := r.BotChecker.IsLinkPreviewRequest(ua)
	if err != nil {
		log.Printf("error checking if user-agent header %s is a preview request: %s", ua, err)
		return
	}

	if isPreview {
		log.Printf("not tracking redirect for link with id %s as it is preview request", record.LinkId)
		return
	}

	err = r.Database.AddLinkClick(c, record.LinkId)
	if err != nil {
		log.Printf("error adding link click for link with id %s: %s", record.LinkId, err)
		return
	}

	if record.NumberOfTimesClicked > r.MaxNumberOfEmailAlerts {
		log.Printf("not sending email notification for link with id %s as number of clicks exceeded", record.LinkId)
		return
	}

	err = r.sendEmail(record)
	if err != nil {
		log.Printf("error notifying that link with id %s was clicked: %s", record.LinkId, err)
		return
	}

	log.Printf("sent email notification for link with id %s", record.LinkId)
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
