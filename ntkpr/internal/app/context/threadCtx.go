package context

import (
	"sort"
	"strings"

	"github.com/haochend413/ntkpr/internal/models"
)

/* Threads */

type ThreadContext struct {
	Name    ContextPtr
	Threads []*models.Thread
	Order   ContextOrder
	Cursor  uint
}

type ThreadContextMgr struct {
	previousContext ContextPtr
	currentContext  ContextPtr
	Contexts        []*ThreadContext
}

// NewThreadContextMgr creates a new context manager with all contexts initialized
func NewThreadContextMgr() *ThreadContextMgr {
	return &ThreadContextMgr{
		previousContext: None,
		currentContext:  Default,
		Contexts: []*ThreadContext{
			{Name: Default, Threads: make([]*models.Thread, 0), Cursor: 0, Order: CreateAt}, // Default context
			{Name: Recent, Threads: make([]*models.Thread, 0), Cursor: 0, Order: UpdateAt},  // Recent context
			{Name: Search, Threads: make([]*models.Thread, 0), Cursor: 0, Order: CreateAt},  // Search context
		},
	}
}

func (cm *ThreadContextMgr) SwitchContext(c ContextPtr) {
	if c != cm.currentContext {
		cm.previousContext = cm.currentContext
		cm.currentContext = c
	}
}

func (cm *ThreadContextMgr) GetCurrentThreads() []*models.Thread {
	return cm.Contexts[cm.currentContext].Threads
}

func (cm *ThreadContextMgr) GetCurrentContext() ContextPtr {
	return cm.currentContext
}

func (cm *ThreadContextMgr) GetPreviousContext() ContextPtr {
	return cm.previousContext
}

func (cm *ThreadContextMgr) RefreshDefaultContext(threads []*models.Thread) {
	cm.Contexts[Default].Threads = threads
}

func (cm *ThreadContextMgr) RefreshRecentContext() {
	threads := cm.Contexts[Default].Threads
	threadsCopy := make([]*models.Thread, len(threads))
	copy(threadsCopy, threads)
	sort.Slice(threadsCopy, func(i, j int) bool {
		return threadsCopy[i].UpdatedAt.After(threadsCopy[j].UpdatedAt)
	})
	recentCount := 20
	if len(threadsCopy) < recentCount {
		recentCount = len(threadsCopy)
	}
	cm.Contexts[Recent].Threads = threadsCopy[:recentCount]
}

func (cm *ThreadContextMgr) RefreshSearchContext(q string) {
	c := cm.currentContext
	if cm.currentContext == Search {
		c = cm.previousContext
	}
	threads := cm.Contexts[c].Threads

	if q == "" {
		cm.Contexts[Search].Threads = threads
		return
	}

	query := strings.ToLower(q)
	filteredThreads := make([]*models.Thread, 0)
	for _, thread := range threads {
		if strings.Contains(strings.ToLower(thread.Name), query) {
			filteredThreads = append(filteredThreads, thread)
			continue
		}
	}
	sort.Slice(filteredThreads, func(i, j int) bool {
		return filteredThreads[i].CreatedAt.Before(filteredThreads[j].CreatedAt)
	})
	cm.Contexts[Search].Threads = filteredThreads
}

func (cm *ThreadContextMgr) GetCurrentCursor() uint {
	return cm.Contexts[cm.currentContext].Cursor
}

func (cm *ThreadContextMgr) SetCurrentCursor(cursor uint) {
	cm.Contexts[cm.currentContext].Cursor = cursor
}

func (cm *ThreadContextMgr) GetCurrentThread() *models.Thread {
	threads := cm.GetCurrentThreads()
	cursor := cm.GetCurrentCursor()
	if len(threads) == 0 || int(cursor) >= len(threads) {
		return nil
	}
	return threads[cursor]
}

func (cm *ThreadContextMgr) GetThreadCount() int {
	return len(cm.GetCurrentThreads())
}

func (cm *ThreadContextMgr) AddThreadToDefault(thread *models.Thread) {
	threads := cm.Contexts[Default].Threads
	insertPos := len(threads)
	for i, t := range threads {
		if t.CreatedAt.After(thread.CreatedAt) {
			insertPos = i
			break
		}
	}
	if insertPos == len(threads) {
		cm.Contexts[Default].Threads = append(threads, thread)
	} else {
		cm.Contexts[Default].Threads = append(threads[:insertPos], append([]*models.Thread{thread}, threads[insertPos:]...)...)
	}
}

func (cm *ThreadContextMgr) RemoveThreadFromDefault(threadID uint) {
	threads := cm.Contexts[Default].Threads
	for i, thread := range threads {
		if thread.ID == threadID {
			cm.Contexts[Default].Threads = append(threads[:i], threads[i+1:]...)
			break
		}
	}
}

func (cm *ThreadContextMgr) SortCurrentContext() {
	threads := cm.Contexts[cm.currentContext].Threads
	order := cm.Contexts[cm.currentContext].Order

	switch order {
	case CreateAt:
		sort.Slice(threads, func(i, j int) bool {
			return threads[i].CreatedAt.Before(threads[j].CreatedAt)
		})
	case UpdateAt:
		sort.Slice(threads, func(i, j int) bool {
			return threads[i].UpdatedAt.After(threads[j].UpdatedAt)
		})
	default:
		sort.Slice(threads, func(i, j int) bool {
			return threads[i].CreatedAt.Before(threads[j].CreatedAt)
		})
	}
}

// UpdateContext switches context, saves current cursor, sorts, and returns new cursor
func (cm *ThreadContextMgr) UpdateContext(newContext ContextPtr, currentCursor uint) uint {
	c0 := cm.currentContext
	cm.Contexts[c0].Cursor = currentCursor
	cm.SwitchContext(newContext)
	cm.SortCurrentContext()
	return cm.Contexts[newContext].Cursor
}

// Search performs search and switches to search context
func (cm *ThreadContextMgr) Search(query string) {
	cm.RefreshSearchContext(query)
	cm.SwitchContext(Search)
}

// SelectItem sets the cursor and returns the item at that position
func (cm *ThreadContextMgr) SelectItem(cursor int) *models.Thread {
	threads := cm.GetCurrentThreads()
	if len(threads) == 0 || cursor >= len(threads) || cursor < 0 {
		return nil
	}
	cm.SetCurrentCursor(uint(cursor))
	return threads[cursor]
}

func (cm *ThreadContextMgr) GetCursors() map[ContextPtr]uint {
	return map[ContextPtr]uint{
		Default: cm.Contexts[Default].Cursor,
		Recent:  cm.Contexts[Recent].Cursor,
		Search:  cm.Contexts[Search].Cursor,
	}
}

func (cm *ThreadContextMgr) SetCursors(m map[ContextPtr]uint) {
	cm.Contexts[Default].Cursor = m[Default]
	cm.Contexts[Recent].Cursor = m[Recent]
	cm.Contexts[Search].Cursor = m[Search]
}
