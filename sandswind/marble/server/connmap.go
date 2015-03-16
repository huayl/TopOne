package server

import (
	"sync"
)

type ConnMap struct {
	lock *sync.RWMutex
	mp   map[interface{}]interface{}
}

// new map.
func NewConnMap() *ConnMap {
	return &ConnMap{
		lock: new(sync.RWMutex),
		mp:   make(map[interface{}]interface{}),
	}
}

// return  value.
func (cm *ConnMap) Get(k interface{}) interface{} {
	cm.lock.RLock()
	defer cm.lock.RUnlock()
	if val, ok := cm.mp[k]; ok {
		return val
	}
	return nil
}

// set item.
func (cm *ConnMap) Set(k interface{}, v interface{}) bool {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	if val, ok := cm.mp[k]; !ok {
		cm.mp[k] = v
	} else if val != v {
		cm.mp[k] = v
	} else {
		return false
	}
	return true
}

// return true if k is existed.
func (cm *ConnMap) Check(k interface{}) bool {
	cm.lock.RLock()
	defer cm.lock.RUnlock()
	if _, ok := cm.mp[k]; !ok {
		return false
	}
	return true
}

// delete item.
func (cm *ConnMap) Delete(k interface{}) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	delete(cm.mp, k)
}

// return all items
func (cm *ConnMap) Items() map[interface{}]interface{} {
	cm.lock.RLock()
	defer cm.lock.RUnlock()
	return cm.mp
}

// return size
func (cm *ConnMap) Size() int {
	cm.lock.RLock()
	defer cm.lock.RUnlock()
	return len(cm.mp)
}
