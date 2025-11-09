package api

type TransmutationRequestDto struct {
	Description string `json:"description"`
}

type TransmutationResponseDto struct {
	ID          int                   `json:"id"`
	AlchemistID int                   `json:"alchemist_id"`
	Description string                `json:"description"`
	Status      string                `json:"status"`
	CreatedAt   string                `json:"created_at"`
	Alchemist   *AlchemistResponseDto `json:"alchemist,omitempty"`
}

type TransmutationTaskResponseDto struct {
	Alchemist *AlchemistResponseDto `json:"alchemist"`
	Status    string                `json:"status"`
}
