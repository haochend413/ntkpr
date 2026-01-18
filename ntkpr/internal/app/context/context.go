package context

import "github.com/haochend413/ntkpr/internal/app/data"

// Context is a wrapper that helps ordering and filtering the raw lists of notes, and store lists from the map.
// In the future, maybe we can add a config menu that allows for multiple orders.
// Context also defines current list, which in combination with cursor, defines current item pointers.

// This should be improved. How ? What is it that I want ?
// How do we define recent activity ? Is "recent note" still meaningful?
// Should this happen globally, across all ? or separate into layers ?

// Think.

// Well, we will have a "changelog" window to demonstrate logs.
// Then how should context work ? search is definitely useful. We probably need recent ? Or should search be in a different regime.

// OK. Enable contexts for all thread, branch and notes. At least for now.

// NONONO. forget about it. This is silly. let's only do context for notes for now, with minimum change.
// Wait. But in that case, we still need a wrapper to render lists from maps. This layer is still required, it's just that we remove the unecessary functionalities.

// This can Actually be done as a general type. However, I am expecting something different for branch / thread / notes, thus let's do it this way.

// Good. I think a thin app layer should work. However, this means that I will have to expand the API layers of my context pkg.
// context should handle all data related stuff : inner logic of sort and order, data, and cursor.

// Right now there is a problem : different threads are disconnected with its branches and notes. Switching threads will not trigger switch of branch and notes.

// Context requires further refining. Cursors should also be a part of context state ? When refreshing.

// Right now the problem is that this context is designed to only have one list. Switching from default to default is never meaningful.
// However, right now we might have several lists co-existing in our app.
// Let's keep that simple and just pass in cursor as well. In this case we enable switching to a different set of notes. (?)
// Maybe we need another layer of wrappers called data manager, and then context is just wrapping around data manageer.
// What should each of them handle ?

// DataMgr should handle the switching logic between threads, branches and notes. It keeps record of all threads, and exposing current threads, branches and notes.
// ContextMgr will Demonstrate based on that.
// DataMgr will also be responsible for syncing with database, instead of being handled directly by App.

// Ok let's save it first. First have the functionalities, and then based on needs, decide whether we need this module.
// Right now, just use the data somthting.

type ContextPtr int

const (
	None    ContextPtr = -1
	Default ContextPtr = 0
	Recent  ContextPtr = 1
	Search  ContextPtr = 2
)

// This is replicative, but might be useful in the future.
type ContextOrder int

const (
	CreateAt ContextOrder = 0 // default , time order
	UpdateAt ContextOrder = 1 // recent, most recently updated
)

// Everything should be fetched from ContextMgr.
type ContextMgr struct {
	DataMgr          *data.DataMgr
	NoteContextMgr   *NoteContextMgr
	BranchContextMgr *BranchContextMgr
	ThreadContextMgr *ThreadContextMgr
}

// This function needs to be implemented when we take everything into account.
func NewContextMgr() {}

// RefreshThreadsContext should not take very long ? Is it really useful to separate it into many pieces ?
// Wait, there is the cursor problem...Yes
func (cm *ContextMgr) RefreshThreadsContext() {
	// load in stuff from DataMgr.
	threads := cm.DataMgr.GetThreads()
	branches := cm.DataMgr.GetActiveBranchList()
	notes := cm.DataMgr.GetActiveNoteList()

	//refresh contexts
	cm.NoteContextMgr.RefreshDefaultContext(notes)
	cm.BranchContextMgr.RefreshDefaultContext(branches)
	cm.ThreadContextMgr.RefreshDefaultContext(threads)
}

func (cm *ContextMgr) RefreshBranchesContext() {
	branches := cm.DataMgr.GetActiveBranchList()
	notes := cm.DataMgr.GetActiveNoteList()
	cm.NoteContextMgr.RefreshDefaultContext(notes)
	cm.BranchContextMgr.RefreshDefaultContext(branches)
}

func (cm *ContextMgr) RefreshNotesContext() {
	notes := cm.DataMgr.GetActiveNoteList()
	cm.NoteContextMgr.RefreshDefaultContext(notes)
}
