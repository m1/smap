package main

import (
	"encoding/json"
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/m1/smap"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"time"
)

var (
	rootCmd         *cobra.Command
	verbose         bool
	jsonPrint       bool
	maxWorkers      int
	ignoreRobotsTxt bool
	userAgent       string
)

func main() {
	rootCmd = &cobra.Command{
		Run:   smap,
		Use:   "smap [url]",
		Short: "smap is a site-mapping engine.",
		Long:  "smap is a site-mapping engine written in Go.",
		Args:  cobra.MinimumNArgs(1),
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose printing")
	rootCmd.PersistentFlags().BoolVar(&jsonPrint, "json", false, "json output")
	rootCmd.PersistentFlags().IntVarP(&maxWorkers, "workers", "w", 50, "How many workers to use")
	rootCmd.PersistentFlags().BoolVar(&ignoreRobotsTxt, "robots", false, "Ignores robots.txt")
	rootCmd.PersistentFlags().StringVarP(&userAgent, "user-agent", "u", "", "User agent to use for the crawler")

	rootCmd.Execute()
}

func smap(_ *cobra.Command, args []string) {
	u, err := url.Parse(args[0])
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if u.Path != "" && u.Path != "/" {
		fmt.Println("needs to be base url")
		os.Exit(1)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		fmt.Println("needs to http or https")
		os.Exit(1)
	}

	c, err := client.New(&client.Config{
		MaxWorkers:      maxWorkers,
		IgnoreRobotsTxt: ignoreRobotsTxt,
		UserAgent:       userAgent,
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var spin *spinner.Spinner
	if verbose {
		spin = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		spin.Suffix = fmt.Sprintf(" Crawling %s", u.String())
		spin.Writer = os.Stderr
		spin.Start()
	}

	siteMap, err := c.Crawl(u)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if verbose && spin != nil {
		spin.Stop()
	}

	if !jsonPrint {
		for _, v := range siteMap {
			fmt.Println(fmt.Sprintf("Path: %s", v.URL.Path))
			fmt.Println(fmt.Sprintf("Redirect: %t", v.IsRedirect))

			redirectUrl := "null"
			if v.IsRedirect {
				redirectUrl = v.RedirectsTo.Path
			}

			fmt.Println(fmt.Sprintf("Redirect Url: %s", redirectUrl))

			fmt.Println("Links:")
			for _, l := range v.Links {
				fmt.Println(fmt.Sprintf("\t%s", l.URL.Path))
			}

			fmt.Println("Linked From:")
			for _, l := range v.LinkedFrom {
				fmt.Println(fmt.Sprintf("\t%s", l.URL.Path))
			}

			fmt.Println()
		}
	}

	js, err := json.Marshal(&siteMap)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(string(js))
	return
}
