package main

import "sync"

type UserProvider interface {
	AuthenticateNewUser() int32
}

type Users struct {
	availableUserID int32
	lock            sync.RWMutex
}

func NewUsers() *Users {
	return &Users{
		availableUserID: 1,
	}
}

func (u *Users) AuthenticateNewUser() int32 {
	u.lock.Lock()
	id := u.availableUserID
	u.availableUserID++
	u.lock.Unlock()
	return id
}
