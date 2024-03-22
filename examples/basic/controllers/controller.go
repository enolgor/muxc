package controllers

import (
	"errors"
	"slices"
	"sync"
)

type Controller interface {
	CreatePet(pet *Pet) (*Pet, error)
	ReadPet(id int64) (*Pet, error)
	UpdatePet(pet *Pet) error
	DeletePet(id int64) error
	ListPets() ([]*Pet, error)
}

type controller struct {
	pets    map[int64]*Pet
	lock    sync.Mutex
	counter int64
}

type Pet struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Breed string `json:"breed"`
}

var ErrPetNotFound error = errors.New("pet not found")

func NewController() Controller {
	return &controller{
		pets:    make(map[int64]*Pet),
		lock:    sync.Mutex{},
		counter: 1,
	}
}

func (ctrl *controller) ListPets() ([]*Pet, error) {
	ctrl.lock.Lock()
	pets := make([]*Pet, len(ctrl.pets))
	i := 0
	for _, v := range ctrl.pets {
		pets[i] = v
		i++
	}
	slices.SortFunc(pets, func(a *Pet, b *Pet) int {
		return int(a.ID - b.ID)
	})
	ctrl.lock.Unlock()
	return pets, nil
}

func (ctrl *controller) CreatePet(pet *Pet) (*Pet, error) {
	ctrl.lock.Lock()
	pet.ID = ctrl.counter
	ctrl.counter += 1
	ctrl.pets[pet.ID] = pet
	ctrl.lock.Unlock()
	return pet, nil
}

func (ctrl *controller) ReadPet(id int64) (*Pet, error) {
	pet, ok := ctrl.pets[id]
	if !ok {
		return nil, ErrPetNotFound
	}
	return pet, nil
}

func (ctrl *controller) UpdatePet(pet *Pet) error {
	_, ok := ctrl.pets[pet.ID]
	if !ok {
		return ErrPetNotFound
	}
	ctrl.lock.Lock()
	ctrl.pets[pet.ID] = pet
	ctrl.lock.Unlock()
	return nil
}

func (ctrl *controller) DeletePet(id int64) error {
	_, ok := ctrl.pets[id]
	if !ok {
		return ErrPetNotFound
	}
	ctrl.lock.Lock()
	delete(ctrl.pets, id)
	ctrl.lock.Unlock()
	return nil
}
