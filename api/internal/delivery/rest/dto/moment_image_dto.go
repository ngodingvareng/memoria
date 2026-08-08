package dto

import "github.com/ngodingvareng/memoria/internal/usecase"

type MomentImageResponse struct {
	ID  string `json:"id" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	URL string `json:"url" example:"https://storage.example.com/moments/.../abc123.jpg?X-Amz-..."`
}

func NewMomentImageResponse(img usecase.MomentImageWithURL) MomentImageResponse {
	return MomentImageResponse{
		ID:  img.ID.String(),
		URL: img.URL,
	}
}
