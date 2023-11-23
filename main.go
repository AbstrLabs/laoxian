// Copyright 2022 The golang.design Initiative Authors.
// All rights reserved. Use of this source code is governed
// by a MIT license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.design/x/clipboard"
	"golang.design/x/hotkey"

	zmq "github.com/go-zeromq/zmq4"
)

func main() {
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	socket := zmq.NewReq(ctx, zmq.WithDialerRetry(time.Second))
	defer socket.Close()

	fmt.Printf("Connecting to hello world server...")
	if err := socket.Dial("tcp://localhost:5555"); err != nil {
		fmt.Errorf("dialing: %w", err)
	}

	sendToGPT := func(str string) string {
		// Send hello.
		m := zmq.NewMsgString(str)
		fmt.Println("sending ", m)
		if err := socket.Send(m); err != nil {
			return "{\"error\": \"client send error\"}"
		}

		// Wait for reply.
		r, err := socket.Recv()
		if err != nil {
			return "{\"error\": \"client receive error\"}"
		}
		ret := string(r.Bytes())
		fmt.Println("received ", ret)

		return ret
	}

	w, before_content := ui(sendToGPT)

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
				before_content.SetText(content)

				w.Show()
			}
		}
	}()

	w.ShowAndRun()
}

func ui(sendToGPT func(str string) string) (fyne.Window, *widget.Label) {
	myApp := app.New()
	myWindow := myApp.NewWindow("laoxian")
	myWindow.Resize(fyne.NewSize(400, 300))

	//text
	before_label := widget.NewLabel("Before")
	before_content := widget.NewLabel("...")
	before_content.Wrapping = fyne.TextWrapWord

	after_label := widget.NewLabel("After")
	after_content := widget.NewLabel("...")
	after_content.Wrapping = fyne.TextWrapWord

	//feature list
	style_list := []string{"Professional", "Casual", "Formal", "Friendly", "Diplomatic"}
	context_list := []string{"Email", "Slack", "Discord", "Telegram", "Facebook", "Instagram"}
	//gpt features
	keyword_label := widget.NewLabel("Keyword")
	keyword_value := widget.NewEntry()
	style_label := widget.NewLabel("Style")
	style_value := widget.NewSelect(style_list, func(value string) {
		log.Println("Select style to", value)
	})
	context_label := widget.NewLabel("Context")
	context_value := widget.NewSelect(context_list, func(value string) {
		log.Println("Select context to", value)
	})

	//button
	reply_button := widget.NewButton("Submit", func() {
		msg := map[string]interface{}{
			"template": "reply",
			"params": map[string]interface{}{
				"keyword": keyword_value.Text,
				"style":   style_value.Selected,
				"content": before_content.Text,
				"context": context_value.Selected,
			},
		}
		str, _ := json.Marshal(msg)

		//TODO: make it go func
		response := sendToGPT(string(str))
		fmt.Println("done")

		var dat map[string]string
		json.Unmarshal([]byte(response), &dat)

		result := dat["completion"]

		after_content.SetText(result)

		clipboard.Write(clipboard.FmtText, []byte(result))
		myWindow.Show()
	})

	rewrite_button := widget.NewButton("Submit", func() {
		msg := map[string]interface{}{
			"template": "rewrite",
			"params": map[string]interface{}{
				"keyword": keyword_value.Text,
				"style":   style_value.Selected,
				"content": before_content.Text,
				"context": context_value.Selected,
			},
		}
		str, _ := json.Marshal(msg)
		response := sendToGPT(string(str))
		fmt.Println("done")

		var dat map[string]string
		json.Unmarshal([]byte(response), &dat)

		result := dat["completion"]

		after_content.SetText(result)

		clipboard.Write(clipboard.FmtText, []byte(result))
		myWindow.Show()
	})

	//grid
	grid_reply := container.New(layout.NewVBoxLayout(), keyword_label, keyword_value, style_label, style_value, context_label, context_value, reply_button)
	grid_rewrite := container.New(layout.NewVBoxLayout(), keyword_label, keyword_value, style_label, style_value, context_label, context_value, rewrite_button)
	grid_content := container.New(layout.NewVBoxLayout(), before_label, before_content, after_label, after_content)

	//mode
	// dark_mode := widget.NewButton("Dark", func() {
	// 	myApp.Settings().SetTheme(theme.DarkTheme())
	// })

	// light_mode := widget.NewButton("Light", func() {
	// 	myApp.Settings().SetTheme(theme.LightTheme())
	// })

	// //tabs
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("", theme.HomeIcon(), grid_content),
		container.NewTabItem("Reply", grid_reply),
		container.NewTabItem("Rewrite", grid_rewrite),
	)

	tabs.SetTabLocation(container.TabLocationLeading)

	myWindow.SetContent(container.NewMax(tabs))
	// myWindow.ShowAndRun()

	return myWindow, before_content
}
