package files

import (
	"os/user"
	"strconv"
	"sync"
)

type ownerCache struct {
	mu     sync.Mutex
	users  map[uint32]string
	groups map[uint32]string
}

func newOwnerCache() *ownerCache {
	return &ownerCache{users: map[uint32]string{}, groups: map[uint32]string{}}
}

func (c *ownerCache) names(uid, gid uint32) (string, string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lookup(c.users, uid, userName), c.lookup(c.groups, gid, groupName)
}

func (c *ownerCache) lookup(cache map[uint32]string, id uint32, resolve func(string) string) string {
	if name, ok := cache[id]; ok {
		return name
	}
	name := resolve(strconv.FormatUint(uint64(id), 10))
	cache[id] = name
	return name
}

func userName(id string) string {
	u, err := user.LookupId(id)
	if err != nil {
		return id
	}
	return u.Username
}

func groupName(id string) string {
	g, err := user.LookupGroupId(id)
	if err != nil {
		return id
	}
	return g.Name
}
