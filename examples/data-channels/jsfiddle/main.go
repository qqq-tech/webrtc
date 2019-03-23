// +build js,wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/pions/webrtc"
	"github.com/pions/webrtc/examples/internal/signal"
)

func main() {
	// Configure and create a new PeerConnection.
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		handleError(err)
	}

	// Create DataChannel.
	sendChannel, err := pc.CreateDataChannel("foo", nil)
	if err != nil {
		handleError(err)
	}
	sendChannel.OnClose(func() {
		fmt.Println("sendChannel has closed")
	})
	sendChannel.OnOpen(func() {
		fmt.Println("sendChannel has opened")
	})
	sendChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		log(fmt.Sprintf("Message from DataChannel %s payload %s", sendChannel.Label(), string(msg.Data)))
	})

	// Add handlers for setting up the connection.
	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log(fmt.Sprint(state))
	})
	pc.OnICECandidate(func(candidate *string) {
		if candidate != nil {
			encodedDescr := signal.Encode(pc.LocalDescription())
			el := getElementByID("localSessionDescription")
			el.Set("value", encodedDescr)
		}
	})
	pc.OnNegotiationNeeded(func() {
		offer, err := pc.CreateOffer(nil)
		if err != nil {
			handleError(err)
		}
		pc.SetLocalDescription(offer)
	})

	// Set up global callbacks which will be triggered on button clicks.
	js.Global().Set("sendMessage", js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		go func() {
			el := getElementByID("message")
			message := el.Get("value").String()
			if message == "" {
				js.Global().Call("alert", "Message must not be empty")
				return
			}
			if err := sendChannel.SendText(message); err != nil {
				handleError(err)
			}
		}()
		return js.Undefined()
	}))
	js.Global().Set("startSession", js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		go func() {
			el := getElementByID("remoteSessionDescription")
			sd := el.Get("value").String()
			if sd == "" {
				js.Global().Call("alert", "Session Description must not be empty")
				return
			}

			descr := webrtc.SessionDescription{}
			signal.Decode(sd, &descr)
			if err := pc.SetRemoteDescription(descr); err != nil {
				handleError(err)
			}
		}()
		return js.Undefined()
	}))

	// Stay alive
	select {}
}

func log(msg string) {
	el := getElementByID("logs")
	el.Set("innerHTML", el.Get("innerHTML").String()+msg+"<br>")
}

func handleError(err error) {
	log("Unexpected error. Check console.")
	panic(err)
}

func getElementByID(id string) js.Value {
	return js.Global().Get("document").Call("getElementById", id)
}