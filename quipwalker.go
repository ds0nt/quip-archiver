package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mduvall/go-quip"
)

var accessToken = os.Getenv("QUIP_ACCESS_TOKEN")
var client *quip.Client

func walk(id string, threads, folders chan map[string]string) {
	folder := client.GetFolder(&quip.GetFolderParams{Id: id})

	for i := range folder.Children {
		if folder.Children[i]["thread_id"] != "" {
			threads <- folder.Children[i]
		} else {
			folders <- folder.Children[i]
			walk(folder.Children[i]["folder_id"], threads, folders)
		}
	}
}

func main() {
	client = quip.NewClient(accessToken)
	threadData := make(chan *quip.Thread)
	threads := make(chan map[string]string)
	folders := make(chan map[string]string)

	threadDataProcessed := make(chan bool)
	threadsProcessed := make(chan bool)
	foldersProcessed := make(chan bool)

	// Process Each SubFolder
	go func() {
		for {
			fldr, more := <-folders
			if more {
				fmt.Println("received folder", fldr)
			} else {
				fmt.Println("received all folders")
				foldersProcessed <- true
				return
			}
		}
	}()

	// Process Thread Ids
	go func() {
		for {
			thrd, more := <-threads
			if more {
				threadData <- client.GetThread(thrd["thread_id"])
			} else {
				fmt.Println("received all threads")
				threadsProcessed <- true
				return
			}
		}
	}()

	// Process Thread Data
	go func() {
		for {
			thread, more := <-threadData
			if more {
				htmlData := []byte(thread.Html)
				err := ioutil.WriteFile("threads/"+thread.Thread["id"]+"-"+thread.Thread["title"], htmlData, 0644)
				if err != nil {
					fmt.Errorf("error writing file %v\n", err)
				}
				fmt.Println("wrote file " + thread.Thread["link"])
			} else {
				fmt.Println("received all threads")
				threadDataProcessed <- true
				return
			}
		}
	}()
	walk("dWPAOAJR9ec", threads, folders)
	close(threads)
	close(folders)

	<-foldersProcessed
	<-threadsProcessed
	<-threadDataProcessed
}
