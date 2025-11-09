package server

import (
	"backend-avanzada/api"
	"backend-avanzada/models"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const auditEntityTransmutation = "transmutation"

var (
	errTransmutationInProgress = errors.New("transmutation already running for this alchemist")
	errAlchemistNotFound       = errors.New("alchemist not found")

	allowedTransmutationStatuses = map[string]bool{
		"IN_PROGRESS": true,
		"COMPLETED":   true,
		"FAILED":      true,
		"CANCELLED":   true,
	}
)

func (s *Server) HandleTransmutations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetAllTransmutations(w, r)
	case http.MethodPost:
		s.handleCreateTransmutation(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) HandleTransmutationsWithId(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetTransmutationByID(w, r)
	case http.MethodPost:
		s.handleStartTransmutation(w, r)
	case http.MethodPatch:
		s.handleUpdateTransmutationStatus(w, r)
	case http.MethodDelete:
		s.handleCancelTransmutation(w, r)
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
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	s.logger.Info(http.StatusOK, r.URL.Path, start)
}

func (s *Server) handleCreateTransmutation(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var req api.TransmutationRequestDto
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, err)
		return
	}
	if req.AlchemistID == nil {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, fmt.Errorf("alchemist_id is required"))
		return
	}
	s.respondStartTransmutation(w, r, *req.AlchemistID, req.Description, start)
}

func (s *Server) handleStartTransmutation(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	idStr := strings.TrimSpace(mux.Vars(r)["id"])
	if idStr == "" {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, fmt.Errorf("missing alchemist id in path"))
		return
	}
	alchemistID, err := strconv.Atoi(idStr)
	if err != nil {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, err)
		return
	}
	var req api.TransmutationRequestDto
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, err)
		return
	}
	if req.AlchemistID != nil && *req.AlchemistID != alchemistID {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, fmt.Errorf("body alchemist_id does not match path id"))
		return
	}
	s.respondStartTransmutation(w, r, alchemistID, req.Description, start)
}

