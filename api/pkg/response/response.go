package response

import "github.com/randhir3-cloud/GK-Circle-v2/api/models"

type ResponseFinalScore struct {
	FinalScore []models.FinalScoreBoard
}

type ResponseFinalScoreForAdmin struct {
	FinalScore []models.FinalScoreBoardAdmin
}
