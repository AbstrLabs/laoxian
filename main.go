// Copyright 2022 The golang.design Initiative Authors.
// All rights reserved. Use of this source code is governed
// by a MIT license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.design/x/clipboard"
	"golang.design/x/hotkey"

	zmq "github.com/pebbe/zmq4"
)

func main() {
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}

	socket, _ := zmq.NewSocket(zmq.REQ)
	defer socket.Close()

	fmt.Println("Connecting to hello world server...")
	socket.Connect("tcp://localhost:5555")

	sendToGPT := func(str string) string {
		socket.Send(string(str), 0)
		// Wait for reply:
		reply, _ := socket.Recv(0)

		fmt.Println("Received ", reply)
		return reply
	}

	w, home_label := ui(sendToGPT)

	go func() {
		// Register a desired hotkey.
		hk := hotkey.New([]hotkey.Modifier{hotkey.ModCmd}, hotkey.KeyL)
		if err := hk.Register(); err != nil {
			panic("hotkey registration failed")
		}
		// Start listen hotkey event whenever it is ready.
		for range hk.Keydown() {
			data := clipboard.Read(clipboard.FmtText)
			if data != nil {
				content := string(data)
				home_label.SetText(content)

				w.Show()
			}
		}
	}()

	w.ShowAndRun()
}

func ui(sendToGPT func(str string) string) (fyne.Window, *widget.Label) {
	myApp := app.New()
	myWindow := myApp.NewWindow("laoxian")

	//reply features
	label1 := widget.NewLabel("Keyword")
	value1 := widget.NewEntry()
	label2 := widget.NewLabel("Style")
	value2 := widget.NewSelect([]string{"Professional", "Casual"}, func(value string) {
		log.Println("Select style to", value)
	})
	label3 := widget.NewLabel("Context")
	value3 := widget.NewSelect([]string{"Email", "Slack"}, func(value string) {
		log.Println("Select context to", value)
	})
	home_label := widget.NewLabel("Home tab")

	button := widget.NewButton("Submit", func() {
		msg := map[string]interface{}{
			"template": "reply",
			"params": map[string]interface{}{
				"keyword": value1.Text,
				"style":   value2.Selected,
				"content": home_label.Text,
				"context": value3.Selected,
			},
		}
		str, _ := json.Marshal(msg)
		respond := sendToGPT(string(str))
		fmt.Println("done")
		clipboard.Write(clipboard.FmtText, []byte(respond))
	})

	//grid
	grid_reply := container.New(layout.NewGridLayout(2), label1, value1, label2, value2, label3, value3, button)
	grid_rewrite := container.New(layout.NewGridLayout(2), label1, value1, label2, value2, label3, value3, button)

	// //tabs
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("", theme.HomeIcon(), home_label),
		container.NewTabItem("Reply", grid_reply),
		container.NewTabItem("Rewrite", grid_rewrite),
	)

	// themes := container.NewGridWithColumns(2,
	// 	widget.NewButton("Dark", func() {
	// 		myApp.Settings().SetTheme(theme.DarkTheme())
	// 	}),
	// 	widget.NewButton("Light", func() {
	// 		myApp.Settings().SetTheme(theme.LightTheme())
	// 	}),
	// )

	tabs.SetTabLocation(container.TabLocationLeading)

	myWindow.SetContent(tabs)
	// myWindow.ShowAndRun()

	return myWindow, home_label
}
