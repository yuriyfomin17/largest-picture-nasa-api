package models

type NasaPhotos struct {
	Photos []NasaPhoto `json:"photos"`
}
type NasaPhoto struct {
	ImageSrc string `json:"img_src"`
}
