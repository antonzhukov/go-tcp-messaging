package main

import "sync"

type UserProvider interface {
	AuthenticateNewUser() int32
}

type Users struct {
	users           map[int32]struct{}
	availableUserID int32
	lock            sync.RWMutex
}

func NewUsers() *Users {
	return &Users{
		users: make(map[int32]struct{}),
	}
}

func (u *Users) AuthenticateNewUser() int32 {
	u.lock.Lock()
	id := u.availableUserID
	u.users[u.availableUserID] = struct{}{}
	u.availableUserID++
	u.lock.Unlock()
	return id
}
