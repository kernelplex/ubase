package extension

import "sync"

// MenuItem represents a link in the admin sidebar navigation.
type MenuItem struct {
    Label string
    Href  string
}

var (
    mu        sync.RWMutex
    menuItems []MenuItem
)

// RegisterMenuItem adds a new item to the admin sidebar menu.
func RegisterMenuItem(item MenuItem) {
    mu.Lock()
    menuItems = append(menuItems, item)
    mu.Unlock()
}

// ListMenuItems returns a copy of all registered menu items.
func ListMenuItems() []MenuItem {
    mu.RLock()
    defer mu.RUnlock()
    out := make([]MenuItem, len(menuItems))
    copy(out, menuItems)
    return out
}

