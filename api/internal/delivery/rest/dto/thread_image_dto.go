package dto

import "github.com/ngodingvareng/memoria/internal/usecase"

type ThreadImageResponse struct {
	ID  string `json:"id"  example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	URL string `json:"url" example:"https://storage.example.com/threads/.../abc123.jpg?X-Amz-..."`
}

func NewThreadImageResponse(img usecase.ThreadImageWithURL) ThreadImageResponse {
	return ThreadImageResponse{
		ID:  img.ID.String(),
		URL: img.URL,
	}
}
