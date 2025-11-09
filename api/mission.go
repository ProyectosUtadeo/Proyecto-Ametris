package api

type MissionRequestDto struct {
	Titulo      string `json:"title"`
	Descripcion string `json:"description"`
	Estado      string `json:"status"`
	AsignadoAID *int   `json:"assigned_to_id"`
}

type MissionResponseDto struct {
	ID          int    `json:"mission_id"`
	Titulo      string `json:"title"`
	Descripcion string `json:"description"`
	Estado      string `json:"status"`
	AsignadoAID *int   `json:"assigned_to_id"`
}
