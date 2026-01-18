package context

import (
	"sort"
	"strings"

	"github.com/haochend413/ntkpr/internal/models"
)

/* Branches*/

type BranchContext struct {
	Name     ContextPtr
	Branches []*models.Branch
	Order    ContextOrder
	Cursor   uint
}

type BranchContextMgr struct {
	previousContext ContextPtr
	currentContext  ContextPtr
	Contexts        []*BranchContext
}

// NewContextMgr creates a new context manager with all contexts initialized
// We might not need recent for branch, and we might put search into a different category. But now, let's keep things real.
func NewBranchContextMgr() *BranchContextMgr {
	return &BranchContextMgr{
		previousContext: None,
		currentContext:  Default,
		Contexts: []*BranchContext{
			{Name: Default, Branches: make([]*models.Branch, 0), Cursor: 0, Order: CreateAt}, // Default context
			{Name: Recent, Branches: make([]*models.Branch, 0), Cursor: 0, Order: UpdateAt},  // Recent context
			{Name: Search, Branches: make([]*models.Branch, 0), Cursor: 0, Order: CreateAt},  // Search context
		},
	}
}

func (cm *BranchContextMgr) SwitchContext(c ContextPtr) {
	// make sure they are different
	if c != cm.currentContext {
		cm.previousContext = cm.currentContext
		cm.currentContext = c
	}
}

func (cm *BranchContextMgr) GetCurrentBranches() []*models.Branch {
	return cm.Contexts[cm.currentContext].Branches
}

func (cm *BranchContextMgr) GetCurrentContext() ContextPtr {
	return cm.currentContext
}

func (cm *BranchContextMgr) GetPreviousContext() ContextPtr {
	return cm.previousContext
}

func (cm *BranchContextMgr) RefreshDefaultContext(bs []*models.Branch) {
	cm.Contexts[Default].Branches = bs
}

func (cm *BranchContextMgr) RefreshRecentContext() {
	bs := cm.Contexts[Default].Branches
	// Create a copy to avoid modifying the original slice
	bscp := make([]*models.Branch, len(bs))
	copy(bscp, bs)
	//sort by UpdatedAt (most recent first)
	sort.Slice(bscp, func(i, j int) bool {
		return bscp[i].UpdatedAt.After(bscp[j].UpdatedAt)
	})
	// this might need a bit tweeking
	recentCount := 20
	if len(bscp) < recentCount {
		recentCount = len(bscp)
	}
	cm.Contexts[Recent].Branches = bscp[:recentCount]
}

func (cm *BranchContextMgr) RefreshSearchContext(q string) {
	// first, get the current note list
	// I am not sure whether this is correct.
	c := cm.currentContext
	if cm.currentContext == Search {
		c = cm.previousContext
	}
	bs := cm.Contexts[c].Branches

	//loop through and search
	if q == "" {
		cm.Contexts[Search].Branches = bs
		return
	}

	query := strings.ToLower(q)
	filteredBranches := make([]*models.Branch, 0)
	for _, b := range bs {
		if strings.Contains(strings.ToLower(b.Name), query) {
			filteredBranches = append(filteredBranches, b)
			continue
		}
	}
	// Sort by CreatedAt to maintain chronological order
	sort.Slice(filteredBranches, func(i, j int) bool {
		return filteredBranches[i].CreatedAt.Before(filteredBranches[j].CreatedAt)
	})
	cm.Contexts[Search].Branches = filteredBranches
}

// GetCurrentCursor returns the cursor position in the current context
func (cm *BranchContextMgr) GetCurrentCursor() uint {
	return cm.Contexts[cm.currentContext].Cursor
}

// SetCurrentCursor sets the cursor position in the current context
func (cm *BranchContextMgr) SetCurrentCursor(cursor uint) {
	cm.Contexts[cm.currentContext].Cursor = cursor
}

