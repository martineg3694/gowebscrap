package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
)

var (
	tag     string
	output  string
	urlFile string
)

var scrapeCmd = &cobra.Command{
	Use:   "scrape [url]",
	Short: "Scrape content from one or multiple websites",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var urls []string

		if urlFile != "" {
			file, err := os.Open(urlFile)
			if err != nil {
				fmt.Printf("Error opening file: %v\n", err)
				return
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					urls = append(urls, line)
				}
			}
		} else if len(args) == 1 {
			urls = append(urls, args[0])
		} else {
			fmt.Println("Please provide a URL or use --file flag.")
			return
		}

		results := make(map[string][]string)

		for _, url := range urls {
			fmt.Printf("Scraping: %s (tag: %s)\n", url, tag)

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				fmt.Printf("Request error: %v\n", err)
				continue
			}
			req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("Failed to GET %s: %v\n", url, err)
				continue
			}
			if err != nil {
				fmt.Printf("Failed to GET %s: %v\n", url, err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				fmt.Printf("Non-200 response from %s: %d\n", url, resp.StatusCode)
				continue
			}

			doc, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				fmt.Printf("Error parsing HTML from %s: %v\n", url, err)
				continue
			}

			var contents []string
			doc.Find(tag).Each(func(i int, s *goquery.Selection) {
				text := strings.TrimSpace(s.Text())
				if text != "" {
					contents = append(contents, text)
				}
			})

			results[url] = contents
		}

		// Output
		if output != "" {
			if strings.HasSuffix(output, ".json") {
				file, err := os.Create(output)
				if err != nil {
					fmt.Printf("Error creating file: %v\n", err)
					return
				}
				defer file.Close()

				encoder := json.NewEncoder(file)
				encoder.SetIndent("", "  ")
				err = encoder.Encode(results)
				if err != nil {
					fmt.Printf("Error writing JSON: %v\n", err)
				} else {
					fmt.Printf("Results saved to %s\n", output)
				}
			} else {
				file, err := os.Create(output)
				if err != nil {
					fmt.Printf("Error creating file: %v\n", err)
					return
				}
				defer file.Close()

				for url, items := range results {
					file.WriteString(fmt.Sprintf("URL: %s\n", url))
					for _, item := range items {
						file.WriteString(fmt.Sprintf(" - %s\n", item))
					}
					file.WriteString("\n")
				}

				fmt.Printf("Results saved to %s\n", output)
			}
		} else {
			// Print to console
			for url, items := range results {
				fmt.Printf("Results from %s:\n", url)
				for _, item := range items {
					fmt.Printf(" - %s\n", item)
				}
			}
		}
	},
}

func init() {
	scrapeCmd.Flags().StringVarP(&tag, "tag", "t", "h1", "HTML tag to scrape (e.g. h1, p, a)")
	scrapeCmd.Flags().StringVarP(&output, "output", "o", "", "Output file (.txt or .json)")
	scrapeCmd.Flags().StringVarP(&urlFile, "file", "f", "", "File containing list of URLs to scrape (1 per line)")
	rootCmd.AddCommand(scrapeCmd)
}
