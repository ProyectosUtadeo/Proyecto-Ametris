package server

import (
	"backend-avanzada/api"
	"backend-avanzada/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func (s *Server) HandleTransmutations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetAllTransmutations(w, r)
	case http.MethodPost:
		// opcional si quisieras POST /transmutations sin id
		w.WriteHeader(http.StatusMethodNotAllowed)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) HandleTransmutationsWithId(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleStartTransmutation(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleGetAllTransmutations(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	items, err := s.TransmutationRepository.FindAll()
	if err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	result := make([]*api.TransmutationResponseDto, 0, len(items))
	for _, t := range items {
		result = append(result, t.ToResponseDto(true))
	}
	resp, err := json.Marshal(result)
	if err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
	s.logger.Info(http.StatusOK, r.URL.Path, start)
}

func (s *Server) handleStartTransmutation(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req api.TransmutationRequestDto
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, err)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	alchemistID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, err)
		return
	}

	// Verifica que el alquimista existe
	alch, err := s.AlchemistRepository.FindById(int(alchemistID))
	if err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	if alch == nil {
		s.HandleError(w, http.StatusNotFound, r.URL.Path, fmt.Errorf("alchemist with id %d not found", alchemistID))
		return
	}

	// Crea registro inmediatamente con estado IN_PROGRESS (visible en el frontend)
	t := &models.Transmutation{
		Description: req.Description,
		Status:      "IN_PROGRESS",
		AlchemistID: alch.ID,
		Alchemist:   alch,
	}
	saved, err := s.TransmutationRepository.Save(t)
	if err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}

	// Duración: reutilizamos los tiempos existentes de config (como en Kill)
	var duration time.Duration
	if req.Description == "" {
		duration = time.Duration(s.Config.KillDuration) * time.Second
	} else {
		duration = time.Duration(s.Config.KillDurationWithDescription) * time.Second
	}

	// Tarea asíncrona: al terminar, marca COMPLETED
	task := func(_ *models.Kill) error {
		// ignoramos el parámetro del tipo Kill; solo usamos la misma firma
		return s.TransmutationRepository.UpdateStatus(saved.ID, "COMPLETED")
	}
	// Usamos el ID del alquimista como clave de tarea para evitar duplicados simultáneos
	if _, ok := s.taskQueue.tasks[int(alch.ID)]; ok {
		s.HandleError(w, http.StatusConflict, r.URL.Path, fmt.Errorf("transmutation already running for id %d", alch.ID))
		return
	}
	s.taskQueue.StartTask(int(alch.ID), duration, task, nil)

	// Respuesta al cliente
	resp := &api.TransmutationTaskResponseDto{
		Alchemist: alch.ToResponseDto(),
		Status:    "Started",
	}
	js, err := json.Marshal(resp)
	if err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(js)
	s.logger.Info(http.StatusAccepted, r.URL.Path, start)
}
