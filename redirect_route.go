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
		log.Printf("fail to get redirect url for path %s: %s", path, err)
		c.Redirect(http.StatusTemporaryRedirect, r.ErrorRedirectUrl)
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, url)

	record, err := r.Database.GetRedirectRecord(c, path)
	if err != nil {
		log.Printf("fail to get redirect data for path %s: %s", path, err)
		return
	}

	err = r.Database.AddLinkClick(c, record.LinkId)
	if err != nil {
		log.Printf("fail to add link click for link with id %s: %s", record.LinkId, err)
	}

	if record.NumberOfTimesClicked < r.MaxNumberOfEmailAlerts {
		subject := "LinkUp link clicked"
		content_template := "LinkUp link %s has been clicked!"
		content := fmt.Sprintf(content_template, fmt.Sprintf("%s/%s", r.RedirectUri, path))

		ser := SendEmailRequest{Email: record.UserEmail, Subject: subject, Content: content}
		err = r.Emailer.SendEmail(ser)
		if err != nil {
			log.Printf("fail to notify %s that link %s was clicked: %s", record.UserEmail, record.RedirectUrl, err)
		}

		log.Printf("sent notification email for link with id %s", record.LinkId)
	}
}
