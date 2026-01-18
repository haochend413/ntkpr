package ui

import (
	"github.com/haochend413/ntkpr/state"
)

// Distribute state in json on startup
func (m *Model) DistributeState(s *state.UIState) {
	// TODO: Implement state restoration for new three-table layout
	// This needs to be redesigned for the new hierarchy navigation
}

// Collect end state on termination
func (m Model) CollectState() *state.State {
	s := &state.State{}
	// TODO: Implement state collection for new three-table layout
	return s
}

func HelpText() string {
	help :=
		`ntkpr is a note / diary management tool built with Golang and Bubbletea framework. 

It allows you to manage your logs in a structured manner. The general idea comes from version control tools in software development. 
You can separate your tasks into different threads, and within one thread you can create multiple branches, each containing a list of notes. 
Each branch / thread contains one summary page that you can use to provide info for this specific branch / thread. 

I recommand creating several short notes instead of one long note. 

**********************

Some important keybindings: 

q : switch to upper table 
w : switch to lower table 
tab : switch between windows 
e : go to edit window 
n : create new item 
ctrl-s : save current edit 

For full usage guidance, please check github : https://github.com/haochend413/ntkpr	
	`
	return help
}
