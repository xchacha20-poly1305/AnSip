// Package freeauth provides auth API.
package freeauth

import (
	"fmt"
	"sync"
)

// FreeAuth is a custom auth user map.
type FreeAuth struct {
	userMap sync.Map
}

func NewFreeAuth() *FreeAuth {
	return &FreeAuth{}
}

// Verify verifies user information. If not found this user, it will create.
func (f *FreeAuth) Verify(username, password string) (success, isNew bool) {
	storedPassword, loaded := f.userMap.Load(username)
	if loaded {
		return storedPassword.(string) == password, false
	}
	f.userMap.Store(username, password)
	return true, true
}

// Delete deletes user based username.
func (f *FreeAuth) Delete(username, password string) error {
	storedPassword, loaded := f.userMap.Load(username)
	if !loaded {
		return fmt.Errorf("freeauth: not found user: %s", username)
	}
	if storedPassword.(string) != password {
		return fmt.Errorf("freeauth: invalid password: %s", password)
	}
	f.userMap.Delete(username)
	return nil
}
