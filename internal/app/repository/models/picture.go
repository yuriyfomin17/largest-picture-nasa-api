package models

// Picture is domain
type Picture struct {
	Sol    int    `bun:",pk"`
	ImgSrc string `bun:"img_src,notnull"`
	Size   int    `bun:"size,notnull"`
}
