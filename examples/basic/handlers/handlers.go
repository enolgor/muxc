package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/enolgor/muxc/examples/basic/controllers"
)

func ListPets(ctrl controllers.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		pets, err := ctrl.ListPets()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		enc := json.NewEncoder(w)
		enc.Encode(pets)
	}
}

func ReadPet(ctrl controllers.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		id, err := strconv.ParseInt(req.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pet, err := ctrl.ReadPet(id)
		if errors.Is(err, controllers.ErrPetNotFound) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		enc := json.NewEncoder(w)
		enc.Encode(pet)
	}
}

func CreatePet(ctrl controllers.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		dec := json.NewDecoder(req.Body)
		defer req.Body.Close()
		pet := &controllers.Pet{}
		var err error
		if err = dec.Decode(pet); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if pet, err = ctrl.CreatePet(pet); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		enc := json.NewEncoder(w)
		enc.Encode(pet)
	}
}

func UpdatePet(ctrl controllers.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		dec := json.NewDecoder(req.Body)
		defer req.Body.Close()
		pet := &controllers.Pet{}
		var err error
		if err = dec.Decode(pet); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err = ctrl.UpdatePet(pet); err != nil {
			if errors.Is(err, controllers.ErrPetNotFound) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func DeletePet(ctrl controllers.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		id, err := strconv.ParseInt(req.PathValue("id"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = ctrl.DeletePet(id)
		if errors.Is(err, controllers.ErrPetNotFound) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
