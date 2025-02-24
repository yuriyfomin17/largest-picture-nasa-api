package pgrepo

import (
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/domain"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/repository/models"
)

func toDomainPicture(picture models.Picture) domain.Picture {
	return domain.NewPicture(domain.NewPictureData{
		Sol:  picture.Sol,
		Size: picture.Size,
		Url:  picture.ImgSrc,
	})
}

func domainToPicture(domainPicture domain.Picture) models.Picture {
	return models.Picture{
		ImgSrc: domainPicture.GetUrl(),
		Size:   domainPicture.GetSize(),
		Sol:    domainPicture.GetSol(),
	}
}
