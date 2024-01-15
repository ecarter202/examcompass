package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

const (
	baseURL      = "https://www.examcompass.com/comptia-network-plus-certification-practice-test-%d-exam-n10-008"
	baseFilepath = "/tmp/files"
)

var browser string

func main() {
	err := os.Mkdir(baseFilepath, 0777)
	if err != nil {
		log.Fatalf("making files dir [ERR: %s]", err)
	}

	browser = launcher.New().
		Headless(false).
		Bin("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome").
		MustLaunch()

	for i := 1; i <= 23; i++ {
		page := fmt.Sprintf(baseURL, i)

		scrapePage(page)
	}
}

func scrapePage(URL string) {

	fmt.Println("starting page", URL)

	page := rod.New().
		ControlURL(browser).
		MustConnect().
		MustPage(URL)

	defer page.Close()

	var counter = 0

	for {
		counter++
		fmt.Println("QUESTION", counter)

		nextBtn, nbErr := page.Element("button.btn-next")
		finishBtn, fbErr := page.Element("button.btn-finish")

		e := proto.NetworkResponseReceived{}
		wait := page.WaitEvent(&e)

		// ISSUE HERE
		// this should be false after getting to the last question
		fmt.Println("VISIBLE:", nextBtn.MustVisible())

		if nextBtn != nil && nbErr == nil {
			fmt.Println("NEXT")
			nextBtn.MustClick()
			wait()
		} else if finishBtn != nil && fbErr == nil {
			fmt.Println("FINISHING")
			finishBtn.MustClick()
			wait()
		} else {
			fmt.Println("STORING HTML")
			html := page.MustHTML()
			if !strings.Contains(html, "Your answer to this question is incorrect or incomplete.") {

				u, err := url.Parse(URL)
				if err != nil {
					log.Fatalf("SOB [ERR: %s]", err)
				}

				filename := strings.ReplaceAll(u.Path, "/", "")
				filepath := fmt.Sprintf("%s/%s.html", baseFilepath, filename)
				if err = os.WriteFile(filepath, []byte(html), 0755); err != nil {
					log.Fatalf("unable to write file [ERR: %s]", err)
				}

				os.Exit(1) // delete me
			} else {
				fmt.Println("UH OH! unable to see results text...")
				break
			}
		}
	}

	time.Sleep(time.Second * 3)
}

func String(item interface{}) string {
	b, err := json.MarshalIndent(item, "", "    ")
	if err != nil {
		return ""
	}

	return "\n" + string(b) + "\n"
}