func (s *Server) respondStartTransmutation(w http.ResponseWriter, r *http.Request, alchemistID int, description string, start time.Time) {
	t, err := s.startTransmutation(alchemistID, description)
	if err != nil {
		switch {
		case errors.Is(err, errAlchemistNotFound):
			s.HandleError(w, http.StatusNotFound, r.URL.Path, err)
		case errors.Is(err, errTransmutationInProgress):
			s.HandleError(w, http.StatusConflict, r.URL.Path, err)
		default:
			s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(t.ToResponseDto(true)); err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	s.logger.Info(http.StatusAccepted, r.URL.Path, start)
}

func (s *Server) startTransmutation(alchemistID int, description string) (*models.Transmutation, error) {
	alch, err := s.AlchemistRepository.FindById(alchemistID)
	if err != nil {
		return nil, err
	}
	if alch == nil {
		return nil, errAlchemistNotFound
	}
	if _, ok := s.taskQueue.tasks[int(alch.ID)]; ok {
		return nil, errTransmutationInProgress
	}
	desc := strings.TrimSpace(description)
	if desc == "" {
		desc = "Generic transmutation"
	}
	t := &models.Transmutation{
		Description: desc,
		Status:      "IN_PROGRESS",
		AlchemistID: alch.ID,
		Alchemist:   alch,
	}
	saved, err := s.TransmutationRepository.Save(t)
	if err != nil {
		return nil, err
	}
	saved.Alchemist = alch
	if err := s.createTransmutationAudit("TRANSMUTATION_STARTED", saved.ID, fmt.Sprintf("Transmutación #%d iniciada por %s", saved.ID, alch.Name)); err != nil {
		_ = s.TransmutationRepository.Delete(saved)
		return nil, err
	}
	duration := s.transmutationDuration(desc)
	task := func(_ *models.Kill) error {
		if err := s.TransmutationRepository.UpdateStatus(saved.ID, "COMPLETED"); err != nil {
			return err
		}
		return s.createTransmutationAudit("TRANSMUTATION_COMPLETED", saved.ID, fmt.Sprintf("Transmutación #%d completada para %s", saved.ID, alch.Name))
	}
	s.taskQueue.StartTask(int(alch.ID), duration, task, nil)
	return saved, nil
}

func (s *Server) transmutationDuration(description string) time.Duration {
	if strings.TrimSpace(description) == "" || description == "Generic transmutation" {
		return time.Duration(s.Config.TransmutationDuration) * time.Second
	}
	return time.Duration(s.Config.TransmutationDurationHigh) * time.Second
}

func (s *Server) handleGetTransmutationByID(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	id, err := strconv.Atoi(strings.TrimSpace(mux.Vars(r)["id"]))
	if err != nil {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, err)
		return
	}
	t, err := s.TransmutationRepository.FindById(id)
	if err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	if t == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(t.ToResponseDto(true)); err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	s.logger.Info(http.StatusOK, r.URL.Path, start)
}

func (s *Server) handleUpdateTransmutationStatus(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	id, err := strconv.Atoi(strings.TrimSpace(mux.Vars(r)["id"]))
	if err != nil {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, err)
		return
	}
	var req api.TransmutationRequestDto
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, err)
		return
	}
	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status == "" {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, fmt.Errorf("status is required"))
		return
	}
	if !allowedTransmutationStatuses[status] {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, fmt.Errorf("invalid status %s", status))
		return
	}
	t, err := s.TransmutationRepository.FindById(id)
	if err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	if t == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	current := strings.ToUpper(strings.TrimSpace(t.Status))
	if current == status {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(t.ToResponseDto(true)); err != nil {
			s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
			return
		}
		s.logger.Info(http.StatusOK, r.URL.Path, start)
		return
	}
	if current == "IN_PROGRESS" && status != "IN_PROGRESS" {
		s.taskQueue.CancelTask(int(t.AlchemistID))
	}
	if err := s.TransmutationRepository.UpdateStatus(t.ID, status); err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	t.Status = status
	if err := s.createTransmutationAudit("TRANSMUTATION_STATUS_UPDATED", t.ID, fmt.Sprintf("Transmutación #%d actualizada a %s", t.ID, status)); err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(t.ToResponseDto(true)); err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	s.logger.Info(http.StatusOK, r.URL.Path, start)
}

func (s *Server) handleCancelTransmutation(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	id, err := strconv.Atoi(strings.TrimSpace(mux.Vars(r)["id"]))
	if err != nil {
		s.HandleError(w, http.StatusBadRequest, r.URL.Path, err)
		return
	}
	t, err := s.TransmutationRepository.FindById(id)
	if err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	if t == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	status := strings.ToUpper(strings.TrimSpace(t.Status))
	if status == "COMPLETED" || status == "FAILED" {
		s.HandleError(w, http.StatusConflict, r.URL.Path, fmt.Errorf("transmutation %d can no longer be cancelled", id))
		return
	}
	if status == "IN_PROGRESS" {
		s.taskQueue.CancelTask(int(t.AlchemistID))
	}
	if err := s.TransmutationRepository.UpdateStatus(t.ID, "CANCELLED"); err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	t.Status = "CANCELLED"
	if err := s.createTransmutationAudit("TRANSMUTATION_CANCELLED", t.ID, fmt.Sprintf("Transmutación #%d cancelada", t.ID)); err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(t.ToResponseDto(true)); err != nil {
		s.HandleError(w, http.StatusInternalServerError, r.URL.Path, err)
		return
	}
	s.logger.Info(http.StatusOK, r.URL.Path, start)
}

func (s *Server) createTransmutationAudit(action string, entityID uint, description string) error {
	_, err := s.AuditRepository.Save(&models.Audit{
		Action:      action,
		Entity:      auditEntityTransmutation,
		EntityID:    entityID,
		Description: description,
	})
	return err
}