// GetCurrentBranch returns the branch at the current cursor position, or nil if invalid
func (cm *BranchContextMgr) GetCurrentBranch() *models.Branch {
	bs := cm.GetCurrentBranches()
	cursor := cm.GetCurrentCursor()
	if len(bs) == 0 || int(cursor) >= len(bs) {
		return nil
	}
	return bs[cursor]
}

// GetBranchCount returns the number of branches in the current context
func (cm *BranchContextMgr) GetBranchCount() int {
	return len(cm.GetCurrentBranches())
}

// This overall class might needs restructuring.
func (cm *BranchContextMgr) AddBranchToDefault(b *models.Branch) {
	bs := cm.Contexts[Default].Branches
	// Find the correct position to insert the note to maintain chronological order
	insertPos := len(bs)
	for i, n := range bs {
		if n.CreatedAt.After(b.CreatedAt) {
			insertPos = i
			break
		}
	}
	// Insert at the correct position
	if insertPos == len(bs) {
		cm.Contexts[Default].Branches = append(bs, b)
	} else {
		cm.Contexts[Default].Branches = append(bs[:insertPos], append([]*models.Branch{b}, bs[insertPos:]...)...)
	}
}

// RemoveNoteFromDefault removes a note from default context by ID
func (cm *BranchContextMgr) RemoveBranchFromDefault(ID uint) {
	bs := cm.Contexts[Default].Branches
	for i, b := range bs {
		if b.ID == ID {
			cm.Contexts[Default].Branches = append(bs[:i], bs[i+1:]...)
			break
		}
	}
}

// SortCurrentContext sorts the current context notes by CreatedAt (chronological order)
func (cm *BranchContextMgr) SortCurrentContext() {
	bs := cm.Contexts[cm.currentContext].Branches
	order := cm.Contexts[cm.currentContext].Order

	// well this is in place ! careful!
	switch order {
	case CreateAt:
		sort.Slice(bs, func(i, j int) bool {
			return bs[i].CreatedAt.Before(bs[j].CreatedAt)
		})
	case UpdateAt:
		sort.Slice(bs, func(i, j int) bool {
			return bs[i].UpdatedAt.After(bs[j].UpdatedAt)
		})
	default:
		sort.Slice(bs, func(i, j int) bool {
			return bs[i].CreatedAt.Before(bs[j].CreatedAt)
		})
	}
}

// only for data storage purposes. Do not use in coding.
func (cm *BranchContextMgr) GetCursors() map[ContextPtr]uint {
	return map[ContextPtr]uint{
		Default: cm.Contexts[Default].Cursor,
		Recent:  cm.Contexts[Recent].Cursor,
		Search:  cm.Contexts[Search].Cursor,
	}
}

func (cm *BranchContextMgr) SetCursors(m map[ContextPtr]uint) {
	cm.Contexts[Default].Cursor = m[Default]
	cm.Contexts[Recent].Cursor = m[Recent]
	cm.Contexts[Search].Cursor = m[Search]
}

// UpdateContext switches context, saves current cursor, sorts, and returns new cursor
func (cm *BranchContextMgr) UpdateContext(newContext ContextPtr, currentCursor uint) uint {
	c0 := cm.currentContext
	cm.Contexts[c0].Cursor = currentCursor
	cm.SwitchContext(newContext)
	cm.SortCurrentContext()
	return cm.Contexts[newContext].Cursor
}

// Search performs search and switches to search context
func (cm *BranchContextMgr) Search(query string) {
	cm.RefreshSearchContext(query)
	cm.SwitchContext(Search)
}

// SelectItem sets the cursor and returns the item at that position
func (cm *BranchContextMgr) SelectItem(cursor int) *models.Branch {
	branches := cm.GetCurrentBranches()
	if len(branches) == 0 || cursor >= len(branches) || cursor < 0 {
		return nil
	}
	cm.SetCurrentCursor(uint(cursor))
	return branches[cursor]
}
