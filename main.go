// Copyright 2022 The golang.design Initiative Authors.
// All rights reserved. Use of this source code is governed
// by a MIT license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
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

	w := app.New().NewWindow("golang.design/x/hotkey")

	label := widget.NewLabel("Hello golang.design!")
	button := widget.NewButton("Hi!", func() { label.SetText("Welcome :)") })
	w.SetContent(container.NewVBox(label, button))

	socket, _ := zmq.NewSocket(zmq.REQ)
	defer socket.Close()

	fmt.Println("Connecting to hello world server...")
	socket.Connect("tcp://localhost:5555")

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
				label.SetText(content)

				msg := map[string]interface{}{"template": "reply", "params": map[string]interface{}{"message_type": "email", "message_content": content, "key_idea": "i disagree"}}
				str, _ := json.Marshal(msg)
				socket.Send(string(str), 0)
				// Wait for reply:
				reply, _ := socket.Recv(0)

				fmt.Println("Received ", reply)
				label.SetText(reply)

				w.Show()
			}
		}
	}()

	w.ShowAndRun()
}
