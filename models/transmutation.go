package models

import (
	"backend-avanzada/api"

	"gorm.io/gorm"
)

type Transmutation struct {
	gorm.Model
	Description string
	Status      string
	AlchemistID uint
	Alchemist   *Alchemist // Ajusta si tu struct se llama distinto
}

func (t *Transmutation) ToResponseDto(includeAlchemist bool) *api.TransmutationResponseDto {
	dto := &api.TransmutationResponseDto{
		ID:          int(t.ID),
		AlchemistID: int(t.AlchemistID),
		Description: t.Description,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt.String(),
	}
	if includeAlchemist && t.Alchemist != nil {
		dto.Alchemist = t.Alchemist.ToResponseDto() // ajusta al nombre real de tu método
	}
	return dto
}
