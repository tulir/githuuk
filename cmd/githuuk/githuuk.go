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
			fmt.Printf("%s pushed %d commits to %s branch %s\n", evt.Sender.Login, len(evt.Commits), evt.Repository.FullName, evt.Ref.Name())
		case *githuuk.PullRequestEvent:
			fmt.Printf("%s %s a pull request with %d changes in %s\n", evt.Sender.Login, evt.Action, evt.NumberOfChanges, evt.Repository.FullName)
		case *githuuk.PingEvent:
			fmt.Printf("Ping received for %s hook %s\n", evt.Repository.FullName, evt.Hook.Name)
		case *githuuk.CreateEvent:
			fmt.Printf("%s created %s %s in %s\n", evt.Sender.Login, evt.RefType, evt.Ref.Name(), evt.Repository.FullName)
		case *githuuk.DeleteEvent:
			fmt.Printf("%s deleted %s %s in %s\n", evt.Sender.Login, evt.RefType, evt.Ref.Name(), evt.Repository.FullName)
		default:
			data, _ := json.MarshalIndent(rawEvent, "", "  ")
			fmt.Printf("Unknown event type: %s. Data: %s\n", rawEvent.GetType(), string(data))
		}
	}
}
