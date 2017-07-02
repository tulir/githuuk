// githuuk - A GitHub webhook receiver written in Go.
// Copyright (C) 2017 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"encoding/json"
	"fmt"

	"maunium.net/go/githuuk"
	flag "maunium.net/go/mauflag"
)

var port = flag.MakeFull("p", "port", "The port on which to listen for GitHub webhooks.", "80").Uint16()
var secret = flag.MakeFull("s", "secret", "The GitHub webhook secret.", "").String()
var wantHelp, _ = flag.MakeHelpFlag()

func main() {
	flag.SetHelpTitles("githuuk - A very simple command-line utility to test webhooks.", "githuuk [-h] [-p port] [-s secret]")
	err := flag.Parse()
	if err != nil {
		fmt.Println(err)
		flag.PrintHelp()
		return
	} else if *wantHelp {
		flag.PrintHelp()
		return
	}

	server := githuuk.NewServer()
	server.Port = *port
	server.Secret = *secret
	server.AsyncListenAndServe()

	for rawEvent := range server.Events {
		switch evt := rawEvent.(type) {
		case *githuuk.PushEvent:
			fmt.Println(evt.Repository.Owner.Login, evt.Repository.Name, evt.Ref.Name(), evt.HeadCommit.ID)
		case *githuuk.PullRequestEvent:
			fmt.Println(evt.Repository.Owner.Login, evt.Repository.Name, evt.Action, evt.NumberOfChanges)
		case *githuuk.PingEvent:
			fmt.Println(evt.Repository.Owner.Login, evt.Repository.Name, evt.Hook.Name, evt.Hook.ID)
		default:
			fmt.Println("Unknown event type", rawEvent.GetType())
			data, _ := json.MarshalIndent(rawEvent, "", "  ")
			fmt.Println(string(data))
		}
	}
}
